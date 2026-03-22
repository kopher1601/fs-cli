package components

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kopher1601/fs-cli/internal/ui"
)

// flashDismissMsg is sent when the flash should be hidden.
type flashDismissMsg struct {
	gen int // only dismiss if generation matches
}

// Flash displays temporary messages.
type Flash struct {
	message string
	level   ui.FlashLevel
	visible bool
	width   int
	gen     int // incremented on each Show
}

// NewFlash creates a new Flash component.
func NewFlash() Flash {
	return Flash{}
}

// SetSize updates the flash width.
func (f *Flash) SetSize(w int) {
	f.width = w
}

// Show displays a flash message and returns a command to auto-dismiss.
func (f *Flash) Show(msg string, level ui.FlashLevel) tea.Cmd {
	f.gen++
	f.message = msg
	f.level = level
	f.visible = true
	gen := f.gen
	return tea.Tick(4*time.Second, func(time.Time) tea.Msg {
		return flashDismissMsg{gen: gen}
	})
}

// Update handles flash messages.
func (f *Flash) Update(msg tea.Msg) {
	if m, ok := msg.(flashDismissMsg); ok {
		// Only dismiss if no newer message was shown
		if m.gen == f.gen {
			f.visible = false
			f.message = ""
		}
	}
}

// View renders the flash message.
func (f Flash) View() string {
	if !f.visible || f.message == "" {
		return ""
	}

	var style = ui.StyleFlashInfo
	prefix := "ℹ "
	switch f.level {
	case ui.FlashSuccess:
		style = ui.StyleFlashSuccess
		prefix = "✓ "
	case ui.FlashWarn:
		style = ui.StyleFlashWarn
		prefix = "⚠ "
	case ui.FlashError:
		style = ui.StyleFlashError
		prefix = "✗ "
	}

	return style.Width(f.width).Render(prefix + f.message)
}
