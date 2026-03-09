package helper

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/elastic/go-libaudit/v2/aucoalesce"
)

type EventAction string

const (
	EventActionUnknown     EventAction = "unknown"
	EventActionExec        EventAction = "exec"
	EventActionReadDir     EventAction = "read dir"
	EventActionReadFile    EventAction = "read file"
	EventActionWriteDir    EventAction = "write dir"
	EventActionWriteFile   EventAction = "write file"
	EventActionTCPListen   EventAction = "listen on"
	EventActionTCPConnect  EventAction = "connect to"
	EventActionOpenSockets EventAction = "open socket"
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
	case strings.Contains(blockers, "net.bind_tcp"):
		return EventActionTCPListen
	case strings.Contains(blockers, "net.connect_tcp"):
		return EventActionTCPConnect
	case strings.Contains(blockers, "scope.abstract_unix_socket"):
		return EventActionOpenSockets
	case strings.Contains(blockers, "scope.signal"):
		return EventActionSendSignals
	case strings.Contains(blockers, "fs.write_file"),
		strings.Contains(blockers, "fs.truncate"),
		strings.Contains(blockers, "fs.remove_file"),
		strings.Contains(blockers, "fs.ioctl_dev"):
		return EventActionWriteFile
	case strings.Contains(blockers, "fs.make_"),
		strings.Contains(blockers, "fs.remove_dir"),
		strings.Contains(blockers, "fs.refer"):
		return EventActionWriteDir
	case strings.Contains(blockers, "fs.execute"):
		return EventActionExec
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
	}

	return EventActionUnknown
}

func GetEventActionTarget(event *aucoalesce.Event) (EventAction, string) {
	action := GetEventAction(event)
	switch action {
	case EventActionExec, EventActionReadDir, EventActionReadFile, EventActionWriteDir, EventActionWriteFile:
		return action, event.Data["path"]
	case EventActionTCPListen:
		return action, event.Data["src"]
	case EventActionTCPConnect:
		return action, event.Data["dest"]
	case EventActionOpenSockets:
		path, err := hex.DecodeString(event.Data["path"])
		if err != nil {
			return action, event.Data["path"]
		}
		if len(path) > 0 && path[0] == 0 {
			path = fmt.Appendf(nil, "@%s", path)
		}
		return action, string(path)
	case EventActionSendSignals:
		return action, event.Data["syscall"]
	}

	return EventActionUnknown, event.Data["syscall"]
}

func FormatEventLog(event *aucoalesce.Event) string {
	action, target := GetEventActionTarget(event)
	switch action {
	case EventActionTCPListen:
		if event.Data["saddr"] != "" {
			target = fmt.Sprintf("%s:%s", event.Data["saddr"], target)
		} else {
			target = fmt.Sprintf("%s port", target)
		}
	case EventActionTCPConnect:
		target = fmt.Sprintf("%s:%s", event.Data["daddr"], event.Data["dest"])
	}

	return fmt.Sprintf("[DENIED] %s (PID: %s) tried to %s %s (Reason: %s)",
		event.Process.Exe, event.Process.PID, action, target, event.Data["blockers"])
}

func FormatEventMenu(event *aucoalesce.Event) string {
	action, target := GetEventActionTarget(event)
	switch action {
	case EventActionTCPListen:
		target = fmt.Sprintf("%s port", target)
	case EventActionTCPConnect:
		target = fmt.Sprintf("%s port", target)
	}

	return fmt.Sprintf("%s (PID: %s) %s %s (%s)",
		event.Process.Exe, event.Process.PID, action, target, event.Data["blockers"])
}

func CleanEvent(event *aucoalesce.Event) {
	event.Warnings = nil
}
