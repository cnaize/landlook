package app

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/cnaize/landbox"
	"github.com/elastic/go-libaudit/v2/aucoalesce"

	"github.com/cnaize/landlook/app/helper"
	"github.com/cnaize/landlook/app/journal"
)

type State struct {
	Command []string
	ROPaths landbox.Paths
	RWPaths landbox.Paths
	Options landbox.Options
	EnvVars []string
	Journal *journal.Journal
}

func NewState() *State {
	return &State{}
}

func (s *State) AllowEvent(event *aucoalesce.Event) error {
	action, target := helper.GetEventActionTarget(event)
	switch action {
	case helper.EventActionExec, helper.EventActionReadDir, helper.EventActionReadFile:
		s.AddROPaths(target)
	case helper.EventActionWriteDir, helper.EventActionWriteFile:
		s.AddRWPaths(target)
	case helper.EventActionTCPListen:
		s.AddTCPListenPorts(target)
	case helper.EventActionTCPConnect:
		s.AddTCPConnectPorts(target)
	case helper.EventActionOpenSockets:
		s.Options.DenySockets = false
	case helper.EventActionSendSignals:
		s.Options.DenySignals = false
	default:
		return fmt.Errorf("invalid action: %s", helper.GetEventAction(event))
	}

	return nil
}

func (s *State) AddROPaths(paths ...string) {
	s.ROPaths = slices.AppendSeq(s.ROPaths, func(yield func(string) bool) {
		for _, path := range paths {
			if path == "" {
				continue
			}

			if !yield(path) {
				return
			}
		}
	})
}

func (s *State) AddRWPaths(paths ...string) {
	s.RWPaths = slices.AppendSeq(s.RWPaths, func(yield func(string) bool) {
		for _, path := range paths {
			if path == "" {
				continue
			}

			if !yield(path) {
				return
			}
		}
	})
}

func (s *State) AddTCPListenPorts(ports ...string) {
	s.Options.TCPListen = slices.AppendSeq(s.Options.TCPListen, func(yield func(uint16) bool) {
		for _, port := range ports {
			if port, err := strconv.Atoi(port); err == nil {
				if !yield(uint16(port)) {
					return
				}
			}
		}
	})
}

func (s *State) AddTCPConnectPorts(ports ...string) {
	s.Options.TCPConnect = slices.AppendSeq(s.Options.TCPConnect, func(yield func(uint16) bool) {
		for _, port := range ports {
			if port, err := strconv.Atoi(port); err == nil {
				if !yield(uint16(port)) {
					return
				}
			}
		}
	})
}

func (s *State) AddEnvVars(envs ...string) {
	s.EnvVars = slices.AppendSeq(s.EnvVars, func(yield func(string) bool) {
		for _, env := range envs {
			if !strings.Contains(env, "=") {
				continue
			}

			if !yield(env) {
				return
			}
		}
	})
}
