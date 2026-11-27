package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

func TestHeaderBarFitsTerminalWidth(t *testing.T) {
	widths := []int{1, 2, 5, 10, 19, 20, 23, 24, 30, 40, 60, 67, 68, 80, 100, 120, 149, 150, 200}
	for _, width := range widths {
		t.Run(fmt.Sprint(width), func(t *testing.T) {
			m := Model{
				phase:      PhaseListCard,
				termWidth:  width,
				termHeight: 30,
				remoteIP:   "203.0.113.42",
			}
			out := m.headerBar()
			if out == "" {
				t.Fatal("headerBar returned no content")
			}
			for lineNo, line := range strings.Split(strings.TrimRight(out, "\n"), "\n") {
				if got := ansi.StringWidth(line); got > width {
					t.Errorf("line %d is %d cells wide; terminal is %d", lineNo+1, got, width)
				}
			}
			if width >= 8 && !strings.Contains(strings.ToLower(ansi.Strip(out)), "geoyuan") {
				t.Error("headerBar lost the geoyuan wordmark")
			}
		})
	}
}

func TestRenderLogoFitsAvailableWidth(t *testing.T) {
	for width := 0; width <= 200; width++ {
		out := RenderLogo(width)
		if width == 0 {
			if out != "" {
				t.Errorf("RenderLogo(0) = %q; want empty output", out)
			}
			continue
		}
		for lineNo, line := range strings.Split(strings.TrimRight(out, "\n"), "\n") {
			if got := ansi.StringWidth(line); got > width {
				t.Fatalf("RenderLogo(%d) line %d is %d cells wide", width, lineNo+1, got)
			}
		}
	}
}

func TestRenderLogoSelectsCompleteTier(t *testing.T) {
	tests := []struct {
		name     string
		width    int
		contains string
		lines    int
	}{
		{"braille", 100, "⣿", 6},
		{"fallback", 40, "geoyuan.com", 3},
		{"compact", 20, "GEOYUAN", 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := RenderLogo(tt.width)
			plain := ansi.Strip(out)
			if !strings.Contains(plain, tt.contains) {
				t.Errorf("RenderLogo(%d) does not contain %q", tt.width, tt.contains)
			}
			if got := len(strings.Split(strings.TrimRight(out, "\n"), "\n")); got != tt.lines {
				t.Errorf("RenderLogo(%d) has %d lines; want %d", tt.width, got, tt.lines)
			}
		})
	}
}

func TestRenderLogoCentersFallback(t *testing.T) {
	const width = 40
	out := RenderLogo(width)
	for lineNo, line := range strings.Split(strings.TrimRight(out, "\n"), "\n") {
		plain := ansi.Strip(line)
		leading := len(plain) - len(strings.TrimLeft(plain, " "))
		trailing := width - ansi.StringWidth(plain)
		if difference := leading - trailing; difference < -1 || difference > 1 {
			t.Errorf("line %d has %d leading and %d available trailing cells", lineNo+1, leading, trailing)
		}
	}
}

func TestLogoAssetsHaveExpectedCellWidths(t *testing.T) {
	tests := []struct {
		name string
		art  string
		want int
	}{
		{"braille", geoyuanLogoBraille, 58},
		{"halfblock", geoyuanLogoHalfblock, 64},
		{"fallback", geoyuanLogoFallback, 30},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := maxLogoWidth(tt.art); got != tt.want {
				t.Errorf("%s width = %d; want %d", tt.name, got, tt.want)
			}
			for lineNo, line := range strings.Split(ansi.Strip(tt.art), "\n") {
				if tt.name == "braille" && strings.Contains(line, "m") {
					t.Errorf("%s line %d contains a visible marker", tt.name, lineNo+1)
				}
			}
		})
	}
}

func TestBrailleLogoHasNoBlankScanlines(t *testing.T) {
	leftBits := [...]int{0, 1, 2, 6}
	rightBits := [...]int{3, 4, 5, 7}
	for lineNo, line := range strings.Split(strings.TrimRight(ansi.Strip(geoyuanLogoBraille), "\n"), "\n") {
		for dotRow := 0; dotRow < 4; dotRow++ {
			hasDot := false
			for _, r := range line {
				code := r - '\u2800'
				if code&(1<<leftBits[dotRow]) != 0 || code&(1<<rightBits[dotRow]) != 0 {
					hasDot = true
					break
				}
			}
			if !hasDot {
				t.Errorf("Braille line %d dot row %d is blank", lineNo+1, dotRow+1)
			}
		}
	}
}

func TestBrailleLogoUsesCyberBlueGradient(t *testing.T) {
	out := RenderLogo(100)
	for _, want := range []string{"\x1b[1m", "38;2;30;64;175", "38;2;34;211;238", "38;2;99;102;241"} {
		if !strings.Contains(out, want) {
			t.Errorf("Braille logo missing %q", want)
		}
	}
	for _, oldColor := range []string{"38;2;34;197;94", "38;2;249;115;22", "38;2;234;179;8"} {
		if strings.Contains(out, oldColor) {
			t.Errorf("Braille logo still uses old color %q", oldColor)
		}
	}
	if strings.Contains(out, "█") {
		t.Error("Braille logo contains full-block artwork")
	}
}

func TestDisplayWidthUsesTerminalCells(t *testing.T) {
	if got := displayWidth("\x1b[31m界🙂\x1b[0m"); got != 4 {
		t.Errorf("displayWidth() = %d; want 4", got)
	}
	if got := padTo("界", 4); got != "界  " {
		t.Errorf("padTo() = %q; want %q", got, "界  ")
	}
}

func TestHeaderBarRespondsToResize(t *testing.T) {
	m := Model{phase: PhaseMenu, termWidth: 160, termHeight: 30}
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 40, Height: 30})
	resized := updated.(Model)
	plain := ansi.Strip(resized.headerBar())
	if !strings.Contains(plain, "geoyuan.com") {
		t.Error("narrow terminal did not select the complete fallback logo")
	}
	for lineNo, line := range strings.Split(strings.TrimRight(resized.headerBar(), "\n"), "\n") {
		if got := ansi.StringWidth(line); got > 40 {
			t.Errorf("resized line %d is %d cells wide; want at most 40", lineNo+1, got)
		}
	}
}

func TestShortTerminalUsesCompactHeader(t *testing.T) {
	m := Model{phase: PhaseMenu, termWidth: 120, termHeight: 19}
	out := m.headerBar()
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) > 3 {
		t.Errorf("compact header has %d lines; want at most 3", len(lines))
	}
	if !strings.Contains(ansi.Strip(out), "GEOYUAN") {
		t.Error("compact header lost the GEOYUAN wordmark")
	}
}

func TestHelpOverlayKeepsResponsiveHeader(t *testing.T) {
	m := Model{phase: PhaseHelp, prevPhase: PhaseMenu, termWidth: 100, termHeight: 40}
	out := m.viewHelpOverlay()
	header := out
	if idx := strings.Index(out, "\n\n"); idx >= 0 {
		header = out[:idx+2]
	}
	if !strings.Contains(strings.ToLower(ansi.Strip(header)), "geoyuan") {
		t.Error("help overlay lost the header wordmark")
	}
	for lineNo, line := range strings.Split(strings.TrimRight(header, "\n"), "\n") {
		if got := ansi.StringWidth(line); got > 100 {
			t.Errorf("help header line %d is %d cells wide; want at most 100", lineNo+1, got)
		}
	}
}

func TestStatusLineFitsTerminalWidth(t *testing.T) {
	widths := []int{1, 2, 5, 10, 19, 20, 23, 24, 30, 40, 50, 60, 80, 100, 120, 200}
	phases := []Phase{PhaseRegister, PhaseQuizLoading, PhaseListCard, PhaseMarket}
	for _, phase := range phases {
		for _, width := range widths {
			m := Model{phase: phase, termWidth: width, termHeight: 30}
			out := m.statusLine()
			if out == "" {
				t.Fatalf("statusLine returned no content for phase %v at width %d", phase, width)
			}
			for lineNo, line := range strings.Split(strings.TrimRight(out, "\n"), "\n") {
				if got := ansi.StringWidth(line); got > width {
					t.Errorf("phase %v width %d line %d is %d cells wide", phase, width, lineNo+1, got)
				}
			}
		}
	}
}

func TestStatusLineReportsOfflineChain(t *testing.T) {
	m := Model{phase: PhaseMenu, termWidth: 100, termHeight: 30}
	plain := ansi.Strip(m.statusLine())
	if !strings.Contains(plain, "Offline") {
		t.Error("offline statusLine did not show Offline")
	}
	if strings.Contains(plain, "Amoy") {
		t.Error("offline statusLine still claimed an Amoy connection")
	}
}

func TestStatusLineShowsQuizLoadingBadge(t *testing.T) {
	m := Model{phase: PhaseQuizLoading, termWidth: 100, termHeight: 30}
	if plain := ansi.Strip(m.statusLine()); !strings.Contains(plain, "LOADING") {
		t.Error("quiz-loading statusLine did not show LOADING")
	}
}

func TestShortWalletAddr(t *testing.T) {
	tests := []struct {
		addr string
		want string
	}{
		{"", ""},
		{"0x1234", ""},
		{"0x1234567890abcdef1234567890abcdef12345678", "0x1234…5678"},
	}
	for _, tt := range tests {
		if got := shortWalletAddr(tt.addr); got != tt.want {
			t.Errorf("shortWalletAddr(%q) = %q; want %q", tt.addr, got, tt.want)
		}
	}
}

func TestRegisterAndMintIncludeStatusLine(t *testing.T) {
	t.Run("register", func(t *testing.T) {
		m := Model{phase: PhaseRegister, termWidth: 100, termHeight: 40}
		plain := ansi.Strip(m.View())
		if !strings.Contains(plain, "REGISTER") || !strings.Contains(plain, "Offline") {
			t.Error("register view lost its phase/status footer")
		}
	})
	t.Run("mint", func(t *testing.T) {
		m := Model{phase: PhaseMintCard, termWidth: 100, termHeight: 40}
		m.mintView.step = 1
		plain := ansi.Strip(m.View())
		if !strings.Contains(plain, "MINT") || !strings.Contains(plain, "Offline") {
			t.Error("mint view lost its phase/status footer")
		}
	})
}
