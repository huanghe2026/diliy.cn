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
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"geoyuan.cn/ssh/internal/chat"
	"geoyuan.cn/ssh/internal/questiongen"
	"geoyuan.cn/ssh/internal/questions"
	"geoyuan.cn/ssh/internal/store"
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
)

// ═══════════════════════════════════════════════
//  🎨 Color Palette — 全部移到 styles.go 统一管理
// ═══════════════════════════════════════════════

// ═══════════════════════════════════════════════
//  📐 Reusable Styles — 全部移到 styles.go 统一管理
// ═══════════════════════════════════════════════

var usernameRe = regexp.MustCompile(`^[a-z0-9_]{2,20}$`)

// ═══════════════════════════════════════════════
//  Model
// ═══════════════════════════════════════════════
type Model struct {
	store     *store.Store
	generator *questiongen.Generator
	fp        string
	user      *store.User
	phase     Phase

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

	// 状态
	Done   bool
	UserID int64
	Score  int
	Total  int
	Passed bool
}

// New 构造 TUI Model
func New(s *store.Store, gen *questiongen.Generator, fp string, user *store.User, chatHub *chat.Hub, ip, host, geo string) Model {
	m := Model{
		store:      s,
		generator:  gen,
		fp:         fp,
		user:       user,
		ChatHub:    chatHub,
		remoteIP:   ip,
		remoteHost: host,
		geoLoc:     geo,
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
		ti.CharLimit = 20
		ti.Prompt = "用户名: "
		ti.Focus()
		m.nameInput = ti
		m.phase = PhaseRegister
	}
	return m
}

// Init 实现 tea.Model
func (m Model) Init() tea.Cmd {
	switch m.phase {
	case PhaseRegister:
		return textinput.Blink
	case PhaseChat:
		return tea.Batch(textinput.Blink, receiveMsgs(m.sub))
	case PhaseWelcome:
		return tea.Tick(1800*time.Millisecond, func(t time.Time) tea.Msg {
			return welcomeDoneMsg{}
		})
	}
	return nil
}

type (
	chatMsgWrapper struct{ Msg chat.Message }
	chatClosedMsg  struct{}
	welcomeDoneMsg struct{}
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
	case welcomeDoneMsg:
		if m.phase == PhaseWelcome {
			m.phase = PhaseChat
			return m, tea.Batch(textinput.Blink, receiveMsgs(m.sub))
		}
	case chatMsgWrapper:
		if m.phase == PhaseChat {
			m.chatLog = append(m.chatLog, msg.Msg)
			return m, receiveMsgs(m.sub)
		}
	case chatClosedMsg:
		m.Done = true
		return m, tea.Quit

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.cleanupChat()
			m.Done = true
			return m, tea.Quit
		}
		switch m.phase {
		case PhaseRegister:
			return m.updateRegister(msg)
		case PhaseQuiz:
			return m.updateQuiz(msg)
		case PhaseResult:
			if msg.String() == "enter" || msg.String() == "q" || msg.String() == "esc" {
				if m.passed {
					m.enterChat()
					return m, tea.Batch(textinput.Blink, receiveMsgs(m.sub))
				}
				m.Done = true
				return m, tea.Quit
			}
		case PhaseWelcome:
			if msg.String() == "enter" || msg.String() == "q" || msg.String() == "esc" {
				m.enterChat()
				return m, tea.Batch(textinput.Blink, receiveMsgs(m.sub))
			}
		case PhaseRetry:
			return m.updateRetry(msg)
		case PhaseChat:
			return m.updateChat(msg)
		}
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
	case PhaseChat:
		return m.viewChat()
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
//  🏗️  Header Bar (所有页面共享)
// ═══════════════════════════════════════════════
func (m Model) headerBar(title string) string {
	var b strings.Builder

	// 左侧：logo + title
	b.WriteString(styleLogo.Render(" ◆ geoyuan"))
	b.WriteString(styleHeader.Render("  " + title))

	// 右侧：用户信息 + 在线
	right := m.userInfoString()
	b.WriteString(strings.Repeat(" ", 3))
	b.WriteString(right)

	// 分隔线
	b.WriteString("\n" + lipgloss.NewStyle().
		Foreground(colorBrand).
		Render(strings.Repeat("━", 80)) + "\n")

	return b.String()
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
//  📝 注册页 — 霓虹 Cyberpunk 风
// ═══════════════════════════════════════════════
func (m Model) viewRegister() string {
	var b strings.Builder
	b.WriteString(m.headerBar("· Join Geoyuan ·"))

	// 🌈 彩虹渐变 Geoyuan Logo
	b.WriteString(RenderRainbowLogo())
	b.WriteString(RenderNeonTagline())

	b.WriteString("\n")

	// 🛰️ 欢迎卡 (glow box)
	b.WriteString(RenderWelcomeCard())
	b.WriteString("\n")

	// 📍 连接信息（地理位置优先，降级 IP）
	if m.remoteIP != "" {
		var ipInfo string
		if m.geoLoc != "" {
			ipInfo = fmt.Sprintf("  📍 Connected from %s", m.geoLoc)
			// 本地连接 (127.0.0.1) 就不显示 IP 了
			if m.remoteIP != "127.0.0.1" && m.remoteIP != "::1" {
				ipInfo += fmt.Sprintf("  ·  %s", m.remoteIP)
			}
		} else {
			ipInfo = fmt.Sprintf("  📍 Connected from %s", m.remoteIP)
		}
		b.WriteString(lipgloss.NewStyle().Foreground(cNeonCyan).Render(ipInfo))
		b.WriteString("\n\n")
	}

	// 📝 注册输入卡
	ti := m.nameInput
	ti.Prompt = "  " + lipgloss.NewStyle().Foreground(cNeonCyan).Bold(true).Render(">") + " "
	nameInputLine := ti.View()

	registerCard := fmt.Sprintf(
		"  %s\n\n  %s\n\n  %s",
		lipgloss.NewStyle().Foreground(cNeonGold).Bold(true).Render("⚡  New Buddy Registration"),
		lipgloss.NewStyle().Foreground(cMuted).Render("  Enter your unique handle (2-20 chars, a-z/0-9/_)"),
		nameInputLine,
	)
	b.WriteString(styleCardAccent.Render(registerCard))

	if m.inputErr != "" {
		b.WriteString("\n  " + styleErr.Render("✗  "+m.inputErr))
	}

	b.WriteString("\n  " + lipgloss.NewStyle().Foreground(cMuted).Render("↵ Enter to confirm  ·  Ctrl+C to exit  ·  Try :redeem <token>"))
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
		m.user = &store.User{ID: uid, Username: name, PubKeyFingerprint: m.fp, Status: store.StatusApproved}
	} else if m.user.Status != store.StatusApproved {
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
			Type: chat.MsgSystem,
			Content: fmt.Sprintf("⚡ %s 通过 Web 测试（%d/%d）正式加入 geoyuan！", m.user.Username, wp.Score, wp.Total),
			Time: time.Now(),
		})
	}

	// 如果当前不在聊天室，进入
	if m.phase != PhaseChat {
		m.enterChat()
	}

	return m, true
}

// ═══════════════════════════════════════════════
//  📋 答题页
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

	// 📊 霓虹进度条
	b.WriteString(renderFancyProgress(m.cursor+1, total))
	b.WriteString("\n\n")

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

	b.WriteString(styleCard.Render(card))
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(cMuted).Render("  ↑↓/jk to navigate  ·  1-4 to select & advance  ·  ⏎ to confirm  ·  Ctrl+C to exit"))
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
//  🏆 结果页 — 霓虹风
// ═══════════════════════════════════════════════
func (m Model) viewResult() string {
	var b strings.Builder
	total := len(m.questions)
	pct := 0
	if total > 0 {
		pct = m.score * 100 / total
	}

	b.WriteString(m.headerBar("·  RECRUITMENT RESULT  ·"))

	if m.passed {
		// ── PASS: 霓虹恭喜画面 ──
		b.WriteString(RenderCongratsPass(m.score, total))
		b.WriteString("\n")
		b.WriteString(renderScoreDashboard(m.score, total, pct, m.passed))
		b.WriteString("\n\n")

		b.WriteString(styleSuccessText.Render(
			"  ✓  Welcome aboard, "+m.user.Username+"!\n"+
				"  ✓  Private chat room unlocked\n"+
				"  ✓  All members are waiting for you",
		))
		b.WriteString("\n\n")
		b.WriteString(lipgloss.NewStyle().Foreground(cNeonCyan).Render("  ⏎  Enter to enter private chat  ·  q/esc to exit"))
	} else {
		// ── FAIL: 霓虹未通过画面 ──
		b.WriteString(RenderTryAgainFail(m.score, total))
		b.WriteString("\n")
		b.WriteString(renderScoreDashboard(m.score, total, pct, m.passed))
		b.WriteString("\n\n")

		b.WriteString(styleDangerText.Render(
			fmt.Sprintf("  ✗  Not yet — need ≥ %d / %d to pass", PassThreshold, total)))
		b.WriteString("\n\n")
		b.WriteString(lipgloss.NewStyle().Foreground(cMuted).Render("  Reconnect via ssh diliy.cn to try again  ·  ⏎/q to exit"))
	}

	// 答题回顾（前 8 题）
	b.WriteString("\n\n")
	reviewHeader := lipgloss.NewStyle().Foreground(cNeonGold).Bold(true).Render(
		"┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━  QUIZ REVIEW (first 8)  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓\n",
	)
	b.WriteString(reviewHeader)
	b.WriteString("\n")
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
		b.WriteString(fmt.Sprintf("%s  %d. %s\n", mark, i+1, q.Q))
		if !ok {
			b.WriteString(styleExplain.Render(fmt.Sprintf(
				"  Your answer: %s   Correct: %s   💡 %s",
				letter(ans), letter(q.Answer), q.Explain)) + "\n")
		}
		b.WriteString("\n")
	}
	if len(m.questions) > limit {
		b.WriteString(styleHint.Render(fmt.Sprintf("  ... and %d more questions omitted", len(m.questions)-limit)))
	}

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
//  👋 欢迎页（已通过用户）
// ═══════════════════════════════════════════════
func (m Model) viewWelcome() string {
	var b strings.Builder
	b.WriteString(m.headerBar("·  WELCOME BACK  ·"))

	// 🌈 小彩虹 logo
	b.WriteString(RenderRainbowLogo() + "\n")

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
	b.WriteString(styleCardSuccess.Render(userCard))

	// ✨ 即将进入聊天室
	b.WriteString("\n  " + lipgloss.NewStyle().
		Foreground(cNeonCyan).
		Bold(true).
		Render("⚡  Entering private chat room..."))

	return b.String()
}

// ═══════════════════════════════════════════════
//  🔁 重测选择
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

	b.WriteString(styleCardDanger.Render(fmt.Sprintf(
		"  ⚠  %s，你尚未通过准入测试\n\n  当前状态: pending（未通过）",
		m.user.Username,
	)))

	b.WriteString("\n  你可以：\n\n")

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
		b.WriteString(marker + l + "\n")
	}

	b.WriteString("\n" + styleHint.Render("↑↓ 选择 · 回车确认"))
	return b.String()
}

// ═══════════════════════════════════════════════
//  💬 聊天室
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
					Type: chat.MsgSystem,
					Content: fmt.Sprintf("👥 在线 (%d): %s", len(users), strings.Join(users, " · ")),
					Time: time.Now(),
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

	// 聊天记录
	start := 0
	if len(m.chatLog) > 80 {
		start = len(m.chatLog) - 80
	}
	for _, msg := range m.chatLog[start:] {
		b.WriteString(renderChatMessage(msg, m.user))
		b.WriteString("\n")
	}
	if start > 0 {
		b.WriteString(styleHint.Render(fmt.Sprintf("... (省略 %d 条历史)", start)) + "\n")
	}

	// 输入框
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorAccent).
		Padding(0, 1).
		Render(m.chatInput.View()))

	b.WriteString("\n" + styleHint.Render("Enter 发送 · /w 私聊 · :help 命令 · :quit 退出 · Ctrl+C 断开"))

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
//  🔧 工具函数
// ═══════════════════════════════════════════════
func letter(i int) string {
	if i < 0 || i > 3 {
		return "-"
	}
	return string(rune('A' + i))
}
