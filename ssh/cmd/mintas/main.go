// cmd/mintas — 运维/验证工具：以指定 userID 的派生钱包身份 mint 一张卡。
//
// 仅在 Polygon Amoy 测试网使用。用于验证「每个用户独立钱包」的签名归属，
// 不参与线上服务。
//
// 用法：
//
//	cd ssh
//	GEOYUAN_WALLET_SEED=<hex> go run ./cmd/mintas -uid 3 -lat 31.2304 -lon 121.4737 -title "Shanghai Test"
//
// 依赖环境变量（或项目根 .env）：
//
//	GEOYUAN_RPC_URL / GEOYUAN_CONTRACT / CARD_CHAIN_KEY / GEOYUAN_WALLET_SEED
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"geoyuan.cn/ssh/internal/card"
	"geoyuan.cn/ssh/internal/chain"
	"geoyuan.cn/ssh/internal/wallet"
)

func main() {
	loadDotEnv()

	uid := flag.Int64("uid", 0, "userID（决定派生地址）")
	lat := flag.Float64("lat", 31.2304, "纬度")
	lon := flag.Float64("lon", 121.4737, "经度")
	title := flag.String("title", "Ops Test Card", "卡片标题")
	desc := flag.String("desc", "minted by cmd/mintas", "描述")
	color := flag.String("color", "#00f5ff", "颜色")
	addrOnly := flag.Bool("addr-only", false, "只打印派生地址，不发交易")
	flag.Parse()

	if *uid <= 0 {
		log.Fatal("-uid 必须是正整数")
	}

	cfg := chain.DefaultConfig()
	client, err := chain.New(cfg)
	if err != nil {
		log.Fatalf("connect rpc: %v", err)
	}
	privKey := cfg.PrivateKey
	if privKey == "" {
		privKey = os.Getenv("CARD_CHAIN_KEY")
	}
	svc, err := card.New(client, privKey)
	if err != nil {
		log.Fatalf("card service: %v", err)
	}
	mgr, err := wallet.New(svc, os.Getenv("GEOYUAN_WALLET_SEED"))
	if err != nil {
		log.Fatalf("wallet manager: %v", err)
	}
	if !mgr.Ready() {
		log.Fatal("用户独立钱包未启用（缺 GEOYUAN_WALLET_SEED 或合约/链不可用）")
	}

	addr := mgr.AddressFor(*uid)
	fmt.Printf("userID=%d wallet=%s\n", *uid, addr)
	if bal, err := client.Balance(addr); err == nil {
		fmt.Printf("balance=%.6f MATIC\n", bal)
	}
	if *addrOnly {
		return
	}

	res, err := mgr.Mint(*uid, card.MintParams{
		Lat: *lat, Lon: *lon,
		Title: *title, Description: *desc, Color: *color,
	})
	if err != nil {
		log.Fatalf("mint: %v", err)
	}
	fmt.Printf("minted token=#%s h3=0x%x tx=%s block=%d\n",
		res.TokenID.Text(10), res.H3Index, res.TxHash, res.Block)
	if bal, err := client.Balance(addr); err == nil {
		fmt.Printf("balance after=%.6f MATIC\n", bal)
	}
}

// loadDotEnv — 与 main.go 同款极简 .env 读取
func loadDotEnv() {
	for _, path := range []string{".env", "/home/huanghe/myweb/diliy/.env"} {
		f, err := os.Open(path)
		if err != nil {
			continue
		}
		defer f.Close()
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())
			if line == "" || strings.HasPrefix(line, "#") || !strings.Contains(line, "=") {
				continue
			}
			kv := strings.SplitN(line, "=", 2)
			k := strings.TrimSpace(kv[0])
			v := strings.Trim(strings.TrimSpace(kv[1]), "\"'")
			if strings.Contains(strings.ToUpper(k), "GEOYUAN") || k == "CARD_CHAIN_KEY" {
				if os.Getenv(k) == "" {
					os.Setenv(k, v)
				}
			}
		}
		return
	}
}
