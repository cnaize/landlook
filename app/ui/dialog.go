package ui

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

var _ tea.Model = (*Dialog)(nil)

type Dialog struct {
	input textinput.Model
}

func NewDialog() *Dialog {
	d := &Dialog{
		input: textinput.New(),
	}
	d.input.Prompt = "\nDo you want to continue? [Y/n] "
	d.input.CharLimit = 3
	d.input.Focus()

	return d
}

func (d *Dialog) Init() tea.Cmd {
	return textinput.Blink
}

func (d *Dialog) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter":
			answer := strings.ToLower(d.input.Value())
			if answer == "" || answer == "y" || answer == "yes" {
				return d, tea.Quit
			} else {
				return d, tea.Interrupt
			}
		case "ctrl+c":
			return d, tea.Interrupt
		}
	}

	var cmd tea.Cmd
	d.input, cmd = d.input.Update(msg)
	return d, cmd
}

func (d *Dialog) View() tea.View {
	return tea.NewView(d.input.View())
}
