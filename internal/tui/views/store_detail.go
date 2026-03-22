package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dustin/go-humanize"
	"github.com/kopher1601/fs-cli/internal/api"
	"github.com/kopher1601/fs-cli/internal/model"
	"github.com/kopher1601/fs-cli/internal/tui/components"
	"github.com/kopher1601/fs-cli/internal/ui"
)

// StoreDetailModel displays details for a single store.
type StoreDetailModel struct {
	store  *api.FileSearchStore
	keys   ui.KeyMap
	width  int
	height int
}

// NewStoreDetailModel creates a new StoreDetailModel.
func NewStoreDetailModel(keys ui.KeyMap) StoreDetailModel {
	return StoreDetailModel{keys: keys}
}

// SetStore sets the store to display.
func (m *StoreDetailModel) SetStore(store *api.FileSearchStore) {
	m.store = store
}

// SetSize updates the view dimensions.
func (m *StoreDetailModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

// Update handles messages.
func (m StoreDetailModel) Update(msg tea.Msg) (StoreDetailModel, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(msg, m.keys.Delete):
			if m.store != nil {
				return m, func() tea.Msg {
					return ConfirmDeleteMsg{
						ResourceName: m.store.Name,
						DisplayName:  m.store.DisplayName,
						Force:        false,
					}
				}
			}
		case key.Matches(msg, m.keys.ForceDel):
			if m.store != nil {
				return m, func() tea.Msg {
					return ConfirmDeleteMsg{
						ResourceName: m.store.Name,
						DisplayName:  m.store.DisplayName,
						Force:        true,
					}
				}
			}
		}
	}
	return m, nil
}

// View renders the store detail.
func (m StoreDetailModel) View() string {
	if m.store == nil {
		return "  No store selected"
	}

	s := m.store
	var b strings.Builder

	b.WriteString("\n")
	writeLine(&b, "Name", s.Name)
	writeLine(&b, "Display Name", s.DisplayName)
	writeLine(&b, "Created", fmt.Sprintf("%s (%s)", s.CreateTime, model.FormatAge(s.CreateTime)))
	writeLine(&b, "Updated", fmt.Sprintf("%s (%s)", s.UpdateTime, model.FormatAge(s.UpdateTime)))
	b.WriteString("\n")
	writeLine(&b, "Active Documents", fmt.Sprintf("%d", s.ActiveDocumentsCount))
	writeLine(&b, "Pending Documents", fmt.Sprintf("%d", s.PendingDocumentsCount))
	writeLine(&b, "Failed Documents", fmt.Sprintf("%d", s.FailedDocumentsCount))
	writeLine(&b, "Total Size", humanize.IBytes(uint64(s.SizeBytes)))
	b.WriteString("\n")

	return b.String()
}

func writeLine(b *strings.Builder, label, value string) {
	l := ui.StyleSubtle.Render(fmt.Sprintf("  %-20s", label+":"))
	v := ui.StyleTitle.Render(value)
	b.WriteString(l + " " + v + "\n")
}

// Hints returns the key hints.
func (m StoreDetailModel) Hints() []components.KeyHint {
	return []components.KeyHint{
		{Key: "esc", Desc: "back"},
		{Key: "d", Desc: "delete"},
	}
}

// BreadcrumbItems returns the breadcrumb path.
func (m StoreDetailModel) BreadcrumbItems() []string {
	name := "Store"
	if m.store != nil {
		name = m.store.DisplayName
		if name == "" {
			name = model.ShortName(m.store.Name)
		}
	}
	return []string{"Stores", name}
}
