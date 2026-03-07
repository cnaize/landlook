package app

import (
	"slices"
	"strings"

	"github.com/cnaize/landbox"
)

type State struct {
	Command []string
	ROPaths landbox.Paths
	RWPaths landbox.Paths
	Options landbox.Options
	EnvVars []string
	Journal *Journal
}

func NewState() *State {
	return &State{}
}

func (s *State) AddROPaths(paths ...string) {
	s.ROPaths = slices.AppendSeq(s.ROPaths, func(yield func(string) bool) {
		for _, path := range paths {
			if len(path) < 1 {
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
			if len(path) < 1 {
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
