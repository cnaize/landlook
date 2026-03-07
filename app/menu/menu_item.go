package menu

import (
	"charm.land/bubbles/v2/list"
	"github.com/elastic/go-libaudit/v2/aucoalesce"

	"github.com/cnaize/landlook/app/helper"
)

var _ list.Item = (*MenuItem)(nil)

type MenuItem struct {
	allow bool
	event *aucoalesce.Event
}

func NewMenuItem(allow bool, event *aucoalesce.Event) *MenuItem {
	return &MenuItem{
		allow: allow,
		event: event,
	}
}

func (i *MenuItem) Title() string {
	if i.allow {
		return "[ALLOW]"
	}

	return "[DENY]"
}

func (i *MenuItem) Description() string {
	return helper.FormatEventMenu(i.event)

}

func (i *MenuItem) FilterValue() string {
	return i.Title() + i.Description()
}
