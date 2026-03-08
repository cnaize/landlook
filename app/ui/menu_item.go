package ui

import (
	"fmt"

	"charm.land/bubbles/v2/list"
	"github.com/elastic/go-libaudit/v2/aucoalesce"

	"github.com/cnaize/landlook/app/helper"
)

var _ list.Item = (*MenuItem)(nil)

type MenuItem struct {
	Allow bool
	Event *aucoalesce.Event
}

func NewMenuItem(allow bool, event *aucoalesce.Event) *MenuItem {
	return &MenuItem{
		Allow: allow,
		Event: event,
	}
}

func (i *MenuItem) Title() string {
	if i.Allow {
		return "[ALLOW]"
	}

	return "[DENY]"
}

func (i *MenuItem) Description() string {
	return helper.FormatEventMenu(i.Event)

}

func (i *MenuItem) FilterValue() string {
	return fmt.Sprintf("%s %s", i.Title(), i.Description())
}
