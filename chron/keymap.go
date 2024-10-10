package chron

import (
	"github.com/charmbracelet/bubbles/key"
)

type chronDefaultKeyMap struct {
	Up           key.Binding
	Down         key.Binding
	Left         key.Binding
	Right        key.Binding
	SelectToggle key.Binding

	PageDown key.Binding
	PageUp   key.Binding

	// Filter allows the user to start typing and filter the rows.
	Filter key.Binding

	// FilterBlur is the key that stops the user's input from typing into the filter.
	FilterBlur key.Binding

	// FilterClear will clear the filter while it's blurred.
	FilterClear key.Binding

	AdditionalShortHelpKeys func() []key.Binding
	AdditionalFullHelpKeys  func() []key.Binding
}

func createChronDefaultKeyMap() chronDefaultKeyMap {
	return chronDefaultKeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "move up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "move down"),
		),
		Right: key.NewBinding(
			key.WithKeys("shift+right"),
			key.WithHelp("shift+→", "scroll right"),
		),
		Left: key.NewBinding(
			key.WithKeys("shift+left"),
			key.WithHelp("shift+←", "scroll left"),
		),

		SelectToggle: key.NewBinding(
			key.WithKeys(" ", "enter"),
			key.WithHelp("<space>/enter", "select row"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("ctrl+u", "pgup"),
			key.WithHelp("ctrl+u/page up", "Page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("ctrl+d", "pgdown"),
			key.WithHelp("ctrl+d/page down", "Page down"),
		),
		Filter: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "filter"),
		),
		FilterBlur: key.NewBinding(
			key.WithKeys("enter", "esc"),
			key.WithHelp("enter/esc", "unfocus"),
		),
		FilterClear: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "clear filter"),
		),
	}
}

// With supplying 'extraKeybindings' we can add extra bindings that we might need for that 'model'
func (km chronDefaultKeyMap) ShortHelp() []key.Binding {
	kb := []key.Binding{
		km.Up, km.Down, km.Left, km.Right, km.SelectToggle, km.PageDown, km.PageUp, km.Filter, km.FilterBlur, km.FilterClear,
	}
	if km.AdditionalShortHelpKeys != nil {
		kb = append(kb, km.AdditionalShortHelpKeys()...)
	}
	return kb
}

// TODO: Fix that we can supply more than just one row here
func (km chronDefaultKeyMap) FullHelp() [][]key.Binding {
	kb := [][]key.Binding{
		{km.Up, km.Down, km.Left, km.Right, km.SelectToggle},
		{km.PageDown, km.PageUp, km.Filter, km.FilterBlur, km.FilterClear},
	}
	if km.AdditionalFullHelpKeys != nil {
		kb = append(kb, km.AdditionalFullHelpKeys())
	}
	return kb

}
