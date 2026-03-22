package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kopher1601/fs-cli/internal/api"
	"github.com/kopher1601/fs-cli/internal/tui/components"
	"github.com/kopher1601/fs-cli/internal/tui/views"
	"github.com/kopher1601/fs-cli/internal/ui"
)

// tickMsg is sent on auto-refresh intervals.
type tickMsg time.Time

// deleteProgressMsg reports one item deleted in a batch.
type deleteProgressMsg struct {
	remaining []string
	force     bool
	err       error
}

// deleteCompleteMsg is sent when batch delete is done.
type deleteCompleteMsg struct{}

// navEntry stores a view and its associated context for back navigation.
type navEntry struct {
	view  ui.ViewType
	store *api.FileSearchStore
	doc   *api.Document
}

// App is the root bubbletea model.
type App struct {
	client     *api.Client
	keys       ui.KeyMap
	activeView ui.ViewType
	navStack   []navEntry

	// Context
	currentStore *api.FileSearchStore
	currentDoc   *api.Document

	// Views
	storesView      views.StoresModel
	docsView        views.DocumentsModel
	storeDetailView views.StoreDetailModel
	docDetailView   views.DocDetailModel
	uploadView      views.UploadModel
	opsView         views.OperationsModel
	helpView        views.HelpModel

	// Operation tracker
	opTracker *views.OperationTracker

	// Components
	breadcrumb components.Breadcrumb
	statusBar  components.StatusBar
	flash      components.Flash
	confirm    components.Confirm

	// Filter
	filterInput textinput.Model
	filtering   bool

	// Create store
	createInput textinput.Model
	creating    bool

	// Pending batch delete
	pendingBatch *views.BatchDeleteMsg

	// Delete progress
	deleting     bool
	deleteTotal  int
	deleteDone   int
	deleteFailed int

	// Dimensions
	width  int
	height int
}

// NewApp creates a new App model.
func NewApp(client *api.Client) App {
	keys := ui.DefaultKeyMap()

	fi := textinput.New()
	fi.Placeholder = "Filter..."
	fi.CharLimit = 100

	ci := textinput.New()
	ci.Placeholder = "Store display name"
	ci.CharLimit = 512

	tracker := views.NewOperationTracker()

	return App{
		client:          client,
		keys:            keys,
		activeView:      ui.ViewStores,
		storesView:      views.NewStoresModel(client, keys),
		docsView:        views.NewDocumentsModel(client, keys),
		storeDetailView: views.NewStoreDetailModel(keys),
		docDetailView:   views.NewDocDetailModel(keys),
		uploadView:      views.NewUploadModel(client, keys),
		opsView:         views.NewOperationsModel(client, keys, tracker),
		helpView:        views.NewHelpModel(keys),
		opTracker:       tracker,
		breadcrumb:      components.NewBreadcrumb(),
		statusBar:       components.NewStatusBar(),
		flash:           components.NewFlash(),
		confirm:         components.NewConfirm(),
		filterInput:     fi,
		createInput:     ci,
	}
}

// Init returns the initial commands.
func (a App) Init() tea.Cmd {
	return tea.Batch(
		a.storesView.Init(),
		a.tickCmd(),
	)
}

func (a App) tickCmd() tea.Cmd {
	return tea.Tick(10*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Update handles all messages.
func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.updateSizes()
		return a, nil

	case tea.KeyMsg:
		if key.Matches(msg, a.keys.ForceQuit) {
			return a, tea.Quit
		}

		if a.confirm.IsVisible() {
			cmd := a.confirm.Update(msg)
			return a, cmd
		}

		if a.filtering {
			return a.updateFilter(msg)
		}

		if a.creating {
			return a.updateCreate(msg)
		}

		// Upload 뷰는 텍스트 입력이 있으므로 글로벌 키를 우회
		if a.activeView == ui.ViewUpload {
			return a.updateActiveView(msg)
		}

		switch {
		case key.Matches(msg, a.keys.Quit):
			if len(a.navStack) > 0 {
				return a.navigateBack()
			}
			return a, tea.Quit

		case key.Matches(msg, a.keys.Back):
			if len(a.navStack) > 0 {
				return a.navigateBack()
			}

		case key.Matches(msg, a.keys.Filter):
			a.filtering = true
			a.filterInput.Reset()
			a.filterInput.Focus()
			return a, a.filterInput.Cursor.BlinkCmd()

		case key.Matches(msg, a.keys.Refresh):
			return a, func() tea.Msg { return ui.RefreshMsg{} }

		case key.Matches(msg, a.keys.Ops):
			return a, func() tea.Msg {
				return ui.NavigateMsg{View: ui.ViewOperations}
			}

		case key.Matches(msg, a.keys.Help):
			return a, func() tea.Msg {
				return ui.NavigateMsg{View: ui.ViewHelp}
			}
		}

		return a.updateActiveView(msg)

	case tickMsg:
		cmds = append(cmds, a.tickCmd())
		switch a.activeView {
		case ui.ViewStores:
			var cmd tea.Cmd
			a.storesView, cmd = a.storesView.Update(ui.RefreshMsg{})
			cmds = append(cmds, cmd)
		case ui.ViewDocuments:
			var cmd tea.Cmd
			a.docsView, cmd = a.docsView.Update(ui.RefreshMsg{})
			cmds = append(cmds, cmd)
		}
		return a, tea.Batch(cmds...)

	case ui.NavigateMsg:
		return a.handleNavigate(msg)

	case ui.BackMsg:
		return a.navigateBack()

	case ui.ErrMsg:
		cmd := a.flash.Show(msg.Err.Error(), ui.FlashError)
		return a, cmd

	case ui.FlashMsg:
		cmd := a.flash.Show(msg.Message, msg.Level)
		// Also refresh data after successful operations
		var refreshCmd tea.Cmd
		switch a.activeView {
		case ui.ViewStores:
			a.storesView, refreshCmd = a.storesView.Update(ui.RefreshMsg{})
		case ui.ViewDocuments:
			a.docsView, refreshCmd = a.docsView.Update(ui.RefreshMsg{})
		}
		return a, tea.Batch(cmd, refreshCmd)

	case views.BatchDeleteMsg:
		label := fmt.Sprintf("Delete %d documents", msg.Count)
		a.confirm.Show(label, "", msg.Force)
		a.pendingBatch = &msg
		return a, nil

	case deleteProgressMsg:
		a.deleteDone++
		if msg.err != nil {
			a.deleteFailed++
		}
		flashCmd := a.flash.Show(
			fmt.Sprintf("Deleting... %d/%d", a.deleteDone, a.deleteTotal),
			ui.FlashInfo,
		)
		nextCmd := a.deleteNext(msg.remaining, msg.force)
		return a, tea.Batch(flashCmd, nextCmd)

	case deleteCompleteMsg:
		a.deleting = false
		total := a.deleteTotal
		failed := a.deleteFailed
		var cmd tea.Cmd
		if failed > 0 {
			cmd = a.flash.Show(
				fmt.Sprintf("Deleted %d/%d documents (%d failed)", total-failed, total, failed),
				ui.FlashWarn,
			)
		} else {
			cmd = a.flash.Show(
				fmt.Sprintf("Deleted %d documents", total),
				ui.FlashSuccess,
			)
		}
		var refreshCmd tea.Cmd
		a.docsView, refreshCmd = a.docsView.Update(ui.RefreshMsg{})
		return a, tea.Batch(cmd, refreshCmd)

	case views.ConfirmDeleteMsg:
		a.confirm.Show(msg.DisplayName, msg.ResourceName, msg.Force)
		return a, nil

	case components.ConfirmResult:
		if msg.Confirmed {
			// Batch delete
			if a.pendingBatch != nil {
				batch := a.pendingBatch
				a.pendingBatch = nil
				return a, a.startBatchDelete(batch.ResourceNames, batch.Force)
			}
			return a, a.executeDelete(msg.ResourceName, msg.Force)
		}
		a.pendingBatch = nil
		cmd := a.flash.Show("Cancelled", ui.FlashInfo)
		return a, cmd

	case views.StoreDeletedMsg:
		cmd := a.flash.Show("Store deleted", ui.FlashSuccess)
		var refreshCmd tea.Cmd
		a.storesView, refreshCmd = a.storesView.Update(ui.RefreshMsg{})
		return a, tea.Batch(cmd, refreshCmd)

	case views.DocDeletedMsg:
		cmd := a.flash.Show("Document deleted", ui.FlashSuccess)
		var refreshCmd tea.Cmd
		a.docsView, refreshCmd = a.docsView.Update(ui.RefreshMsg{})
		// Navigate back from detail to documents list
		if a.activeView == ui.ViewDocDetail {
			a2, backCmd := a.navigateBack()
			return a2, tea.Batch(cmd, refreshCmd, backCmd)
		}
		return a, tea.Batch(cmd, refreshCmd)

	case views.ShowCreateStoreMsg:
		a.creating = true
		a.createInput.Reset()
		a.createInput.Focus()
		return a, a.createInput.Cursor.BlinkCmd()

	case views.UploadStartedMsg:
		if msg.Operation != nil {
			storeName := ""
			if a.currentStore != nil {
				storeName = a.currentStore.Name
			}
			a.opTracker.Track(msg.Operation, storeName, "Upload")
			cmd := a.flash.Show("Upload started, tracking operation...", ui.FlashSuccess)
			a2, backCmd := a.navigateBack()
			pollCmd := a.opsView.StartPolling()
			return a2, tea.Batch(cmd, backCmd, pollCmd)
		}
		return a, nil

	case views.OpPollTickMsg:
		var cmd tea.Cmd
		a.opsView, cmd = a.opsView.Update(msg)
		// Check if any completed operations and notify
		return a, cmd

	default:
		a.flash.Update(msg)
		return a.updateActiveView(msg)
	}
}

func (a App) updateActiveView(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch a.activeView {
	case ui.ViewStores:
		var cmd tea.Cmd
		a.storesView, cmd = a.storesView.Update(msg)
		return a, cmd
	case ui.ViewDocuments:
		var cmd tea.Cmd
		a.docsView, cmd = a.docsView.Update(msg)
		return a, cmd
	case ui.ViewStoreDetail:
		var cmd tea.Cmd
		a.storeDetailView, cmd = a.storeDetailView.Update(msg)
		return a, cmd
	case ui.ViewDocDetail:
		var cmd tea.Cmd
		a.docDetailView, cmd = a.docDetailView.Update(msg)
		return a, cmd
	case ui.ViewUpload:
		var cmd tea.Cmd
		a.uploadView, cmd = a.uploadView.Update(msg)
		return a, cmd
	case ui.ViewOperations:
		var cmd tea.Cmd
		a.opsView, cmd = a.opsView.Update(msg)
		return a, cmd
	case ui.ViewHelp:
		var cmd tea.Cmd
		a.helpView, cmd = a.helpView.Update(msg)
		return a, cmd
	}
	return a, nil
}

func (a App) updateFilter(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		a.filtering = false
		a.filterInput.Blur()
		a.applyFilter(a.filterInput.Value())
		return a, nil
	case tea.KeyEsc:
		a.filtering = false
		a.filterInput.Blur()
		a.filterInput.Reset()
		a.applyFilter("")
		return a, nil
	}

	var cmd tea.Cmd
	a.filterInput, cmd = a.filterInput.Update(msg)
	a.applyFilter(a.filterInput.Value())
	return a, cmd
}

func (a *App) applyFilter(f string) {
	switch a.activeView {
	case ui.ViewStores:
		a.storesView.SetFilter(f)
	case ui.ViewDocuments:
		a.docsView.SetFilter(f)
	}
}

func (a App) updateCreate(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		name := a.createInput.Value()
		a.creating = false
		a.createInput.Blur()
		if name == "" {
			cmd := a.flash.Show("Store name cannot be empty", ui.FlashWarn)
			return a, cmd
		}
		return a, a.createStore(name)
	case tea.KeyEsc:
		a.creating = false
		a.createInput.Blur()
		return a, nil
	}

	var cmd tea.Cmd
	a.createInput, cmd = a.createInput.Update(msg)
	return a, cmd
}

func (a App) handleNavigate(msg ui.NavigateMsg) (tea.Model, tea.Cmd) {
	// Push current state
	a.navStack = append(a.navStack, navEntry{
		view:  a.activeView,
		store: a.currentStore,
		doc:   a.currentDoc,
	})

	a.activeView = msg.View

	var cmd tea.Cmd
	switch msg.View {
	case ui.ViewDocuments:
		if msg.Store != nil {
			a.currentStore = msg.Store
			cmd = a.docsView.SetStore(msg.Store)
			a.docsView.SetSize(a.width, a.contentHeight())
		}
	case ui.ViewStoreDetail:
		if msg.Store != nil {
			a.currentStore = msg.Store
			a.storeDetailView.SetStore(msg.Store)
			a.storeDetailView.SetSize(a.width, a.contentHeight())
		}
	case ui.ViewDocDetail:
		if msg.Store != nil && msg.Doc != nil {
			a.currentStore = msg.Store
			a.currentDoc = msg.Doc
			a.docDetailView.SetDoc(msg.Store, msg.Doc)
			a.docDetailView.SetSize(a.width, a.contentHeight())
		}
	case ui.ViewUpload:
		if msg.Store != nil {
			a.currentStore = msg.Store
			a.uploadView.SetStore(msg.Store)
			a.uploadView.SetSize(a.width, a.contentHeight())
		}
	case ui.ViewOperations:
		a.opsView.SetSize(a.width, a.contentHeight())
		a.opsView.Refresh()
		if a.opTracker.ActiveCount() > 0 {
			cmd = a.opsView.StartPolling()
		}
	case ui.ViewHelp:
		a.helpView.SetSize(a.width, a.contentHeight())
	}

	return a, cmd
}

func (a App) navigateBack() (tea.Model, tea.Cmd) {
	if len(a.navStack) == 0 {
		return a, nil
	}
	prev := a.navStack[len(a.navStack)-1]
	a.navStack = a.navStack[:len(a.navStack)-1]
	a.activeView = prev.view
	a.currentStore = prev.store
	a.currentDoc = prev.doc
	return a, nil
}

func (a App) contentHeight() int {
	h := a.height - 4
	if h < 1 {
		h = 1
	}
	return h
}

func (a *App) updateSizes() {
	h := a.contentHeight()

	a.breadcrumb.SetSize(a.width)
	a.statusBar.SetSize(a.width)
	a.flash.SetSize(a.width)
	a.confirm.SetSize(a.width)

	a.storesView.SetSize(a.width, h)
	a.docsView.SetSize(a.width, h)
	a.storeDetailView.SetSize(a.width, h)
	a.docDetailView.SetSize(a.width, h)
	a.uploadView.SetSize(a.width, h)
	a.opsView.SetSize(a.width, h)
	a.helpView.SetSize(a.width, h)
}

func (a *App) executeDelete(resourceName string, force bool) tea.Cmd {
	a.flash.Show("Deleting...", ui.FlashInfo)
	client := a.client
	isDoc := strings.Contains(resourceName, "/documents/")
	return func() tea.Msg {
		var err error
		if isDoc {
			err = client.DeleteDocument(context.Background(), resourceName, force)
		} else {
			err = client.DeleteStore(context.Background(), resourceName, force)
		}
		if err != nil {
			return ui.ErrMsg{Err: err}
		}
		if isDoc {
			return views.DocDeletedMsg{}
		}
		return views.StoreDeletedMsg{}
	}
}

func (a *App) startBatchDelete(names []string, force bool) tea.Cmd {
	a.deleting = true
	a.deleteTotal = len(names)
	a.deleteDone = 0
	a.deleteFailed = 0
	return a.deleteNext(names, force)
}

func (a App) deleteNext(names []string, force bool) tea.Cmd {
	if len(names) == 0 {
		return func() tea.Msg { return deleteCompleteMsg{} }
	}
	client := a.client
	name := names[0]
	rest := names[1:]
	return func() tea.Msg {
		err := client.DeleteDocument(context.Background(), name, force)
		return deleteProgressMsg{remaining: rest, force: force, err: err}
	}
}

func (a *App) createStore(displayName string) tea.Cmd {
	a.flash.Show("Creating store...", ui.FlashInfo)
	client := a.client
	return func() tea.Msg {
		_, err := client.CreateStore(context.Background(), displayName)
		if err != nil {
			return ui.ErrMsg{Err: err}
		}
		return ui.FlashMsg{Message: fmt.Sprintf("Store '%s' created", displayName), Level: ui.FlashSuccess}
	}
}

// View renders the full application.
func (a App) View() string {
	if a.width == 0 {
		return "Loading..."
	}

	var sections []string

	// Breadcrumb
	a.updateBreadcrumb()
	sections = append(sections, a.breadcrumb.View())

	// Separator
	sep := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#374151")).
		Render(strings.Repeat("─", a.width))
	sections = append(sections, sep)

	// Content
	switch a.activeView {
	case ui.ViewStores:
		sections = append(sections, a.storesView.View())
	case ui.ViewDocuments:
		sections = append(sections, a.docsView.View())
	case ui.ViewStoreDetail:
		sections = append(sections, a.storeDetailView.View())
	case ui.ViewDocDetail:
		sections = append(sections, a.docDetailView.View())
	case ui.ViewUpload:
		sections = append(sections, a.uploadView.View())
	case ui.ViewOperations:
		sections = append(sections, a.opsView.View())
	case ui.ViewHelp:
		sections = append(sections, a.helpView.View())
	default:
		sections = append(sections, "  View not implemented yet")
	}

	// Confirm dialog or status bar
	if a.confirm.IsVisible() {
		sections = append(sections, a.confirm.View())
	} else if a.filtering {
		filterLine := ui.StyleKeyHint.Render("/") + ui.StyleKeyDesc.Render(" ") + a.filterInput.View()
		sections = append(sections, filterLine)
	} else if a.creating {
		createLine := ui.StyleKeyHint.Render("New Store: ") + a.createInput.View()
		sections = append(sections, createLine)
	} else {
		a.updateStatusBar()
		sections = append(sections, a.statusBar.View())
	}

	// Flash
	if fv := a.flash.View(); fv != "" {
		sections = append(sections, fv)
	}

	return strings.Join(sections, "\n")
}

func (a *App) updateBreadcrumb() {
	switch a.activeView {
	case ui.ViewStores:
		a.breadcrumb.SetItems(a.storesView.BreadcrumbItems())
	case ui.ViewDocuments:
		a.breadcrumb.SetItems(a.docsView.BreadcrumbItems())
	case ui.ViewStoreDetail:
		a.breadcrumb.SetItems(a.storeDetailView.BreadcrumbItems())
	case ui.ViewDocDetail:
		a.breadcrumb.SetItems(a.docDetailView.BreadcrumbItems())
	case ui.ViewUpload:
		a.breadcrumb.SetItems(a.uploadView.BreadcrumbItems())
	case ui.ViewOperations:
		a.breadcrumb.SetItems(a.opsView.BreadcrumbItems())
	case ui.ViewHelp:
		a.breadcrumb.SetItems(a.helpView.BreadcrumbItems())
	default:
		a.breadcrumb.SetItems([]string{"Unknown"})
	}
}

func (a *App) updateStatusBar() {
	switch a.activeView {
	case ui.ViewStores:
		a.statusBar.SetHints(a.storesView.Hints())
	case ui.ViewDocuments:
		a.statusBar.SetHints(a.docsView.Hints())
	case ui.ViewStoreDetail:
		a.statusBar.SetHints(a.storeDetailView.Hints())
	case ui.ViewDocDetail:
		a.statusBar.SetHints(a.docDetailView.Hints())
	case ui.ViewUpload:
		a.statusBar.SetHints(a.uploadView.Hints())
	case ui.ViewOperations:
		a.statusBar.SetHints(a.opsView.Hints())
	case ui.ViewHelp:
		a.statusBar.SetHints(a.helpView.Hints())
	}
}
