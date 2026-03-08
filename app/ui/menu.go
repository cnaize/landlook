package ui

import (
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/elastic/go-libaudit/v2/aucoalesce"
)

var _ tea.Model = (*Menu)(nil)

type Menu struct {
	List list.Model
	main lipgloss.Style
}

func NewMenu(events []*aucoalesce.Event) *Menu {
	items := make([]list.Item, len(events))
	for i, event := range events {
		items[i] = &MenuItem{
			Allow: false,
			Event: event,
		}
	}

	m := Menu{
		List: list.New(items, NewItemDelegate(), 0, 0),
		main: lipgloss.NewStyle().Margin(0),
	}
	m.List.Title = "Change Permissions"

	return &m
}

func (m *Menu) Init() tea.Cmd {
	return nil
}

func (m *Menu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "space":
			if m.List.FilterState() != list.Filtering {
				item := m.List.SelectedItem().(*MenuItem)
				item.Allow = !item.Allow
			}
		case "ctrl+c":
			return m, tea.Interrupt
		}
	case tea.WindowSizeMsg:
		h, v := m.main.GetFrameSize()
		m.List.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.List, cmd = m.List.Update(msg)

	return m, cmd
}

func (m *Menu) View() tea.View {
	view := tea.NewView(m.main.Render(m.List.View()))
	view.AltScreen = true

	return view
}
