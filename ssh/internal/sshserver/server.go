// Package sshserver 基于 golang.org/x/crypto/ssh 实现 geoyuan 准入 SSH 服务。
//
// 设计要点：
//   - PublicKeyCallback 对任意公钥（含未知指纹）都放行，准入与否由答题结果决定，
//     不在认证阶段拒绝连接（对齐 pico.sh 的开放入口模式）。
//   - 会话成功后将 ssh channel 的 I/O 交给 bubbletea 跑答题 TUI。
//   - 通过答题的用户自动进入聊天室（Chat Hub）。
//   - host key 首次运行生成 ed25519 并持久化，重启复用。
package sshserver

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/binary"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/crypto/ssh"

	"geoyuan.cn/ssh/internal/chat"
	"geoyuan.cn/ssh/internal/geo"
	"geoyuan.cn/ssh/internal/questiongen"
	"geoyuan.cn/ssh/internal/store"
	"geoyuan.cn/ssh/internal/tui"
)

// 资源保护参数：公网 22 端口会被扫描器高频连接，必须限流，
// 否则单机容易被大量空闲连接拖垮。
const (
	maxConcurrentConns = 64               // 同时处理的 TCP 连接上限
	handshakeTimeout   = 15 * time.Second // 必须在该时间内完成握手
	maxSessionDuration = 60 * time.Minute // 单次会话最长时长（聊天室允许更久）
)

// Server 持有存储、AI 题目生成器、聊天中心与监听地址
type Server struct {
	store     *store.Store
	generator *questiongen.Generator
	chatHub   *chat.Hub
	addr      string
}

// New 构造 Server
func New(s *store.Store, gen *questiongen.Generator, chatHub *chat.Hub, addr string) *Server {
	return &Server{store: s, generator: gen, chatHub: chatHub, addr: addr}
}

// ListenAndServe 加载/生成 host key 并开始监听，阻塞直到出错。
func (s *Server) ListenAndServe(hostKeyPath string) error {
	signer, err := loadOrCreateHostKey(hostKeyPath)
	if err != nil {
		return fmt.Errorf("host key: %w", err)
	}

	config := &ssh.ServerConfig{
		// 允许任意公钥进入；准入判定交给答题结果
		PublicKeyCallback: func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			fp := ssh.FingerprintSHA256(key)
			log.Printf("conn from %s user=%s fp=%s", conn.RemoteAddr(), conn.User(), fp[:16]+"...")
			return &ssh.Permissions{
				Extensions: map[string]string{"fp": fp},
			}, nil
		},
		// 不允许密码 / 键盘交互认证，只接受公钥
		NoClientAuth: false,
	}
	config.AddHostKey(signer)

	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", s.addr, err)
	}
	log.Printf("geoyuan ssh quiz listening on %s", s.addr)
	if s.generator != nil {
		log.Printf("AI question generator ready (model configured)")
	} else {
		log.Printf("AI question generator disabled (no API key), using fallback题库")
	}
	if s.chatHub != nil {
		log.Printf("chat hub ready")
	}

	sem := make(chan struct{}, maxConcurrentConns)
	for {
		nConn, err := ln.Accept()
		if err != nil {
			log.Printf("accept error: %v", err)
			continue
		}
		select {
		case sem <- struct{}{}:
			go func() {
				defer func() { <-sem }()
				s.handleConn(nConn, config)
			}()
		default:
			// 超过上限直接拒绝，避免扫描器把连接池占满
			nConn.Close()
		}
	}
}

// handleConn 完成 SSH 握手并路由 session 通道
func (s *Server) handleConn(nConn net.Conn, config *ssh.ServerConfig) {
	defer nConn.Close()
	remoteAddr := nConn.RemoteAddr().String()

	// 握手阶段必须有超时，防止半开连接长期占用
	_ = nConn.SetDeadline(time.Now().Add(handshakeTimeout))
	conn, chans, reqs, err := ssh.NewServerConn(nConn, config)
	if err != nil {
		// 握手失败（无公钥、超时、版本不匹配等）直接断开，不记录
		return
	}
	defer conn.Close()
	_ = nConn.SetDeadline(time.Time{})

	// 丢弃连接级全局请求（keepalive 等）
	go ssh.DiscardRequests(reqs)

	// 一个连接只允许一个 session，其余拒绝
	sessionStarted := false
	for newChan := range chans {
		if newChan.ChannelType() != "session" {
			newChan.Reject(ssh.UnknownChannelType, "only session supported")
			continue
		}
		if sessionStarted {
			newChan.Reject(ssh.Prohibited, "only one session per connection")
			continue
		}
		sessionStarted = true
		ch, requests, err := newChan.Accept()
		if err != nil {
			log.Printf("accept channel: %v", err)
			continue
		}
		fp := ""
		if conn.Permissions != nil {
			fp = conn.Permissions.Extensions["fp"]
		}
		go s.handleSession(ch, requests, fp, remoteAddr)
	}
}

// ptyReq 解析 "pty-req" 请求载荷
type ptyReq struct {
	Term        string
	Width       uint32
	Height      uint32
	PixelWidth  uint32
	PixelHeight uint32
	Modes       string
}

// winChange 解析 "window-change" 请求载荷
type winChange struct {
	Width       uint32
	Height      uint32
	PixelWidth  uint32
	PixelHeight uint32
}

// handleSession 把 channel I/O 接入 bubbletea 运行 TUI
func (s *Server) handleSession(ch ssh.Channel, requests <-chan *ssh.Request, fp string, remoteAddr string) {
	defer ch.Close()

	// 解析 remoteAddr → IP + 主机名（Reverse DNS，同步做一次，快的）
	ip, host := parseRemote(remoteAddr)
	go func() {
		// 异步补全主机名（如果一开始没有）
		if host == "" {
			if h, err := net.LookupAddr(ip); err == nil && len(h) > 0 {
				host = h[0]
			}
		}
	}()

	// 🌍 查 IP 地理位置（带缓存 + 2s 超时，失败降级显示 IP）
	var geoLoc string
	if loc := geo.Resolve(ip); loc != nil {
		geoLoc = loc.Display()
	}
	log.Printf("conn from %s geo=%q host=%q", ip, geoLoc, host)

	// 查用户（新用户为 nil）
	user, err := s.store.GetUserByFingerprint(fp)
	if err != nil {
		log.Printf("get user by fp %s: %v", fp, err)
		io.WriteString(ch, "服务内部错误，请稍后重试。\r\n")
		return
	}

	model := tui.New(s.store, s.generator, fp, user, s.chatHub, ip, host, geoLoc)
	prog := tea.NewProgram(
		model,
		tea.WithInput(ch),
		tea.WithOutput(ch),
		// NOTE: 不用 WithAltScreen() — 让内容直接输出到主屏幕，用户可滚动
		// 之前加了 alt screen 但内容有 30+ 行，小终端（24行）会被截断
	)

	// 限制单次会话时长，避免空闲/无人值守连接长期占用资源
	sessionTimer := time.AfterFunc(maxSessionDuration, prog.Quit)
	defer sessionTimer.Stop()

	// 在独立 goroutine 中处理 channel 请求：PTY / shell / 窗口变化
	go func() {
		for req := range requests {
			switch req.Type {
			case "pty-req":
				var pty ptyReq
				ssh.Unmarshal(req.Payload, &pty)
				if req.WantReply {
					req.Reply(true, nil)
				}
				if pty.Width > 0 && pty.Height > 0 {
					prog.Send(tea.WindowSizeMsg{Width: int(pty.Width), Height: int(pty.Height)})
				}
			case "shell":
				if req.WantReply {
					req.Reply(true, nil)
				}
			case "window-change":
				var w winChange
				ssh.Unmarshal(req.Payload, &w)
				if w.Width > 0 && w.Height > 0 {
					prog.Send(tea.WindowSizeMsg{Width: int(w.Width), Height: int(w.Height)})
				}
				if req.WantReply {
					req.Reply(true, nil)
				}
			default:
				// exec 等非交互请求一律拒绝
				if req.WantReply {
					req.Reply(false, nil)
				}
			}
		}
	}()

	// 阻塞运行直到用户退出或断开
	if _, err := prog.Run(); err != nil {
		log.Printf("tui run error: %v", err)
	}

	// 通知客户端正常退出
	status := make([]byte, 4)
	binary.BigEndian.PutUint32(status, 0)
	ch.SendRequest("exit-status", false, status)
}

// loadOrCreateHostKey 加载已有的 ed25519 host key，不存在则生成并持久化。
func loadOrCreateHostKey(path string) (ssh.Signer, error) {
	if data, err := os.ReadFile(path); err == nil {
		return ssh.ParsePrivateKey(data)
	}

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generate ed25519: %w", err)
	}

	block, err := ssh.MarshalPrivateKey(priv, "geoyuan-host-key")
	if err != nil {
		return nil, fmt.Errorf("marshal private key: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, err
	}
	if err := os.WriteFile(path, pem.EncodeToMemory(block), 0o600); err != nil {
		return nil, err
	}
	// 同时写一份公钥，便于运维核对指纹（失败不影响启动）
	if pubKey, err := ssh.NewPublicKey(pub); err == nil {
		authorized := ssh.MarshalAuthorizedKey(pubKey)
		_ = os.WriteFile(path+".pub", authorized, 0o644)
	}
	return ssh.NewSignerFromKey(priv)
}

// parseRemote 从 remoteAddr（形如 "1.2.3.4:5678" 或 "[::1]:1234"）中提取 IP 和尝试 Reverse DNS
func parseRemote(remoteAddr string) (ip, host string) {
	hostPort, _, err := net.SplitHostPort(remoteAddr)
	if err == nil {
		ip = hostPort
	} else {
		ip = remoteAddr
	}
	// 快速做一次 Reverse DNS（带超时保护）
	addrs, err := net.LookupAddr(ip)
	if err == nil && len(addrs) > 0 {
		host = addrs[0]
	}
	return ip, host
}
