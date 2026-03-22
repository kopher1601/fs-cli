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

// DocDetailModel displays details for a single document.
type DocDetailModel struct {
	store  *api.FileSearchStore
	doc    *api.Document
	keys   ui.KeyMap
	width  int
	height int
}

// NewDocDetailModel creates a new DocDetailModel.
func NewDocDetailModel(keys ui.KeyMap) DocDetailModel {
	return DocDetailModel{keys: keys}
}

// SetDoc sets the document to display.
func (m *DocDetailModel) SetDoc(store *api.FileSearchStore, doc *api.Document) {
	m.store = store
	m.doc = doc
}

// SetSize updates the view dimensions.
func (m *DocDetailModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

// Update handles messages.
func (m DocDetailModel) Update(msg tea.Msg) (DocDetailModel, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(msg, m.keys.Delete):
			if m.doc != nil {
				return m, func() tea.Msg {
					return ConfirmDeleteMsg{
						ResourceName: m.doc.Name,
						DisplayName:  m.doc.DisplayName,
						Force:        false,
						IsDocument:   true,
					}
				}
			}
		}
	}
	return m, nil
}

// View renders the document detail.
func (m DocDetailModel) View() string {
	if m.doc == nil {
		return "  No document selected"
	}

	d := m.doc
	var b strings.Builder

	b.WriteString("\n")
	writeLine(&b, "Name", d.Name)
	writeLine(&b, "Display Name", d.DisplayName)
	writeLine(&b, "State", renderState(d.State))
	writeLine(&b, "MIME Type", d.MimeType)
	writeLine(&b, "Size", humanize.IBytes(uint64(d.SizeBytes)))
	writeLine(&b, "Created", fmt.Sprintf("%s (%s)", d.CreateTime, model.FormatAge(d.CreateTime)))
	writeLine(&b, "Updated", fmt.Sprintf("%s (%s)", d.UpdateTime, model.FormatAge(d.UpdateTime)))

	if len(d.CustomMetadata) > 0 {
		b.WriteString("\n")
		b.WriteString(ui.StyleSubtle.Render("  Custom Metadata:") + "\n")
		for _, line := range model.FormatMetadata(d.CustomMetadata) {
			b.WriteString(ui.StyleTitle.Render(line) + "\n")
		}
	}

	b.WriteString("\n")

	return b.String()
}

func renderState(state string) string {
	switch state {
	case "STATE_ACTIVE":
		return ui.StyleStateActive.Render("ACTIVE")
	case "STATE_PENDING":
		return ui.StyleStatePending.Render("PENDING")
	case "STATE_FAILED":
		return ui.StyleStateFailed.Render("FAILED")
	default:
		return state
	}
}

// Hints returns the key hints.
func (m DocDetailModel) Hints() []components.KeyHint {
	return []components.KeyHint{
		{Key: "esc", Desc: "back"},
		{Key: "d", Desc: "delete"},
	}
}

// BreadcrumbItems returns the breadcrumb path.
func (m DocDetailModel) BreadcrumbItems() []string {
	storeName := "Store"
	docName := "Document"
	if m.store != nil {
		storeName = m.store.DisplayName
		if storeName == "" {
			storeName = model.ShortName(m.store.Name)
		}
	}
	if m.doc != nil {
		docName = m.doc.DisplayName
		if docName == "" {
			docName = model.ShortName(m.doc.Name)
		}
	}
	return []string{"Stores", storeName, "Documents", docName}
}
