package components

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kopher1601/fs-cli/internal/ui"
)

// Table is a custom table component with selection, sorting and filtering.
type Table struct {
	columns  []string
	rows     [][]string
	cursor   int
	offset   int
	width    int
	height   int
	filter   string
	filtered []int          // indices into rows that match filter
	selected map[int]bool   // multi-select: original row indices
	keys     ui.KeyMap
}

// NewTable creates a new Table with the given columns.
func NewTable(columns []string, keys ui.KeyMap) Table {
	return Table{
		columns: columns,
		keys:    keys,
	}
}

// SetSize updates the table dimensions.
func (t *Table) SetSize(w, h int) {
	t.width = w
	t.height = h
}

// SetRows replaces the table data and clears selections.
func (t *Table) SetRows(rows [][]string) {
	t.rows = rows
	t.selected = nil
	t.applyFilter()
	if t.cursor >= len(t.visibleRows()) {
		t.cursor = max(0, len(t.visibleRows())-1)
	}
}

// ToggleSelect toggles the selection of the current row.
func (t *Table) ToggleSelect() {
	idx := t.SelectedIndex()
	if idx < 0 {
		return
	}
	if t.selected == nil {
		t.selected = make(map[int]bool)
	}
	if t.selected[idx] {
		delete(t.selected, idx)
	} else {
		t.selected[idx] = true
	}
}

// SelectedIndices returns all multi-selected row indices.
// If nothing is selected, returns nil.
func (t *Table) SelectedIndices() []int {
	if len(t.selected) == 0 {
		return nil
	}
	indices := make([]int, 0, len(t.selected))
	for idx := range t.selected {
		indices = append(indices, idx)
	}
	return indices
}

// SelectedCount returns the number of selected rows.
func (t *Table) SelectedCount() int {
	return len(t.selected)
}

// ClearSelection clears all multi-selections.
func (t *Table) ClearSelection() {
	t.selected = nil
}

// SelectedIndex returns the original row index of the currently selected row.
func (t *Table) SelectedIndex() int {
	vis := t.visibleRows()
	if t.cursor < 0 || t.cursor >= len(vis) {
		return -1
	}
	return vis[t.cursor]
}

// SelectedRow returns the currently selected row data.
func (t *Table) SelectedRow() []string {
	idx := t.SelectedIndex()
	if idx < 0 || idx >= len(t.rows) {
		return nil
	}
	return t.rows[idx]
}

// RowCount returns the number of visible rows.
func (t *Table) RowCount() int {
	return len(t.visibleRows())
}

// SetFilter sets the filter string.
func (t *Table) SetFilter(f string) {
	t.filter = f
	t.applyFilter()
	t.cursor = 0
	t.offset = 0
}

func (t *Table) applyFilter() {
	if t.filter == "" {
		t.filtered = nil
		return
	}
	lower := strings.ToLower(t.filter)
	t.filtered = nil
	for i, row := range t.rows {
		for _, cell := range row {
			if strings.Contains(strings.ToLower(cell), lower) {
				t.filtered = append(t.filtered, i)
				break
			}
		}
	}
}

func (t *Table) visibleRows() []int {
	if t.filtered != nil {
		return t.filtered
	}
	indices := make([]int, len(t.rows))
	for i := range t.rows {
		indices[i] = i
	}
	return indices
}

// Update handles key events for table navigation.
func (t *Table) Update(msg tea.KeyMsg) {
	vis := t.visibleRows()
	count := len(vis)
	if count == 0 {
		return
	}

	switch {
	case key.Matches(msg, t.keys.Up):
		if t.cursor > 0 {
			t.cursor--
		}
	case key.Matches(msg, t.keys.Down):
		if t.cursor < count-1 {
			t.cursor++
		}
	case key.Matches(msg, t.keys.Top):
		t.cursor = 0
	case key.Matches(msg, t.keys.Bottom):
		t.cursor = count - 1
	case msg.String() == " ":
		t.ToggleSelect()
		// Move cursor down after toggle
		if t.cursor < count-1 {
			t.cursor++
		}
	}

	// Adjust scroll offset
	maxVisible := t.maxVisibleRows()
	if t.cursor < t.offset {
		t.offset = t.cursor
	}
	if t.cursor >= t.offset+maxVisible {
		t.offset = t.cursor - maxVisible + 1
	}
}

func (t *Table) maxVisibleRows() int {
	if t.height <= 2 {
		return 1
	}
	return t.height - 2
}

// View renders the table.
func (t *Table) View() string {
	if t.width == 0 {
		return ""
	}

	vis := t.visibleRows()
	colWidths := t.calcColumnWidths()

	var b strings.Builder

	// Header
	header := t.renderRow(t.columns, colWidths, true, -1, false)
	b.WriteString(header)
	b.WriteString("\n")

	// Separator
	sep := strings.Repeat("─", t.width)
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#374151")).Render(sep))
	b.WriteString("\n")

	// Rows
	maxRows := t.maxVisibleRows()
	if len(vis) == 0 {
		empty := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			Italic(true).
			Render("  No items found")
		b.WriteString(empty)
		b.WriteString("\n")
	} else {
		for i := t.offset; i < len(vis) && i < t.offset+maxRows; i++ {
			row := t.rows[vis[i]]
			isCursor := i == t.cursor
			line := t.renderRow(row, colWidths, false, vis[i], isCursor)
			b.WriteString(line)
			b.WriteString("\n")
		}
	}

	// Fill remaining lines
	rendered := len(vis) - t.offset
	if rendered > maxRows {
		rendered = maxRows
	}
	if rendered < 0 {
		rendered = 0
	}
	for i := rendered; i < maxRows; i++ {
		b.WriteString("\n")
	}

	return b.String()
}

func (t *Table) renderRow(cells []string, widths []int, isHeader bool, rowIdx int, isCursor bool) string {
	var parts []string
	for i, cell := range cells {
		if i >= len(widths) {
			break
		}
		w := widths[i]
		truncated := truncate(cell, w)
		padded := pad(truncated, w)
		parts = append(parts, padded)
	}

	line := "  " + strings.Join(parts, "  ")

	if isHeader {
		return ui.StyleTableHeader.Width(t.width).Render(line)
	}

	isChecked := t.selected != nil && t.selected[rowIdx]

	// Prefix: ► for cursor, ✓ for selected, space for neither
	prefix := "  "
	if isCursor && isChecked {
		prefix = "►✓"
	} else if isCursor {
		prefix = "► "
	} else if isChecked {
		prefix = " ✓"
	}
	line = prefix + line[2:]

	if isCursor {
		return ui.StyleTableRowSelected.Width(t.width).Render(line)
	}
	if isChecked {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A78BFA")).
			Width(t.width).
			Render(line)
	}
	return ui.StyleTableRow.Width(t.width).Render(line)
}

// displayWidth returns the visual width of a string, accounting for
// wide characters (CJK, emoji, etc.).
func displayWidth(s string) int {
	return lipgloss.Width(s)
}

func (t *Table) calcColumnWidths() []int {
	if len(t.columns) == 0 {
		return nil
	}

	available := t.width - 2 - (len(t.columns)-1)*2
	if available < len(t.columns) {
		available = len(t.columns)
	}

	maxWidths := make([]int, len(t.columns))
	for i, col := range t.columns {
		maxWidths[i] = displayWidth(col)
	}
	for _, row := range t.rows {
		for i, cell := range row {
			if i < len(maxWidths) {
				w := displayWidth(cell)
				if w > maxWidths[i] {
					maxWidths[i] = w
				}
			}
		}
	}

	total := 0
	for _, w := range maxWidths {
		total += w
	}

	widths := make([]int, len(t.columns))
	if total == 0 {
		each := available / len(t.columns)
		for i := range widths {
			widths[i] = each
		}
		return widths
	}

	remaining := available
	for i, mw := range maxWidths {
		w := (mw * available) / total
		if w < 4 {
			w = 4
		}
		if w > mw {
			w = mw
		}
		widths[i] = w
		remaining -= w
	}

	if remaining > 0 && len(widths) > 1 {
		widths[1] += remaining
	}

	return widths
}

// truncate cuts a string to fit within maxW display columns.
func truncate(s string, maxW int) string {
	if displayWidth(s) <= maxW {
		return s
	}
	if maxW <= 3 {
		// Just take runes until we fit
		w := 0
		for i, r := range s {
			rw := displayWidth(string(r))
			if w+rw > maxW {
				return s[:i]
			}
			w += rw
		}
		return s
	}
	// Reserve 3 columns for "..."
	target := maxW - 3
	w := 0
	for i, r := range s {
		rw := displayWidth(string(r))
		if w+rw > target {
			return s[:i] + "..."
		}
		w += rw
	}
	return s
}

// pad adds spaces to reach w display columns.
func pad(s string, w int) string {
	dw := displayWidth(s)
	if dw >= w {
		return s
	}
	return s + strings.Repeat(" ", w-dw)
}
