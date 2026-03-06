package main

import (
	"context"
	"fmt"
	"os"

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
		Usage:     "secure command inspection tool",
		ArgsUsage: "command [arguments]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        app.AppFlagLogLevel,
				Usage:       "set zerolog level",
				Value:       "info",
				DefaultText: "info",
			},
			&cli.StringFlag{
				Name:        app.AppFlagROPaths,
				Usage:       "allow to read/exec path",
				DefaultText: "deny all",
			},
			&cli.StringFlag{
				Name:        app.AppFlagRWPaths,
				Usage:       "allow to read/write/exec path",
				DefaultText: "deny all",
			},
			&cli.Uint16SliceFlag{
				Name:        app.AppFlagTCPListen,
				Aliases:     []string{"l"},
				Usage:       "allow to listen tcp ports",
				DefaultText: "deny all",
			},
			&cli.Uint16SliceFlag{
				Name:        app.AppFlagTCPConnect,
				Aliases:     []string{"c"},
				Usage:       "allow to connect tcp ports",
				DefaultText: "deny all",
			},
			&cli.BoolFlag{
				Name:        app.AppFlagDenySockets,
				Usage:       "allow to open abstract sockets",
				Value:       false,
				DefaultText: "deny",
			},
			&cli.BoolFlag{
				Name:        app.AppFlagDenySignals,
				Usage:       "allow to send signals",
				Value:       false,
				DefaultText: "deny",
			},
			&cli.StringSliceFlag{
				Name:        app.AppFlagAddEnvs,
				Aliases:     []string{"e"},
				Usage:       "add environment variables",
				DefaultText: "empty list",
			},
			&cli.BoolFlag{
				Name:        app.AppFlagAddSelf,
				Usage:       fmt.Sprintf("add command itself to --%s", app.AppFlagROPaths),
				Value:       true,
				DefaultText: "true",
			},
			&cli.BoolFlag{
				Name:        app.AppFlagAddDeps,
				Usage:       fmt.Sprintf("add command dependencies to --%s", app.AppFlagROPaths),
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
		logger.Err(err).Msg("failed to run")
	}
}
