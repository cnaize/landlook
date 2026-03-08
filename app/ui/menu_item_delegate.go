package ui

import (
	"fmt"
	"io"

	"charm.land/bubbles/v2/list"
	"charm.land/lipgloss/v2"
)

var _ list.ItemDelegate = (*ItemDelegate)(nil)

type ItemDelegate struct {
	list.DefaultDelegate
}

func NewItemDelegate() *ItemDelegate {
	return &ItemDelegate{
		DefaultDelegate: list.NewDefaultDelegate(),
	}
}

func (d *ItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item := listItem.(*MenuItem)

	titleStyle, descStyle := d.Styles.NormalTitle, d.Styles.NormalDesc
	if m.Index() == index {
		titleStyle, descStyle = d.Styles.SelectedTitle, d.Styles.SelectedDesc
	}

	titleStyle = titleStyle.Foreground(lipgloss.Red)
	if item.Allow {
		titleStyle = titleStyle.Foreground(lipgloss.Green)
	}

	fmt.Fprintf(w, "%s\n%s", titleStyle.Render(item.Title()), descStyle.Render(item.Description()))
}
