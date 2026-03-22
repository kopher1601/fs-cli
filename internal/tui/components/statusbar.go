package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/kopher1601/fs-cli/internal/ui"
)

// KeyHint represents a single key-action hint.
type KeyHint struct {
	Key  string
	Desc string
}

// StatusBar renders key hints at the bottom of the screen.
type StatusBar struct {
	hints []KeyHint
	width int
}

// NewStatusBar creates a new StatusBar.
func NewStatusBar() StatusBar {
	return StatusBar{}
}

// SetSize updates the status bar width.
func (s *StatusBar) SetSize(w int) {
	s.width = w
}

// SetHints sets the key hints to display.
func (s *StatusBar) SetHints(hints []KeyHint) {
	s.hints = hints
}

// View renders the status bar.
func (s StatusBar) View() string {
	if len(s.hints) == 0 {
		return ""
	}

	var parts []string
	for _, h := range s.hints {
		k := ui.StyleKeyHint.Render(h.Key)
		d := ui.StyleKeyDesc.Render(":" + h.Desc)
		parts = append(parts, k+d)
	}

	content := strings.Join(parts, "  ")
	return lipgloss.NewStyle().
		Width(s.width).
		Background(lipgloss.Color("#1F2937")).
		Padding(0, 1).
		Render(content)
}
