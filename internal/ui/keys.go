package ui

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines the global key bindings.
type KeyMap struct {
	Up        key.Binding
	Down      key.Binding
	Top       key.Binding
	Bottom    key.Binding
	Enter     key.Binding
	Back      key.Binding
	Quit      key.Binding
	ForceQuit key.Binding
	Refresh   key.Binding
	Filter    key.Binding
	Help      key.Binding
	NextPage  key.Binding
	PrevPage  key.Binding
	Create    key.Binding
	Delete    key.Binding
	ForceDel  key.Binding
	Detail    key.Binding
	Upload    key.Binding
	Ops key.Binding
}

// DefaultKeyMap returns the default key bindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Top: key.NewBinding(
			key.WithKeys("g"),
			key.WithHelp("g", "top"),
		),
		Bottom: key.NewBinding(
			key.WithKeys("G"),
			key.WithHelp("G", "bottom"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q"),
			key.WithHelp("q", "quit"),
		),
		ForceQuit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "force quit"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("ctrl+r"),
			key.WithHelp("ctrl+r", "refresh"),
		),
		Filter: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "filter"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		NextPage: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "next page"),
		),
		PrevPage: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "prev page"),
		),
		Create: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "create"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete"),
		),
		ForceDel: key.NewBinding(
			key.WithKeys("D"),
			key.WithHelp("D", "force delete"),
		),
		Detail: key.NewBinding(
			key.WithKeys("y"),
			key.WithHelp("y", "detail"),
		),
		Upload: key.NewBinding(
			key.WithKeys("u"),
			key.WithHelp("u", "upload"),
		),
		Ops: key.NewBinding(
			key.WithKeys("o"),
			key.WithHelp("o", "operations"),
		),
	}
}
