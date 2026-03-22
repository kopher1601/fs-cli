package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/kopher1601/fs-cli/internal/ui"
)

// Breadcrumb renders a navigation path.
type Breadcrumb struct {
	items []string
	width int
}

// NewBreadcrumb creates a new Breadcrumb.
func NewBreadcrumb() Breadcrumb {
	return Breadcrumb{}
}

// SetSize updates the breadcrumb width.
func (b *Breadcrumb) SetSize(w int) {
	b.width = w
}

// SetItems sets the breadcrumb path items.
func (b *Breadcrumb) SetItems(items []string) {
	b.items = items
}

// View renders the breadcrumb.
func (b Breadcrumb) View() string {
	if len(b.items) == 0 {
		return ""
	}

	appName := ui.StyleHeader.Render(" fs-cli ")
	sep := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#4B5563")).
		Render(" │ ")

	var pathParts []string
	for i, item := range b.items {
		if i == len(b.items)-1 {
			pathParts = append(pathParts, ui.StyleBreadcrumbActive.Render(item))
		} else {
			pathParts = append(pathParts, ui.StyleBreadcrumb.Render(item))
			pathParts = append(pathParts, lipgloss.NewStyle().
				Foreground(lipgloss.Color("#4B5563")).
				Render(" > "))
		}
	}

	left := appName + sep + strings.Join(pathParts, "")
	helpHint := ui.StyleKeyHint.Render("?") + ui.StyleKeyDesc.Render(":Help")

	leftLen := lipgloss.Width(left)
	rightLen := lipgloss.Width(helpHint)
	gap := b.width - leftLen - rightLen
	if gap < 1 {
		gap = 1
	}

	return left + strings.Repeat(" ", gap) + helpHint
}
