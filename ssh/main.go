// Command geoyuan-ssh 是 geoyuan team AI 招募测试 + 私密聊天室 + 链上卡牌 Mint 的 SSH 服务入口。
//
// 用法：
//
//	GEOYUAN_ADDR=:22 go run .        # 监听 22（需 root / setcap）
//	go run .                          # 默认监听 :2222
//
// 环境变量：
//
//	GEOYUAN_ADDR              监听地址（默认 :2222）
//	GEOYUAN_DATA_DIR          数据目录（默认 ./data）
//	GEOYUAN_LLM_API_KEY       API Key（设置则启用 AI 动态出题）
//	GEOYUAN_LLM_BASE_URL      API 地址（默认 https://api.openai.com/v1）
//	GEOYUAN_LLM_MODEL         模型名（默认 gpt-4o-mini）
//	GEOYUAN_RPC_URL           Polygon RPC（默认 Amoy PublicNode）
//	GEOYUAN_CONTRACT          GeoyuanCard 合约地址
//	CARD_CHAIN_KEY            部署钱包私钥（用于链上 Mint 签名，空则只做 H3 预览）
//
// 当前会自动加载项目根目录的 .env（如果存在），方便本地开发。
package main

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"geoyuan.cn/ssh/internal/card"
	"geoyuan.cn/ssh/internal/chain"
	"geoyuan.cn/ssh/internal/chat"
	"geoyuan.cn/ssh/internal/questiongen"
	"geoyuan.cn/ssh/internal/sshserver"
	"geoyuan.cn/ssh/internal/store"
	"geoyuan.cn/ssh/internal/wallet"
)

func main() {
	// 🧠 资源保护：2vCPU/1.8G 小机器，Go heap 必须主动归还 OS
	// 默认 GOGC=100（heap 到上次 2 倍才 GC → 峰值 400M+ 不归还）
	// 调到 50：heap 增长 50% 就触发 GC + 更激进归还
	debug.SetGCPercent(50)
	log.Printf("Go GOGC set to 50 (heap GC + aggressive return to OS)")

	loadDotEnv()

	addr := os.Getenv("GEOYUAN_ADDR")
	if addr == "" {
		addr = ":2222"
	}

	dataDir := "data"
	if env := os.Getenv("GEOYUAN_DATA_DIR"); env != "" {
		dataDir = env
	}
	if err := os.MkdirAll(dataDir, 0o700); err != nil {
		log.Fatalf("create data dir: %v", err)
	}

	st, err := store.Open(filepath.Join(dataDir, "geoyuan.db"))
	if err != nil {
		log.Fatalf("open store: %v", err)
	}
	defer st.Close()

	// AI 题目生成器（可选，无 API Key 时 generator 为 nil，走 fallback）
	cfg := questiongen.DefaultConfig()
	var gen *questiongen.Generator
	if cfg.APIKey != "" {
		gen = questiongen.New(cfg)
	}

	// 全局聊天室 Hub
	chatHub := chat.NewHub()

	// Polygon RPC + GeoyuanCard + 用户钱包（可选，缺任何条件都降级为 H3 预览）
	walletMgr := initWalletManager()

	srv := sshserver.New(st, gen, chatHub, walletMgr, addr)
	if err := srv.ListenAndServe(filepath.Join(dataDir, "ssh_host_ed25519_key")); err != nil {
		log.Fatalf("serve: %v", err)
	}
}

// initWalletManager 初始化链上服务 + 用户钱包管理器，失败降级为 nil（TUI 走 H3 预览）
func initWalletManager() *wallet.Manager {
	cfg := chain.DefaultConfig()
	if cfg.ContractAddr == "" {
		log.Printf("GeoyuanCard 合约地址未配置（GEOYUAN_CONTRACT），链上 Mint 不可用")
		return nil
	}
	client, err := chain.New(cfg)
	if err != nil {
		log.Printf("Polygon RPC 连接失败 (%s)：%v，降级为 H3 预览", cfg.RPCURL, err)
		return nil
	}
	log.Printf("Polygon RPC 已连接 chain_id=%d contract=%s", cfg.ChainID, cfg.ContractAddr)

	privKey := cfg.PrivateKey
	if privKey == "" {
		privKey = os.Getenv("CARD_CHAIN_KEY")
	}
	svc, err := card.New(client, privKey)
	if err != nil {
		log.Printf("GeoyuanCard 服务初始化失败：%v，降级为 H3 预览", err)
		return nil
	}

	mgr, err := wallet.New(svc, os.Getenv("GEOYUAN_WALLET_SEED"))
	if err != nil {
		log.Printf("用户钱包初始化失败：%v，降级为部署钱包代签", err)
		return nil
	}
	if mgr.Ready() {
		log.Printf("用户独立钱包已启用（BIP44 m/44'/60'/0'/0/{userID}）")
	} else {
		log.Printf("⚠️  GEOYUAN_WALLET_SEED 未配置：所有用户共用部署钱包 %s，资产不隔离", svc.FromAddr())
	}
	if svc.Ready() {
		log.Printf("GeoyuanCard 链上 Mint 已就绪 gas-signer=%s", svc.FromAddr())
	}
	return mgr
}

// loadDotEnv — 读取项目根目录的 .env（简单 KEY=VALUE 解析，支持 # 注释）
func loadDotEnv() {
	candidates := []string{
		".env",
		"/home/huanghe/myweb/diliy/.env",
	}
	for _, path := range candidates {
		f, err := os.Open(path)
		if err != nil {
			continue
		}
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			// 支持 KEY=VALUE 和 KEY: VALUE 两种写法
			var key, val string
			if strings.Contains(line, "=") {
				parts := strings.SplitN(line, "=", 2)
				key = strings.TrimSpace(parts[0])
				val = strings.Trim(strings.TrimSpace(parts[1]), "\"'")
			} else if strings.Contains(line, ":") {
				parts := strings.SplitN(line, ":", 2)
				key = strings.TrimSpace(parts[0])
				val = strings.Trim(strings.TrimSpace(parts[1]), "\"'")
			} else {
				continue
			}
			// 只注入 geoyuan 相关变量 + Polygon 私钥，避免覆盖其他系统变量
			if shouldPassToEnv(key) {
				os.Setenv(key, val)
			}
		}
		log.Printf("已加载 .env: %s", path)
		return
	}
	log.Println("未找到 .env，使用默认环境变量")
}

// shouldPassToEnv — .env 中只挑 geoyuan / polygon / deploy 相关的变量注入
func shouldPassToEnv(key string) bool {
	upper := strings.ToUpper(key)
	return strings.Contains(upper, "GEOYUAN") ||
		strings.Contains(upper, "CARD_CHAIN_KEY") ||
		strings.Contains(upper, "DEPLOY_WALLET")
}
