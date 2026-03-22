package views

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kopher1601/fs-cli/internal/tui/components"
	"github.com/kopher1601/fs-cli/internal/ui"
)

// HelpModel displays the help screen.
type HelpModel struct {
	keys   ui.KeyMap
	width  int
	height int
}

// NewHelpModel creates a new HelpModel.
func NewHelpModel(keys ui.KeyMap) HelpModel {
	return HelpModel{keys: keys}
}

// SetSize updates the view dimensions.
func (m *HelpModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

// Update handles messages.
func (m HelpModel) Update(msg tea.Msg) (HelpModel, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		if key.Matches(msg, m.keys.Back) || key.Matches(msg, m.keys.Quit) || key.Matches(msg, m.keys.Help) {
			return m, func() tea.Msg { return ui.BackMsg{} }
		}
	}
	return m, nil
}

// View renders the help screen.
func (m HelpModel) View() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(ui.StyleTitle.Render("  Keyboard Shortcuts") + "\n\n")

	writeSection(&b, "Global", []helpEntry{
		{"↑/k", "Move up"},
		{"↓/j", "Move down"},
		{"g", "Go to top"},
		{"G", "Go to bottom"},
		{"n", "Next page"},
		{"p", "Previous page"},
		{"/", "Filter"},
		{"?", "Toggle help"},
		{"o", "Operations"},
		{"Ctrl+R", "Refresh"},
		{"q/Esc", "Back / Quit"},
		{"Ctrl+C", "Force quit"},
	})

	writeSection(&b, "Stores View", []helpEntry{
		{"Enter", "Open documents"},
		{"c", "Create store"},
		{"d", "Delete store"},
		{"D", "Force delete store"},
		{"y", "Store detail"},
	})

	writeSection(&b, "Documents View", []helpEntry{
		{"Enter", "Document detail"},
		{"u", "Upload file"},
		{"d", "Delete document"},
		{"D", "Force delete document"},
	})

	return b.String()
}

type helpEntry struct {
	key  string
	desc string
}

func writeSection(b *strings.Builder, title string, entries []helpEntry) {
	b.WriteString(ui.StyleKeyHint.Render("  "+title) + "\n")
	for _, e := range entries {
		k := ui.StyleKeyHint.Render("    " + pad(e.key, 12))
		d := ui.StyleKeyDesc.Render(e.desc)
		b.WriteString(k + d + "\n")
	}
	b.WriteString("\n")
}

func pad(s string, w int) string {
	if len(s) >= w {
		return s
	}
	return s + strings.Repeat(" ", w-len(s))
}

// Hints returns the key hints.
func (m HelpModel) Hints() []components.KeyHint {
	return []components.KeyHint{
		{Key: "esc", Desc: "back"},
	}
}

// BreadcrumbItems returns the breadcrumb path.
func (m HelpModel) BreadcrumbItems() []string {
	return []string{"Help"}
}
