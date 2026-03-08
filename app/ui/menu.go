package ui

import (
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"github.com/elastic/go-libaudit/v2/aucoalesce"
)

var _ tea.Model = (*Menu)(nil)

type Menu struct {
	List list.Model
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
	}
	m.List.SetShowTitle(false)

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
		m.List.SetSize(msg.Width, msg.Height)
	}

	var cmd tea.Cmd
	m.List, cmd = m.List.Update(msg)

	return m, cmd
}

func (m *Menu) View() tea.View {
	return tea.NewView(m.List.View())
}
