package views

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kopher1601/fs-cli/internal/api"
	"github.com/kopher1601/fs-cli/internal/model"
	"github.com/kopher1601/fs-cli/internal/tui/components"
	"github.com/kopher1601/fs-cli/internal/ui"
)

// DocsLoadedMsg is sent when documents are fetched.
type DocsLoadedMsg struct {
	Documents []api.Document
	NextToken string
}

// DocDeletedMsg is sent after a document is deleted.
type DocDeletedMsg struct{}

// DocumentsModel is the view model for the documents list.
type DocumentsModel struct {
	table     components.Table
	client    *api.Client
	keys      ui.KeyMap
	store     *api.FileSearchStore
	docs      []api.Document
	pageToken string
	nextToken string
	loading   bool
	width     int
	height    int
}

// NewDocumentsModel creates a new DocumentsModel.
func NewDocumentsModel(client *api.Client, keys ui.KeyMap) DocumentsModel {
	table := components.NewTable(model.DocColumns(), keys)
	return DocumentsModel{
		table:  table,
		client: client,
		keys:   keys,
	}
}

// SetStore sets the store context and triggers a data load.
func (m *DocumentsModel) SetStore(store *api.FileSearchStore) tea.Cmd {
	m.store = store
	m.pageToken = ""
	m.nextToken = ""
	m.docs = nil
	return m.fetchDocs("")
}

// SetSize updates the view dimensions.
func (m *DocumentsModel) SetSize(w, h int) {
	m.width = w
	m.height = h
	m.table.SetSize(w, h)
}

// SetFilter applies a text filter to the table.
func (m *DocumentsModel) SetFilter(f string) {
	m.table.SetFilter(f)
}

// SelectedDoc returns the currently selected document.
func (m *DocumentsModel) SelectedDoc() *api.Document {
	idx := m.table.SelectedIndex()
	if idx < 0 || idx >= len(m.docs) {
		return nil
	}
	return &m.docs[idx]
}

// Update handles messages.
func (m DocumentsModel) Update(msg tea.Msg) (DocumentsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case DocsLoadedMsg:
		m.loading = false
		m.docs = msg.Documents
		m.nextToken = msg.NextToken
		rows := make([][]string, len(m.docs))
		for i, d := range m.docs {
			rows[i] = model.DocRowToStrings(model.DocToRow(&d))
		}
		m.table.SetRows(rows)

	case DocDeletedMsg:
		return m, m.fetchDocs(m.pageToken)

	case ui.RefreshMsg:
		if m.store != nil {
			return m, m.fetchDocs(m.pageToken)
		}

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Enter):
			if d := m.SelectedDoc(); d != nil {
				return m, func() tea.Msg {
					return ui.NavigateMsg{View: ui.ViewDocDetail, Store: m.store, Doc: d}
				}
			}

		case key.Matches(msg, m.keys.Delete), key.Matches(msg, m.keys.ForceDel):
			return m, m.deleteSelected(true)

		case key.Matches(msg, m.keys.Upload):
			if m.store != nil {
				return m, func() tea.Msg {
					return ui.NavigateMsg{View: ui.ViewUpload, Store: m.store}
				}
			}

		case key.Matches(msg, m.keys.NextPage):
			if m.nextToken != "" {
				m.pageToken = m.nextToken
				m.loading = true
				return m, m.fetchDocs(m.nextToken)
			}

		case key.Matches(msg, m.keys.PrevPage):
			if m.pageToken != "" {
				m.pageToken = ""
				m.loading = true
				return m, m.fetchDocs("")
			}

		default:
			m.table.Update(msg)
		}
	}

	return m, nil
}

// View renders the documents table.
func (m DocumentsModel) View() string {
	return m.table.View()
}

// Hints returns the key hints for the status bar.
func (m DocumentsModel) Hints() []components.KeyHint {
	hints := []components.KeyHint{
		{Key: "space", Desc: "select"},
		{Key: "u", Desc: "upload"},
		{Key: "d", Desc: "delete"},
		{Key: "enter", Desc: "detail"},
		{Key: "esc", Desc: "back"},
	}
	if m.table.SelectedCount() > 0 {
		hints[2] = components.KeyHint{Key: "d", Desc: fmt.Sprintf("delete(%d)", m.table.SelectedCount())}
	}
	return hints
}

// BreadcrumbItems returns the breadcrumb path.
func (m DocumentsModel) BreadcrumbItems() []string {
	storeName := "Store"
	if m.store != nil {
		storeName = m.store.DisplayName
		if storeName == "" {
			storeName = model.ShortName(m.store.Name)
		}
	}
	return []string{"Stores", storeName, "Documents"}
}

func (m DocumentsModel) deleteSelected(force bool) tea.Cmd {
	// Multi-select: use selected indices
	indices := m.table.SelectedIndices()
	if len(indices) > 0 {
		var names []string
		for _, idx := range indices {
			if idx < len(m.docs) {
				names = append(names, m.docs[idx].Name)
			}
		}
		count := len(names)
		return func() tea.Msg {
			return BatchDeleteMsg{
				ResourceNames: names,
				Force:         force,
				Count:         count,
			}
		}
	}
	// Single: use cursor
	if d := m.SelectedDoc(); d != nil {
		return func() tea.Msg {
			return ConfirmDeleteMsg{
				ResourceName: d.Name,
				DisplayName:  d.DisplayName,
				Force:        force,
				IsDocument:   true,
			}
		}
	}
	return nil
}

// BatchDeleteMsg is sent when multiple documents need to be deleted.
type BatchDeleteMsg struct {
	ResourceNames []string
	Force         bool
	Count         int
}

func (m DocumentsModel) fetchDocs(pageToken string) tea.Cmd {
	client := m.client
	storeName := m.store.Name
	return func() tea.Msg {
		resp, err := client.ListDocuments(context.Background(), storeName, 20, pageToken)
		if err != nil {
			return ui.ErrMsg{Err: err}
		}
		return DocsLoadedMsg{
			Documents: resp.Documents,
			NextToken: resp.NextPageToken,
		}
	}
}
