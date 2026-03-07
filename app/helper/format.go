package helper

import (
	"fmt"

	"github.com/elastic/go-libaudit/v2/aucoalesce"
)

func FormatEvent(event *aucoalesce.Event) string {
	path := event.Data["path"]
	syscall := event.Data["syscall"]
	blockers := event.Data["blockers"]

	return fmt.Sprintf("[DENIED] %s (PID: %s) tried to %s %s (Reason: %s)",
		event.Process.Exe, event.Process.PID, SyscallToAction(syscall), path, blockers)
}

func FormatEventMenu(event *aucoalesce.Event) string {
	path := event.Data["path"]
	syscall := event.Data["syscall"]
	blockers := event.Data["blockers"]

	return fmt.Sprintf("%s (PID: %s) %s %s (%s)",
		event.Process.Exe, event.Process.PID, SyscallToAction(syscall), path, blockers)
}

func SyscallToAction(syscall string) string {
	switch syscall {
	case "openat", "open":
		return "read"
	case "mkdirat", "mkdir":
		return "create directory"
	case "connect":
		return "connect to network"
	}

	return "access"
}
