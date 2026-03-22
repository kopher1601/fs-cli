package views

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dustin/go-humanize"
	"github.com/kopher1601/fs-cli/internal/api"
	"github.com/kopher1601/fs-cli/internal/model"
	"github.com/kopher1601/fs-cli/internal/tui/components"
	"github.com/kopher1601/fs-cli/internal/ui"
)

// UploadStartedMsg is sent when an upload begins.
type UploadStartedMsg struct {
	Operation *api.Operation
}

// UploadProgressMsg reports upload progress.
type UploadProgressMsg struct {
	BytesSent  int64
	TotalBytes int64
}

// UploadCompleteMsg is sent when the upload HTTP request finishes.
type UploadCompleteMsg struct {
	Operation *api.Operation
	Err       error
}

const (
	fieldFilePath = iota
	fieldDisplayName
	fieldMaxTokens
	fieldOverlapTokens
	fieldCount
)

// UploadModel is the view model for the upload form.
type UploadModel struct {
	client    *api.Client
	keys      ui.KeyMap
	store     *api.FileSearchStore
	inputs    []textinput.Model
	focus     int
	uploading bool
	spinner   spinner.Model
	progress  float64 // 0.0 ~ 1.0
	fileSize  int64
	bytesSent int64
	fileName  string
	width     int
	height    int
}

// NewUploadModel creates a new UploadModel.
func NewUploadModel(client *api.Client, keys ui.KeyMap) UploadModel {
	inputs := make([]textinput.Model, fieldCount)

	inputs[fieldFilePath] = textinput.New()
	inputs[fieldFilePath].Placeholder = "/path/to/file"
	inputs[fieldFilePath].CharLimit = 512
	inputs[fieldFilePath].Width = 50

	inputs[fieldDisplayName] = textinput.New()
	inputs[fieldDisplayName].Placeholder = "(optional)"
	inputs[fieldDisplayName].CharLimit = 512
	inputs[fieldDisplayName].Width = 50

	inputs[fieldMaxTokens] = textinput.New()
	inputs[fieldMaxTokens].Placeholder = "800"
	inputs[fieldMaxTokens].CharLimit = 10
	inputs[fieldMaxTokens].Width = 20

	inputs[fieldOverlapTokens] = textinput.New()
	inputs[fieldOverlapTokens].Placeholder = "200"
	inputs[fieldOverlapTokens].CharLimit = 10
	inputs[fieldOverlapTokens].Width = 20

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(ui.ColorPrimary)

	return UploadModel{
		client:  client,
		keys:    keys,
		inputs:  inputs,
		spinner: s,
	}
}

// SetStore sets the target store and resets the form.
func (m *UploadModel) SetStore(store *api.FileSearchStore) {
	m.store = store
	m.focus = 0
	m.uploading = false
	m.progress = 0
	m.bytesSent = 0
	m.fileSize = 0
	m.fileName = ""
	for i := range m.inputs {
		m.inputs[i].Reset()
		m.inputs[i].Blur()
	}
	m.inputs[m.focus].Focus()
}

// SetSize updates the view dimensions.
func (m *UploadModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

// Update handles messages.
func (m UploadModel) Update(msg tea.Msg) (UploadModel, tea.Cmd) {
	switch msg := msg.(type) {
	case UploadProgressMsg:
		m.bytesSent = msg.BytesSent
		m.fileSize = msg.TotalBytes
		if msg.TotalBytes > 0 {
			m.progress = float64(msg.BytesSent) / float64(msg.TotalBytes)
		}
		return m, nil

	case UploadCompleteMsg:
		m.uploading = false
		if msg.Err != nil {
			return m, func() tea.Msg {
				return ui.ErrMsg{Err: msg.Err}
			}
		}
		op := msg.Operation
		return m, func() tea.Msg {
			return UploadStartedMsg{Operation: op}
		}

	case spinner.TickMsg:
		if m.uploading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil

	case tea.KeyMsg:
		// Uploading: only allow esc to cancel
		if m.uploading {
			return m, nil
		}

		switch {
		case msg.Type == tea.KeyTab || msg.Type == tea.KeyDown:
			m.inputs[m.focus].Blur()
			m.focus = (m.focus + 1) % fieldCount
			m.inputs[m.focus].Focus()
			return m, m.inputs[m.focus].Cursor.BlinkCmd()

		case msg.Type == tea.KeyShiftTab || msg.Type == tea.KeyUp:
			m.inputs[m.focus].Blur()
			m.focus = (m.focus - 1 + fieldCount) % fieldCount
			m.inputs[m.focus].Focus()
			return m, m.inputs[m.focus].Cursor.BlinkCmd()

		case msg.Type == tea.KeyEnter:
			filePath := m.inputs[fieldFilePath].Value()
			if filePath == "" {
				return m, func() tea.Msg {
					return ui.FlashMsg{Message: "File path is required", Level: ui.FlashWarn}
				}
			}
			// Validate file exists before starting
			expandedPath := api.ExpandHome(filePath)
			info, err := os.Stat(expandedPath)
			if err != nil {
				return m, func() tea.Msg {
					return ui.FlashMsg{Message: fmt.Sprintf("File not found: %s", filePath), Level: ui.FlashError}
				}
			}
			m.uploading = true
			m.progress = 0
			m.bytesSent = 0
			m.fileSize = info.Size()
			m.fileName = info.Name()
			return m, tea.Batch(m.spinner.Tick, m.startUpload())

		case key.Matches(msg, m.keys.Back):
			return m, func() tea.Msg { return ui.BackMsg{} }
		}

		var cmd tea.Cmd
		m.inputs[m.focus], cmd = m.inputs[m.focus].Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m UploadModel) startUpload() tea.Cmd {
	client := m.client
	storeName := m.store.Name
	filePath := m.inputs[fieldFilePath].Value()
	displayName := m.inputs[fieldDisplayName].Value()
	maxTokensStr := m.inputs[fieldMaxTokens].Value()
	overlapStr := m.inputs[fieldOverlapTokens].Value()

	return func() tea.Msg {
		config := &api.UploadConfig{
			DisplayName: displayName,
		}

		maxTokens, _ := strconv.Atoi(maxTokensStr)
		overlap, _ := strconv.Atoi(overlapStr)
		if maxTokens > 0 || overlap > 0 {
			config.ChunkingConfig = &api.ChunkingConfig{
				WhiteSpaceConfig: &api.WhiteSpaceConfig{
					MaxTokensPerChunk: maxTokens,
					MaxOverlapTokens:  overlap,
				},
			}
		}

		op, err := client.UploadFile(context.Background(), storeName, filePath, config)
		if err != nil {
			return UploadCompleteMsg{Err: err}
		}
		return UploadCompleteMsg{Operation: op}
	}
}

// View renders the upload form.
func (m UploadModel) View() string {
	storeName := "Store"
	if m.store != nil {
		storeName = m.store.DisplayName
		if storeName == "" {
			storeName = model.ShortName(m.store.Name)
		}
	}

	labels := []string{
		"  File Path:        ",
		"  Display Name:     ",
		"  Max Tokens/Chunk: ",
		"  Overlap Tokens:   ",
	}

	s := "\n"
	s += ui.StyleTitle.Render("  Upload to: "+storeName) + "\n\n"

	for i, input := range m.inputs {
		cursor := "  "
		if i == m.focus && !m.uploading {
			cursor = "► "
		}
		label := ui.StyleSubtle.Render(labels[i])
		s += cursor + label + input.View() + "\n"
	}

	s += "\n"

	if m.uploading {
		// Spinner + progress bar
		s += "  " + m.spinner.View() + " Uploading " + m.fileName + "...\n"
		s += "  " + m.renderProgressBar() + "\n"
		s += "  " + ui.StyleSubtle.Render(fmt.Sprintf(
			"%s / %s",
			humanize.IBytes(uint64(m.bytesSent)),
			humanize.IBytes(uint64(m.fileSize)),
		)) + "\n"
	} else {
		s += ui.StyleSubtle.Render("  tab:Next  enter:Upload  esc:Cancel") + "\n"
	}

	return s
}

func (m UploadModel) renderProgressBar() string {
	barWidth := 40
	if m.width > 0 && m.width-10 < barWidth {
		barWidth = m.width - 10
	}
	if barWidth < 10 {
		barWidth = 10
	}

	filled := int(m.progress * float64(barWidth))
	if filled > barWidth {
		filled = barWidth
	}

	pct := int(m.progress * 100)
	if pct > 100 {
		pct = 100
	}

	bar := lipgloss.NewStyle().Foreground(ui.ColorPrimary).Render(strings.Repeat("█", filled))
	empty := lipgloss.NewStyle().Foreground(ui.ColorBorder).Render(strings.Repeat("░", barWidth-filled))

	return fmt.Sprintf("%s%s %d%%", bar, empty, pct)
}

// Hints returns the key hints.
func (m UploadModel) Hints() []components.KeyHint {
	if m.uploading {
		return []components.KeyHint{
			{Key: "...", Desc: "uploading"},
		}
	}
	return []components.KeyHint{
		{Key: "tab", Desc: "next"},
		{Key: "enter", Desc: "upload"},
		{Key: "esc", Desc: "cancel"},
	}
}

// BreadcrumbItems returns the breadcrumb path.
func (m UploadModel) BreadcrumbItems() []string {
	storeName := "Store"
	if m.store != nil {
		storeName = m.store.DisplayName
		if storeName == "" {
			storeName = model.ShortName(m.store.Name)
		}
	}
	return []string{"Stores", storeName, "Upload"}
}
