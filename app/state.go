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

// TODO: ADD OTHER ACTIONS!!!
func (s *State) AllowEvent(event *aucoalesce.Event) error {
	switch helper.GetEventAction(event) {
	case helper.EventActionRead, helper.EventActionExec:
		s.AddROPaths(event.Data["path"])
	case helper.EventActionWrite:
		s.AddRWPaths(event.Data["path"])
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
