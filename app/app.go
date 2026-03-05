package app

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/cnaize/landbox"
	"github.com/rs/zerolog"
	"github.com/urfave/cli/v3"

	"github.com/cnaize/landlook/lib/get"
)

const (
	AppFlagROPaths     = "ro"
	AppFlagRWPaths     = "rw"
	AppFlagTCPListen   = "tcp-listen"
	AppFlagTCPConnect  = "tcp-connect"
	AppFlagDenySockets = "sockets"
	AppFlagDenySignals = "signals"
	AppFlagAddEnvs     = "env"
	AppFlagAddSelf     = "add-self"
	AppFlagAddDeps     = "add-deps"
)

type App struct {
	state  *State
	logger zerolog.Logger
}

func NewApp(logger zerolog.Logger) *App {
	return &App{
		state:  NewState(),
		logger: logger,
	}
}

func (a *App) RunLoop(ctx context.Context, cmd *cli.Command) error {
	args := cmd.Args().Slice()
	if len(args) < 1 {
		return fmt.Errorf("missing command to run")
	}

	binPath, err := exec.LookPath(args[0])
	if err != nil {
		return fmt.Errorf("find binary: %w", err)
	}

	// fill state
	a.state.Command = args
	a.state.AddROPaths(strings.Split(cmd.String(AppFlagROPaths), ":")...)
	a.state.AddRWPaths(strings.Split(cmd.String(AppFlagRWPaths), ":")...)
	a.state.Options = landbox.Options{
		TCPListen:   cmd.Uint16Slice(AppFlagTCPListen),
		TCPConnect:  cmd.Uint16Slice(AppFlagTCPConnect),
		DenySockets: cmd.Bool(AppFlagDenySockets),
		DenySignals: cmd.Bool(AppFlagDenySignals),
		EnableDebug: true,
	}

	// add self
	if cmd.Bool(AppFlagAddSelf) {
		a.state.AddROPaths(binPath)
	}

	// add deps
	if cmd.Bool(AppFlagAddDeps) {
		deps, err := get.BinaryDeps(ctx, binPath)
		if err != nil {
			return fmt.Errorf("get binary deps: %w", err)
		}

		a.state.AddROPaths(deps...)
	}

	return a.run(ctx, a.state)
}

func (a *App) run(ctx context.Context, state *State) error {
	sandbox := landbox.NewSandbox(state.ROPaths, state.RWPaths, &state.Options)
	defer sandbox.Close()

	output, _ := sandbox.CommandContext(ctx, state.Command[0], state.Command[1:]...).CombinedOutput()
	a.logger.Info().Msg(string(output))

	return nil
}
