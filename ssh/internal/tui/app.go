// Package tui 实现 geoyuan AI 招募 + 私密聊天室的终端交互界面（bubbletea）。
//
// 视觉设计：
//   - 统一 header bar：logo + 阶段标题 + 用户信息（用户名/IP/在线）
//   - 彩色 256 色 palette，渐变感的欢迎页 / 结果页
//   - 结果页：恭喜画面 + 分数仪表盘 + 答题回顾
//   - 聊天室：消息气泡风格，区分自己/他人/DM
package tui

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"geoyuan.cn/ssh/internal/card"
	"geoyuan.cn/ssh/internal/chat"
	"geoyuan.cn/ssh/internal/questiongen"
	"geoyuan.cn/ssh/internal/questions"
	"geoyuan.cn/ssh/internal/store"
	"geoyuan.cn/ssh/internal/wallet"
)

// PassThreshold 通过阈值
const PassThreshold = 24
const QuizCount = 30

// Phase TUI 阶段
type Phase int

const (
	PhaseRegister Phase = iota
	PhaseQuiz
	PhaseResult
	PhaseWelcome
	PhaseRetry
	PhaseChat
	PhaseMenu        // 主菜单（Welcome 后进入）
	PhaseMintCard    // Mint GeoyuanCard 视图
	PhaseMyCards     // 我的卡片列表
	PhaseMarket      // 全市场挂售浏览
	PhaseListCard    // 把自己的卡挂售
	PhaseQuitConfirm // ⚠️ 退出确认弹窗（防误触）
	PhaseHelp        // ? 帮助浮层（覆盖任意页面之上）
)

// ═══════════════════════════════════════════════
//  🎨 Color Palette — 全部移到 styles.go 统一管理
// ═══════════════════════════════════════════════

// ═══════════════════════════════════════════════
//  📐 Reusable Styles — 全部移到 styles.go 统一管理
// ═══════════════════════════════════════════════

var usernameRe = regexp.MustCompile(`^[a-z0-9_]{2,20}$`)

// ═══════════════════════════════════════════════
//
//	Model
//
// ═══════════════════════════════════════════════
type Model struct {
	store     *store.Store
	generator *questiongen.Generator
	fp          string
	user        *store.User
	phase       Phase
	prevPhase   Phase // 帮助浮层关闭后回去的阶段

	// 🧭 终端尺寸（WindowSizeMsg 注入）— 用来让 View() 自适应
	termWidth  int
	termHeight int

	// 连接信息（UI 展示用）
	remoteIP   string
	remoteHost string
	geoLoc     string // 地理位置 (e.g. "Hangzhou, China")

	// 注册
	nameInput textinput.Model
	inputErr  string

	// 答题
	questions []questions.Question
	cursor    int
	selected  int
	answers   []int

	// 动画状态（spinner 帧 + cursor blink 帧，全局共享）
	spinnerFrame   uint
	cursorBlink    uint

	// 结果
	score  int
	passed bool

	// 重测
	retrySel int

	// 聊天室
	ChatHub   *chat.Hub
	sub       *chat.Subscription
	chatLog   []chat.Message
	chatInput textinput.Model

	// 🎴 Mint — GeoyuanCard 铸造视图
	mintView   PhaseMint
	menuCursor int // 主菜单当前选中项

	cardService *card.Service   // 只读服务（MyCards / NeighborMatch）
	walletMgr   *wallet.Manager // 用户独立钱包（mint 签名 + gas）

	// 🃏 MyCards / 邻居匹配
	myCardsCache    []card.Card
	myCardsErr      error
	myCardsLoading  bool
	neighborCache   []card.Card
	neighborErr     error
	neighborLoading bool
	neighborK       int
	neighborH3      uint64
	showNeighbors   bool

	// 市场
	marketListings  []card.MarketListing
	marketErr       string
	marketLoading   bool
	marketListingSel int // 市场列表当前选中的 listing
	marketOffset    uint64
	marketCount     uint64

	// 挂售
	listCardCursor int // My Cards 里选卡挂售的 cursor
	listPriceWei   string // 用户输入的挂售价（wei 或 MATIC）
	listInputFocus bool
	listInput      textinput.Model
	listCardLoading bool
	listCardErr    string
	listCardDone   bool

	// 状态
	Done   bool
	UserID int64
	Score  int
	Total  int
	Passed bool
}

// New 构造 TUI Model
func New(s *store.Store, gen *questiongen.Generator, fp string, user *store.User, chatHub *chat.Hub, walletMgr *wallet.Manager, ip, host, geo string) Model {
	var svc *card.Service
	if walletMgr != nil {
		svc = walletMgr.Service()
	}
	m := Model{
		store:       s,
		generator:   gen,
		fp:          fp,
		user:        user,
		ChatHub:     chatHub,
		cardService: svc,
		walletMgr:   walletMgr,
		remoteIP:    ip,
		remoteHost:  host,
		geoLoc:      geo,
	}
	switch {
	case user != nil && user.Status == store.StatusApproved:
		m.phase = PhaseWelcome
		m.enterChat()
	case user != nil:
		m.phase = PhaseRetry
	default:
		ti := textinput.New()
		ti.Placeholder = "2-20 位小写字母/数字/下划线"
		ti.CharLimit = 120 // 同时容纳 :redeem + 32 字符 token；用户名校验仍限 20
		ti.Prompt = "用户名: "
		ti.Focus()
		m.nameInput = ti
		m.phase = PhaseRegister
	}
	return m
}

// Init 实现 tea.Model — 初始拿到终端尺寸 + 阶段启动命令
func (m Model) Init() tea.Cmd {
	// 先触发一个 WindowSizeMsg 来拿终端尺寸
	cmds := []tea.Cmd{tea.EnterAltScreen}

	switch m.phase {
	case PhaseRegister:
		cmds = append(cmds, textinput.Blink)
	case PhaseChat:
		cmds = append(cmds, textinput.Blink, receiveMsgs(m.sub))
	}
	// 全局 spinner + cursor blink tick（任何阶段都要）
	cmds = append(cmds, tickSpinner())
	return tea.Batch(cmds...)
}

// tickSpinner 全局动画 tick —— 60fps 足够视觉流畅
func tickSpinner() tea.Cmd {
	return tea.Tick(120*time.Millisecond, func(t time.Time) tea.Msg {
		return spinnerTickMsg{t}
	})
}

type spinnerTickMsg struct{ t time.Time }

// spinnerFrames 动画帧
var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// spinner 返回当前帧 + 颜色（青蓝色）
func (m Model) spinner() string {
	frame := spinnerFrames[m.spinnerFrame%uint(len(spinnerFrames))]
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("45")). // 亮青蓝
		Bold(true).
		Render(frame)
}

// cursorMarker 返回闪烁的 "▶ " 当前选中标记
func (m Model) cursorMarker() string {
	// blink 周期 4 帧（480ms）：前 2 帧实心 ▶ ，后 2 帧半透明 ▷
	if m.cursorBlink%4 < 2 {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("226")). // 亮黄
			Bold(true).
			Render("▶ ")
	}
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")). // 暗灰（半透明感）
		Render("▷ ")
}

type (
	chatMsgWrapper struct{ Msg chat.Message }
	chatClosedMsg  struct{}
	welcomeDoneMsg struct{}
	mintResultMsg  struct {
		result *card.MintResult
		err    error
	}
	myCardsLoadedMsg struct {
		cards []card.Card
		err   error
	}
	neighborLoadedMsg struct {
		cards []card.Card
		err   error
	}
	marketLoadedMsg struct {
		listings []card.MarketListing
		count    uint64
		err      error
	}
	listCardDoneMsg struct {
		txHash string
		err    error
	}
	unlistDoneMsg struct {
		txHash string
		err    error
	}
	buyDoneMsg struct {
		txHash string
		err    error
	}
)

func receiveMsgs(sub *chat.Subscription) tea.Cmd {
	return func() tea.Msg {
		if sub == nil {
			return nil
		}
		msg, ok := <-sub.C
		if !ok {
			return chatClosedMsg{}
		}
		return chatMsgWrapper{Msg: msg}
	}
}

// Update 实现 tea.Model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// 🌀 动画 tick —— 全局，每 120ms
	case spinnerTickMsg:
		m.spinnerFrame++
		m.cursorBlink++
		// 每 tick 继续下一个（不中断）
		return m, tickSpinner()

	// 🧭 终端尺寸变化 — Bubble Tea 自动发
	case tea.WindowSizeMsg:
		m.termWidth = msg.Width
		m.termHeight = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case welcomeDoneMsg:
		if m.phase == PhaseWelcome {
			m.phase = PhaseMenu
			m.menuCursor = 0
			return m, nil
		}
	case chatMsgWrapper:
		if m.phase == PhaseChat {
			m.chatLog = append(m.chatLog, msg.Msg)
			return m, receiveMsgs(m.sub)
		}
	case chatClosedMsg:
		m.Done = true
		return m, tea.Quit
	case mintResultMsg:
		m.mintView = m.mintView.WithResult(msg.result, msg.err)
		return m, nil
	case myCardsLoadedMsg:
		m.myCardsCache = msg.cards
		m.myCardsErr = msg.err
		m.myCardsLoading = false
		return m, nil
	case neighborLoadedMsg:
		m.neighborCache = msg.cards
		m.neighborErr = msg.err
		m.neighborLoading = false
		return m, nil
	case marketLoadedMsg:
		m.marketListings = msg.listings
		m.marketCount = msg.count
		if msg.err != nil {
			m.marketErr = msg.err.Error()
		} else {
			m.marketErr = ""
		}
		m.marketLoading = false
		if m.marketListingSel >= len(m.marketListings) && len(m.marketListings) > 0 {
			m.marketListingSel = 0
		}
		return m, nil
	case listCardDoneMsg:
		m.listCardLoading = false
		if msg.err != nil {
			m.listCardErr = "挂售失败: " + msg.err.Error()
		} else {
			m.listCardErr = ""
			m.listCardDone = true
		}
		return m, nil
	case unlistDoneMsg:
		m.listCardLoading = false
		if msg.err != nil {
			m.listCardErr = "下架失败: " + msg.err.Error()
		} else {
			m.listCardDone = true
		}
		return m, nil
	case buyDoneMsg:
		m.marketLoading = false
		if msg.err != nil {
			m.marketErr = "购买失败: " + msg.err.Error()
		} else {
			m.marketErr = ""
			// 买完自动刷新
			m.marketOffset = 0
			m.marketListingSel = 0
			m.marketLoading = true
			return m, m.loadMarketCmd()
		}
		return m, nil
	}
	return m, nil
}

// handleKey 分发键盘事件到对应阶段
func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "ctrl+c" {
		m.cleanupChat()
		m.Done = true
		return m, tea.Quit
	}

	// ═══ 全局拦截 ═══
	// 1) ? 键（非 Register/Quiz/QuitConfirm 阶段）→ 弹出帮助浮层
	if msg.String() == "/" {
		// Bubble Tea 对 ? 的处理：有的终端发 "/"，有的发 shift+/
		// msg.String() 对 / 的值取决于键盘布局，我们也映射一下 shift 组合
	}
	switch msg.String() {
	case "?", "/":
		// Register/Quiz 阶段先不打扰（正在输入/答题）
		if m.phase != PhaseRegister && m.phase != PhaseQuiz && m.phase != PhaseQuitConfirm && m.phase != PhaseHelp {
			m.prevPhase = m.phase
			m.phase = PhaseHelp
			return m, nil
		}
	case "escape", "esc":
		// Esc 在 Menu/Welcome/Result/Chat 直接退出 → 改为确认弹窗
		if m.phase == PhaseMenu || m.phase == PhaseChat || m.phase == PhaseWelcome || m.phase == PhaseResult || m.phase == PhaseRetry {
			m.prevPhase = m.phase
			m.phase = PhaseQuitConfirm
			return m, nil
		}
	}

	// ═══ 分发到各阶段 ═══
	switch m.phase {
	case PhaseRegister:
		return m.updateRegister(msg)
	case PhaseQuiz:
		return m.updateQuiz(msg)
	case PhaseResult:
		if msg.String() == "enter" {
			m.phase = PhaseMenu
			m.menuCursor = 0
			return m, nil
		}
	case PhaseWelcome:
		if msg.String() == "enter" {
			m.phase = PhaseMenu
			m.menuCursor = 0
			return m, nil
		}
	case PhaseRetry:
		return m.updateRetry(msg)
	case PhaseMenu:
		return m.updateMenu(msg)
	case PhaseMintCard:
		return m.updateMint(msg)
	case PhaseMyCards:
		return m.updateMyCards(msg)
	case PhaseMarket:
		return m.updateMarket(msg)
	case PhaseListCard:
		return m.updateListCard(msg)
	case PhaseChat:
		return m.updateChat(msg)
	case PhaseQuitConfirm:
		return m.updateQuitConfirm(msg)
	case PhaseHelp:
		return m.updateHelp(msg)
	}
	return m, nil
}

// View 实现 tea.Model
func (m Model) View() string {
	switch m.phase {
	case PhaseRegister:
		return m.viewRegister()
	case PhaseQuiz:
		return m.viewQuiz()
	case PhaseResult:
		return m.viewResult()
	case PhaseWelcome:
		return m.viewWelcome()
	case PhaseRetry:
		return m.viewRetry()
	case PhaseMenu:
		return m.viewMenu()
	case PhaseMintCard:
		return m.mintView.View(m.termWidth, m.termHeight, m.geoLoc)
	case PhaseMyCards:
		return m.viewMyCards()
	case PhaseMarket:
		return m.viewMarket()
	case PhaseListCard:
		return m.viewListCard()
	case PhaseChat:
		return m.viewChat()
	case PhaseQuitConfirm:
		// 半透明遮罩 + 确认框
		return m.viewQuitConfirm()
	case PhaseHelp:
		// 先渲染底层页面，再叠加帮助框
		return m.viewHelpOverlay()
	}
	return ""
}

// cleanupChat 清理
func (m *Model) cleanupChat() {
	if m.ChatHub != nil && m.sub != nil {
		m.ChatHub.Leave(m.sub)
		m.sub = nil
	}
}

// enterChat 进入聊天室
func (m *Model) enterChat() {
	if m.ChatHub != nil && m.user != nil {
		m.sub = m.ChatHub.Join(m.user.ID, m.user.Username)
		m.phase = PhaseChat
		m.chatLog = nil
		ti := textinput.New()
		ti.Prompt = "> "
		ti.CharLimit = 200
		ti.Placeholder = "输入消息 · /w <user> <msg> 私聊 · :help 命令列表"
		ti.Focus()
		m.chatInput = ti
	}
}

// ═══════════════════════════════════════════════
//
//	🏗️  Header Bar (所有页面共享)
//
// ═══════════════════════════════════════════════
// breadcrumb 返回当前导航路径（供 headerBar 展示）
func (m Model) breadcrumb() string {
	switch m.phase {
	case PhaseWelcome:
		return "  Home  →  Welcome Back"
	case PhaseMenu:
		return "  Home"
	case PhaseMintCard:
		return "  Home  →  Mint Card"
	case PhaseMyCards:
		if m.showNeighbors {
			return "  Home  →  My Cards  →  Neighbors"
		}
		return "  Home  →  My Cards"
	case PhaseMarket:
		return "  Home  →  Marketplace"
	case PhaseListCard:
		return "  Home  →  My Cards  →  List for Sale"
	case PhaseChat:
		return "  Home  →  Crew Chat"
	}
	return "  Home"
}

// statusLine claude code / qodercli 风格单行状态条
// 紧凑 dim 色 · phase · wallet · chain · keys 五段
func (m Model) statusLine() string {
	dim := lipgloss.NewStyle().Foreground(cMuted).Faint(true)
	gold := lipgloss.NewStyle().Foreground(cNeonGold).Bold(true)
	cyan := lipgloss.NewStyle().Foreground(cNeonCyan).Bold(true)
	green := lipgloss.NewStyle().Foreground(cNeonGreen).Bold(true)

	phase := phaseBadge(m.phase)

	wallet := ""
	if m.user != nil {
		addr := m.userWalletAddr()
		if addr != "" && len(addr) > 10 {
			wallet = cyan.Render("🔑 " + addr[:6] + "…" + addr[len(addr)-4:])
		}
	}
	chain := green.Render("✅ Amoy")
	keys := dim.Render("↑↓ Nav · Enter Select · ? Help · Esc Back")

	sep := dim.Render("·")
	parts := []string{gold.Render("◆"), phase}
	if wallet != "" {
		parts = append(parts, wallet)
	}
	parts = append(parts, chain, keys)

	return "\n" + strings.Join(parts, " "+sep+" ") + "\n"
}

// phaseBadge 把 Phase 转成简短徽章文本
func phaseBadge(p Phase) string {
	gold := lipgloss.NewStyle().Foreground(cNeonGold).Bold(true)
	switch p {
	case PhaseRegister:
		return gold.Render("📝 REGISTER")
	case PhaseQuiz:
		return gold.Render("🧠 QUIZ")
	case PhaseResult:
		return gold.Render("🏁 RESULT")
	case PhaseMenu:
		return gold.Render("🏠 MENU")
	case PhaseMintCard:
		return gold.Render("🎴 MINT")
	case PhaseMyCards:
		return gold.Render("🃏 MY CARDS")
	case PhaseMarket:
		return gold.Render("🛒 MARKET")
	case PhaseListCard:
		return gold.Render("📋 LIST FOR SALE")
	case PhaseChat:
		return gold.Render("💬 CHAT")
	case PhaseQuitConfirm:
		return gold.Render("⚠ QUIT?")
	case PhaseHelp:
		return gold.Render("❓ HELP")
	}
	return gold.Render("HOME")
}

// headerBar 顶部条：logo + 面包屑 + 用户信息 + 全局返回提示
func (m Model) headerBar(title string) string {
	var b strings.Builder

	// 左侧：logo + 面包屑（替代原来的 title 参数）
	b.WriteString(styleLogo.Render(" ◆ geoyuan"))
	b.WriteString(styleHeader.Render(m.breadcrumb()))

	// 右侧：用户信息
	right := m.userInfoString()
	b.WriteString(strings.Repeat(" ", 2))
	b.WriteString(right)

	// 最右：全局 Esc 返回提示（不在 Menu/Chat 阶段显示）
	if m.phase != PhaseMenu && m.phase != PhaseChat {
		b.WriteString(strings.Repeat(" ", 2))
		b.WriteString(lipgloss.NewStyle().Foreground(cMuted).Render(" Esc←Back"))
	}

	// 分隔线
	b.WriteString("\n" + lipgloss.NewStyle().
		Foreground(colorBrand).
		Render(strings.Repeat("━", 80)) + "\n")

	return b.String()
}

// hintBar 底部统一导航提示（所有 View 共用）
func hintBar(keys ...string) string {
	if len(keys) == 0 {
		return ""
	}
	return "\n" + lipgloss.NewStyle().Foreground(cMuted).Render(
		strings.Join(keys, "   ·   ")) + "\n"
}

func (m Model) userInfoString() string {
	var parts []string

	if m.user != nil {
		parts = append(parts, styleChatSelfR.Render("👤 "+m.user.Username))
	}

	if m.remoteIP != "" {
		ip := m.remoteIP
		if len(ip) > 16 {
			ip = ip[:16] + ".."
		}
		parts = append(parts, styleHint.Render("📍 "+ip))
	}

	if m.remoteHost != "" && m.remoteHost != m.remoteIP {
		host := m.remoteHost
		if len(host) > 20 {
			host = host[:20] + ".."
		}
		parts = append(parts, styleHint.Render("🕸 "+host))
	}

	if m.ChatHub != nil {
		count := m.ChatHub.Count()
		if count > 0 {
			parts = append(parts, styleOnlineChip.Render(fmt.Sprintf("● %d 在线", count)))
		}
	}

	return strings.Join(parts, "  ")
}

// ═══════════════════════════════════════════════
//
//	📝 注册页 — 按终端高度自适应，任何尺寸都全屏完美
//
// ═══════════════════════════════════════════════
func (m Model) viewRegister() string {
	var b strings.Builder
	h := m.termHeight
	if h == 0 {
		h = 24
	}

	// 顶部：大 logo + tagline
	b.WriteString(RenderGlowLogo())
	b.WriteString(RenderNeonTagline())
	b.WriteString("\n")

	// 注册表单卡（带 label + 下划线 + 字数计数 + 校验 hint）
	name := m.nameInput.Value()
	hint := ""
	if name != "" {
		if len(name) < 3 {
			hint = "too short — minimum 3 characters"
		} else if len(name) > 20 {
			hint = "too long — maximum 20 characters"
		} else if !regexp.MustCompile(`^[a-zA-Z0-9_-]+$`).MatchString(name) {
			hint = "only letters, numbers, hyphen and underscore allowed"
		}
	}
	if m.inputErr != "" {
		hint = m.inputErr
	}
	b.WriteString(RenderRegisterCard(name, len(name), 20, hint))
	b.WriteString("\n")

	// 底部连接信息
	b.WriteString(m.connectionLine())

	return b.String()
}

// registerInput 已废弃 — viewRegister 直接用 RenderRegisterCard 完整替换
func (m Model) registerInput() string { return "" }

// connectionLine 一行版连接信息
func (m Model) connectionLine() string {
	loc := m.remoteIP
	switch {
	case m.remoteIP == "127.0.0.1" || m.remoteIP == "::1":
		loc = "localhost"
	case m.geoLoc != "":
		loc = m.geoLoc + " · " + m.remoteIP
	}
	return lipgloss.NewStyle().Foreground(cNeonCyan).Render("  📍 " + loc)
}

// oneLineWelcome 一行欢迎（Geoyuan 彩虹高亮）— 保留给小屏幕 fallback
func (m Model) oneLineWelcome() string {
	return "  🛰️  Welcome, Buddy — you've found the gateway to " + RainbowText("Geoyuan")
}

// registerInput 注册输入区（已废弃，viewRegister 直接调 RenderRegisterCard）
func (m Model) registerInput_DEPRECATED() string { return "" }

// headerBarCompact — 精简版 header，去掉右侧用户信息 + 80 宽分隔线（已废弃，保留兼容）
func (m Model) headerBarCompact() string {
	var b strings.Builder
	b.WriteString(styleLogo.Render(" ◆ geoyuan"))
	b.WriteString(styleHeader.Render("  ·  Join Geoyuan  ·"))
	b.WriteString("\n")
	// 短分隔线（只画 40 字符，避免宽终端看起来空）
	b.WriteString(lipgloss.NewStyle().Foreground(colorBrand).Render(strings.Repeat("━", 40)) + "\n")
	return b.String()
}

func (m Model) updateRegister(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.nameInput, cmd = m.nameInput.Update(msg)
	if msg.String() != "enter" {
		return m, cmd
	}
	name := strings.TrimSpace(m.nameInput.Value())

	// ── 支持在注册页直接输入 :redeem <token> 跳过注册 ──
	if strings.HasPrefix(name, ":redeem ") || strings.HasPrefix(name, "/redeem ") ||
		strings.HasPrefix(name, ":auth ") || strings.HasPrefix(name, "/auth ") {
		tok := strings.TrimSpace(strings.SplitN(name, " ", 2)[1])
		nm, ok := m.tryRedeem(tok)
		if ok {
			if nm.phase == PhaseWelcome {
				return nm, tea.Tick(1800*time.Millisecond, func(t time.Time) tea.Msg {
					return welcomeDoneMsg{}
				})
			}
			return nm, nil
		}
		return nm, cmd
	}

	if !usernameRe.MatchString(name) {
		m.inputErr = "用户名不合法：仅支持小写字母、数字、下划线，长度 2-20"
		return m, cmd
	}
	exist, err := m.store.GetUserByUsername(name)
	if err != nil {
		m.inputErr = "校验失败：" + err.Error()
		return m, cmd
	}
	if exist != nil {
		m.inputErr = "该用户名已被占用"
		return m, cmd
	}
	uid, err := m.store.CreateUser(name, m.fp)
	if err != nil {
		m.inputErr = "注册失败：" + err.Error()
		return m, cmd
	}
	m.user = &store.User{ID: uid, Username: name, PubKeyFingerprint: m.fp, Status: store.StatusPending}
	m.startQuiz()
	return m, nil
}

// tryRedeem 通用 token 兑换逻辑（注册页 / 重试页 / 聊天室 共用）
// 返回 (修改后的 Model, 是否成功触发跳转)
func (m Model) tryRedeem(token string) (Model, bool) {
	token = strings.TrimSpace(token)
	if token == "" {
		if m.phase == PhaseRegister {
			m.inputErr = "📝 用法: :redeem <token>"
		} else {
			m.chatLog = append(m.chatLog, chat.Message{
				Type: chat.MsgSystem, Content: "📝 用法: :redeem <token>", Time: time.Now(),
			})
		}
		return m, false
	}

	if m.store == nil {
		return m, false
	}

	wp, err := m.store.GetWebPassByToken(token)
	if err != nil || wp == nil {
		msg := "❌ token 无效或已过期。请在 diliy.cn 重新测试获取新 token。"
		if m.phase == PhaseRegister {
			m.inputErr = msg
		} else {
			m.chatLog = append(m.chatLog, chat.Message{
				Type: chat.MsgSystem, Content: msg, Time: time.Now(),
			})
		}
		return m, false
	}

	if !wp.Passed {
		msg := fmt.Sprintf("⚠ 该 token 对应的 Web 测试未通过（%d/%d）。", wp.Score, wp.Total)
		if m.phase == PhaseRegister {
			m.inputErr = msg
		} else {
			m.chatLog = append(m.chatLog, chat.Message{
				Type: chat.MsgSystem, Content: msg, Time: time.Now(),
			})
		}
		return m, false
	}

	// ✅ 通过！标记 used + 确保用户存在 + 升级 approved
	_ = m.store.MarkWebPassUsed(token)

	if m.user == nil {
		// 新用户：用 web_pass.Username 自动注册
		name := wp.Username
		exist, _ := m.store.GetUserByUsername(name)
		if exist != nil {
			// 用户名已被别人占（极小概率），用 fp 前缀
			name = fmt.Sprintf("user_%s", m.fp[:8])
			for i := 2; ; i++ {
				if e, _ := m.store.GetUserByUsername(name); e == nil {
					break
				}
				name = fmt.Sprintf("user_%s_%d", m.fp[:8], i)
			}
		}
		uid, err := m.store.CreateUser(name, m.fp)
		if err != nil {
			msg := "注册失败：" + err.Error()
			if m.phase == PhaseRegister {
				m.inputErr = msg
			}
			return m, false
		}
		m.user = &store.User{ID: uid, Username: name, PubKeyFingerprint: m.fp, Status: store.StatusPending}
	}
	if m.user.Status != store.StatusApproved {
		_, _ = m.store.DB().Exec("UPDATE users SET status = ? WHERE id = ?", store.StatusApproved, m.user.ID)
		m.user.Status = store.StatusApproved
	}

	// 🎉 成功！
	successMsg := fmt.Sprintf("🎉 兑换成功！Web pass (得分 %d/%d) 已绑定 %s。\n   现在你是 geoyuan 正式成员了 ✨",
		wp.Score, wp.Total, m.user.Username)

	if m.phase == PhaseRegister {
		// 注册页直接进欢迎页 → 聊天室
		m.inputErr = ""
		m.nameInput.SetValue("")
		m.chatLog = append(m.chatLog, chat.Message{
			Type: chat.MsgSystem, Content: successMsg, Time: time.Now(),
		})
		m.phase = PhaseWelcome
		return m, true
	}

	// 非注册页：发消息 + 进聊天室
	m.chatLog = append(m.chatLog, chat.Message{
		Type: chat.MsgSystem, Content: successMsg, Time: time.Now(),
	})

	if m.ChatHub != nil {
		m.ChatHub.Broadcast(chat.Message{
			Type:    chat.MsgSystem,
			Content: fmt.Sprintf("⚡ %s 通过 Web 测试（%d/%d）正式加入 geoyuan！", m.user.Username, wp.Score, wp.Total),
			Time:    time.Now(),
		})
	}

	// 如果当前不在聊天室，进入
	if m.phase != PhaseChat {
		m.enterChat()
	}

	return m, true
}

// ═══════════════════════════════════════════════
//
//	📋 答题页
//
// ═══════════════════════════════════════════════
func (m Model) updateQuiz(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if len(msg.Runes) > 0 {
		for _, r := range msg.Runes {
			if m.phase != PhaseQuiz {
				break
			}
			m.applyQuizKey(string(r))
		}
		return m, nil
	}
	m.applyQuizKey(msg.String())
	return m, nil
}

func (m *Model) applyQuizKey(s string) {
	switch s {
	case "up", "k":
		if m.selected > 0 {
			m.selected--
		}
	case "down", "j":
		if m.selected < 3 {
			m.selected++
		}
	case "1", "2", "3", "4":
		m.selected = int(s[0] - '1')
		m.commitAndAdvance()
	case "enter", "l", "right":
		m.commitAndAdvance()
	}
}

func (m Model) viewQuiz() string {
	var b strings.Builder
	q := m.questions[m.cursor]
	total := len(m.questions)

	b.WriteString(m.headerBar(fmt.Sprintf("·  AI RECRUITMENT QUIZ  ·  %d / %d", m.cursor+1, total)))

	var content strings.Builder
	// 📊 霓虹进度条
	content.WriteString(renderFancyProgress(m.cursor+1, total))
	content.WriteString("\n\n")

	// ❓ 题目卡片
	card := fmt.Sprintf("%s\n\n", styleQuestion.Render(fmt.Sprintf("❓  Q%d.  %s", m.cursor+1, q.Q)))

	for i, opt := range q.Options {
		var line string
		letter := string(rune('A' + i))
		marker := "  "
		if i == m.selected {
			line = styleOptionSelected.Render(fmt.Sprintf("%s ➤ %s. %s", marker, letter, opt))
		} else {
			line = styleOption.Render(fmt.Sprintf("%s    %s. %s", marker, letter, opt))
		}
		card += line + "\n"
	}

	content.WriteString(styleCard.Render(card))
	content.WriteString("\n")
	content.WriteString(lipgloss.NewStyle().Foreground(cMuted).Render("  ↑↓/jk to navigate  ·  1-4 to select & advance  ·  ⏎ to confirm  ·  Ctrl+C to exit"))

	b.WriteString(lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(cNeonPurple).
		Padding(0, 2).
		Width(72).
		Render(content.String()))

	b.WriteString(m.statusLine())
	return b.String()
}

func (m *Model) startQuiz() {
	// 用 generator 配置的 timeout，fallback 60s
	timeout := 60 * time.Second
	if m.generator != nil {
		timeout = m.generator.Cfg().Timeout
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var qs []questions.Question
	if m.generator != nil {
		var err error
		qs, err = m.generator.Generate(ctx, QuizCount)
		if err != nil {
			log.Printf("[tui] startQuiz fallback: %v", err)
			qs = questions.All()
		}
	} else {
		qs = questions.All()
	}

	m.questions = qs
	m.answers = make([]int, len(qs))
	for i := range m.answers {
		m.answers[i] = -1
	}
	m.cursor = 0
	m.selected = 0
	m.phase = PhaseQuiz
}

func (m *Model) commitAndAdvance() {
	m.answers[m.cursor] = m.selected
	if m.cursor < len(m.questions)-1 {
		m.cursor++
		m.selected = 0
	} else {
		m.finishQuiz()
	}
}

func (m *Model) finishQuiz() {
	total := len(m.questions)
	score := 0
	for i, q := range m.questions {
		if m.answers[i] == q.Answer {
			score++
		}
	}
	passed := score >= PassThreshold
	_ = m.store.RecordAttempt(m.user.ID, score, total, passed)
	m.score = score
	m.passed = passed
	m.Total = total
	m.Score = score
	m.Passed = passed
	m.UserID = m.user.ID
	if passed {
		m.user.Status = store.StatusApproved
	}
	m.phase = PhaseResult
}

// ═══════════════════════════════════════════════
//
//	🏆 结果页 — 霓虹风
//
// ═══════════════════════════════════════════════
func (m Model) viewResult() string {
	var b strings.Builder
	total := len(m.questions)
	pct := 0
	if total > 0 {
		pct = m.score * 100 / total
	}

	b.WriteString(m.headerBar("·  RECRUITMENT RESULT  ·"))

	var content strings.Builder

	if m.passed {
		// ── PASS: 霓虹恭喜画面 ──
		content.WriteString(RenderCongratsPass(m.score, total))
		content.WriteString("\n")
		content.WriteString(renderScoreDashboard(m.score, total, pct, m.passed))
		content.WriteString("\n\n")

		content.WriteString(styleSuccessText.Render(
			"  ✓  Welcome aboard, " + m.user.Username + "!\n" +
				"  ✓  Private chat room unlocked\n" +
				"  ✓  All members are waiting for you",
		))
		content.WriteString("\n\n")
		content.WriteString(lipgloss.NewStyle().Foreground(cNeonCyan).Render("  ⏎  Enter to open main menu  ·  q/esc to exit"))
	} else {
		// ── FAIL: 霓虹未通过画面 ──
		content.WriteString(RenderTryAgainFail(m.score, total))
		content.WriteString("\n")
		content.WriteString(renderScoreDashboard(m.score, total, pct, m.passed))
		content.WriteString("\n\n")

		content.WriteString(styleDangerText.Render(
			fmt.Sprintf("  ✗  Not yet — need ≥ %d / %d to pass", PassThreshold, total)))
		content.WriteString("\n\n")
		content.WriteString(lipgloss.NewStyle().Foreground(cMuted).Render("  Reconnect via ssh diliy.cn to try again  ·  ⏎/q to exit"))
	}

	// 答题回顾（前 8 题）
	content.WriteString("\n\n")
	reviewHeader := lipgloss.NewStyle().Foreground(cNeonGold).Bold(true).Render(
		"┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━  QUIZ REVIEW (first 8)  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓\n",
	)
	content.WriteString(reviewHeader)
	content.WriteString("\n")
	limit := 8
	if len(m.questions) < limit {
		limit = len(m.questions)
	}
	for i := 0; i < limit; i++ {
		q := m.questions[i]
		ans := m.answers[i]
		ok := ans == q.Answer
		mark := styleMarkBad.Render("  ✗")
		if ok {
			mark = styleMarkOK.Render("  ✓")
		}
		content.WriteString(fmt.Sprintf("%s  %d. %s\n", mark, i+1, q.Q))
		if !ok {
			content.WriteString(styleExplain.Render(fmt.Sprintf(
				"  Your answer: %s   Correct: %s   💡 %s",
				letter(ans), letter(q.Answer), q.Explain)) + "\n")
		}
		content.WriteString("\n")
	}
	if len(m.questions) > limit {
		content.WriteString(styleHint.Render(fmt.Sprintf("  ... and %d more questions omitted", len(m.questions)-limit)))
	}

	b.WriteString(lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(cNeonPurple).
		Padding(0, 2).
		Width(72).
		Render(content.String()))

	b.WriteString(m.statusLine())
	return b.String()
}

// renderScoreDashboard 渲染一个精致的分数仪表盘
func renderScoreDashboard(score, total, pct int, passed bool) string {
	var b strings.Builder

	color := colorDanger
	borderStyle := styleCardDanger
	bigScoreColor := styleScoreBig
	if passed {
		color = colorSuccess
		borderStyle = styleCardSuccess
		bigScoreColor = lipgloss.NewStyle().Bold(true).Foreground(colorGold).Align(lipgloss.Center)
	}

	// 大分数
	b.WriteString(bigScoreColor.Render(fmt.Sprintf("%d / %d", score, total)))
	b.WriteString("\n")

	// 百分比 + 进度条
	var pctStyle lipgloss.Style
	if passed {
		pctStyle = lipgloss.NewStyle().Foreground(colorSuccess).Bold(true)
	} else {
		pctStyle = lipgloss.NewStyle().Foreground(colorDanger).Bold(true)
	}
	b.WriteString(pctStyle.Render(fmt.Sprintf("%d%%", pct)))
	b.WriteString("  ")
	b.WriteString(renderFancyProgress(score, total))

	// 通过线标记
	threshPct := PassThreshold * 100 / total
	b.WriteString(fmt.Sprintf("\n\n目标线: ≥ %d 分 (%d%%)", PassThreshold, threshPct))

	_ = color
	return borderStyle.Render(b.String())
}

// renderFancyProgress 精致渐变进度条
func renderFancyProgress(cur, total int) string {
	if total <= 0 {
		return ""
	}
	width := 40
	filled := cur * width / total
	if filled > width {
		filled = width
	}

	// 渐变填充
	filledStr := strings.Repeat("█", filled)
	emptyStr := strings.Repeat("░", width-filled)

	filledColor := colorDanger
	if cur*100/total >= 80 {
		filledColor = colorGold
	}
	if cur*100/total >= PassThreshold*100/total {
		filledColor = colorSuccess
	}

	bar := lipgloss.NewStyle().Foreground(filledColor).Render(filledStr) +
		lipgloss.NewStyle().Foreground(colorDim).Render(emptyStr)

	return fmt.Sprintf("[%s]  %d/%d  (%d%%)", bar, cur, total, cur*100/total)
}

// ═══════════════════════════════════════════════
//
//	👋 欢迎页（已通过用户，等用户按 Enter 手动进入主菜单）
//
// ═══════════════════════════════════════════════
func (m Model) viewWelcome() string {
	var b strings.Builder
	b.WriteString(m.headerBar("·  WELCOME BACK  ·"))

	var content strings.Builder

	// 🌈 小彩虹 logo
	content.WriteString(RenderRainbowLogo() + "\n")

	n, _ := m.store.CountApproved()
	totalAttempts, _ := m.store.CountAttemptsForUser(m.user.ID)

	// 🧑 用户详情卡
	fp := m.fp
	if len(fp) > 32 {
		fp = fp[:32] + "..."
	}
	userCard := RenderUserCard(
		m.user.Username,
		fp,
		m.remoteIP,
		true, // approved
		n,
		totalAttempts,
	)
	content.WriteString(styleCardSuccess.Render(userCard))

	// ✨ 手动导航提示（不再自动 skip）
	content.WriteString("\n  " + lipgloss.NewStyle().
		Foreground(cNeonCyan).
		Bold(true).
		Render("✓  验证通过！按 Enter 进入主菜单"))

	// 底部 hint
	content.WriteString(hintBar("⏎ Enter → Main Menu", "q / Esc → Exit"))

	b.WriteString(lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(cNeonPurple).
		Padding(0, 2).
		Width(72).
		Render(content.String()))

	b.WriteString(m.statusLine())
	return b.String()
}

// ═══════════════════════════════════════════════
//
//	🔁 重测选择
//
// ═══════════════════════════════════════════════
func (m Model) updateRetry(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "left", "up", "h", "k", "tab":
		if m.retrySel > 0 {
			m.retrySel--
		}
	case "right", "down", "l", "j":
		if m.retrySel < 2 {
			m.retrySel++
		}
	case "enter":
		if m.retrySel == 0 {
			m.startQuiz()
			return m, nil
		} else if m.retrySel == 1 {
			// 退回注册页（支持 :redeem token）
			ti := textinput.New()
			ti.Placeholder = "输入 :redeem <token> 或新用户名"
			ti.CharLimit = 120
			ti.Prompt = "> "
			ti.Focus()
			m.nameInput = ti
			m.inputErr = ""
			m.phase = PhaseRegister
			return m, nil
		}
		m.Done = true
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) viewRetry() string {
	var b strings.Builder
	b.WriteString(m.headerBar("· 再次挑战 ·"))

	var content strings.Builder
	content.WriteString(styleCardDanger.Render(fmt.Sprintf(
		"  ⚠  %s，你尚未通过准入测试\n\n  当前状态: pending（未通过）",
		m.user.Username,
	)))

	content.WriteString("\n  你可以：\n\n")

	labels := []string{
		"    🎯  重新挑战 AI 题目",
		"    🪙  用 Web token 兑换 (:redeem <token>)",
		"    🚪  退出连接",
	}
	for i, l := range labels {
		marker := "  "
		if i == m.retrySel {
			marker = styleCursor.Render("> ")
		}
		content.WriteString(marker + l + "\n")
	}

	content.WriteString("\n" + styleHint.Render("↑↓ 选择 · 回车确认"))

	b.WriteString(lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(cNeonPurple).
		Padding(0, 2).
		Width(72).
		Render(content.String()))

	b.WriteString(m.statusLine())
	return b.String()
}

// ═══════════════════════════════════════════════
//
//	💬 聊天室
//
// ═══════════════════════════════════════════════
func (m Model) updateChat(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.chatInput, cmd = m.chatInput.Update(msg)

	if msg.String() != "enter" {
		return m, cmd
	}

	input := strings.TrimSpace(m.chatInput.Value())
	if input == "" {
		return m, cmd
	}

	if strings.HasPrefix(input, ":") {
		parts := strings.Fields(input)
		switch strings.ToLower(parts[0]) {
		case ":quit", ":exit", ":q":
			m.cleanupChat()
			m.Done = true
			return m, tea.Quit
		case ":help":
			m.chatLog = append(m.chatLog, chat.Message{
				Type: chat.MsgSystem,
				Content: "━━━ 命令列表 ━━━\n" +
					"  :quit / :exit / :q     退出聊天室\n" +
					"  :help                  显示此帮助\n" +
					"  :online                查看在线成员\n" +
					"  :w <user> <msg>        发送私聊消息\n" +
					"  :dm <user> <msg>       同上别名\n" +
					"  /me <action>           发送动作消息",
				Time: time.Now(),
			})
		case ":online":
			if m.ChatHub != nil {
				users := m.ChatHub.Users()
				m.chatLog = append(m.chatLog, chat.Message{
					Type:    chat.MsgSystem,
					Content: fmt.Sprintf("👥 在线 (%d): %s", len(users), strings.Join(users, " · ")),
					Time:    time.Now(),
				})
			}
		case ":w", ":dm":
			if len(parts) < 3 {
				m.chatLog = append(m.chatLog, chat.Message{
					Type:    chat.MsgSystem,
					Content: "⚠ 用法: :w <用户名> <消息内容>",
					Time:    time.Now(),
				})
			} else if m.ChatHub != nil && m.user != nil {
				toName := parts[1]
				content := strings.Join(parts[2:], " ")
				m.ChatHub.SendPrivate(m.user.ID, m.user.Username, toName, content)
			}
		case ":redeem", "/redeem", ":auth", "/auth":
			m, _ = m.tryRedeem(strings.Join(parts[1:], " "))
		default:
			m.chatLog = append(m.chatLog, chat.Message{
				Type:    chat.MsgSystem,
				Content: "⚠ 未知命令。输入 :help 查看帮助。",
				Time:    time.Now(),
			})
		}
	} else if strings.HasPrefix(input, "/me ") {
		// 动作消息
		action := strings.TrimPrefix(input, "/me ")
		msg := chat.Message{
			Type:    chat.MsgChat,
			UserID:  m.user.ID,
			User:    m.user.Username,
			Content: action,
			Time:    time.Now(),
		}
		// 自定义格式（斜体动作）
		m.ChatHub.Broadcast(chat.Message{
			Type:    chat.MsgSystem,
			Content: fmt.Sprintf("  ~~ %s %s", m.user.Username, action),
			Time:    time.Now(),
		})
		_ = msg
	} else if m.ChatHub != nil && m.user != nil {
		m.ChatHub.SendChat(m.user.ID, m.user.Username, input)
	}

	m.chatInput.SetValue("")
	return m, cmd
}

func (m Model) viewChat() string {
	var b strings.Builder
	count := 0
	if m.ChatHub != nil {
		count = m.ChatHub.Count()
	}

	// Header
	b.WriteString(m.headerBar(fmt.Sprintf("· 💬 私密聊天室  |  在线 %d ·", count)))

	var content strings.Builder

	// 聊天记录
	start := 0
	if len(m.chatLog) > 80 {
		start = len(m.chatLog) - 80
	}
	for _, msg := range m.chatLog[start:] {
		content.WriteString(renderChatMessage(msg, m.user))
		content.WriteString("\n")
	}
	if start > 0 {
		content.WriteString(styleHint.Render(fmt.Sprintf("... (省略 %d 条历史)", start)) + "\n")
	}

	// 输入框
	content.WriteString("\n")
	content.WriteString(lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorAccent).
		Padding(0, 1).
		Render(m.chatInput.View()))

	content.WriteString("\n" + styleHint.Render("Enter 发送 · /w 私聊 · :help 命令 · :quit 退出 · Ctrl+C 断开"))

	b.WriteString(lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(cNeonPurple).
		Padding(0, 2).
		Width(72).
		Render(content.String()))

	b.WriteString(m.statusLine())
	return b.String()
}

// renderChatMessage 渲染单条聊天消息，区分自己/他人/DM/系统
func renderChatMessage(msg chat.Message, currentUser *store.User) string {
	ts := msg.Time.Format("15:04:05")
	tsStyle := lipgloss.NewStyle().Foreground(colorDim)

	switch msg.Type {
	case chat.MsgSystem:
		return styleChatSystemR.Render(fmt.Sprintf("%s", msg.Content))

	case chat.MsgPrivate:
		sender := styleChatDMR.Render(msg.User)
		return fmt.Sprintf("%s %s → DM %s: %s",
			tsStyle.Render("["+ts+"]"),
			sender,
			msg.TargetName,
			msg.Content)

	default: // MsgChat
		var nameStyle lipgloss.Style
		if currentUser != nil && msg.User == currentUser.Username {
			nameStyle = styleChatSelfR
		} else {
			nameStyle = styleChatOtherR
		}
		return fmt.Sprintf("%s %s: %s",
			tsStyle.Render("["+ts+"]"),
			nameStyle.Render(msg.User),
			msg.Content)
	}
}

// ═══════════════════════════════════════════════
//
//	🔧 工具函数
//
// ═══════════════════════════════════════════════
func letter(i int) string {
	if i < 0 || i > 3 {
		return "-"
	}
	return string(rune('A' + i))
}

// ═══════════════════════════════════════════════
//  📋 Main Menu — 主菜单（Welcome 之后进入）
// ═══════════════════════════════════════════════

var menuItems = []struct {
	icon  string
	label string
	desc  string
}{
	{"🎴", "Mint a Geoyuan Card", "Create your permanent geodigital asset on Polygon"},
	{"🃏", "My Cards", "View your minted cards and H3 cells"},
	{"🛒", "Marketplace", "Browse & trade cards listed by the community"},
	{"💬", "Crew Chat", "Talk with other Geoyuan builders"},
	{"🌐", "Open geoyuan.com", "Learn more about the Network State"},
	{"🚪", "Exit", "Disconnect from SSH"},
}

func (m Model) viewMenu() string {
	var b strings.Builder
	b.WriteString(m.headerBar(" · Main Menu"))
	b.WriteString("\n")

	// 大 logo + 欢迎标语
	b.WriteString(RenderGlowLogo())
	b.WriteString(lipgloss.NewStyle().Foreground(cNeonPurple).Bold(true).
		Render("  ◆ Welcome back, Buddy — pick your mission") + "\n\n")

	// Panel Card 网格：3 列 × 2 行
	cols := 3
	padX := 2 // 两个 panel 之间的水平间距

	// 先构建 2 行 panel 网格
	linesPerCard := 5
	rows := [][]string{{}, {}} // row 0, row 1 (3 + 3)

	for i, item := range menuItems {
		row := i / cols
		if row > 1 {
			row = 1 // 安全 fallback
		}
		idxInRow := i % cols

		// panel border + 内容
		selected := i == m.menuCursor
		panel := renderMenuPanel(item, idxInRow, selected)
		rows[row] = append(rows[row], panel)
	}

	// 拼 panel 网格：逐行，每行把 3 个 panel 水平拼起来
	for _, row := range rows {
		var rowLines [][]string // 每 panel 的行
		for _, panel := range row {
			rowLines = append(rowLines, strings.Split(panel, "\n"))
		}
		for l := 0; l < linesPerCard; l++ {
			line := ""
			for pi, panelLines := range rowLines {
				if pi > 0 {
					line += strings.Repeat(" ", padX)
				}
				if l < len(panelLines) {
					line += panelLines[l]
				} else {
					// 补空格对齐
					line += strings.Repeat(" ", 28)
				}
			}
			b.WriteString(line + "\n")
		}
		b.WriteString("\n") // panel 行之间空一行
	}

	// 底部键位 hint
	b.WriteString(m.statusLine())
	return b.String()
}

// renderMenuPanel 渲染单个菜单 panel card
// idxInRow: 0/1/2 第几个位置，决定数字快捷键 badge
// selected: 是否当前选中
func renderMenuPanel(item struct {
	icon, label, desc string
}, idxInRow int, selected bool) string {
	numBadge := lipgloss.NewStyle().Foreground(cMuted).Render(fmt.Sprintf("[%d]", idxInRow+1))

	// 边框颜色 + 背景
	borderFG := cBorder
	bg := lipgloss.Color("#1a1a2e")
	titleColor := cText
	descColor := cMuted
	iconStyle := lipgloss.NewStyle().Foreground(cText)

	if selected {
		borderFG = cNeonCyan
		titleColor = cNeonCyan
		descColor = cNeonGreen
		iconStyle = lipgloss.NewStyle().Foreground(cNeonGold).Bold(true)
	}

	border := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderFG).
		Padding(0, 1).
		Width(28).
		Background(bg)

	titleS := lipgloss.NewStyle().Foreground(titleColor).Bold(true)
	descS := lipgloss.NewStyle().Foreground(descColor).Faint(true)

	// panel 内容
	content := fmt.Sprintf("%s %s\n%s\n%s",
		iconStyle.Render(item.icon),
		titleS.Render(truncate(item.label, 20)),
		"",
		descS.Render(truncate(item.desc, 24)),
	)
	if selected {
		content = numBadge + "  " + content
	} else {
		content = "    " + content
	}

	// 每 panel 5 行高（border 占 3 行 + padding）
	return border.Render(content)
}

func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max-1]) + "…"
}

func (m Model) updateMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	cols := 3
	rows := (len(menuItems) + cols - 1) / cols

	switch msg.String() {
	case "up", "k":
		// 上一行：cursor - cols
		if m.menuCursor >= cols {
			m.menuCursor -= cols
		} else {
			// 第一行 wrap 到最后一行同列
			m.menuCursor = m.menuCursor + (rows-1)*cols
			if m.menuCursor >= len(menuItems) {
				m.menuCursor = len(menuItems) - 1
			}
		}
	case "down", "j":
		// 下一行：cursor + cols
		if m.menuCursor+cols < len(menuItems) {
			m.menuCursor += cols
		} else {
			// 最后一行 wrap 到第一行同列
			m.menuCursor = m.menuCursor % cols
		}
	case "left", "h":
		if m.menuCursor > 0 {
			m.menuCursor--
		} else {
			m.menuCursor = len(menuItems) - 1
		}
	case "right", "l":
		if m.menuCursor < len(menuItems)-1 {
			m.menuCursor++
		} else {
			m.menuCursor = 0
		}
	case "1", "2", "3", "4", "5", "6":
		// 数字快捷键：直接跳到对应项并自动 Enter 触发
		idx, _ := strconv.Atoi(msg.String())
		if idx >= 1 && idx <= len(menuItems) {
			m.menuCursor = idx - 1
			// 自动执行和 Enter 一样的跳转逻辑
			return m.enterMenuSelection()
		}
	case "enter", " ":
		return m.enterMenuSelection()
	case "q", "esc":
		m.cleanupChat()
		m.Done = true
		return m, tea.Quit
	}
	return m, nil
}

// enterMenuSelection 执行主菜单当前选中项的跳转
// （数字快捷键 1-6 也走这里，抽出来避免 switch 重复）
func (m Model) enterMenuSelection() (tea.Model, tea.Cmd) {
	switch m.menuCursor {
	case 0: // 1 → Mint
		m.mintView = NewPhaseMint(m.mintFn(), m.chainReady(), m.userWalletAddr(), m.remoteIP, m.geoLoc)
		m.phase = PhaseMintCard
		return m, nil
	case 1: // 2 → My Cards
		m.myCardsLoading = true
		m.myCardsErr = nil
		m.showNeighbors = false
		m.phase = PhaseMyCards
		return m, m.loadMyCardsCmd()
	case 2: // 3 → Marketplace
		m.marketLoading = true
		m.marketErr = ""
		m.marketOffset = 0
		m.marketListingSel = 0
		m.phase = PhaseMarket
		return m, m.loadMarketCmd()
	case 3: // 4 → Crew Chat
		m.enterChat()
		return m, tea.Batch(textinput.Blink, receiveMsgs(m.sub))
	case 4: // 5 → geoyuan.com — 显示 URL 后退出
		fmt.Print("\n  🌐 Visit https://www.geoyuan.com/\n\n")
		m.Done = true
		return m, tea.Quit
	case 5: // 6 → Exit
		m.cleanupChat()
		m.Done = true
		return m, tea.Quit
	}
	return m, nil
}

// ═══════════════════════════════════════════════
//
//	🎴 Mint 视图处理器
//
// ═══════════════════════════════════════════════
func (m Model) updateMint(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.mintView, cmd = m.mintView.Update(msg)
	if m.mintView.Done() {
		m.phase = PhaseMenu
		m.menuCursor = 0
	}
	return m, cmd
}

// ═══════════════════════════════════════════════
//
//	🃏 My Cards 视图（链上 myCards + metadata）
//
// ═══════════════════════════════════════════════
func (m Model) loadMyCardsCmd() tea.Cmd {
	svc := m.cardService
	addr := m.userWalletAddr()
	if addr == "" {
		addr = "0x0000000000000000000000000000000000000000"
	}
	return func() tea.Msg {
		if svc == nil {
			return myCardsLoadedMsg{err: fmt.Errorf("GeoyuanCard service not available (chain not ready)")}
		}
		cards, err := svc.MyCards(addr)
		if err != nil {
			return myCardsLoadedMsg{err: err}
		}
		return myCardsLoadedMsg{cards: cards}
	}
}

func (m Model) loadNeighborCmd(h3Index uint64, k int) tea.Cmd {
	svc := m.cardService
	addr := m.userWalletAddr()
	if addr == "" {
		addr = "0x0000000000000000000000000000000000000000"
	}
	return func() tea.Msg {
		if svc == nil {
			return neighborLoadedMsg{err: fmt.Errorf("GeoyuanCard service not available")}
		}
		cards, err := svc.NeighborMatch(addr, h3Index, k)
		if err != nil {
			return neighborLoadedMsg{err: err}
		}
		return neighborLoadedMsg{cards: cards}
	}
}

// userWalletAddr — 当前用户的钱包地址（每次按需计算，redeem/注册后会变化）
// 未启用用户独立钱包时回退到部署钱包地址。
func (m Model) userWalletAddr() string {
	if m.walletMgr != nil && m.user != nil && m.user.ID > 0 {
		if a := m.walletMgr.AddressFor(m.user.ID); a != "" {
			return a
		}
	}
	if m.cardService != nil {
		return m.cardService.FromAddr()
	}
	return ""
}

// mintFn — 返回绑定到当前用户钱包的 mint 函数
func (m Model) mintFn() func(card.MintParams) (*card.MintResult, error) {
	mgr := m.walletMgr
	uid := int64(0)
	if m.user != nil {
		uid = m.user.ID
	}
	return func(p card.MintParams) (*card.MintResult, error) {
		if mgr == nil {
			return nil, fmt.Errorf("链上服务不可用")
		}
		return mgr.Mint(uid, p)
	}
}

// chainReady — 链上 mint 是否可用
func (m Model) chainReady() bool {
	return m.walletMgr != nil && m.walletMgr.Ready()
}

func (m Model) viewMyCards() string {
	var b strings.Builder
	b.WriteString(m.headerBar(" · My Cards"))
	b.WriteString("\n")
	neon := lipgloss.NewStyle().Foreground(lipgloss.Color("201")).Bold(true)
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	green := lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true)
	red := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	cyan := lipgloss.NewStyle().Foreground(lipgloss.Color("45")).Bold(true)

	var content strings.Builder

	if m.myCardsLoading {
		content.WriteString("  " + m.spinner() + "  Loading your cards from Polygon Amoy...\n")
		content.WriteString(dim.Render("  (this may take 10-30 seconds)\n"))
	} else if m.myCardsErr != nil {
		content.WriteString(red.Render(fmt.Sprintf("  ✗ Failed to load cards: %v", m.myCardsErr)) + "\n\n")
		content.WriteString(dim.Render("  Press R to retry · Esc to return\n"))
	} else if !m.showNeighbors {
		// 我的卡片列表
		content.WriteString(neon.Render("  🃏 Your Geoyuan Cards") + "\n")
		if m.chainReady() {
			content.WriteString(green.Render("  ✓ Connected to Polygon Amoy") + "\n")
		} else {
			content.WriteString(dim.Render("  (chain not ready — showing cache only)") + "\n")
		}
		if wa := m.userWalletAddr(); wa != "" {
			content.WriteString(dim.Render("  👛 Wallet: "+wa) + "\n")
		}
		content.WriteString("\n")

		if len(m.myCardsCache) == 0 {
			content.WriteString(dim.Render("  📭 You don't own any cards yet.\n"))
			content.WriteString("  Mint your first card from the Main Menu → 🎴 Mint.\n\n")
		} else {
			for _, c := range m.myCardsCache {
				content.WriteString(fmt.Sprintf("  🎴 Token #%s\n", c.TokenID))
				if c.Title != "" {
					content.WriteString(fmt.Sprintf("     Title:   %s\n", c.Title))
				}
				if c.Desc != "" {
					content.WriteString(fmt.Sprintf("     Desc:    %s\n", truncateLine(c.Desc, 60)))
				}
				if c.Color != "" {
					content.WriteString(fmt.Sprintf("     Color:   %s ", c.Color))
					content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(strings.TrimPrefix(c.Color, "#"))).Render("■"))
					content.WriteString("\n")
				}
				if c.H3Index != 0 {
					content.WriteString(fmt.Sprintf("     H3:      0x%x (res=%d)\n", c.H3Index, c.Res))
					content.WriteString(fmt.Sprintf("     Lat/Lon: %.4f, %.4f\n", c.Lat, c.Lon))
				}
				content.WriteString("\n")
			}
		}

		// 邻居匹配入口（基于第一张卡的 H3 cell）
		if len(m.myCardsCache) > 0 && m.myCardsCache[0].H3Index != 0 {
			first := m.myCardsCache[0]
			content.WriteString(dim.Render("  ─────────────────────────────────────\n"))
			content.WriteString("  Press " + cyan.Render("M") + " to find neighbor matches around your first card\n")
			content.WriteString(fmt.Sprintf("    Center cell: 0x%x (k=%d)\n", first.H3Index, m.neighborK))
		}
		content.WriteString(hintBar("R reload · M find neighbors · L list this card · Esc/← Back · q quit"))
	} else {
		// 邻居匹配结果
		content.WriteString(neon.Render(fmt.Sprintf("  🧭 Neighbor Match — k=%d around 0x%x", m.neighborK, m.neighborH3)) + "\n\n")
		if m.neighborLoading {
			content.WriteString("  " + m.spinner() + "  Loading neighbor matches...\n")
		} else if m.neighborErr != nil {
			content.WriteString(red.Render(fmt.Sprintf("  ✗ Neighbor query failed: %v", m.neighborErr)) + "\n\n")
		} else if len(m.neighborCache) == 0 {
			content.WriteString(dim.Render("  🗺️  No matching cards found in this area yet.\n"))
			content.WriteString("  More travelers minting nearby cards over time → matches will appear.\n\n")
		} else {
			content.WriteString(green.Render(fmt.Sprintf("  Found %d neighbor card(s):\n\n", len(m.neighborCache))))
			for _, c := range m.neighborCache {
				title := c.Title
				if title == "" {
					title = "(no title)"
				}
				content.WriteString(fmt.Sprintf("  🧩 #%s  %s\n", c.TokenID, title))
				if c.H3Index != 0 {
					content.WriteString(fmt.Sprintf("     cell: 0x%x  lat=%.4f lon=%.4f  res=%d\n", c.H3Index, c.Lat, c.Lon, c.Res))
				}
			}
			content.WriteString("\n")
		}
		content.WriteString(hintBar("Esc/← Back to My Cards · q quit"))
	}

	b.WriteString(lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(cNeonPurple).
		Padding(0, 2).
		Width(72).
		Render(content.String()))

	b.WriteString(m.statusLine())
	return b.String()
}

func truncateLine(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}

func (m Model) updateMyCards(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.phase = PhaseMenu
		m.showNeighbors = false
		return m, nil
	case "r", "R":
		if !m.myCardsLoading {
			m.myCardsLoading = true
			m.myCardsErr = nil
			return m, m.loadMyCardsCmd()
		}
	case "m", "M":
		if !m.showNeighbors && len(m.myCardsCache) > 0 && m.myCardsCache[0].H3Index != 0 && !m.neighborLoading {
			m.neighborK = 2
			m.neighborH3 = m.myCardsCache[0].H3Index
			m.neighborLoading = true
			m.neighborErr = nil
			m.showNeighbors = true
			return m, m.loadNeighborCmd(m.neighborH3, m.neighborK)
		}
	case "l", "L":
		if !m.showNeighbors && len(m.myCardsCache) > 0 {
			// 进入挂售 PhaseListCard（选卡 → 输入价 → 链上 list）
			m.listCardCursor = 0
			m.listCardErr = ""
			m.listCardDone = false
			m.listCardLoading = false
			m.listPriceWei = ""
			ti := textinput.New()
			ti.Prompt = "MATIC 价格: "
			ti.CharLimit = 20
			ti.Placeholder = "0.50 = 0.5 MATIC"
			ti.Focus()
			m.listInput = ti
			m.phase = PhaseListCard
			return m, nil
		}
	}
	return m, nil
}

// ═══════════════════════════════════════════════
//
//	🛒 Marketplace — 全市场挂售浏览 + 购买
//
// ═══════════════════════════════════════════════

const marketPageSize = 10

func (m Model) loadMarketCmd() tea.Cmd {
	svc := m.cardService
	offset := m.marketOffset
	return func() tea.Msg {
		if svc == nil {
			return marketLoadedMsg{err: fmt.Errorf("链上服务不可用")}
		}
		count, err := svc.ListingCount()
		if err != nil {
			return marketLoadedMsg{err: err}
		}
		listings, err := svc.MarketListings(marketPageSize, offset)
		if err != nil {
			return marketLoadedMsg{err: err}
		}
		return marketLoadedMsg{listings: listings, count: count}
	}
}

func (m Model) viewMarket() string {
	var b strings.Builder
	b.WriteString(m.headerBar(" · Marketplace"))
	b.WriteString("\n")
	neon := lipgloss.NewStyle().Foreground(lipgloss.Color("201")).Bold(true)
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	green := lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true)
	red := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)

	var content strings.Builder

	if m.marketLoading {
		content.WriteString("  " + m.spinner() + "  Loading Geoyuan Marketplace...\n")
	} else if m.marketErr != "" {
		content.WriteString(red.Render(fmt.Sprintf("  ✗ %v", m.marketErr)) + "\n\n")
		content.WriteString(dim.Render("  R retry · Esc back\n"))
	} else {
		content.WriteString(neon.Render(fmt.Sprintf("  🛒 Geoyuan Marketplace   (%d listing(s) total)", m.marketCount)) + "\n\n")

		if len(m.marketListings) == 0 {
			content.WriteString(dim.Render("  📭 No cards listed yet.\n"))
			content.WriteString("  Be the first — go to My Cards, select a card, press L to list.\n\n")
		} else {
			for i, l := range m.marketListings {
				prefix := "  "
				if i == m.marketListingSel {
					prefix = m.cursorMarker()
				}
				content.WriteString(prefix)
				content.WriteString(fmt.Sprintf("#%s  %s", l.TokenID, l.Title))
				content.WriteString("   ")
				content.WriteString(green.Render(fmt.Sprintf("◎ %s MATIC", l.PriceMATIC)))
				content.WriteString("\n")
				content.WriteString("    ")
				content.WriteString(dim.Render(fmt.Sprintf("Seller: %s  ·  cell: 0x%x  ·  %.4f,%.4f",
					shortAddr(l.Seller), l.H3Index, l.Lat, l.Lon)))
				content.WriteString("\n\n")
			}
		}

		// 分页
		if m.marketCount > marketPageSize {
			page := m.marketOffset / marketPageSize
			totalPages := (m.marketCount + marketPageSize - 1) / marketPageSize
			content.WriteString(dim.Render(fmt.Sprintf("  Page %d/%d", page+1, totalPages)) + "\n")
		}

		content.WriteString(hintBar("↑↓/jk select · Enter or B buy · N next page · P prev page · R reload · Esc back"))
	}

	b.WriteString(lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(cNeonPurple).
		Padding(0, 2).
		Width(72).
		Render(content.String()))

	b.WriteString(m.statusLine())
	return b.String()
}

func (m Model) updateMarket(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.phase = PhaseMenu
		return m, nil
	case "r", "R":
		if !m.marketLoading {
			m.marketLoading = true
			m.marketErr = ""
			return m, m.loadMarketCmd()
		}
	case "up", "k":
		if m.marketListingSel > 0 {
			m.marketListingSel--
		}
	case "down", "j":
		if m.marketListingSel < len(m.marketListings)-1 {
			m.marketListingSel++
		}
	case "n", "N":
		if m.marketOffset+marketPageSize < m.marketCount && !m.marketLoading {
			m.marketOffset += marketPageSize
			m.marketListingSel = 0
			m.marketLoading = true
			m.marketErr = ""
			return m, m.loadMarketCmd()
		}
	case "p", "P":
		if m.marketOffset > 0 && !m.marketLoading {
			m.marketOffset -= marketPageSize
			if m.marketOffset < 0 {
				m.marketOffset = 0
			}
			m.marketListingSel = 0
			m.marketLoading = true
			m.marketErr = ""
			return m, m.loadMarketCmd()
		}
	case "b", "B", "enter":
		if len(m.marketListings) == 0 {
			return m, nil
		}
		// B / Enter 都可以购买（enter 在 marketplace 视图里没有其他含义）
		// 买选中的 listing
		if !m.chainReady() {
			m.marketErr = "链上钱包未就绪，无法购买"
			return m, nil
		}
		l := m.marketListings[m.marketListingSel]
		return m, m.buyCmd(l.TokenID, l.PriceWei)
	}
	return m, nil
}

// ═══════════════════════════════════════════════
//
//	📦 List Card — 选卡 → 输入价 → 挂售
//
// ═══════════════════════════════════════════════

func (m Model) viewListCard() string {
	var b strings.Builder
	b.WriteString(m.headerBar(" · List Card for Sale"))
	b.WriteString("\n")
	neon := lipgloss.NewStyle().Foreground(lipgloss.Color("201")).Bold(true)
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	green := lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true)
	cyan := lipgloss.NewStyle().Foreground(lipgloss.Color("45")).Bold(true)
	red := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)

	var content strings.Builder

	if m.listCardLoading {
		content.WriteString("  " + m.spinner() + "  Submitting listing to Polygon Amoy...\n")
		content.WriteString(dim.Render("  (waiting for receipt, 10-60s)") + "\n")
		content.WriteString(hintBar("waiting..."))
	} else if m.listCardDone {
		content.WriteString(green.Render("  ✓ Listing successful!") + "\n\n")
		content.WriteString(dim.Render("  Your card is now on the marketplace."))
		content.WriteString(hintBar("Esc or Enter → Back to My Cards"))
	} else {
		// 选卡列表
		content.WriteString(neon.Render("  🃏 Pick a card to list") + "\n\n")
		if len(m.myCardsCache) == 0 {
			content.WriteString(red.Render("  ✗ No cards to list. Go mint first.\n"))
			content.WriteString(hintBar("Esc → Back to My Cards"))
		} else {
			for i, c := range m.myCardsCache {
				prefix := "  "
				if i == m.listCardCursor {
					prefix = m.cursorMarker()
				}
				content.WriteString(prefix)
				content.WriteString(fmt.Sprintf("#%s  %s", c.TokenID, c.Title))
				content.WriteString("\n")
				content.WriteString(dim.Render(fmt.Sprintf("    cell: 0x%x  ·  %.4f,%.4f", c.H3Index, c.Lat, c.Lon)))
				content.WriteString("\n\n")
			}

			// 价格输入
			content.WriteString(cyan.Render("  💰 Set price (MATIC):") + "\n\n")
			content.WriteString("  " + m.listInput.View() + "\n")
		}

		if m.listCardErr != "" {
			content.WriteString("\n\n  " + red.Render(m.listCardErr))
		}

		content.WriteString(hintBar("↑↓ pick card · Enter confirm · Esc cancel"))
	}

	b.WriteString(lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(cNeonPurple).
		Padding(0, 2).
		Width(72).
		Render(content.String()))

	b.WriteString(m.statusLine())
	return b.String()
}

func (m Model) updateListCard(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.listCardLoading {
		// loading 阶段也允许 Esc/q 紧急返回 Menu（防止死等链上）
		if msg.String() == "esc" || msg.String() == "q" {
			m.listCardLoading = false
			m.listCardErr = "(已取消挂售)"
			return m, nil
		}
		return m, nil
	}
	if m.listCardDone {
		if msg.String() == "esc" || msg.String() == "enter" {
			m.phase = PhaseMyCards
			m.listCardDone = false
			return m, m.loadMyCardsCmd()
		}
		return m, nil
	}

	var cmd tea.Cmd
	m.listInput, cmd = m.listInput.Update(msg)

	switch msg.String() {
	case "esc", "q":
		m.phase = PhaseMyCards
		return m, nil
	case "up", "k":
		if m.listCardCursor > 0 {
			m.listCardCursor--
		}
	case "down", "j":
		if m.listCardCursor < len(m.myCardsCache)-1 {
			m.listCardCursor++
		}
	case "enter":
		if len(m.myCardsCache) == 0 {
			return m, nil
		}
		priceStr := strings.TrimSpace(m.listInput.Value())
		if priceStr == "" {
			m.listCardErr = "请输入价格（0 = 免费）"
			return m, cmd
		}
		if !m.chainReady() {
			m.listCardErr = "链上钱包未就绪"
			return m, cmd
		}
		card := m.myCardsCache[m.listCardCursor]
		return m, m.listCardCmd(card.TokenID, priceStr)
	}
	return m, cmd
}

// ═══════════════════════════════════════════════
//
//	🔧 Async cmds — list/buy 链上操作
//
// ═══════════════════════════════════════════════

func (m Model) listCardCmd(tokenIDstr, priceMATICStr string) tea.Cmd {
	mgr := m.walletMgr
	if mgr == nil {
		return func() tea.Msg { return listCardDoneMsg{err: fmt.Errorf("钱包管理器不可用")} }
	}
	uid := int64(0)
	if m.user != nil {
		uid = m.user.ID
	}
	tid := new(big.Int)
	tid.SetString(tokenIDstr, 10)
	if tid.Sign() == 0 {
		return func() tea.Msg { return listCardDoneMsg{err: fmt.Errorf("无效 token ID: %s", tokenIDstr)} }
	}
	priceWei, err := maticToWei(priceMATICStr)
	if err != nil {
		return func() tea.Msg { return listCardDoneMsg{err: err} }
	}

	return func() tea.Msg {
		_, _, err := mgr.List(uid, tid, priceWei)
		if err != nil {
			return listCardDoneMsg{err: err}
		}
		return listCardDoneMsg{}
	}
}

func (m Model) buyCmd(tokenIDstr, priceWeiStr string) tea.Cmd {
	mgr := m.walletMgr
	if mgr == nil {
		return func() tea.Msg { return buyDoneMsg{err: fmt.Errorf("钱包管理器不可用")} }
	}
	uid := int64(0)
	if m.user != nil {
		uid = m.user.ID
	}
	tid := new(big.Int)
	tid.SetString(tokenIDstr, 10)
	if tid.Sign() == 0 {
		return func() tea.Msg { return buyDoneMsg{err: fmt.Errorf("无效 token ID")} }
	}
	valueWei, ok := new(big.Int).SetString(priceWeiStr, 10)
	if !ok {
		return func() tea.Msg { return buyDoneMsg{err: fmt.Errorf("无效价格: %s", priceWeiStr)} }
	}

	return func() tea.Msg {
		_, _, err := mgr.Buy(uid, tid, valueWei)
		if err != nil {
			return buyDoneMsg{err: err}
		}
		return buyDoneMsg{}
	}
}

// ── 小工具 ──
func shortAddr(addr string) string {
	if len(addr) <= 10 {
		return addr
	}
	return addr[:6] + "…" + addr[len(addr)-4:]
}

// maticToWei — MATIC 字符串 → wei big.Int
func maticToWei(s string) (*big.Int, error) {
	f, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return nil, fmt.Errorf("价格解析失败: %v", err)
	}
	if f < 0 {
		return nil, fmt.Errorf("价格不能为负")
	}
	// MATIC 18 位精度
	wei := new(big.Float).Mul(big.NewFloat(f), big.NewFloat(1e18))
	out := new(big.Int)
	wei.Int(out)
	return out, nil
}

// ═══════════════════════════════════════════════
//
//	⚠️ 退出确认弹窗（PhaseQuitConfirm）
//
// ═══════════════════════════════════════════════

func (m Model) viewQuitConfirm() string {
	// 先渲染底层页面（让弹窗显得是"浮层"）
	var bg string
	switch m.prevPhase {
	case PhaseMenu:
		bg = m.viewMenu()
	case PhaseWelcome:
		bg = m.viewWelcome()
	case PhaseMyCards:
		bg = m.viewMyCards()
	case PhaseMarket:
		bg = m.viewMarket()
	case PhaseListCard:
		bg = m.viewListCard()
	case PhaseMintCard:
		bg = m.mintView.View(m.termWidth, m.termHeight, m.geoLoc)
	case PhaseChat:
		bg = m.viewChat()
	default:
		bg = m.viewMenu()
	}

	// 半透明遮罩：用大量空格盖上去（近似效果）
	// 实际上 Bubble Tea 做真正 overlay 复杂，这里简化为底部弹出 modal
	var b strings.Builder
	b.WriteString(strings.Repeat("\n", 2)) // 留空

	warn := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	green := lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true)
	neon := lipgloss.NewStyle().Foreground(lipgloss.Color("201")).Bold(true)

	// 画一个居中的 modal 框
	boxTop := "┌────────────────────────────────────────────────────┐"
	boxMid := "│ " + warn.Render("⚠  Are you sure you want to quit Geoyuan?") + " │"
	boxLine := "│" + strings.Repeat(" ", 52) + "│"
	yesLine := "│   " + green.Render("y / Enter") + "  → Quit TUI                    │"
	noLine := "│   " + neon.Render("n / Esc") + "  → Stay here (back to " + phaseLabel(m.prevPhase) + ") │"
	boxBot := "└────────────────────────────────────────────────────┘"

	b.WriteString(boxTop + "\n")
	b.WriteString(boxMid + "\n")
	b.WriteString(boxLine + "\n")
	b.WriteString("│  This will disconnect your SSH session.\n")
	b.WriteString("│  No card / wallet data is ever stored.\n")
	b.WriteString(boxLine + "\n")
	b.WriteString(yesLine + "\n")
	b.WriteString(noLine + "\n")
	b.WriteString(boxBot + "\n")
	b.WriteString(hintBar("y / Enter → Quit", "n / Esc → Stay"))

	// 用 bg 作为底色（简化：直接返回 modal，不叠加到底层）
	_ = bg
	return b.String()
}

func (m Model) updateQuitConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y", "enter":
		m.cleanupChat()
		m.Done = true
		return m, tea.Quit
	case "n", "N", "esc":
		m.phase = m.prevPhase
		return m, nil
	}
	return m, nil
}

// phaseLabel 把 Phase 转成人眼可读的短标签
func phaseLabel(p Phase) string {
	switch p {
	case PhaseMenu:
		return "Main Menu"
	case PhaseWelcome:
		return "Welcome"
	case PhaseMyCards:
		return "My Cards"
	case PhaseMarket:
		return "Marketplace"
	case PhaseListCard:
		return "List for Sale"
	case PhaseMintCard:
		return "Mint Card"
	case PhaseChat:
		return "Crew Chat"
	}
	return "previous page"
}

// ═══════════════════════════════════════════════
//
//	? 帮助浮层（PhaseHelp）
//
// ═══════════════════════════════════════════════

func (m Model) viewHelpOverlay() string {
	var b strings.Builder
	neon := lipgloss.NewStyle().Foreground(lipgloss.Color("201")).Bold(true)
	cyan := lipgloss.NewStyle().Foreground(lipgloss.Color("45")).Bold(true)
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	green := lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true)

	// 顶部标题
	b.WriteString("\n  " + neon.Render("◆ GEoYUAN  KEYBOARD REFERENCE"))
	b.WriteString("  " + dim.Render("(press any key or Esc to close)"))
	b.WriteString("\n\n")

	// 全局（所有页面）
	b.WriteString("  " + cyan.Render("GLOBAL (any page)") + "\n")
	b.WriteString("  " + green.Render("?" + strings.Repeat(" ", 14)) + "Toggle this help\n")
	b.WriteString("  " + green.Render("Esc") + strings.Repeat(" ", 13) + "Back to previous / Quit confirm\n")
	b.WriteString("  " + green.Render("Ctrl+C") + strings.Repeat(" ", 10) + "Force quit\n")
	b.WriteString("\n")

	// 主菜单
	b.WriteString("  " + cyan.Render("MAIN MENU") + "\n")
	b.WriteString("  " + green.Render("↑ / ↓  or  j / k") + strings.Repeat(" ", 5) + "Navigate options\n")
	b.WriteString("  " + green.Render("1 - 6") + strings.Repeat(" ", 13) + "Jump directly to item\n")
	b.WriteString("  " + green.Render("Enter") + strings.Repeat(" ", 13) + "Select current\n")
	b.WriteString("\n")

	// My Cards
	b.WriteString("  " + cyan.Render("MY CARDS") + "\n")
	b.WriteString("  " + green.Render("R") + strings.Repeat(" ", 14) + "Reload from chain\n")
	b.WriteString("  " + green.Render("M") + strings.Repeat(" ", 14) + "Find neighbor cards\n")
	b.WriteString("  " + green.Render("L") + strings.Repeat(" ", 14) + "List this card for sale\n")
	b.WriteString("  " + green.Render("↑ / ↓") + strings.Repeat(" ", 10) + "Scroll cards\n")
	b.WriteString("\n")

	// Marketplace
	b.WriteString("  " + cyan.Render("MARKETPLACE") + "\n")
	b.WriteString("  " + green.Render("↑ / ↓  or  j / k") + strings.Repeat(" ", 5) + "Select listing\n")
	b.WriteString("  " + green.Render("Enter  or  B") + strings.Repeat(" ", 8) + "Buy selected\n")
	b.WriteString("  " + green.Render("N / P") + strings.Repeat(" ", 12) + "Next / Prev page\n")
	b.WriteString("  " + green.Render("R") + strings.Repeat(" ", 14) + "Reload listings\n")
	b.WriteString("\n")

	// List Card
	b.WriteString("  " + cyan.Render("LIST CARD FOR SALE") + "\n")
	b.WriteString("  " + green.Render("↑ / ↓  or  j / k") + strings.Repeat(" ", 5) + "Pick a card to list\n")
	b.WriteString("  " + green.Render("Enter") + strings.Repeat(" ", 13) + "Confirm listing\n")
	b.WriteString("  " + green.Render("Esc") + strings.Repeat(" ", 13) + "Cancel / go back\n")
	b.WriteString("\n")

	// Quiz
	b.WriteString("  " + cyan.Render("QUIZ") + "\n")
	b.WriteString("  " + green.Render("1 - 4") + strings.Repeat(" ", 13) + "Select answer\n")
	b.WriteString("  " + green.Render("Enter") + strings.Repeat(" ", 13) + "Next question\n")
	b.WriteString("\n")

	// Chat
	b.WriteString("  " + cyan.Render("CREW CHAT") + "\n")
	b.WriteString("  " + green.Render("Enter") + strings.Repeat(" ", 13) + "Send message\n")
	b.WriteString("  " + green.Render("Esc") + strings.Repeat(" ", 13) + "Leave chat\n")
	b.WriteString("\n")

	b.WriteString(dim.Render("  Press any key to close  ·  Enjoy Geoyuan! 🌍"))
	return b.String()
}

func (m Model) updateHelp(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// 任意键关闭帮助浮层，回到之前的页面
	m.phase = m.prevPhase
	return m, nil
}
