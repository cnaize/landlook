package app

import (
	"fmt"

	"github.com/elastic/go-libaudit/aucoalesce"
)

func FormatEvent(event *aucoalesce.Event) string {
	path := event.Data["path"]
	syscall := event.Data["syscall"]
	blockers := event.Data["blockers"]

	action := "access"
	switch syscall {
	case "openat", "open":
		action = "read"
	case "mkdirat", "mkdir":
		action = "create directory"
	case "connect":
		action = "connect to network"
	}

	return fmt.Sprintf("[DENIED] %s (PID: %s) tried to %s %s (Reason: %s)",
		event.Process.Exe, event.Process.PID, action, path, blockers)
}

func CleanEvent(event *aucoalesce.Event) *aucoalesce.Event {
	event.Warnings = nil

	return event
}
