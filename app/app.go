package app

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/cnaize/landbox"
	"github.com/rs/zerolog"
	"github.com/urfave/cli/v3"

	"github.com/cnaize/landlook/lib/get"
)

const (
	AppFlagLogLevel    = "log-level"
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

func (a *App) Run(ctx context.Context, cmd *cli.Command) error {
	// set log level
	logLevel, err := zerolog.ParseLevel(cmd.String(AppFlagLogLevel))
	if err != nil {
		a.logger.Warn().Str(AppFlagLogLevel, cmd.String(AppFlagLogLevel)).Msg(`invalid log level, setting to "debug"`)
		logLevel = zerolog.DebugLevel
	}
	a.logger = a.logger.Level(logLevel)

	// check app args
	args := cmd.Args().Slice()
	if len(args) < 1 {
		return fmt.Errorf("missing command to run")
	}

	// get binary path
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

	// add envs
	a.state.AddEnvVars(cmd.StringSlice(AppFlagAddEnvs)...)

	// add self
	if cmd.Bool(AppFlagAddSelf) {
		a.state.AddROPaths(binPath)
	}

	// add deps
	if cmd.Bool(AppFlagAddDeps) {
		deps, err := get.BinaryDeps(ctx, binPath)
		if err != nil {
			a.logger.Warn().Err(err).Msg("failed to add deps")
		} else {
			a.state.AddROPaths(deps...)
		}
	}

	return a.run(ctx, a.state)
}

func (a *App) run(ctx context.Context, state *State) error {
	sandbox := landbox.NewSandbox(state.ROPaths, state.RWPaths, &state.Options)
	defer sandbox.Close()

	// create command
	cmd := sandbox.CommandContext(ctx, state.Command[0], state.Command[1:]...)
	cmd.Env = append(cmd.Env, a.state.EnvVars...)

	// print debug
	a.logger.Debug().Strs("cmd", cmd.Args[1:]).Strs("env", cmd.Env).Msg("run command")

	// start journaling
	state.Journal = NewJournal(a.logger)
	defer state.Journal.Stop()

	if err := state.Journal.Start(ctx); err != nil {
		return fmt.Errorf("journal start: %w", err)
	}

	// run command
	output, _ := cmd.CombinedOutput()
	a.logger.Info().Msg(string(output))

	// wait a bit
	time.Sleep(time.Second)

	return nil
}
