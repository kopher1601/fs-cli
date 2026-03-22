package components

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ConfirmResult is sent when the user answers the confirmation.
type ConfirmResult struct {
	Confirmed    bool
	ResourceName string
	Force        bool
}

// Confirm is a confirmation dialog component.
type Confirm struct {
	message      string
	resourceName string
	force        bool
	visible      bool
	width        int
}

// NewConfirm creates a new Confirm dialog.
func NewConfirm() Confirm {
	return Confirm{}
}

// SetSize updates the dialog width.
func (c *Confirm) SetSize(w int) {
	c.width = w
}

// Show displays the confirmation dialog.
func (c *Confirm) Show(displayName, resourceName string, force bool) {
	action := "Delete"
	if force {
		action = "Force delete"
	}
	c.message = fmt.Sprintf("%s '%s'? (y/N)", action, displayName)
	c.resourceName = resourceName
	c.force = force
	c.visible = true
}

// IsVisible returns whether the dialog is showing.
func (c *Confirm) IsVisible() bool {
	return c.visible
}

// Hide hides the dialog.
func (c *Confirm) Hide() {
	c.visible = false
}

// Update handles key events.
func (c *Confirm) Update(msg tea.KeyMsg) tea.Cmd {
	if !c.visible {
		return nil
	}

	yesKey := key.NewBinding(key.WithKeys("y", "Y"))
	noKey := key.NewBinding(key.WithKeys("n", "N", "esc"))

	switch {
	case key.Matches(msg, yesKey):
		c.visible = false
		name := c.resourceName
		force := c.force
		return func() tea.Msg {
			return ConfirmResult{Confirmed: true, ResourceName: name, Force: force}
		}
	case key.Matches(msg, noKey):
		c.visible = false
		return func() tea.Msg {
			return ConfirmResult{Confirmed: false}
		}
	}
	return nil
}

// View renders the dialog.
func (c Confirm) View() string {
	if !c.visible {
		return ""
	}

	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F59E0B")).
		Bold(true).
		Padding(0, 1)

	return style.Width(c.width).Render("⚠ " + c.message)
}
