package helper

import (
	"fmt"

	"github.com/elastic/go-libaudit/v2/aucoalesce"
)

func FormatEvent(event *aucoalesce.Event) string {
	path := event.Data["path"]
	blockers := event.Data["blockers"]

	return fmt.Sprintf("[DENIED] %s (PID: %s) tried to %s %s (Reason: %s)",
		event.Process.Exe, event.Process.PID, GetEventAction(event), path, blockers)
}

func FormatEventMenu(event *aucoalesce.Event) string {
	path := event.Data["path"]
	blockers := event.Data["blockers"]

	return fmt.Sprintf("%s (PID: %s) %s %s (%s)",
		event.Process.Exe, event.Process.PID, GetEventAction(event), path, blockers)
}
