package tui

import (
	"fmt"
	"math/big"
	"math/rand"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"geoyuan.cn/ssh/internal/geo"
	"geoyuan.cn/ssh/internal/h3"
)

// PhaseMint — Bubble Tea Mint 视图
type PhaseMint struct {
	step         int // 0=预览H3, 1=标题, 2=描述, 3=颜色, 4=确认, 5=完成, 99=退出
	titleIn      textinput.Model
	descIn       textinput.Model
	colorIn      textinput.Model
	h3Index      uint64
	lat, lon     float64
	res          int
	title, desc  string
	color        string
	err          string
	loading      bool
	loadingStart time.Time
}

// NewPhaseMint — 初始化 Mint 视图（自动用 IP 做 H3 预览）
func NewPhaseMint(remoteIP, geoLoc string) PhaseMint {
	pm := PhaseMint{res: h3.DefaultRes}

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
	pm.titleIn.Focus()
	pm.titleIn.Width = 30

	pm.descIn = textinput.New()
	pm.descIn.Placeholder = "My first Geoyuan card"
	pm.descIn.Width = 40

	pm.colorIn = textinput.New()
	pm.colorIn.SetValue(randomNeonColor())
	pm.colorIn.Width = 10

	return pm
}

// randomNeonColor — 霓虹调色板随机一个
func randomNeonColor() string {
	colors := []string{"#c77dff", "#ff6ec7", "#00f5ff", "#00ff9d", "#ffd700", "#ff3864"}
	return colors[rand.Intn(len(colors))]
}

// H3Hex — uint64 → "0x..." hex
func H3Hex(idx uint64) string {
	return fmt.Sprintf("0x%x", idx)
}

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
				if strings.TrimSpace(p.titleIn.Value()) == "" {
					p.err = "Title can't be empty"
					return p, nil
				}
				p.err = ""
				p.step = 2
				return p, nil
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
				return p, nil
			case "esc":
				p.step = 1
				return p, nil
			}
			return p, cmd
		case 3: // 颜色
			var cmd tea.Cmd
			p.colorIn, cmd = p.colorIn.Update(msg)
			switch msg.String() {
			case "enter":
				p.step = 4
				return p, nil
			case "esc":
				p.step = 2
				return p, nil
			}
			return p, cmd
		case 4: // 确认
			switch msg.String() {
			case "y", "Y":
				p.title = p.titleIn.Value()
				p.desc = p.descIn.Value()
				p.color = p.colorIn.Value()
				p.loading = true
				p.loadingStart = time.Now()
				p.step = 5
				return p, nil
			case "esc", "n", "N":
				p.step = 1
				return p, nil
			}
		case 5: // 完成/loading → Esc 返回 Welcome
			switch msg.String() {
			case "enter", "esc":
				p.step = 99
				return p, nil
			}
		}
	}
	return p, nil
}

// Done — 退出 Mint 视图时返回 true
func (p PhaseMint) Done() bool {
	return p.step == 99
}

// ─── View ────────────────────────────────────────────────

func (p PhaseMint) View(width, height int, geoLoc string) string {
	neon := lipgloss.NewStyle().Foreground(lipgloss.Color("201")).Bold(true)
	cyan := lipgloss.NewStyle().Foreground(lipgloss.Color("45")).Bold(true)
	green := lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true)
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))

	var sb strings.Builder
	sb.WriteString(neon.Render("🎴 Mint a Geoyuan Card") + "\n\n")

	switch p.step {
	case 0:
		sb.WriteString(fmt.Sprintf("  📍 Your location:  %s\n", geoLoc))
		sb.WriteString(fmt.Sprintf("  🔷 H3 cell:        %s  (res=%d, ~%.1f km²)\n",
			H3Hex(p.h3Index), p.res, h3.AreaKm2(p.h3Index)))
		sb.WriteString(fmt.Sprintf("  📏 Center:          %.4f°N, %.4f°E\n", p.lat, p.lon))
		// 邻居预览
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
			sb.WriteString(dim.Render("  ✗ "+p.err) + "\n")
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
		sb.WriteString("  " + cyan.Render("Confirm — Mint this card?") + "\n\n")
		sb.WriteString(fmt.Sprintf("    Title:     %s\n", p.titleIn.Value()))
		sb.WriteString(fmt.Sprintf("    Desc:      %s\n", p.descIn.Value()))
		sb.WriteString(fmt.Sprintf("    Color:     %s\n", p.colorIn.Value()))
		sb.WriteString(fmt.Sprintf("    Location:  %s\n", geoLoc))
		sb.WriteString(fmt.Sprintf("    H3 Cell:   %s (res=%d)\n", H3Hex(p.h3Index), p.res))
		sb.WriteString(fmt.Sprintf("    Lat/Lon:   %.4f, %.4f\n", p.lat, p.lon))
		sb.WriteString("\n  [" + green.Render("Y") + "] Yes  [" + dim.Render("Esc") + "] Back\n")

	case 5:
		if p.loading {
			// 简化：本地业务逻辑完成，RPC 还没连 — 展示 preview 结果
			sb.WriteString(green.Render("  ✅ Card prepared for minting!") + "\n\n")
			sb.WriteString(fmt.Sprintf("    Title:     %s\n", p.title))
			sb.WriteString(fmt.Sprintf("    Desc:      %s\n", p.desc))
			sb.WriteString(fmt.Sprintf("    Color:     %s\n", p.color))
			sb.WriteString(fmt.Sprintf("    H3 Cell:   %s (unique, no one else can mint here)\n", H3Hex(p.h3Index)))
			sb.WriteString(fmt.Sprintf("    Lat/Lon:   %.4f°N, %.4f°E\n", p.lat, p.lon))
			sb.WriteString(fmt.Sprintf("    Owner:     your wallet (connect Polygon RPC)\n\n"))
			sb.WriteString("  (Online minting needs Polygon RPC + deployed contract)\n")
			sb.WriteString("  Press " + cyan.Render("Enter") + " or " + dim.Render("Esc") + " to return\n")
			sb.WriteString(fmt.Sprintf("  Token ID would be: #%s\n", big.NewInt(time.Now().UnixNano()).Text(10)))
		}
	}

	return sb.String()
}
