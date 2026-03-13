package style

import "github.com/charmbracelet/bubbles/key"

type GlobalKeyMap struct {
	Quit key.Binding
}

var GlobalKeys = GlobalKeyMap{
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c"),
	),
}
