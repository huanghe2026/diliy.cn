// Package tui 视觉组件 — 霓虹 Cyberpunk 风
// 集中管理所有 ASCII art logo、渐变渲染、视觉组件
package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ═══════════════════════════════════════════════
//  🎨 主题色板（Neon Cyberpunk）
// ═══════════════════════════════════════════════
// 全部用 truecolor hex，fallback 256 色
var (
	// Neon primary
	cNeonPurple = lipgloss.Color("#c77dff")  // 主品牌紫
	cNeonPink   = lipgloss.Color("#ff6ec7")  // 霓虹粉
	cNeonCyan   = lipgloss.Color("#00f5ff")  // 霓虹青
	cNeonGreen  = lipgloss.Color("#00ff9d")  // 霓虹绿
	cNeonGold   = lipgloss.Color("#ffd700")  // 霓虹金
	cNeonRed    = lipgloss.Color("#ff3864")  // 霓虹红

	// Neutrals (dark terminal friendly)
	cText    = lipgloss.Color("#e0e0e0")
	cMuted   = lipgloss.Color("#8b8b8b")
	cDim     = lipgloss.Color("#555555")
	cBorder  = lipgloss.Color("#4a4a6a")
	cSurface = lipgloss.Color("#1a1a2e")
)

// ═══════════════════════════════════════════════
//  🌈 Geoyuan ASCII Logo — 彩虹渐变渲染
// ═══════════════════════════════════════════════

// geoyuanLogoRaw 原始 ASCII art (6 行)
const geoyuanLogoRaw = `
 ██████╗ ███████╗ ██╗   ██╗ ██████╗ ██╗   ██╗ ██████╗ ██╗   ██╗
 ██╔══██╗██╔════╝ ██║   ██║██╔═══██╗██║   ██║██╔═══██╗╚██╗ ██╔╝
 ██████╔╝█████╗   ██║   ██║██║   ██║██║   ██║██║   ██║ ╚████╔╝ 
 ██╔══██╗██╔══╝   ╚██╗ ██╔╝██║   ██║██║   ██║██║   ██║  ╚██╔╝  
 ██║  ██║███████╗  ╚████╔╝ ╚██████╔╝╚██████╔╝╚██████╔╝   ██║   
 ╚═╝  ╚═╝╚══════╝   ╚═══╝   ╚═════╝  ╚═════╝  ╚═════╝    ╚═╝   
`

// neonTagline 霓虹标语 + 装饰
const neonTagline = `
     ◆ ◆ ◆   ♥  AI Team · SSH Recruitment Challenge  ♥   ◆ ◆ ◆
`

// rainbowColors 用于逐字符渐变渲染 Geoyuan logo
var rainbowColors = []lipgloss.Color{
	lipgloss.Color("#ff3864"), // 红
	lipgloss.Color("#ff6ec7"), // 粉
	lipgloss.Color("#c77dff"), // 紫
	lipgloss.Color("#00f5ff"), // 青
	lipgloss.Color("#00ff9d"), // 绿
	lipgloss.Color("#ffd700"), // 金
}

// RenderRainbowLogo 把 Geoyuan ASCII logo 按列做彩虹渐变
func RenderRainbowLogo() string {
	lines := strings.Split(strings.TrimLeft(geoyuanLogoRaw, "\n"), "\n")
	var result strings.Builder
	totalWidth := 65

	for _, line := range lines {
		runes := []rune(line)
		colorIdx := 0
		for i := 0; i < totalWidth; i++ {
			var ch rune = ' '
			if i < len(runes) {
				ch = runes[i]
			}
			// 仅对非空格字符上色
			if ch != ' ' && ch != '\n' {
				c := rainbowColors[colorIdx%len(rainbowColors)]
				result.WriteString(lipgloss.NewStyle().Foreground(c).Render(string(ch)))
			} else {
				result.WriteString(" ")
			}
			colorIdx++
		}
		result.WriteString("\n")
	}
	return result.String()
}

// RenderNeonTagline 渲染霓虹闪烁标语
func RenderNeonTagline() string {
	s := neonTagline
	return lipgloss.NewStyle().
		Bold(true).
		Align(lipgloss.Center).
		Width(80).
		Render(s)
}

// RenderGlowBox 用渐变边框渲染一个发光盒子
// content: 内部内容, width: 盒宽, colors: 边框颜色数组
func RenderGlowBox(content string, width int, colors []lipgloss.Color) string {
	if width < 20 {
		width = 40
	}
	border := lipgloss.ThickBorder()
	top := strings.Repeat(border.Top, width-2)
	bottom := strings.Repeat(border.Bottom, width-2)

	var b strings.Builder
	// 顶边框（渐变）
	for i, r := range []rune("╔" + top + "╗") {
		c := colors[i%len(colors)]
		b.WriteString(lipgloss.NewStyle().Foreground(c).Render(string(r)))
	}
	b.WriteString("\n")

	// 内容行
	contentLines := strings.Split(content, "\n")
	for _, line := range contentLines {
		padded := line
		if len(line) < width-2 {
			padded = line + strings.Repeat(" ", width-2-len(line))
		}
		// 左边框
		b.WriteString(lipgloss.NewStyle().Foreground(colors[0]).Render(string(border.Left)))
		b.WriteString(padded)
		b.WriteString(lipgloss.NewStyle().Foreground(colors[len(colors)-1]).Render(string(border.Right)))
		b.WriteString("\n")
	}

	// 底边框（渐变）
	for i, r := range []rune("╚" + bottom + "╝") {
		c := colors[i%len(colors)]
		b.WriteString(lipgloss.NewStyle().Foreground(c).Render(string(r)))
	}
	b.WriteString("\n")
	return b.String()
}

// RenderWelcomeCard 注册页主欢迎卡
func RenderWelcomeCard() string {
	w := `
  ╭──────────────────────────────────────────────────────────╮
  │                                                          │
  │   🛰️  Welcome, Buddy                                    │
  │                                                          │
  │   You've found the gateway to Geoyuan —                  │
  │   an exclusive team of AI engineers and builders.        │
  │                                                          │
  │   Pass the AI recruitment test to earn your place.       │
  │                                                          │
  │   💡  30 questions  ·  6 min  ·  80% pass threshold       │
  │                                                          │
  ╰──────────────────────────────────────────────────────────╯
`
	return w
}

// RenderCongratsPass 通过恭喜画面
func RenderCongratsPass(score, total int) string {
	arts := []string{
		`
  ╔═══════════════════════════════════════════════════════════╗
  ║                                                           ║
  ║   ✨   🎉   🌟   CONGRATULATIONS  🌟   🎉   ✨             ║
  ║                                                           ║
  ║         You've been promoted to                            ║
  ║                                                           ║
  ║          ╔══════════════════╗                             ║
  ║          ║  Geoyuan MEMBER  ║                             ║
  ║          ╚══════════════════╝                             ║
  ║                                                           ║
  ║     Score: %d / %d   ✓   ✓   ✓                             ║
  ║                                                           ║
  ╚═══════════════════════════════════════════════════════════╝
`,
	}
	return fmt.Sprintf(arts[0], score, total)
}

// RenderTryAgainFail 未通过画面
func RenderTryAgainFail(score, total int) string {
	return fmt.Sprintf(`
  ┌─────────────────────────────────────────────────────────────┐
  │                                                             │
  │          ✦   Not this time!   ✦                             │
  │                                                             │
  │     Score: %d / %d                                          │
  │                                                             │
  │     ✧   Keep learning, keep hacking!   ✧                    │
  │                                                             │
  └─────────────────────────────────────────────────────────────┘
`, score, total)
}

// ═══════════════════════════════════════════════
//  🔧 Legacy Compat — app.go 引用的旧变量名
//  全部映射到新霓虹色
// ═══════════════════════════════════════════════

var (
	// 颜色别名 (backward compat)
	colorBrand      = cNeonPurple
	colorBrandDark  = lipgloss.Color("#9d4edd")
	colorAccent     = cNeonCyan
	colorGold       = cNeonGold
	colorSuccess    = cNeonGreen
	colorDanger     = cNeonRed
	colorWarn       = cNeonGold
	colorText       = cText
	colorMuted      = cMuted
	colorDim        = cDim
	colorBorder     = cBorder
	colorBorderSub  = lipgloss.Color("#3a3a5a")

	colorSelf   = cNeonCyan
	colorOther  = cNeonPurple
	colorDM     = cNeonPink
	colorSystem = cDim

	// 样式别名 (backward compat)
	styleLogo = lipgloss.NewStyle().Bold(true).Foreground(cNeonPurple).Padding(0, 1)

	styleHeader = lipgloss.NewStyle().Bold(true).Foreground(cNeonPurple).MarginBottom(1)

	styleHeaderBar = lipgloss.NewStyle().Foreground(cNeonCyan).Bold(true)

	styleCard = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(cBorder).
			Padding(1, 2).
			MarginBottom(1)

	styleCardAccent = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(cNeonPurple).
			Padding(1, 2).
			MarginBottom(1)

	styleCardSuccess = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(cNeonGreen).
			Padding(1, 2).
			MarginBottom(1)

	styleCardDanger = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(cNeonRed).
			Padding(1, 2).
			MarginBottom(1)

	styleCenter = lipgloss.NewStyle().
			Align(lipgloss.Center).
			Width(80)

	styleHint = lipgloss.NewStyle().
			Foreground(cMuted).
			MarginTop(1).
			Faint(true)

	styleErr = lipgloss.NewStyle().
			Foreground(cNeonRed).
			Bold(true).
			MarginTop(1)

	styleSuccessText = lipgloss.NewStyle().
			Bold(true).
			Foreground(cNeonGreen)

	styleDangerText = lipgloss.NewStyle().
			Bold(true).
			Foreground(cNeonRed)

	styleCursor = lipgloss.NewStyle().
			Foreground(cNeonCyan).
			Bold(true)

	styleQuestion = lipgloss.NewStyle().
			Bold(true).
			Foreground(cNeonCyan).
			PaddingLeft(1).
			MarginBottom(1)

	styleOption = lipgloss.NewStyle().
			Foreground(cText)

	styleOptionSelected = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000000")).
			Background(cNeonCyan).
			Bold(true).
			Padding(0, 1)

	styleMarkOK  = lipgloss.NewStyle().Bold(true).Foreground(cNeonGreen)
	styleMarkBad = lipgloss.NewStyle().Bold(true).Foreground(cNeonRed)
	styleExplain = lipgloss.NewStyle().Foreground(cMuted).PaddingLeft(4)

	styleOnlineChip  = lipgloss.NewStyle().Foreground(cNeonGreen).Bold(true)
	styleOfflineChip = lipgloss.NewStyle().Foreground(cDim)

	styleScoreBig = lipgloss.NewStyle().
			Bold(true).
			Foreground(cNeonGold).
			Align(lipgloss.Center)

	styleChatSelfR   = lipgloss.NewStyle().Foreground(cNeonCyan).Bold(true)
	styleChatOtherR  = lipgloss.NewStyle().Foreground(cNeonPurple).Bold(true)
	styleChatDMR     = lipgloss.NewStyle().Foreground(cNeonPink).Bold(true).Italic(true)
	styleChatSystemR = lipgloss.NewStyle().Foreground(cDim)
)

// RenderUserCard 已通过用户信息详情卡
func RenderUserCard(username, fp, ip string, approved bool, teamSize, totalAttempts int) string {
	status := "PENDING"
	statusColor := lipgloss.NewStyle().Foreground(cNeonRed).Bold(true)
	if approved {
		status = "APPROVED ✓"
		statusColor = lipgloss.NewStyle().Foreground(cNeonGreen).Bold(true)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("  👤  %s\n", username))
	sb.WriteString(fmt.Sprintf("     %s\n", statusColor.Render(status)))
	sb.WriteString(fmt.Sprintf("  🖥️  %s\n", ip))
	sb.WriteString(fmt.Sprintf("  🧬  %s...%s\n", fp[:8], fp[len(fp)-8:]))
	sb.WriteString(fmt.Sprintf("  🤝  Team: %d members\n", teamSize))
	sb.WriteString(fmt.Sprintf("  📊  Quiz attempts: %d\n", totalAttempts))
	return sb.String()
}
