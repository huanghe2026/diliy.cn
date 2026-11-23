// Command geoyuan-ssh 是 geoyuan team AI 招募测试 + 私密聊天室的 SSH 服务入口。
//
// 用法：
//
//	GEOYUAN_ADDR=:22 go run .        # 监听 22（需 root / setcap）
//	go run .                          # 默认监听 :2222
//
// 环境变量（AI 题目生成可选配置）：
//
//	GEOYUAN_LLM_API_KEY    API Key（设置则启用 AI 动态出题，否则用 fallback 题库）
//	GEOYUAN_LLM_BASE_URL   API 地址（默认 https://api.openai.com/v1）
//	GEOYUAN_LLM_MODEL      模型名（默认 gpt-4o-mini）
//
// 数据目录 data/ 下生成 geoyuan.db（SQLite）与 ssh_host_ed25519_key（host key）。
package main

import (
	"log"
	"os"
	"path/filepath"

	"geoyuan.cn/ssh/internal/chat"
	"geoyuan.cn/ssh/internal/questiongen"
	"geoyuan.cn/ssh/internal/sshserver"
	"geoyuan.cn/ssh/internal/store"
)

func main() {
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

	srv := sshserver.New(st, gen, chatHub, addr)
	if err := srv.ListenAndServe(filepath.Join(dataDir, "ssh_host_ed25519_key")); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
