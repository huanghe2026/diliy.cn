package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestKeys_CompleteAudit — 全按键审计：每个 Phase × 关键按键
// 验证 phase 变化是否符合预期（顶层 Esc→QuitConfirm，子页面 Esc→返回上级）
func TestKeys_CompleteAudit(t *testing.T) {
	tests := []struct {
		name      string
		initial   Phase
		key       string
		wantPhase Phase
	}{
		// ═══ 顶层页面：Esc → QuitConfirm ═══
		{"Register Esc→QuitConfirm", PhaseRegister, "esc", PhaseQuitConfirm},
		{"QuizLoading Esc→QuitConfirm", PhaseQuizLoading, "esc", PhaseQuitConfirm},
		{"Quiz Esc→QuitConfirm", PhaseQuiz, "esc", PhaseQuitConfirm},
		{"Result Esc→QuitConfirm", PhaseResult, "esc", PhaseQuitConfirm},
		{"Welcome Esc→QuitConfirm", PhaseWelcome, "esc", PhaseQuitConfirm},
		{"Retry Esc→QuitConfirm", PhaseRetry, "esc", PhaseQuitConfirm},
		{"Menu Esc→QuitConfirm", PhaseMenu, "esc", PhaseQuitConfirm},
		{"Menu q→QuitConfirm", PhaseMenu, "q", PhaseQuitConfirm},

		// ═══ 子页面：Esc → 返回上级 ═══
		{"MyCards Esc→Menu", PhaseMyCards, "esc", PhaseMenu},
		{"MyCards q→Menu", PhaseMyCards, "q", PhaseMenu},
		{"Market Esc→Menu", PhaseMarket, "esc", PhaseMenu},
		{"Market q→Menu", PhaseMarket, "q", PhaseMenu},
		{"ListCard Esc→MyCards", PhaseListCard, "esc", PhaseMyCards},
		{"ListCard q→MyCards", PhaseListCard, "q", PhaseMyCards},

		// ═══ QuitConfirm ═══
		{"QuitConfirm(n=0) n→取消回 prevPhase(0)", PhaseQuitConfirm, "n", PhaseRegister},
		{"QuitConfirm(n=0) Esc→取消回 prevPhase(0)", PhaseQuitConfirm, "esc", PhaseRegister},

		// ═══ Help ═══
		{"任何 Phase ? → Help", PhaseMenu, "?", PhaseHelp},
		{"任何 Phase / → Help", PhaseRegister, "/", PhaseHelp},

		// ═══ Retry 导航（不崩就行） ═══
		{"Retry k 不崩", PhaseRetry, "k", PhaseRetry},
		{"Retry l 不崩", PhaseRetry, "l", PhaseRetry},

		// ═══ Quiz Loading 忽略其他键 ═══
		{"QuizLoading j 忽略", PhaseQuizLoading, "j", PhaseQuizLoading},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Model{phase: tt.initial}
			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
			if tt.key == "esc" {
				msg = tea.KeyMsg{Type: tea.KeyEscape}
			}
			got, _ := m.handleKey(msg)
			gotM := got.(Model)

			if gotM.phase != tt.wantPhase {
				t.Errorf("FAIL: %s\n  key=%q from phase=%v\n  got phase=%v, want phase=%v",
					tt.name, tt.key, tt.initial, gotM.phase, tt.wantPhase)
			}
		})
	}
}

// TestKeys_QuitConfirm_prevPhase — 测试 Esc → QuitConfirm 后 prevPhase 是否正确保存
func TestKeys_QuitConfirm_prevPhase(t *testing.T) {
	topPhases := []Phase{
		PhaseRegister, PhaseQuizLoading, PhaseQuiz, PhaseResult,
		PhaseWelcome, PhaseRetry, PhaseMenu,
	}
	for _, p := range topPhases {
		m := Model{phase: p}
		msg := tea.KeyMsg{Type: tea.KeyEscape}
		got, _ := m.handleKey(msg)
		gotM := got.(Model)
		if gotM.phase != PhaseQuitConfirm {
			t.Errorf("Top phase %v Esc should go QuitConfirm, got %v", p, gotM.phase)
		}
		if gotM.prevPhase != p {
			t.Errorf("Top phase %v Esc should save prevPhase=%v, got %v", p, p, gotM.prevPhase)
		}

		// QuitConfirm Esc 应该回到 prevPhase
		m2 := Model{phase: PhaseQuitConfirm, prevPhase: p}
		got2, _ := m2.updateQuitConfirm(tea.KeyMsg{Type: tea.KeyEscape})
		got2M := got2.(Model)
		if got2M.phase != p {
			t.Errorf("QuitConfirm Esc should restore prevPhase=%v, got %v", p, got2M.phase)
		}
	}
}
