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
//  🌈 Geoyuan ASCII Logo — 3 种风格 + 彩虹渐变
// ═══════════════════════════════════════════════

// geoyuanLogoBig 主 logo — 方形大字 (figlet big 风格) 辨识度最高!
// 27 字符宽，6 行，方形 block 字母一眼认出
const geoyuanLogoBig = `
╔═╗╔═╗╔╗╔╗  ╔════╗  ╔════╗
║║║║║║║║║  ║╔═══╝  ║╔══╗║
║╚╝║║║║║║  ║║  ╔═╗║╚══╝║║
║╔═╗║║║║║║  ║║  ╚═╝║╔══╗║║
║║║║╚╝╚╝║║  ║╚═══╗  ║║║║║║
╚═╝╚═╝  ╚═╝  ╚════╝  ╚╝╚╝╚╝
`

// geoyuanLogoDohone 备选 — Dohone 风格下划线大字
const geoyuanLogoDohone = `
  ____           _                  _         
 / ___| ___   __| | ___   _ _ __   _| |_  ___ 
| |  _ / _ \ / _' |/ _ \ | '_ \ / _' | \ \/ / 
| |_| | (_) | (_| |  __/ | | | | (_| |  >  <  
 \____|\___/ \__,_|\___| |_| |_|\__,_| /_/\_\ 
`

// geoyuanLogoClassic 经典 block 字母 logo (6 行) — 保留用于小屏幕
const geoyuanLogoClassic = `
 ██████╗ ███████╗ ██╗   ██╗ ██████╗ ██╗   ██╗ ██████╗ ██╗   ██╗
 ██╔══██╗██╔════╝ ██║   ██║██╔═══██╗██║   ██║██╔═══██╗╚██╗ ██╔╝
 ██████╔╝█████╗   ██║   ██║██║   ██║██║   ██║██║   ██║ ╚████╔╝ 
 ██╔══██╗██╔══╝   ╚██╗ ██╔╝██║   ██║██║   ██║██║   ██║  ╚██╔╝  
 ██║  ██║███████╗  ╚████╔╝ ╚██████╔╝╚██████╔╝╚██████╔╝   ██║   
 ╚═╝  ╚═╝╚══════╝   ╚═══╝   ╚═════╝  ╚═════╝  ╚═════╝    ╚═╝   
`

// neonTagline 霓虹标语 (更长更有辨识度)
const neonTagline = `
   ◆ ◆ ◆    ♥    AI Team · SSH Recruitment Challenge    ♥    ◆ ◆ ◆
    ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓
`

// rainbowColors 彩虹色带（红→粉→紫→青→绿→金→红 循环）
var rainbowColors = []lipgloss.Color{
	lipgloss.Color("#ff3864"), // 红
	lipgloss.Color("#ff6ec7"), // 粉
	lipgloss.Color("#c77dff"), // 紫
	lipgloss.Color("#00f5ff"), // 青
	lipgloss.Color("#00ff9d"), // 绿
	lipgloss.Color("#ffd700"), // 金
	lipgloss.Color("#ff3864"), // 红 (循环)
	lipgloss.Color("#ff6ec7"), // 粉
}

// ═══════════════════════════════════════════════
//  🔤 核心：逐字符彩虹渐变渲染
// ═══════════════════════════════════════════════

// rainbowChar 给单个 rune 上色
func rainbowChar(ch rune, idx int) string {
	if ch == ' ' || ch == '\n' {
		return string(ch)
	}
	c := rainbowColors[idx%len(rainbowColors)]
	return lipgloss.NewStyle().
		Foreground(c).
		Bold(true).
		Render(string(ch))
}

// RainbowText 给任意字符串做彩虹渐变（公共函数）
func RainbowText(s string) string {
	var b strings.Builder
	runes := []rune(s)
	for i, r := range runes {
		b.WriteString(rainbowChar(r, i))
	}
	return b.String()
}

// RainbowBorder 渲染彩色边框顶/底
func RainbowBorder(width int, borderChar rune) string {
	var b strings.Builder
	for i := 0; i < width; i++ {
		c := rainbowColors[i%len(rainbowColors)]
		b.WriteString(lipgloss.NewStyle().Foreground(c).Render(string(borderChar)))
	}
	return b.String()
}

// RenderRainbowLogo 渲染大尺寸 Geoyuan logo + 彩虹渐变
func RenderRainbowLogo() string {
	// 检测终端宽度 — 选合适尺寸
	lines := strings.Split(strings.TrimLeft(geoyuanLogoBig, "\n"), "\n")
	var result strings.Builder

	for _, line := range lines {
		runes := []rune(line)
		for i, r := range runes {
			result.WriteString(rainbowChar(r, i))
		}
		result.WriteString("\n")
	}
	return result.String()
}

// RenderNeonTagline 渲染彩虹霓虹标语 + 分隔线
func RenderNeonTagline() string {
	lines := strings.Split(strings.TrimLeft(neonTagline, "\n"), "\n")
	var b strings.Builder

	// 标题行：◆ ◆ ◆ ♥ AI Team ...
	runes := []rune(lines[0])
	for i, r := range runes {
		b.WriteString(rainbowChar(r, i))
	}
	b.WriteString("\n")

	// 分隔线 ▓▓▓▓▓...
	if len(lines) > 1 {
		b.WriteString(RainbowBorder(len([]rune(lines[1])), '▓'))
		b.WriteString("\n")
	}

	return b.String()
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

// RenderWelcomeCard 注册页主欢迎卡 — 用 lipgloss Border 保证渲染正确
func RenderWelcomeCard() string {
	// Geoyuan 彩虹高亮的描述
	line1 := "You've found the gateway to " + RainbowText("Geoyuan") + " —"

	content := strings.Join([]string{
		"",
		"  🛰️  Welcome, Buddy",
		"",
		"  " + line1,
		"  an exclusive team of AI engineers and builders.",
		"",
		"  Pass the AI recruitment test to earn your place.",
		"",
		"  💡  30 questions  ·  6 min  ·  80% pass threshold",
	}, "\n")

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(cNeonPurple).
		Padding(0, 2).
		Width(66).
		Render(content)
}

// padTo 把字符串（可能含 ANSI 序列）填充到指定显示宽度
func padTo(s string, width int) string {
	// 粗略估计 ANSI 序列长度 — 去掉 \x1b[...m 和 lipgloss 输出
	ansiStripped := stripANSI(s)
	if len(ansiStripped) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(ansiStripped))
}

// stripANSI 去掉 ANSI 转义序列（简单版）
func stripANSI(s string) string {
	var b strings.Builder
	inEscape := false
	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
			continue
		}
		if inEscape {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEscape = false
			}
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
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
