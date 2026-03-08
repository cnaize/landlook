package app

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/cnaize/landbox"
	"github.com/rs/zerolog"
	"github.com/urfave/cli/v3"

	"github.com/cnaize/landlook/app/journal"
	"github.com/cnaize/landlook/app/ui"
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
	logger zerolog.Logger
}

func NewApp(logger zerolog.Logger) *App {
	return &App{
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

	// check sudo
	if os.Getenv("SUDO_UID") == "" || os.Getenv("SUDO_GID") == "" {
		return fmt.Errorf("empty SUDO_UID/SUDO_GID, did you forget sudo?")
	}

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
	state := NewState()
	state.Command = args
	state.AddROPaths(strings.Split(cmd.String(AppFlagROPaths), ":")...)
	state.AddRWPaths(strings.Split(cmd.String(AppFlagRWPaths), ":")...)
	state.Options = landbox.Options{
		TCPListen:   cmd.Uint16Slice(AppFlagTCPListen),
		TCPConnect:  cmd.Uint16Slice(AppFlagTCPConnect),
		DenySockets: cmd.Bool(AppFlagDenySockets),
		DenySignals: cmd.Bool(AppFlagDenySignals),
		EnableDebug: true,
	}

	// add envs
	state.AddEnvVars(cmd.StringSlice(AppFlagAddEnvs)...)

	// add self
	if cmd.Bool(AppFlagAddSelf) {
		state.AddROPaths(binPath)
	}

	// add deps
	if cmd.Bool(AppFlagAddDeps) {
		deps, err := get.BinaryDeps(ctx, binPath)
		if err != nil {
			a.logger.Warn().Err(err).Msg("failed to add deps")
		} else {
			state.AddROPaths(deps...)
		}
	}

	return a.runLoop(ctx, state)
}

func (a *App) runLoop(ctx context.Context, state *State) error {
	for {
		// run command
		if err := a.run(ctx, state); err != nil {
			return fmt.Errorf("run command: %w", err)
		}

		// show dialog
		if _, err := tea.NewProgram(ui.NewDialog()).Run(); err != nil {
			return fmt.Errorf("run dialog: %w", err)
		}

		// show menu
		menu := ui.NewMenu(state.Journal.GetEvents())
		if _, err := tea.NewProgram(menu).Run(); err != nil {
			return fmt.Errorf("run menu: %w", err)
		}

		// handle state
		for _, item := range menu.List.Items() {
			item := item.(*ui.MenuItem)
			if !item.Allow {
				continue
			}

			if err := state.AllowEvent(item.Event); err != nil {
				a.logger.Err(err).Any("event", item.Event).Msg("failed to allow event")
				return fmt.Errorf("allow event: %w", err)
			}
		}
	}
}

func (a *App) run(ctx context.Context, state *State) error {
	sandbox := landbox.NewSandbox(state.ROPaths, state.RWPaths, &state.Options)
	defer sandbox.Close()

	// create command
	cmd := sandbox.CommandContext(ctx, state.Command[0], state.Command[1:]...)
	cmd.Env = append(cmd.Env, state.EnvVars...)

	// switch user
	uid, err := strconv.Atoi(os.Getenv("SUDO_UID"))
	if err != nil {
		return fmt.Errorf("parse SUDO_UID: %w", err)
	}
	gid, err := strconv.Atoi(os.Getenv("SUDO_GID"))
	if err != nil {
		return fmt.Errorf("parse SUDO_GID: %w", err)
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Credential: &syscall.Credential{Uid: uint32(uid), Gid: uint32(gid)},
	}

	// print debug
	a.logger.Debug().Strs("cmd", cmd.Args[1:]).Strs("env", cmd.Env).Msg("run command")

	// start journaling
	state.Journal = journal.NewJournal(a.logger)
	defer state.Journal.Stop()

	if err := state.Journal.Start(ctx); err != nil {
		return fmt.Errorf("journal start: %w", err)
	}

	// run command
	output, _ := cmd.CombinedOutput()
	fmt.Println(string(output))

	// wait a bit
	time.Sleep(time.Second)

	return nil
}
