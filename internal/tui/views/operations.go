package views

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kopher1601/fs-cli/internal/api"
	"github.com/kopher1601/fs-cli/internal/model"
	"github.com/kopher1601/fs-cli/internal/tui/components"
	"github.com/kopher1601/fs-cli/internal/ui"
)

// OpUpdatedMsg is sent when an operation status changes.
type OpUpdatedMsg struct {
	Op *model.TrackedOp
}

// OpPollTickMsg triggers operation polling.
type OpPollTickMsg struct{}

// OperationTracker manages active operations.
type OperationTracker struct {
	mu         sync.RWMutex
	operations map[string]*model.TrackedOp
}

// NewOperationTracker creates a new tracker.
func NewOperationTracker() *OperationTracker {
	return &OperationTracker{
		operations: make(map[string]*model.TrackedOp),
	}
}

// Track adds an operation to track.
func (ot *OperationTracker) Track(op *api.Operation, storeName, opType string) {
	ot.mu.Lock()
	defer ot.mu.Unlock()
	ot.operations[op.Name] = &model.TrackedOp{
		Operation: op,
		StoreName: storeName,
		OpType:    opType,
	}
}

// Update updates an operation's status.
func (ot *OperationTracker) Update(op *api.Operation) {
	ot.mu.Lock()
	defer ot.mu.Unlock()
	if tracked, ok := ot.operations[op.Name]; ok {
		tracked.Operation = op
	}
}

// All returns all tracked operations.
func (ot *OperationTracker) All() []*model.TrackedOp {
	ot.mu.RLock()
	defer ot.mu.RUnlock()
	ops := make([]*model.TrackedOp, 0, len(ot.operations))
	for _, op := range ot.operations {
		ops = append(ops, op)
	}
	return ops
}

// ActiveCount returns the number of non-done operations.
func (ot *OperationTracker) ActiveCount() int {
	ot.mu.RLock()
	defer ot.mu.RUnlock()
	count := 0
	for _, op := range ot.operations {
		if !op.Operation.Done {
			count++
		}
	}
	return count
}

// OperationsModel is the view model for the operations list.
type OperationsModel struct {
	table   components.Table
	client  *api.Client
	keys    ui.KeyMap
	tracker *OperationTracker
	width   int
	height  int
}

// NewOperationsModel creates a new OperationsModel.
func NewOperationsModel(client *api.Client, keys ui.KeyMap, tracker *OperationTracker) OperationsModel {
	table := components.NewTable(model.OpColumns(), keys)
	return OperationsModel{
		table:   table,
		client:  client,
		keys:    keys,
		tracker: tracker,
	}
}

// SetSize updates the view dimensions.
func (m *OperationsModel) SetSize(w, h int) {
	m.width = w
	m.height = h
	m.table.SetSize(w, h)
}

// Refresh updates the table from the tracker.
func (m *OperationsModel) Refresh() {
	ops := m.tracker.All()
	rows := make([][]string, len(ops))
	for i, op := range ops {
		rows[i] = []string{
			model.ShortName(op.Operation.Name),
			model.ShortName(op.StoreName),
			op.OpType,
			model.OpStatus(op.Operation),
		}
	}
	m.table.SetRows(rows)
}

// PollActive polls all active operations and returns a command.
func (m OperationsModel) PollActive() tea.Cmd {
	ops := m.tracker.All()
	client := m.client
	tracker := m.tracker

	var activeOps []*model.TrackedOp
	for _, op := range ops {
		if !op.Operation.Done {
			activeOps = append(activeOps, op)
		}
	}

	if len(activeOps) == 0 {
		return nil
	}

	return func() tea.Msg {
		for _, tracked := range activeOps {
			updated, err := client.GetOperation(context.Background(), tracked.Operation.Name)
			if err != nil {
				continue
			}
			tracker.Update(updated)
		}
		return OpPollTickMsg{}
	}
}

// StartPolling begins periodic polling.
func (m OperationsModel) StartPolling() tea.Cmd {
	return tea.Tick(3*time.Second, func(time.Time) tea.Msg {
		return OpPollTickMsg{}
	})
}

// Update handles messages.
func (m OperationsModel) Update(msg tea.Msg) (OperationsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case OpPollTickMsg:
		m.Refresh()
		if m.tracker.ActiveCount() > 0 {
			return m, tea.Batch(m.PollActive(), m.StartPolling())
		}
		return m, nil

	case tea.KeyMsg:
		if key.Matches(msg, m.keys.Back) || key.Matches(msg, m.keys.Quit) {
			return m, func() tea.Msg { return ui.BackMsg{} }
		}
		m.table.Update(msg)
	}

	return m, nil
}

// View renders the operations table.
func (m OperationsModel) View() string {
	return m.table.View()
}

// Hints returns the key hints.
func (m OperationsModel) Hints() []components.KeyHint {
	count := m.tracker.ActiveCount()
	suffix := ""
	if count > 0 {
		suffix = " (" + strings.Repeat("●", count) + " active)"
	}
	return []components.KeyHint{
		{Key: "esc", Desc: "back" + suffix},
	}
}

// BreadcrumbItems returns the breadcrumb path.
func (m OperationsModel) BreadcrumbItems() []string {
	return []string{"Operations"}
}
