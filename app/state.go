package app

import (
	"fmt"
	"slices"
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
	switch helper.GetEventAction(event) {
	case helper.EventActionExec, helper.EventActionReadDir, helper.EventActionReadFile:
		s.AddROPaths(event.Data["path"])
	case helper.EventActionWriteDir, helper.EventActionWriteFile:
		s.AddRWPaths(event.Data["path"])
	case helper.EventActionTCPListen:
		// TODO: add tcp listen port here
	case helper.EventActionTCPConnect:
		// TODO: add tcp connect port here
	case helper.EventActionMakeSockets:
		s.Options.DenySockets = false
	case helper.EventActionSendSignals:
		s.Options.DenySignals = false
	default:
		return fmt.Errorf("unknown action: %s", helper.GetEventAction(event))
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
