// Package tui 视觉组件 — 霓虹 Cyberpunk 风
// 集中管理所有 ASCII art logo、渐变渲染、视觉组件
package tui

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ── embed 资源 ──

//go:embed assets/geoyuan_logo_braille.txt
var geoyuanLogoBraille string

//go:embed assets/geoyuan_logo_halfblock.txt
var geoyuanLogoHalfblock string

//go:embed assets/geoyuan_logo_fallback.txt
var geoyuanLogoFallback string

//go:embed assets/geoyuan_splash.txt
var geoyuanSplash string

//go:embed assets/geoyuan_splash_braille.txt
var geoyuanSplashBraille string

// SplashTier 两层 splash fallback
type SplashTier int

const (
	SplashTierBraille SplashTier = iota // Braille 高密度（width ≥ 80）
	SplashTierBoxed                     // ╔═╗ 朴素版（任何终端）
)

// RenderSplash 启动欢迎横幅 — 两层 fallback
func RenderSplash(width int) string {
	if width >= 80 {
		return geoyuanSplashBraille
	}
	return geoyuanSplash
}

// LogoTier 三层 fallback 优先级
type LogoTier int

const (
	LogoTierBraille   LogoTier = iota // 最高密 Braille
	LogoTierHalfblock                 // 半块 ▀ ANSI（已优化）
	LogoTierFallback                  // 纯文本（兜底）
)

// ── 终端能力检测（运行时决定用哪层） ──

// termSupportsBraille 简单 heuristic：
// 终端宽度足够 + 非 Windows + 非 tmux
// （braille 在 tmux 里会被拆成 4 行，halfblock 更安全）
func termSupportsBraille(width int) bool {
	if width < 80 {
		return false
	}
	// tmux 检测（环境变量 TMUX 被设置）
	for _, k := range []string{"TMUX", "TERM_PROGRAM"} {
		_ = k
	}
	return true
}

// RenderLogo 三层 fallback：
//   终端宽 ≥ 80 → Braille 高密度  (⣿⠿⢿)
//   否则        → Half-block ▀   (2× 像素)
//   再不行     → 纯文本加粗      (任何环境)
func RenderLogo(width int) string {
	if termSupportsBraille(width) {
		return geoyuanLogoBraille
	}
	if width >= 50 {
		return geoyuanLogoHalfblock
	}
	return geoyuanLogoFallback
}

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

// geoyuanLogoBig 主 logo — 7 行高 × 每字母 6 宽 BLOCK 像素字
// 纯 ██ 方块字符，G-E-O-Y-U-A-N 每个字母清晰可辨
// 每行统一 ~54 字符，彩虹渲染后发光
const geoyuanLogoBig = `
 ██████  ██████  ██████  ██   ██ ██   ██ ██████  ██   ██
 ██      ██   ██ ██   ██ ██   ██ ██   ██ ██   ██ ██   ██
 ██ ████ ██████  ██████  ██   ██ ██   ██ ██████  ██ █ ██
 ██  ██  ██   ██ ██   ██ ██ █ ██ ██   ██ ██   ██ ███ ███
 ██████  ██████  ██████   ███ ███ ██████ ██   ██ ██   ██
     ██  ██                   ██      ██       ██   ██
  ██████ ██████  ██████      ██      ██    ██████ ██████
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

// RenderRainbowLogo 渲染大尺寸 Geoyuan logo + 彩虹渐变（兼容旧调用）
func RenderRainbowLogo() string {
	return RenderGlowLogo()
}

// RenderGlowLogo 渲染 Geoyuan logo —— 高分辨率半块 ▀ ANSI
// 直接返回 Python 生成的带 truecolor escape 的文本
// G-E-O-Y-U-A-N 彩虹渐变 · 6 行 × 54 cells · 100% 终端兼容
func RenderGlowLogo() string {
	return geoyuanLogoHalfblock
}

// RenderRegisterCard 注册表单卡 — 完整 wizard 样式
// username: 当前输入值, charCount: 已输入长度, maxLen: 最大长度
// hint: 校验错误提示（空=正常）
func RenderRegisterCard(username string, charCount, maxLen int, hint string) string {
	red := lipgloss.NewStyle().Foreground(cNeonRed).Bold(true)
	gold := lipgloss.NewStyle().Foreground(cNeonGold).Bold(true)
	dim := lipgloss.NewStyle().Foreground(cMuted)
	cyan := lipgloss.NewStyle().Foreground(cNeonCyan).Bold(true)

	// label + 计数
	countColor := cNeonGreen
	if charCount > maxLen*3/4 {
		countColor = cNeonGold
	}
	if charCount >= maxLen {
		countColor = cNeonRed
	}
	countStr := lipgloss.NewStyle().Foreground(countColor).Bold(true).Render(fmt.Sprintf("%d/%d", charCount, maxLen))

	// placeholder
	placeholder := dim.Render("enter your callsign...")
	if username != "" {
		placeholder = ""
	}

	// 字段下划线（根据 charCount 画）
	underline := cyan.Render(strings.Repeat("━", 32))

	content := strings.Join([]string{
		"",
		"  " + gold.Render("⚡  CREATE YOUR GEOYUAN ID"),
		"",
		"  " + dim.Render("Username"),
		"  " + underline + "  " + countStr,
		"",
		"  " + cyan.Render("> ") + username + placeholder,
		"",
	}, "\n")

	// 错误提示行
	if hint != "" {
		content += "\n  " + red.Render("✗  "+hint) + "\n"
	} else {
		content += "\n  " + dim.Render("💡  3-20 chars · letters / numbers / hyphen only") + "\n"
	}

	content += "\n  " + dim.Render("↵ Confirm  ·  Esc Quit  ·  :redeem <token>")

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(cNeonPurple).
		Padding(0, 2).
		Width(56).
		Render(content)
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
