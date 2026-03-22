package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	ColorPrimary   = lipgloss.Color("#7C3AED")
	ColorSecondary = lipgloss.Color("#6B7280")
	ColorSuccess   = lipgloss.Color("#10B981")
	ColorWarning   = lipgloss.Color("#F59E0B")
	ColorDanger    = lipgloss.Color("#EF4444")
	ColorInfo      = lipgloss.Color("#3B82F6")
	ColorMuted     = lipgloss.Color("#6B7280")
	ColorBorder    = lipgloss.Color("#374151")
	ColorHighlight = lipgloss.Color("#7C3AED")

	// Header / Breadcrumb
	StyleHeader = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(ColorPrimary).
			Padding(0, 1)

	StyleBreadcrumb = lipgloss.NewStyle().
			Foreground(ColorMuted)

	StyleBreadcrumbActive = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Bold(true)

	// Table
	StyleTableHeader = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorPrimary).
				BorderBottom(true).
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(ColorBorder)

	StyleTableRow = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#D1D5DB"))

	StyleTableRowSelected = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(ColorHighlight).
				Bold(true)

	// Status bar
	StyleStatusBar = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Background(lipgloss.Color("#1F2937")).
			Padding(0, 1)

	StyleKeyHint = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true)

	StyleKeyDesc = lipgloss.NewStyle().
			Foreground(ColorSecondary)

	// Flash messages
	StyleFlashInfo = lipgloss.NewStyle().
			Foreground(ColorInfo).
			Padding(0, 1)

	StyleFlashWarn = lipgloss.NewStyle().
			Foreground(ColorWarning).
			Padding(0, 1)

	StyleFlashError = lipgloss.NewStyle().
			Foreground(ColorDanger).
			Bold(true).
			Padding(0, 1)

	StyleFlashSuccess = lipgloss.NewStyle().
				Foreground(ColorSuccess).
				Padding(0, 1)

	// Document state colors
	StyleStateActive  = lipgloss.NewStyle().Foreground(ColorSuccess)
	StyleStatePending = lipgloss.NewStyle().Foreground(ColorWarning)
	StyleStateFailed  = lipgloss.NewStyle().Foreground(ColorDanger)

	// General
	StyleTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF"))

	StyleSubtle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	StyleBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder)
)
