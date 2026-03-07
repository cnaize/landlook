package menu

import (
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"github.com/elastic/go-libaudit/v2/aucoalesce"
)

var _ tea.Model = (*Menu)(nil)

type Menu struct {
	list list.Model
}

func NewMenu(events []*aucoalesce.Event) *Menu {
	items := make([]list.Item, len(events))
	for i, event := range events {
		items[i] = &MenuItem{
			allow: false,
			event: event,
		}
	}

	m := Menu{
		list: list.New(items, list.NewDefaultDelegate(), 0, 0),
	}
	m.list.Title = "Menu"

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
			item := m.list.SelectedItem().(*MenuItem)
			item.allow = !item.allow
		case "ctrl+c":
			return m, tea.Interrupt
		}
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)

	return m, cmd
}

func (m *Menu) View() tea.View {
	return tea.NewView(m.list.View())
}
