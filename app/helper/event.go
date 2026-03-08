package helper

import (
	"strconv"
	"strings"

	"github.com/elastic/go-libaudit/v2/aucoalesce"
)

type EventAction string

const (
	EventActionUnknown     EventAction = "unknown"
	EventActionRead        EventAction = "read"
	EventActionExec        EventAction = "exec"
	EventActionWrite       EventAction = "write"
	EventActionTCPListen   EventAction = "listen on"
	EventActionTCPConnect  EventAction = "connect to"
	EventActionMakeSockets EventAction = "create socket"
	EventActionSendSignals EventAction = "send signal"
)

func GetEventAction(event *aucoalesce.Event) EventAction {
	blockers := event.Data["blockers"]
	switch {
	case strings.Contains(blockers, "fs.write_file") || strings.Contains(blockers, "fs.make_"):
		return EventActionWrite
	case strings.Contains(blockers, "fs.execute"):
		return EventActionExec
	case strings.Contains(blockers, "fs.read_file") || strings.Contains(blockers, "fs.read_dir"):
		if event.Data["syscall"] == "openat" {
			// check flags
			flags, err := strconv.Atoi(event.Data["a2"])
			if err != nil {
				return EventActionUnknown
			}

			if flags&0x1 != 0 || // O_WRONLY
				flags&0x2 != 0 || // O_RDWR
				flags&0x40 != 0 || // O_CREAT
				flags&0x200 != 0 || // O_TRUNC
				flags&0x400 != 0 { // O_APPEND
				return EventActionWrite
			}
		}
		return EventActionRead
	}

	return EventActionUnknown
}

func CleanEvent(event *aucoalesce.Event) {
	event.Warnings = nil
}
