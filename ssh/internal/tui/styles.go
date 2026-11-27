// Package tui 视觉组件 — 霓虹 Cyberpunk 风
// 集中管理所有 ASCII art logo、渐变渲染、视觉组件
package tui

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
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

// displayWidth uses the same ANSI/grapheme-aware measurement as Lip Gloss.
func displayWidth(s string) int {
	return ansi.StringWidth(s)
}

func trimLogoLine(line string) string {
	visible := strings.TrimRight(ansi.Strip(line), " ")
	if visible == "" {
		return ""
	}
	return ansi.Cut(line, 0, displayWidth(visible))
}

func stripLogoTrailingSpace(logo string) string {
	var b strings.Builder
	for _, rawLine := range strings.Split(logo, "\n") {
		line := trimLogoLine(rawLine)
		if line != "" {
			b.WriteString(line)
			b.WriteString("\n")
		}
	}
	return b.String()
}

const logoWidthMargin = 2

var cyberpunkGradient = [][3]int{
	{37, 99, 235},
	{67, 56, 202},
	{124, 58, 237},
	{168, 85, 247},
	{236, 72, 153},
	{8, 145, 178},
	{34, 211, 238},
	{56, 189, 248},
}

func renderBrailleGradient(art string) string {
	width := maxLogoWidth(art)
	lines := strings.Split(strings.TrimRight(art, "\n"), "\n")
	var b strings.Builder
	for y, line := range lines {
		b.WriteString("\x1b[1m")
		for x, r := range []rune(line) {
			if r == '\u2800' {
				b.WriteRune(r)
				continue
			}
			colorIndex := 0
			if width > 1 {
				colorIndex = x * (len(cyberpunkGradient) - 1) / (width - 1)
				colorIndex = (colorIndex + y) % len(cyberpunkGradient)
			}
			color := cyberpunkGradient[colorIndex]
			fmt.Fprintf(&b, "\x1b[38;2;%d;%d;%dm%c", color[0], color[1], color[2], r)
		}
		b.WriteString("\x1b[0m\n")
	}
	return b.String()
}

// ── 赛博霓虹 glow logo（全块 █ + 紫粉蓝渐变 + 深紫光晕底）──

// neonGradient 赛博霓虹色系（紫→粉→蓝→青循环）
var neonGradient = [][3]int{
	{147, 51, 234}, // 紫
	{168, 85, 247}, // 亮紫
	{217, 70, 239}, // 洋红
	{236, 72, 153}, // 粉
	{99, 102, 241}, // 靛蓝
	{59, 130, 246}, // 蓝
	{34, 211, 238}, // 青
	{147, 51, 234}, // 紫（循环）
}

// geoyuanLogoNeon GEOYUAN small 字体字符画（线条字符 _ / \ |，霓虹灯管效果）
const geoyuanLogoNeon = `
  ___ ___ _____   ___   _  _   _  _
 / __| __/ _ \ \ / / | | |/_\ | \| |
| (_ | _| (_) \ V /| |_| / _ \| .' |
 \___|___\___/ |_|  \___/_/ \_\_|\_|
`

// renderNeonGlowLogo 渲染赛博霓虹 glow logo：全块 █ 无横纹 + 紫粉蓝渐变 + 深紫光晕底。
func renderNeonGlowLogo(width int) string {
	if width <= 0 {
		return ""
	}
	lines := strings.Split(strings.TrimRight(geoyuanLogoNeon, "\n"), "\n")
	maxW := maxLogoWidth(geoyuanLogoNeon)
	if maxW+2 > width {
		if maxLogoWidth(geoyuanLogoFallback)+logoWidthMargin <= width {
			return centerLogo(geoyuanLogoFallback, width)
		}
		return renderCompactLogo(width)
	}

	leftPad := (width - maxW) / 2
	bgOn := "\x1b[48;2;16;8;44m" // 深紫光晕底
	bgOff := "\x1b[49m"

	var b strings.Builder
	for _, line := range lines {
		b.WriteString(bgOn)
		b.WriteString(strings.Repeat(" ", leftPad))
		for x, r := range []rune(line) {
			if r == ' ' {
				b.WriteRune(' ')
				continue
			}
			ci := x * (len(neonGradient) - 1) / (maxW - 1)
			c := neonGradient[ci]
			fmt.Fprintf(&b, "\x1b[1;38;2;%d;%d;%dm\x1b[48;2;%d;%d;%dm%c",
				c[0], c[1], c[2], c[0]/4, c[1]/4, c[2]/4, r)
		}
		b.WriteString(strings.Repeat(" ", width-leftPad-maxW))
		b.WriteString(bgOff)
		b.WriteString("\n")
	}
	return b.String()
}

// renderSlogan 渲染品牌 slogan「Card your life」居中霓虹
func renderSlogan(width int) string {
	if width <= 0 {
		return ""
	}
	text := "◆ Card your life ◆"
	if displayWidth(text) > width {
		text = "Card your life"
		if displayWidth(text) > width {
			return ""
		}
	}
	leftPad := (width - displayWidth(text)) / 2
	style := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#c084fc"))
	return strings.Repeat(" ", leftPad) + style.Render(text) + "\x1b[0m\n"
}

// renderHeroSlogan 放大版 slogan（欢迎页/主页）：霓虹渐变 tagline + 中文定位
func renderHeroSlogan(width int) string {
	if width <= 0 {
		return ""
	}
	tagline := "Card your life"
	subtitle := "把地理位置变成可拥有的永久数字资产"

	var b strings.Builder
	b.WriteString("  ")
	runes := []rune(tagline)
	for i, r := range runes {
		if r == ' ' {
			b.WriteRune(' ')
			continue
		}
		ci := i * (len(neonGradient) - 1) / (len(runes) - 1)
		c := neonGradient[ci]
		fmt.Fprintf(&b, "\x1b[1;38;2;%d;%d;%dm%c", c[0], c[1], c[2], r)
	}
	b.WriteString("\x1b[0m\n")
	b.WriteString("  " + lipgloss.NewStyle().Foreground(cMuted).Render(subtitle) + "\n")
	return b.String()
}

// RenderLogo picks the richest complete asset that fits the terminal width.
func RenderLogo(width int) string {
	if width <= 0 {
		return ""
	}

	if maxLogoWidth(geoyuanLogoBraille)+logoWidthMargin <= width {
		return centerLogo(renderBrailleGradient(geoyuanLogoBraille), width)
	}
	if maxLogoWidth(geoyuanLogoFallback)+logoWidthMargin <= width {
		return centerLogo(geoyuanLogoFallback, width)
	}
	return renderCompactLogo(width)
}

func maxLogoWidth(logo string) int {
	width := 0
	for _, line := range strings.Split(logo, "\n") {
		if w := displayWidth(line); w > width {
			width = w
		}
	}
	return width
}

func centerLogo(logo string, termWidth int) string {
	if termWidth <= 0 {
		return ""
	}

	var b strings.Builder
	for _, raw := range strings.Split(logo, "\n") {
		line := trimLogoLine(raw)
		if line == "" {
			continue
		}
		lineWidth := displayWidth(line)
		if lineWidth > termWidth {
			line = ansi.Truncate(line, termWidth, "")
			lineWidth = displayWidth(line)
		}
		if lineWidth == 0 {
			continue
		}
		padding := (termWidth - lineWidth) / 2
		b.WriteString(strings.Repeat(" ", padding))
		b.WriteString(line)
		b.WriteString("\x1b[0m\n")
	}
	return b.String()
}

func renderCompactLogo(width int) string {
	text := "GEOYUAN"
	if width < displayWidth(text) {
		text = "GY"
		if width < displayWidth(text) {
			text = "G"
		}
	}
	return centerLogo(RainbowText(text), width)
}

// ═══════════════════════════════════════════════
//
//	🎨 主题色板（Earth-Tone · Geoyuan 地理主题）
//
// ═══════════════════════════════════════════════
// 地形绿 → 地质橙 → 矿物金 → 森林翠 → 河流青 → 海洋蓝 → 深海蓝
// 全部用 truecolor hex，fallback 256 色
var (
	// Earth primary
	cNeonPurple = lipgloss.Color("#3b82f6") // 主品牌蓝（海洋）
	cNeonPink   = lipgloss.Color("#1d4ed8") // 深海蓝（原霓虹粉）
	cNeonCyan   = lipgloss.Color("#06b6d4") // 河流青
	cNeonGreen  = lipgloss.Color("#22c55e") // 地形绿
	cNeonGold   = lipgloss.Color("#f59e0b") // 矿物金
	cNeonRed    = lipgloss.Color("#ef4444") // 警示红

	// Neutrals (dark terminal friendly)
	cText    = lipgloss.Color("#e2e8f0") // slate-200
	cMuted   = lipgloss.Color("#94a3b8") // slate-400
	cDim     = lipgloss.Color("#64748b") // slate-500
	cBorder  = lipgloss.Color("#475569") // slate-600
	cSurface = lipgloss.Color("#1e293b") // slate-800
)

// ═══════════════════════════════════════════════
//  🌈 Geoyuan ASCII Logo — 3 种风格 + 彩虹渐变
// ═══════════════════════════════════════════════

// geoyuanLogoBig 主 logo — 7 行高 × 56 列宽
// 100% 纯 ASCII（# + 空格）— 任何字体、任何终端下绝对显示一致
// （Braille / Halfblock 在某些字体下宽度渲染不一致 → 已弃用）
const geoyuanLogoBig = `
######  ######  ######  ##   ## ##   ## ######  ##   ##
 ##      ##   ## ##   ## ##   ## ##   ## ##   ## ##   ##
 ## #### ######  ######  ##   ## ##   ## ######  ## # ##
 ##  ##  ##   ## ##   ## ## # ## ##   ## ##   ## ### ###
 ######  ######  ######   ### ### ###### ##   ## ##   ##
     ##  ##                   ##      ##       ##   ##
  ###### ######  ######      ##      ##    ###### ######
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

// rainbowColors 大地色带（绿→橙→金→翠→青→蓝→深海 循环）
var rainbowColors = []lipgloss.Color{
	lipgloss.Color("#22c55e"), // 地形绿
	lipgloss.Color("#f97316"), // 地质橙
	lipgloss.Color("#f59e0b"), // 矿物金
	lipgloss.Color("#10b981"), // 森林翠
	lipgloss.Color("#06b6d4"), // 河流青
	lipgloss.Color("#3b82f6"), // 海洋蓝
	lipgloss.Color("#1d4ed8"), // 深海蓝
	lipgloss.Color("#22c55e"), // 地形绿(循环)
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

// RenderGlowLogo 渲染嵌入式 GEOYUAN ANSI 像素字。
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
	if padding := width - displayWidth(s); padding > 0 {
		return s + strings.Repeat(" ", padding)
	}
	return s
}

// adaptiveCardWidth 响应式卡片宽度 — 按 termWidth 取 min(72, termWidth-10)
// -4 滚动条安全余量 + -2 左右各 2 列 padding
// 兜底 40 保证窄终端不崩
func adaptiveCardWidth(termWidth int) int {
	w := termWidth - 10
	if w > 72 {
		return 72
	}
	if w < 40 {
		return 40
	}
	return w
}

// stripANSI 去掉 ANSI 转义序列。
func stripANSI(s string) string {
	return ansi.Strip(s)
}

// RenderCongratsPass 🏆 Geoyuan 专属通过画面 — 霓虹赛博风
func RenderCongratsPass(score, total int) string {
	card := fmt.Sprintf(`
  ╔═══════════════════════════════════════════════════════════════════╗
  ║  ✨   🎉   🌟   🏅  MISSION COMPLETE  🏅   🌟   🎉   ✨             ║
  ║                                                                   ║
  ║    🌍  Geoyuan 准入认证通过                                        ║
  ║    ✅  30/30  AI 题库 · 区块链存证                                ║
  ║                                                                   ║
  ║    ┌─────────────────────────────────────────────────────────┐    ║
  ║    │  🎴  Mint 专属 Geo-Card  ·  绑定经纬度 ·  Polygon 链上   │    ║
  ║    │  💬  私密聊天室解锁      ·  全网 geo-peer 匹配          │    ║
  ║    │  💰  Marketplace 挂售     ·  链上 P2P 交易              │    ║
  ║    └─────────────────────────────────────────────────────────┘    ║
  ║                                                                   ║
  ║    Score: %d / %d   (%d%%)   ⚡   Pass ≥ %d                       ║
  ║                                                                   ║
  ║    ┌─ 下一步 ─────────────────────────────────────────────────┐   ║
  ║    │  ⏎  进入主菜单  ·  1 → Mint  ·  2 → Chat  ·  3 → Market │   ║
  ║    └──────────────────────────────────────────────────────────┘   ║
  ╚═══════════════════════════════════════════════════════════════════╝
`, score, total, score*100/total, PassThreshold)

	// 霓虹色渲染
	cyan := lipgloss.NewStyle().Foreground(cNeonCyan).Bold(true)
	gold := lipgloss.NewStyle().Foreground(cNeonGold).Bold(true)
	purple := lipgloss.NewStyle().Foreground(cNeonPurple).Bold(true)
	green := lipgloss.NewStyle().Foreground(cNeonGreen).Bold(true)
	pink := lipgloss.NewStyle().Foreground(cNeonPink).Bold(true)

	// 关键词着色
	card = strings.ReplaceAll(card, "MISSION COMPLETE", pink.Render("MISSION COMPLETE"))
	card = strings.ReplaceAll(card, "Geoyuan", purple.Render("Geoyuan"))
	card = strings.ReplaceAll(card, "Score:", green.Render("Score:"))
	card = strings.ReplaceAll(card, "下一步", cyan.Render("下一步"))
	card = strings.ReplaceAll(card, "🎴 Mint", gold.Render("🎴 Mint"))
	card = strings.ReplaceAll(card, "💬", cyan.Render("💬"))
	card = strings.ReplaceAll(card, "Marketplace", gold.Render("Marketplace"))
	card = strings.ReplaceAll(card, "✅", green.Render("✅"))
	card = strings.ReplaceAll(card, "⏎  进入主菜单", purple.Render("⏎  进入主菜单"))

	return card
}

// RenderTryAgainFail 未通过画面 — 霓虹鼓励版（没有"不通过"，只有"还没"）
func RenderTryAgainFail(score, total int) string {
	card := fmt.Sprintf(`
  ┌───────────────────────────────────────────────────────────────────┐
  │                                                                   │
  │     🌊   NOT YET, BUDDY                                           │
  │                                                                   │
  │     Score: %d / %d   (%d%%)   ⚡   Need ≥ %d / %d                 │
  │                                                                   │
  │     ┌─────────────────────────────────────────────────────────┐   │
  │     │  🔁  重连 ssh diliy.cn  再挑战一次                       │   │
  │     │  🧠  AI 每次动态生成 30 新题                             │   │
  │     │  💡  本地题库无限次练习 · 答案有解析                      │   │
  │     └─────────────────────────────────────────────────────────┘   │
  │                                                                   │
  │     目标线: Pass ≥ %d / %d (%d%%)                                  │
  │                                                                   │
  └───────────────────────────────────────────────────────────────────┘
`, score, total, score*100/total, PassThreshold, total, PassThreshold, total, PassThreshold*100/total)

	red := lipgloss.NewStyle().Foreground(cNeonRed).Bold(true)
	card = strings.ReplaceAll(card, "NOT YET, BUDDY", red.Render("NOT YET, BUDDY"))

	return card
}

// ═══════════════════════════════════════════════
//  🔧 Legacy Compat — app.go 引用的旧变量名
//  全部映射到新霓虹色
// ═══════════════════════════════════════════════

var (
	// 颜色别名 (backward compat — 大地色系)
	colorBrand     = cNeonPurple               // 主品牌海洋蓝
	colorBrandDark = lipgloss.Color("#1e40af") // 深蓝
	colorAccent    = cNeonCyan                 // 河流青
	colorGold      = cNeonGold                 // 矿物金
	colorSuccess   = cNeonGreen                // 地形绿
	colorDanger    = cNeonRed                  // 警示红
	colorWarn      = cNeonGold
	colorText      = cText
	colorMuted     = cMuted
	colorDim       = cDim
	colorBorder    = cBorder
	colorBorderSub = lipgloss.Color("#334155") // slate-700

	colorSelf   = cNeonCyan  // 河流青（自己的消息）
	colorOther  = cNeonGreen // 地形绿（别人的消息）
	colorDM     = cNeonGold  // 矿物金（私信）
	colorSystem = cDim

	// 样式别名 (backward compat) — 大地蓝主题
	styleLogo = lipgloss.NewStyle().Bold(true).Foreground(cNeonPurple).Padding(0, 1)

	styleHeader = lipgloss.NewStyle().Bold(true).Foreground(cNeonPurple).MarginBottom(1)

	styleHeaderBar = lipgloss.NewStyle().Foreground(cNeonGreen).Bold(true) // 地形绿

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
