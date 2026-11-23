// Command geoyuan-api 是 geoyuan 准入系统的 HTTP API 服务。
//
// 提供：
//   POST /api/quiz/questions    — 生成 AI 招募测试题目（动态）
//   POST /api/quiz/submit       — 提交答案，验证，生成 WebPass token
//   POST /api/auth/verify       — SSH 端用 WebPass token 换 approved 状态
//   GET  /api/health            — 健康检查
//   GET  /api/stats             — 统计数据（成员数等）
//
// 与 geoyuan-ssh 共享同一个 SQLite。
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"geoyuan.cn/ssh/internal/questiongen"
	"geoyuan.cn/ssh/internal/questions"
	"geoyuan.cn/ssh/internal/store"
)

// ──────────── 共享状态 ────────────

type App struct {
	store     *store.Store
	generator *questiongen.Generator
}

// ──────────── 请求/响应类型 ────────────

type QuestionOut struct {
	ID      string   `json:"id"`
	Q       string   `json:"q"`
	Options []string `json:"options"`
	// 注意：答案不直接返回，前端用 hash 文本匹配
}

type QuestionsResp struct {
	Questions []QuestionOut `json:"questions"`
	SessionID string        `json:"session_id"`
}

type SubmitReq struct {
	Username    string `json:"username"`
	Email       string `json:"email"`
	SessionID   string `json:"session_id"`
	Answers     []int  `json:"answers"` // 用户选择的选项索引
}

type SubmitResp struct {
	Score     int    `json:"score"`
	Total     int    `json:"total"`
	Passed    bool   `json:"passed"`
	Token     string `json:"token,omitempty"`
	Message   string `json:"message,omitempty"`
}

type VerifyReq struct {
	Token              string `json:"token"`
	PubKeyFingerprint  string `json:"pubkey_fingerprint"`
}

type VerifyResp struct {
	Approved bool   `json:"approved"`
	Username string `json:"username"`
	Message  string `json:"message,omitempty"`
}

// session 缓存（map，进程内即可）
type cachedSession struct {
	Questions []questions.Question
	CreatedAt time.Time
}

// ──────────── CORS 中间件 ────────────

func cors(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next(w, r)
	}
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}

// ──────────── 路由处理 ────────────

func (a *App) handleQuestions(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeErr(w, 405, "method not allowed")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 50*time.Second)
	defer cancel()

	var qs []questions.Question
	if a.generator != nil {
		var err error
		qs, err = a.generator.Generate(ctx, 16) // Web 端 16 题（与当前 quiz.ts 数量对齐）
		if err != nil {
			log.Printf("ai question gen failed, fallback: %v", err)
			qs = questions.All()
			if len(qs) > 16 {
				qs = qs[:16]
			}
		}
	} else {
		qs = questions.All()
		if len(qs) > 16 {
			qs = qs[:16]
		}
	}

	// 生成 session id（简化：每次新的）
	sessionID := fmt.Sprintf("sess-%d", time.Now().UnixNano())

	out := make([]QuestionOut, len(qs))
	for i, q := range qs {
		out[i] = QuestionOut{
			ID:      fmt.Sprintf("q%d", i+1),
			Q:       q.Q,
			Options: q.Options[:],
		}
	}

	// 存入缓存（进程内，简化实现用全局 map）
	mu.Lock()
	sessions[sessionID] = cachedSession{Questions: qs, CreatedAt: time.Now()}
	mu.Unlock()

	writeJSON(w, 200, QuestionsResp{Questions: out, SessionID: sessionID})
}

func (a *App) handleSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeErr(w, 405, "method not allowed")
		return
	}

	var req SubmitReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, 400, "invalid body")
		return
	}

	if req.Username == "" || req.Email == "" {
		writeErr(w, 400, "username and email required")
		return
	}

	// 从 session 取题目
	mu.RLock()
	sess, ok := sessions[req.SessionID]
	mu.RUnlock()

	var score, total int
	var passed bool

	if ok {
		// AI 动态题目：后端直接验答案
		total = len(sess.Questions)
		for i, q := range sess.Questions {
			if i < len(req.Answers) && req.Answers[i] == q.Answer {
				score++
			}
		}
		passed = score*100/total >= 70 // ≥70% 通过
	} else {
		// Fallback：session 过期，信任前端传来的答案数（简化）
		total = len(req.Answers)
		score = 0 // session 丢了就不判分，让用户重试
		passed = false
	}

	if total == 0 {
		writeErr(w, 400, "session not found or expired, please retry")
		return
	}

	// 记录 WebPass
	token, err := a.store.CreateWebPass(req.Username, req.Email, score, total, passed)
	if err != nil {
		writeErr(w, 500, "failed to record result: "+err.Error())
		return
	}

	// 清理 session
	mu.Lock()
	delete(sessions, req.SessionID)
	mu.Unlock()

	resp := SubmitResp{Score: score, Total: total, Passed: passed}
	if passed {
		resp.Token = token
		resp.Message = "✓ 恭喜通过！复制下方 token，SSH 连接后可自动解锁聊天室。"
	} else {
		resp.Message = "未达通过线，欢迎重新尝试或直接 SSH 连接报名。"
	}

	writeJSON(w, 200, resp)
}

func (a *App) handleVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeErr(w, 405, "method not allowed")
		return
	}

	var req VerifyReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, 400, "invalid body")
		return
	}

	wp, err := a.store.GetWebPassByToken(req.Token)
	if err != nil || wp == nil || !wp.Passed {
		writeJSON(w, 200, VerifyResp{Approved: false, Message: "token invalid or not passed"})
		return
	}

	if wp.Used {
		writeJSON(w, 200, VerifyResp{Approved: true, Username: wp.Username, Message: "token already used, still approved"})
		return
	}

	// 标记 used
	if err := a.store.MarkWebPassUsed(req.Token); err != nil {
		log.Printf("mark web pass used: %v", err)
	}

	writeJSON(w, 200, VerifyResp{Approved: true, Username: wp.Username, Message: "verified from web pass"})
}

func (a *App) handleStats(w http.ResponseWriter, r *http.Request) {
	approved, _ := a.store.CountApproved()
	writeJSON(w, 200, map[string]any{
		"approved_members": approved,
	})
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]string{"status": "ok"})
}

// ──────────── main ────────────

// 全局 session 缓存（简化实现，单机足够）
var (
	mu       sync.RWMutex
	sessions = map[string]cachedSession{}
)

func main() {
	addr := os.Getenv("GEOYUAN_API_ADDR")
	if addr == "" {
		addr = ":3000"
	}

	dataDir := "data"
	if env := os.Getenv("GEOYUAN_DATA_DIR"); env != "" {
		dataDir = env
	}

	st, err := store.Open(filepath.Join(dataDir, "geoyuan.db"))
	if err != nil {
		log.Fatalf("open store: %v", err)
	}
	defer st.Close()

	// 题目生成器
	cfg := questiongen.DefaultConfig()
	var gen *questiongen.Generator
	if cfg.APIKey != "" {
		gen = questiongen.New(cfg)
	}

	app := &App{store: st, generator: gen}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", cors(handleHealth))
	mux.HandleFunc("/api/stats", cors(app.handleStats))
	mux.HandleFunc("/api/quiz/questions", cors(app.handleQuestions))
	mux.HandleFunc("/api/quiz/submit", cors(app.handleSubmit))
	mux.HandleFunc("/api/auth/verify", cors(app.handleVerify))

	// 静态文件（如果需要）
	if err := os.MkdirAll("public", 0o755); err == nil {
		mux.Handle("/", http.FileServer(http.Dir("public")))
	}

	log.Printf("geoyuan api listening on %s", addr)
	if gen != nil {
		log.Printf("ai question generator enabled")
	}

	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 120 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
