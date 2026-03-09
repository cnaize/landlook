package ui

import (
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/elastic/go-libaudit/v2/aucoalesce"
)

var _ tea.Model = (*Menu)(nil)

type Menu struct {
	List list.Model
	main lipgloss.Style
	desc lipgloss.Style
	keys map[string]key.Binding
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
		desc: NewItemDelegate().Styles.SelectedDesc.Margin(0),
		keys: map[string]key.Binding{
			"cancel": key.NewBinding(
				key.WithKeys("esc"),
			),
			"toggle": key.NewBinding(
				key.WithKeys("space"),
				key.WithHelp("space", "toggle"),
			),
			"run": key.NewBinding(
				key.WithKeys("r"),
				key.WithHelp("r", "run"),
			),
			"quit": key.NewBinding(
				key.WithKeys("q", "ctrl+c"),
			),
		},
	}
	m.List.Title = "Change Permissions"
	m.List.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			m.keys["toggle"],
			m.keys["run"],
		}
	}
	m.List.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			m.keys["toggle"],
			m.keys["run"],
		}
	}

	return &m
}

func (m *Menu) Init() tea.Cmd {
	m.List.KeyMap.Quit.SetEnabled(false)

	return nil
}

func (m *Menu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var oldWidth int
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if m.List.FilterState() == list.Filtering {
			break
		}

		switch {
		case key.Matches(msg, m.keys["cancel"]) && m.List.FilterState() == list.Unfiltered:
			return m, cmd
		case key.Matches(msg, m.keys["toggle"]):
			item := m.List.SelectedItem().(*MenuItem)
			item.Allow = !item.Allow
		case key.Matches(msg, m.keys["run"]):
			return m, tea.Quit
		case key.Matches(msg, m.keys["quit"]):
			return m, tea.Interrupt
		}
	case tea.WindowSizeMsg:
		h, v := m.main.GetFrameSize()
		desc := m.desc.Width(msg.Width - h).Render(m.List.SelectedItem().(*MenuItem).Description())

		oldWidth = m.List.Width()
		m.List.SetSize(msg.Width-h, msg.Height-v-m.desc.GetVerticalFrameSize()-lipgloss.Height(desc))
	}

	oldItem := m.List.SelectedItem().(*MenuItem)
	m.List, cmd = m.List.Update(msg)
	newItem := m.List.SelectedItem().(*MenuItem)

	if oldWidth != 0 {
		oldDesc := m.desc.Width(oldWidth).Render(oldItem.Description())
		newDesc := m.desc.Width(m.List.Width()).Render(newItem.Description())

		m.List.SetHeight(m.List.Height() + lipgloss.Height(oldDesc) - lipgloss.Height(newDesc))
	} else if oldItem != newItem {
		oldDesc := m.desc.Width(m.List.Width()).Render(oldItem.Description())
		newDesc := m.desc.Width(m.List.Width()).Render(newItem.Description())

		m.List.SetHeight(m.List.Height() + lipgloss.Height(oldDesc) - (lipgloss.Height(newDesc)))
	}

	return m, cmd
}

func (m *Menu) View() tea.View {
	item := m.List.SelectedItem().(*MenuItem)
	desc := m.desc.Width(m.List.Width() - m.desc.GetHorizontalFrameSize()).Render(item.Description())

	view := tea.NewView(m.main.Render(lipgloss.JoinVertical(lipgloss.Left, m.List.View(), desc)))
	view.AltScreen = true

	return view
}
