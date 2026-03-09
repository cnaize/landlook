package helper

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/elastic/go-libaudit/v2/aucoalesce"
)

type EventAction string

const (
	EventActionUnknown     EventAction = "unknown"
	EventActionExec        EventAction = "exec"
	EventActionReadDir     EventAction = "read directory"
	EventActionReadFile    EventAction = "read file"
	EventActionWriteDir    EventAction = "write directory"
	EventActionWriteFile   EventAction = "write file"
	EventActionTCPListen   EventAction = "listen on"
	EventActionTCPConnect  EventAction = "connect to"
	EventActionMakeSockets EventAction = "create socket"
	EventActionSendSignals EventAction = "send signal"
)

func GetEventAction(event *aucoalesce.Event) EventAction {
	checkWriteFlags := func(flags string) bool {
		bits, err := strconv.ParseUint(strings.TrimPrefix(flags, "0x"), 16, 64)
		if err != nil {
			return false
		}

		if bits&0x3 != 0 || // O_ACCMODE
			bits&0x40 != 0 || // O_CREAT
			bits&0x200 != 0 || // O_TRUNC
			bits&0x400 != 0 { // O_APPEND
			return true
		}

		return false
	}

	sysCall := event.Data["syscall"]
	blockers := event.Data["blockers"]
	switch {
	case strings.Contains(blockers, "fs.execute"):
		return EventActionExec
	case strings.Contains(blockers, "fs.write_file") || strings.Contains(blockers, "fs.truncate"):
		return EventActionWriteFile
	case strings.Contains(blockers, "fs.make_"):
		return EventActionWriteDir
	case strings.Contains(blockers, "fs.read_file"):
		if (sysCall == "open" || sysCall == "openat") && checkWriteFlags(event.Data["a2"]) {
			return EventActionWriteFile
		}
		return EventActionReadFile
	case strings.Contains(blockers, "fs.read_dir"):
		if (sysCall == "open" || sysCall == "openat") && checkWriteFlags(event.Data["a2"]) {
			return EventActionWriteDir
		}
		return EventActionReadDir
	case strings.Contains(blockers, "net.bind_tcp"):
		return EventActionTCPListen
	case strings.Contains(blockers, "net.connect_tcp"):
		return EventActionTCPConnect
	case sysCall == "socket":
		return EventActionMakeSockets
	case sysCall == "kill" || sysCall == "tkill" || sysCall == "tgkill":
		return EventActionSendSignals
	}

	return EventActionUnknown
}

func CleanEvent(event *aucoalesce.Event) {
	event.Warnings = nil
}

func FormatEventLog(event *aucoalesce.Event) string {
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
