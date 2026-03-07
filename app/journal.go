package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/elastic/go-libaudit"
	"github.com/elastic/go-libaudit/aucoalesce"
	"github.com/elastic/go-libaudit/auparse"
	"github.com/rs/zerolog"
)

type Journal struct {
	logger zerolog.Logger
	client *libaudit.AuditClient
	reasmr *libaudit.Reassembler
	events []*aucoalesce.Event
}

func NewJournal(logger zerolog.Logger) *Journal {
	return &Journal{
		logger: logger,
	}
}

func (a *Journal) Start(ctx context.Context) error {
	// create client
	client, err := libaudit.NewMulticastAuditClient(nil)
	if err != nil {
		client, err = libaudit.NewAuditClient(nil)
		if err != nil {
			return fmt.Errorf("new client: %w", err)
		}

		if err := client.SetPID(libaudit.WaitForReply); err != nil {
			return fmt.Errorf("set pid: %w", err)
		}
	}

	// create reassembler
	a.reasmr, err = libaudit.NewReassembler(0x20, 3*time.Second, a)
	if err != nil {
		return fmt.Errorf("new reassembler: %w", err)
	}

	// run loop
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if a.reasmr.Maintain() != nil {
					return
				}
			default:
				// receive message
				raw, err := client.Receive(false)
				if err != nil {
					return
				}

				// filter message
				if raw.Type < 1420 || raw.Type > 1425 {
					continue
				}

				// parse message
				msg, err := auparse.Parse(raw.Type, string(raw.Data))
				if err != nil {
					continue
				}

				// save message
				a.reasmr.PushMessage(msg)
			}
		}
	}()

	return nil
}

func (a *Journal) Stop() error {
	var errs error
	if a.client != nil {
		if err := a.client.Close(); err != nil {
			errs = errors.Join(errs, fmt.Errorf("close client: %w", err))
		}
	}

	if a.reasmr != nil {
		if err := a.reasmr.Close(); err != nil {
			errs = errors.Join(errs, fmt.Errorf("close reassembler: %w", err))
		}
	}

	return errs
}

func (a *Journal) ReassemblyComplete(msgs []*auparse.AuditMessage) {
	event, err := aucoalesce.CoalesceMessages(msgs)
	if err != nil {
		a.logger.Debug().Msgf("coalesce messages: %s", err.Error())
		return
	}

	aucoalesce.ResolveIDs(event)
	a.events = append(a.events, event)

	a.logger.Info().Str("path", event.Data["path"]).Msg("new event")
}

func (a *Journal) EventsLost(count int) {
	a.logger.Warn().Int("count", count).Msg("events lost")
}
