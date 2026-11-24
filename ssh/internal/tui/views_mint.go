package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"geoyuan.cn/ssh/internal/card"
	"geoyuan.cn/ssh/internal/geo"
	"geoyuan.cn/ssh/internal/h3"
)

// PhaseMint — Bubble Tea Mint 视图
type PhaseMint struct {
	cardService *card.Service
	step        int // 0=H3预览 1=标题 2=描述 3=颜色 4=确认 5=处理中/完成 99=退出
	titleIn     textinput.Model
	descIn      textinput.Model
	colorIn     textinput.Model
	h3Index     uint64
	lat, lon    float64
	res         int
	title, desc string
	color       string
	err         string
	minting     bool             // true = 异步链上 Mint 已发送，等结果
	mintResult  *card.MintResult // 链上返回
	mintErr     error            // 链上错误
}

// NewPhaseMint — 初始化 Mint 视图（自动用 IP 做 H3 预览）
func NewPhaseMint(svc *card.Service, remoteIP, geoLoc string) PhaseMint {
	pm := PhaseMint{cardService: svc, res: h3.DefaultRes}

	// 从 IP 拿经纬度，拿不到降级杭州
	lat, lon := geo.LatLonFromIP(remoteIP)
	if lat == 0 && lon == 0 {
		lat, lon = 30.2741, 120.1551 // 杭州西湖
	}
	c := h3.Encode(lat, lon, h3.DefaultRes)
	pm.h3Index = c.Index
	pm.lat = c.Lat
	pm.lon = c.Lon
	pm.res = c.Res

	pm.titleIn = textinput.New()
	pm.titleIn.Placeholder = "Morning Lake Walk"
	pm.titleIn.CharLimit = 30
	pm.titleIn.Focus()
	pm.titleIn.Width = 30

	pm.descIn = textinput.New()
	pm.descIn.Placeholder = "My first Geoyuan card"
	pm.descIn.Width = 40

	pm.colorIn = textinput.New()
	pm.colorIn.SetValue("#c77dff")
	pm.colorIn.Width = 10

	return pm
}

// WithResult — Model 收到 mintResultMsg 后更新 PhaseMint
func (p PhaseMint) WithResult(r *card.MintResult, err error) PhaseMint {
	p.minting = false
	p.mintResult = r
	p.mintErr = err
	if err == nil {
		p.step = 5
	} else {
		p.step = 4 // 失败回到确认页让用户重试
	}
	return p
}

// H3Hex — uint64 → "0x..." hex
func H3Hex(idx uint64) string { return fmt.Sprintf("0x%x", idx) }

// Done — 退出 Mint 视图时返回 true
func (p PhaseMint) Done() bool { return p.step == 99 }

// ─── Update ──────────────────────────────────────────────

func (p PhaseMint) Update(msg tea.Msg) (PhaseMint, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch p.step {
		case 0: // H3 预览
			switch msg.String() {
			case "enter", " ":
				p.step = 1
				return p, nil
			case "esc", "q":
				p.step = 99
				return p, nil
			}
		case 1: // 标题
			var cmd tea.Cmd
			p.titleIn, cmd = p.titleIn.Update(msg)
			switch msg.String() {
			case "enter":
				if len([]rune(strings.TrimSpace(p.titleIn.Value()))) < 2 {
					p.err = "Title must be 2-30 characters"
					return p, nil
				}
				p.err = ""
				p.step = 2
				p.titleIn.Blur()
				p.descIn.Focus()
				return p, textinput.Blink
			case "esc":
				p.step = 99
				return p, nil
			}
			return p, cmd
		case 2: // 描述
			var cmd tea.Cmd
			p.descIn, cmd = p.descIn.Update(msg)
			switch msg.String() {
			case "enter":
				p.step = 3
				p.descIn.Blur()
				p.colorIn.Focus()
				return p, textinput.Blink
			case "esc":
				p.step = 1
				p.descIn.Blur()
				p.titleIn.Focus()
				return p, textinput.Blink
			}
			return p, cmd
		case 3: // 颜色
			var cmd tea.Cmd
			p.colorIn, cmd = p.colorIn.Update(msg)
			switch msg.String() {
			case "enter":
				p.step = 4
				p.colorIn.Blur()
				return p, nil
			case "esc":
				p.step = 2
				p.colorIn.Blur()
				p.descIn.Focus()
				return p, textinput.Blink
			}
			return p, cmd
		case 4: // 确认
			if p.minting {
				return p, nil // 正在异步发交易，忽略按键
			}
			switch msg.String() {
			case "y", "Y":
				p.title = p.titleIn.Value()
				p.desc = p.descIn.Value()
				p.color = p.colorIn.Value()
				p.minting = true
				p.err = ""
				p.step = 5
				return p, p.doMintCmd()
			case "esc", "q":
				// 直接退出 Mint 视图（包括 mint 失败后）
				p.step = 99
				return p, nil
			case "n", "N":
				// 回 step 1 重新输入标题
				p.step = 1
				p.titleIn.Focus()
				return p, textinput.Blink
			}
		case 5: // 结果页 — 链上返回 or 离线预览
			switch msg.String() {
			case "enter", "esc", "q":
				p.step = 99
				return p, nil
			}
		}
	}
	return p, nil
}

// doMintCmd — 异步调用 card.Service.Mint（不阻塞 Bubble Tea 主循环）
func (p PhaseMint) doMintCmd() tea.Cmd {
	params := card.MintParams{
		Lat:         p.lat,
		Lon:         p.lon,
		Res:         p.res,
		Title:       p.title,
		Description: p.desc,
		Color:       p.color,
	}
	svc := p.cardService
	return func() tea.Msg {
		if svc == nil {
			return mintResultMsg{err: fmt.Errorf("GeoyuanCard 服务不可用（chain 未连接），降级为 H3 预览")}
		}
		result, err := svc.Mint(params)
		return mintResultMsg{result: result, err: err}
	}
}

// ─── View ────────────────────────────────────────────────

func (p PhaseMint) View(width, height int, geoLoc string) string {
	neon := lipgloss.NewStyle().Foreground(lipgloss.Color("201")).Bold(true)
	cyan := lipgloss.NewStyle().Foreground(lipgloss.Color("45")).Bold(true)
	green := lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true)
	yellow := lipgloss.NewStyle().Foreground(lipgloss.Color("228")).Bold(true)
	red := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))

	var sb strings.Builder
	sb.WriteString(neon.Render("🎴 Mint a Geoyuan Card") + "\n\n")

	switch p.step {
	case 0:
		sb.WriteString(fmt.Sprintf("  📍 Your location:  %s\n", geoLoc))
		sb.WriteString(fmt.Sprintf("  🔷 H3 cell:        %s  (res=%d, ~%.1f km²)\n",
			H3Hex(p.h3Index), p.res, h3.AreaKm2(p.h3Index)))
		sb.WriteString(fmt.Sprintf("  📏 Center:          %.4f°N, %.4f°E\n", p.lat, p.lon))
		ns := h3.Neighbors(p.h3Index)
		sb.WriteString(fmt.Sprintf("  🧭 Neighbors:      %d cells (k=1 ring → 7 total)\n", len(ns)))
		sh := h3.Encode(31.2304, 121.4737, p.res)
		sb.WriteString(fmt.Sprintf("  🗺️  Grid distance to Shanghai: %d hops\n", h3.GridDistance(p.h3Index, sh.Index)))
		sb.WriteString("\n")
		sb.WriteString("  ┌─────────────────────────────────────┐\n")
		sb.WriteString("  │  Press " + cyan.Render("Enter") + " to mint a card here        │\n")
		sb.WriteString("  │  Press " + dim.Render("Esc") + "  to cancel                     │\n")
		sb.WriteString("  └─────────────────────────────────────┘\n")

	case 1:
		sb.WriteString("  " + cyan.Render("Step 1/3") + "  Card title (2-30 chars)\n\n")
		sb.WriteString("  > " + p.titleIn.View() + "\n")
		if p.err != "" {
			sb.WriteString(red.Render("  ✗ "+p.err) + "\n")
		}

	case 2:
		sb.WriteString("  " + cyan.Render("Step 2/3") + "  Short description\n\n")
		sb.WriteString("  > " + p.descIn.View() + "\n")

	case 3:
		sb.WriteString("  " + cyan.Render("Step 3/3") + "  Neon color (hex e.g. #c77dff)\n\n")
		sb.WriteString("  > " + p.colorIn.View() + "\n\n")
		sb.WriteString("  Palette: ")
		for _, c := range []string{"#c77dff", "#ff6ec7", "#00f5ff", "#00ff9d", "#ffd700", "#ff3864"} {
			sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(c[1:])).Render("■ "))
		}
		sb.WriteString("\n")

	case 4:
		// 确认页（同时也是错误重试页）
		sb.WriteString("  " + cyan.Render("Confirm — Mint this card?") + "\n\n")
		sb.WriteString(fmt.Sprintf("    Title:     %s\n", p.titleIn.Value()))
		sb.WriteString(fmt.Sprintf("    Desc:      %s\n", p.descIn.Value()))
		sb.WriteString(fmt.Sprintf("    Color:     %s\n", p.colorIn.Value()))
		sb.WriteString(fmt.Sprintf("    Location:  %s\n", geoLoc))
		sb.WriteString(fmt.Sprintf("    H3 Cell:   %s (res=%d)\n", H3Hex(p.h3Index), p.res))
		sb.WriteString(fmt.Sprintf("    Lat/Lon:   %.4f, %.4f\n", p.lat, p.lon))
		if p.cardService != nil && p.cardService.Ready() {
			sb.WriteString("    Chain:     " + green.Render("Polygon Amoy (live)") + "\n")
		} else {
			sb.WriteString("    Chain:     " + yellow.Render("offline preview") + "\n")
		}
		sb.WriteString("\n  [" + green.Render("Y") + "] Mint  [" + dim.Render("N") + "] Edit  [" + dim.Render("Esc") + "] Cancel\n")
		if p.mintErr != nil {
			sb.WriteString(red.Render(fmt.Sprintf("\n  ✗ Previous attempt failed: %v", p.mintErr)) + "\n")
		}

	case 5:
		if p.minting {
			sb.WriteString(yellow.Render("  ⛓️  Submitting transaction to Polygon Amoy...") + "\n")
			sb.WriteString(dim.Render("     (this may take 30-60 seconds)") + "\n")
			sb.WriteString("\n  Press " + dim.Render("Enter/Esc") + " to wait — it keeps going in background\n")
			return sb.String()
		}
		if p.mintErr != nil {
			sb.WriteString(red.Render("  ✗ Mint failed") + "\n\n")
			sb.WriteString(fmt.Sprintf("    Reason: %v\n", p.mintErr))
			sb.WriteString("\n  Press " + cyan.Render("Enter") + " or " + dim.Render("Esc") + " to return\n")
			return sb.String()
		}
		// 成功
		sb.WriteString(green.Render("  ✅ Card minted successfully on Polygon Amoy!") + "\n\n")
		if p.mintResult != nil {
			if p.mintResult.Preview {
				sb.WriteString(yellow.Render("  (offline preview — contract not ready)") + "\n\n")
			}
			sb.WriteString(fmt.Sprintf("    Title:     %s\n", p.mintResult.Title))
			sb.WriteString(fmt.Sprintf("    Desc:      %s\n", p.mintResult.Description))
			sb.WriteString(fmt.Sprintf("    Color:     %s\n", p.mintResult.Color))
			sb.WriteString(fmt.Sprintf("    H3 Cell:   %s\n", H3Hex(p.mintResult.H3Index)))
			sb.WriteString(fmt.Sprintf("    Lat/Lon:   %.4f, %.4f\n", p.mintResult.Lat, p.mintResult.Lon))
			if p.mintResult.TokenID != nil && p.mintResult.TokenID.Sign() > 0 {
				sb.WriteString(fmt.Sprintf("    Token ID:  #%s\n", p.mintResult.TokenID.Text(10)))
			}
			if !p.mintResult.Preview {
				sb.WriteString(fmt.Sprintf("    Tx Hash:   %s\n", p.mintResult.TxHash))
				sb.WriteString(fmt.Sprintf("    Block:     %d\n", p.mintResult.Block))
			}
		}
		sb.WriteString("\n  Press " + cyan.Render("Enter") + " or " + dim.Render("Esc") + " to return\n")
	}

	return sb.String()
}
