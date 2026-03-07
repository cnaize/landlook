package helper

import "github.com/elastic/go-libaudit/v2/aucoalesce"

func CleanEvent(event *aucoalesce.Event) {
	event.Warnings = nil
}
