package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/appleboy/graceful"
	"github.com/rs/zerolog"
	"github.com/urfave/cli/v3"

	"github.com/cnaize/landlook/app"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger()

	cli := &cli.Command{
		Name:      "landlook",
		Usage:     "landlock security policy generator",
		ArgsUsage: "application [arguments]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        app.AppFlagLogLevel,
				Usage:       "set zerolog level",
				Value:       "error",
				DefaultText: "error",
			},
			&cli.StringFlag{
				Name:        app.AppFlagOutput,
				Aliases:     []string{"o"},
				Usage:       "output file",
				Value:       "landlook.json",
				DefaultText: "landlook.json",
			},
			&cli.StringSliceFlag{
				Name:        app.AppFlagROPaths,
				Usage:       "allow read/exec path",
				DefaultText: "deny all",
			},
			&cli.StringSliceFlag{
				Name:        app.AppFlagRWPaths,
				Usage:       "allow read/exec/write path",
				DefaultText: "deny all",
			},
			&cli.Uint16SliceFlag{
				Name:        app.AppFlagTCPListen,
				Aliases:     []string{"l"},
				Usage:       "allow listen tcp port",
				DefaultText: "deny all",
			},
			&cli.Uint16SliceFlag{
				Name:        app.AppFlagTCPConnect,
				Aliases:     []string{"c"},
				Usage:       "allow connect tcp port",
				DefaultText: "deny all",
			},
			&cli.BoolFlag{
				Name:        app.AppFlagAllowSockets,
				Usage:       "allow open abstract sockets",
				Value:       false,
				DefaultText: "deny",
			},
			&cli.BoolFlag{
				Name:        app.AppFlagAllowSignals,
				Usage:       "allow send signals",
				Value:       false,
				DefaultText: "deny",
			},
			&cli.StringSliceFlag{
				Name:        app.AppFlagAddEnvs,
				Aliases:     []string{"e"},
				Usage:       "add environment variable",
				DefaultText: "empty list",
			},
			&cli.BoolFlag{
				Name:        app.AppFlagAddSelf,
				Usage:       fmt.Sprintf("add application itself to --%s", app.AppFlagROPaths),
				Value:       true,
				DefaultText: "true",
			},
			&cli.BoolFlag{
				Name:        app.AppFlagAddDeps,
				Usage:       fmt.Sprintf("add application dependencies to --%s", app.AppFlagROPaths),
				Value:       true,
				DefaultText: "true",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return app.NewApp(logger).Run(ctx, cmd)
		},
	}

	m := graceful.NewManagerWithContext(ctx, graceful.WithLogger(graceful.NewEmptyLogger()))
	m.AddRunningJob(func(ctx context.Context) error {
		defer cancel()

		return cli.Run(ctx, os.Args)
	})

	<-m.Done()

	for _, err := range m.Errors() {
		if !errors.Is(err, tea.ErrInterrupted) {
			logger.Err(err).Msg("failed to run app")
		}
	}
}
