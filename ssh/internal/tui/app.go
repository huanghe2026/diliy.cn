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
//  🎨 Color Palette
// ═══════════════════════════════════════════════
var (
	// Logo & Brand
	colorBrand      = lipgloss.Color("213") // 亮紫粉
	colorBrandDark  = lipgloss.Color("177") // 深紫粉
	colorAccent     = lipgloss.Color("51")  // 青蓝
	colorGold       = lipgloss.Color("220") // 金色
	colorSuccess    = lipgloss.Color("42")  // 成功绿
	colorDanger     = lipgloss.Color("203") // 危险红
	colorWarn       = lipgloss.Color("220") // 警告金

	// Neutrals
	colorText       = lipgloss.Color("250") // 主文字
	colorMuted      = lipgloss.Color("240") // 次要文字
	colorDim        = lipgloss.Color("237") // 更暗
	colorBorder     = lipgloss.Color("245") // 边框
	colorBorderSub  = lipgloss.Color("238") // 次级边框

	// Chat message colors
	colorSelf       = lipgloss.Color("51")  // 自己发的消息（青蓝）
	colorOther      = lipgloss.Color("213") // 他人消息（紫粉）
	colorDM         = lipgloss.Color("183") // 私聊（洋红）
	colorSystem     = lipgloss.Color("243") // 系统消息（灰）
)

// ═══════════════════════════════════════════════
//  📐 Reusable Styles
// ═══════════════════════════════════════════════
var (
	styleLogo = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorBrand).
			Padding(0, 1)

	styleHeader = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorBrand).
			MarginBottom(1)

	styleHeaderBar = lipgloss.NewStyle().
			Foreground(colorMuted)

	styleCard = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(1, 2).
			MarginBottom(1)

	styleCardAccent = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBrand).
			Padding(1, 2).
			MarginBottom(1)

	styleCardSuccess = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(colorSuccess).
			Padding(1, 2).
			MarginBottom(1)

	styleCardDanger = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorDanger).
			Padding(1, 2).
			MarginBottom(1)

	styleCenter = lipgloss.NewStyle().
			Align(lipgloss.Center).
			Width(80)

	styleHint = lipgloss.NewStyle().
			Foreground(colorMuted).
			MarginTop(1).
			Faint(true)

	styleErr = lipgloss.NewStyle().
			Foreground(colorDanger).
			MarginTop(1)

	styleSuccessText = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorSuccess)

	styleDangerText = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorDanger)

	styleCursor = lipgloss.NewStyle().
			Foreground(colorAccent).
			Bold(true)

	styleQuestion = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorText).
			PaddingLeft(1).
			MarginBottom(1)

	styleOption = lipgloss.NewStyle().
			Foreground(colorText)

	styleOptionSelected = lipgloss.NewStyle().
			Foreground(lipgloss.Color("0")).
			Background(colorAccent).
			Bold(true).
			Padding(0, 1)

	styleMarkOK = lipgloss.NewStyle().Bold(true).Foreground(colorSuccess)
	styleMarkBad = lipgloss.NewStyle().Bold(true).Foreground(colorDanger)
	styleExplain = lipgloss.NewStyle().Foreground(colorMuted).PaddingLeft(4)

	styleOnlineChip = lipgloss.NewStyle().
			Foreground(colorSuccess).
			Bold(true)

	styleOfflineChip = lipgloss.NewStyle().
			Foreground(colorDim)

	styleScoreBig = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorGold).
			Align(lipgloss.Center)

	// 聊天消息样式（renderChatMessage 里用）
	styleChatSelfR = lipgloss.NewStyle().Foreground(colorSelf).Bold(true)
	styleChatOtherR = lipgloss.NewStyle().Foreground(colorOther).Bold(true)
	styleChatDMR   = lipgloss.NewStyle().Foreground(colorDM).Bold(true).Italic(true)
	styleChatSystemR = lipgloss.NewStyle().Foreground(colorSystem)

	// 渐变边框颜色（模拟彩虹边框）
	rainbowColors = []lipgloss.Color{
		lipgloss.Color("213"), // 紫粉
		lipgloss.Color("212"),
		lipgloss.Color("211"),
		lipgloss.Color("177"),
		lipgloss.Color("183"),
		lipgloss.Color("51"),  // 青蓝
		lipgloss.Color("45"),
		lipgloss.Color("42"),  // 绿
	}
)

// geoyuanLogo 品牌 logo
const geoyuanLogo = `
  ██████╗ ███████╗ ██╗   ██╗ ██████╗ ██╗   ██╗ ██████╗ ██╗   ██╗
  ██╔══██╗██╔════╝ ██║   ██║██╔═══██╗██║   ██║██╔═══██╗╚██╗ ██╔╝
  ██████╔╝█████╗   ██║   ██║██║   ██║██║   ██║██║   ██║ ╚████╔╝ 
  ██╔══██╗██╔══╝   ╚██╗ ██╔╝██║   ██║██║   ██║██║   ██║  ╚██╔╝  
  ██║  ██║███████╗  ╚████╔╝ ╚██████╔╝╚██████╔╝╚██████╔╝   ██║   
  ╚═╝  ╚═╝╚══════╝   ╚═══╝   ╚═════╝  ╚═════╝  ╚═════╝    ╚═╝   
       ♥  AI Team · SSH Recruitment Challenge ♥`

// congratulationsArt 恭喜画面
const congratulationsArt = `
  ╔═══════════════════════════════════════════════╗
  ║                                               ║
  ║   🎉  CONGRATULATIONS  🎉                     ║
  ║                                               ║
  ║       You've been promoted to                 ║
  ║            GEORUAN MEMBER                     ║
  ║                                               ║
  ╚═══════════════════════════════════════════════╝
`

const thankYouArt = `
  ┌─────────────────────────────────────────┐
  │                                         │
  │         Thanks for trying!              │
  │                                         │
  │     ✦   Keep learning, keep hacking!   ✦
  │                                         │
  └─────────────────────────────────────────┘
`

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
func New(s *store.Store, gen *questiongen.Generator, fp string, user *store.User, chatHub *chat.Hub, ip, host string) Model {
	m := Model{
		store:     s,
		generator: gen,
		fp:        fp,
		user:      user,
		ChatHub:   chatHub,
		remoteIP:  ip,
		remoteHost: host,
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
//  📝 注册页
// ═══════════════════════════════════════════════
func (m Model) viewRegister() string {
	var b strings.Builder
	b.WriteString(m.headerBar("· 加入 geoyuan ·"))

	// Logo
	b.WriteString(geoyuanLogo + "\n\n")

	// 欢迎卡片
	welcomeText := styleCenter.Render(
		fmt.Sprintf("%s\n\n%s",
			styleHeader.Render("Welcome, traveler."),
			styleHint.Render("You've found the gateway to geoyuan — an exclusive team of AI engineers.\n"+
				"Pass the recruitment test to earn your place."),
		),
	)
	b.WriteString(welcomeText + "\n\n")

	// 连接信息
	if m.remoteIP != "" {
		b.WriteString(styleHint.Render(fmt.Sprintf("📍 Connecting from %s", m.remoteIP)))
		if m.remoteHost != "" && m.remoteHost != m.remoteIP {
			b.WriteString(styleHint.Render(fmt.Sprintf("  ·  %s", m.remoteHost)))
		}
		b.WriteString("\n\n")
	}

	// 注册卡片
	cardContent := "首次连接，请创建一个你的唯一用户名：\n\n" + m.nameInput.View()
	b.WriteString(styleCardAccent.Render(cardContent))

	if m.inputErr != "" {
		b.WriteString("\n" + styleErr.Render("⚠  "+m.inputErr))
	}

	b.WriteString("\n" + styleHint.Render("回车确认 · Ctrl+C 退出"))
	return b.String()
}

func (m Model) updateRegister(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.nameInput, cmd = m.nameInput.Update(msg)
	if msg.String() != "enter" {
		return m, cmd
	}
	name := strings.TrimSpace(m.nameInput.Value())
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

	b.WriteString(m.headerBar(fmt.Sprintf("· AI 招募测试  第 %d / %d ·", m.cursor+1, total)))

	// 进度条
	b.WriteString(renderFancyProgress(m.cursor+1, total))
	b.WriteString("\n\n")

	// 题目卡片
	card := fmt.Sprintf("%s\n\n", styleQuestion.Render(fmt.Sprintf("❓ %d. %s", m.cursor+1, q.Q)))

	for i, opt := range q.Options {
		var line string
		letter := string(rune('A' + i))
		marker := "  "
		if i == m.selected {
			line = styleOptionSelected.Render(fmt.Sprintf("%s %s. %s", marker, letter, opt))
		} else {
			line = styleOption.Render(fmt.Sprintf("%s %s. %s", marker, letter, opt))
		}
		card += line + "\n"
	}

	b.WriteString(styleCard.Render(card))
	b.WriteString("\n")
	b.WriteString(styleHint.Render("↑↓/jk 选择 · 1-4 直选并下一题 · 回车确认 · Ctrl+C 退出"))
	return b.String()
}

func (m *Model) startQuiz() {
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	var qs []questions.Question
	if m.generator != nil {
		var err error
		qs, err = m.generator.Generate(ctx, QuizCount)
		if err != nil {
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
//  🏆 结果页
// ═══════════════════════════════════════════════
func (m Model) viewResult() string {
	var b strings.Builder
	total := len(m.questions)
	pct := 0
	if total > 0 {
		pct = m.score * 100 / total
	}

	b.WriteString(m.headerBar("· 测试结果 ·"))

	if m.passed {
		// ── 恭喜画面 ──
		b.WriteString(styleCenter.Render(congratulationsArt))
		b.WriteString("\n")

		// 分数仪表盘
		b.WriteString(renderScoreDashboard(m.score, total, pct, m.passed))
		b.WriteString("\n\n")

		b.WriteString(styleSuccessText.Render(
			"  ✓ 你已被正式邀请加入 geoyuan team！\n"+
				"  ✓ 私密聊天室已为你解锁\n"+
				"  ✓ 所有成员都在等你",
		))
		b.WriteString("\n\n")
		b.WriteString(styleHint.Render("按 回车 进入私密聊天室  ·  q/esc 退出"))
	} else {
		// ── 未通过 ──
		b.WriteString(styleCenter.Render(thankYouArt))
		b.WriteString("\n")
		b.WriteString(renderScoreDashboard(m.score, total, pct, m.passed))
		b.WriteString("\n\n")

		b.WriteString(styleDangerText.Render(
			fmt.Sprintf("  ✗ 很遗憾未达通过线（需要 ≥ %d 分）", PassThreshold)))
		b.WriteString("\n\n")
		b.WriteString(styleHint.Render("断开后重新 ssh 连接即可再次测试  ·  回车/q 退出"))
	}

	// 答题回顾（前 8 题）
	b.WriteString("\n\n")
	review := "━━━ 答题回顾（前 8 题）━━━\n\n"
	limit := 8
	if len(m.questions) < limit {
		limit = len(m.questions)
	}
	for i := 0; i < limit; i++ {
		q := m.questions[i]
		ans := m.answers[i]
		ok := ans == q.Answer
		mark := styleMarkBad.Render("✗")
		if ok {
			mark = styleMarkOK.Render("✓")
		}
		review += fmt.Sprintf("%s %d. %s\n", mark, i+1, q.Q)
		if !ok {
			review += styleExplain.Render(fmt.Sprintf(
				"你的答案: %s   正确答案: %s   💡 %s",
				letter(ans), letter(q.Answer), q.Explain)) + "\n"
		}
		review += "\n"
	}
	if len(m.questions) > limit {
		review += styleHint.Render(fmt.Sprintf("... 其余 %d 题已省略", len(m.questions)-limit))
	}
	b.WriteString(review)

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
	b.WriteString(m.headerBar("· 欢迎回来 ·"))

	n, _ := m.store.CountApproved()

	welcomeText := fmt.Sprintf(`
  ╭──────────────────────────────────────────╮
  │  🌟  Welcome back, %-22s│
  │                                          │
  │  You are a certified geoyuan member.     │
  │                                          │
  │  📊 Team size: %d members                │
  │  🕐  Just now connected                  │
  │                                          │
  │  ── Entering private chat room ──       │
  ╰──────────────────────────────────────────╯
`, m.user.Username, n)

	// 渲染
	b.WriteString(styleCenter.Render(welcomeText))
	b.WriteString("\n\n")

	if m.remoteIP != "" {
		info := fmt.Sprintf("📍 %s", m.remoteIP)
		if m.remoteHost != "" {
			info += fmt.Sprintf("  ·  %s", m.remoteHost)
		}
		b.WriteString(styleHint.Render(info) + "\n")
	}

	return b.String()
}

// ═══════════════════════════════════════════════
//  🔁 重测选择
// ═══════════════════════════════════════════════
func (m Model) updateRetry(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "left", "up", "h", "k", "tab":
		m.retrySel = 0
	case "right", "down", "l", "j":
		m.retrySel = 1
	case "enter":
		if m.retrySel == 0 {
			m.startQuiz()
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

	b.WriteString("\n  是否现在重新测试？\n\n")

	labels := []string{
		"    🎯  重新挑战 AI 题目",
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
		case ":redeem", ":auth":
			// Web 通过后用 token 解锁
			if len(parts) < 2 {
				m.chatLog = append(m.chatLog, chat.Message{
					Type:    chat.MsgSystem,
					Content: "📝 用法: :redeem <token>  （从 diliy.cn Web 端测试通过后获取）",
					Time:    time.Now(),
				})
			} else if m.store != nil && m.user != nil {
				token := parts[1]
				wp, err := m.store.GetWebPassByToken(token)
				if err != nil || wp == nil {
					m.chatLog = append(m.chatLog, chat.Message{
						Type:    chat.MsgSystem,
						Content: "❌ token 无效或已过期。请在 diliy.cn 重新测试获取新 token。",
						Time:    time.Now(),
					})
				} else if !wp.Passed {
					m.chatLog = append(m.chatLog, chat.Message{
						Type:    chat.MsgSystem,
						Content: fmt.Sprintf("⚠ 该 token 对应的 Web 测试未通过（%d/%d）。", wp.Score, wp.Total),
						Time:    time.Now(),
					})
				} else {
					// 标记 used + 升级 SSH 用户为 approved
					_ = m.store.MarkWebPassUsed(token)
					if m.user.Status != store.StatusApproved {
						_, _ = m.store.DB().Exec("UPDATE users SET status = ? WHERE id = ?", store.StatusApproved, m.user.ID)
						m.user.Status = store.StatusApproved
					}
					m.chatLog = append(m.chatLog, chat.Message{
						Type:    chat.MsgSystem,
						Content: fmt.Sprintf("🎉 兑换成功！Web pass (得分 %d/%d, 用户 %s) 已绑定到你。\n   现在你是 geoyuan 正式成员了 ✨", wp.Score, wp.Total, wp.Username),
						Time:    time.Now(),
					})
					if m.ChatHub != nil {
						m.ChatHub.Broadcast(chat.Message{
							Type:    chat.MsgSystem,
							Content: fmt.Sprintf("⚡ %s 通过 Web 测试（%d/%d）正式加入 geoyuan！", m.user.Username, wp.Score, wp.Total),
							Time:    time.Now(),
						})
					}
				}
			}
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
