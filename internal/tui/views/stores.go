package views

import (
	"context"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kopher1601/fs-cli/internal/api"
	"github.com/kopher1601/fs-cli/internal/model"
	"github.com/kopher1601/fs-cli/internal/tui/components"
	"github.com/kopher1601/fs-cli/internal/ui"
)

// StoresLoadedMsg is sent when stores are fetched.
type StoresLoadedMsg struct {
	Stores    []api.FileSearchStore
	NextToken string
}

// StoreDeletedMsg is sent after a store is deleted.
type StoreDeletedMsg struct{}

// StoresModel is the view model for the stores list.
type StoresModel struct {
	table     components.Table
	client    *api.Client
	keys      ui.KeyMap
	stores    []api.FileSearchStore
	pageToken string
	nextToken string
	loading   bool
	width     int
	height    int
}

// NewStoresModel creates a new StoresModel.
func NewStoresModel(client *api.Client, keys ui.KeyMap) StoresModel {
	table := components.NewTable(model.StoreColumns(), keys)
	return StoresModel{
		table:  table,
		client: client,
		keys:   keys,
	}
}

// Init returns the initial command to load stores.
func (m StoresModel) Init() tea.Cmd {
	return m.fetchStores("")
}

// SetSize updates the view dimensions.
func (m *StoresModel) SetSize(w, h int) {
	m.width = w
	m.height = h
	m.table.SetSize(w, h)
}

// SelectedStore returns the currently selected store.
func (m *StoresModel) SelectedStore() *api.FileSearchStore {
	idx := m.table.SelectedIndex()
	if idx < 0 || idx >= len(m.stores) {
		return nil
	}
	return &m.stores[idx]
}

// SetFilter applies a text filter to the table.
func (m *StoresModel) SetFilter(f string) {
	m.table.SetFilter(f)
}

// Update handles messages.
func (m StoresModel) Update(msg tea.Msg) (StoresModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case StoresLoadedMsg:
		m.loading = false
		m.stores = msg.Stores
		m.nextToken = msg.NextToken
		rows := make([][]string, len(m.stores))
		for i, s := range m.stores {
			rows[i] = model.StoreRowToStrings(model.StoreToRow(&s))
		}
		m.table.SetRows(rows)

	case StoreDeletedMsg:
		return m, m.fetchStores(m.pageToken)

	case ui.RefreshMsg:
		return m, m.fetchStores(m.pageToken)

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Enter):
			if s := m.SelectedStore(); s != nil {
				return m, func() tea.Msg {
					return ui.NavigateMsg{View: ui.ViewDocuments, Store: s}
				}
			}

		case key.Matches(msg, m.keys.Detail):
			if s := m.SelectedStore(); s != nil {
				return m, func() tea.Msg {
					return ui.NavigateMsg{View: ui.ViewStoreDetail, Store: s}
				}
			}

		case key.Matches(msg, m.keys.Delete):
			if s := m.SelectedStore(); s != nil {
				return m, func() tea.Msg {
					return ConfirmDeleteMsg{
						ResourceName: s.Name,
						DisplayName:  s.DisplayName,
						Force:        false,
					}
				}
			}

		case key.Matches(msg, m.keys.ForceDel):
			if s := m.SelectedStore(); s != nil {
				return m, func() tea.Msg {
					return ConfirmDeleteMsg{
						ResourceName: s.Name,
						DisplayName:  s.DisplayName,
						Force:        true,
					}
				}
			}

		case key.Matches(msg, m.keys.Create):
			return m, func() tea.Msg {
				return ShowCreateStoreMsg{}
			}

		case key.Matches(msg, m.keys.NextPage):
			if m.nextToken != "" {
				m.pageToken = m.nextToken
				m.loading = true
				return m, m.fetchStores(m.nextToken)
			}

		case key.Matches(msg, m.keys.PrevPage):
			if m.pageToken != "" {
				m.pageToken = ""
				m.loading = true
				return m, m.fetchStores("")
			}

		default:
			m.table.Update(msg)
		}
	}

	return m, tea.Batch(cmds...)
}

// View renders the stores table.
func (m StoresModel) View() string {
	return m.table.View()
}

// Hints returns the key hints for the status bar.
func (m StoresModel) Hints() []components.KeyHint {
	return []components.KeyHint{
		{Key: "c", Desc: "create"},
		{Key: "d", Desc: "delete"},
		{Key: "enter", Desc: "documents"},
		{Key: "y", Desc: "detail"},
		{Key: "/", Desc: "filter"},
		{Key: "?", Desc: "help"},
	}
}

// BreadcrumbItems returns the breadcrumb path.
func (m StoresModel) BreadcrumbItems() []string {
	return []string{"Stores"}
}

func (m StoresModel) fetchStores(pageToken string) tea.Cmd {
	client := m.client
	return func() tea.Msg {
		resp, err := client.ListStores(context.Background(), 20, pageToken)
		if err != nil {
			return ui.ErrMsg{Err: err}
		}
		return StoresLoadedMsg{
			Stores:    resp.FileSearchStores,
			NextToken: resp.NextPageToken,
		}
	}
}

// ConfirmDeleteMsg is sent when a delete needs confirmation.
type ConfirmDeleteMsg struct {
	ResourceName string
	DisplayName  string
	Force        bool
	IsDocument   bool
}

// ShowCreateStoreMsg is sent when the user wants to create a store.
type ShowCreateStoreMsg struct{}
