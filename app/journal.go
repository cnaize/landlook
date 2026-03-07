package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/elastic/go-libaudit/v2"
	"github.com/elastic/go-libaudit/v2/aucoalesce"
	"github.com/elastic/go-libaudit/v2/auparse"
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

func (j *Journal) Start(ctx context.Context) error {
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
	j.reasmr, err = libaudit.NewReassembler(0x20, 3*time.Second, j)
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
				if j.reasmr.Maintain() != nil {
					return
				}
			default:
				// receive message
				raw, err := client.Receive(false)
				if err != nil {
					return
				}

				// parse message
				msg, err := auparse.Parse(raw.Type, string(raw.Data))
				if err != nil {
					continue
				}

				// push message
				j.reasmr.PushMessage(msg)
			}
		}
	}()

	return nil
}

func (j *Journal) Stop() error {
	var errs error
	if j.client != nil {
		if err := j.client.Close(); err != nil {
			errs = errors.Join(errs, fmt.Errorf("close client: %w", err))
		}
	}

	if j.reasmr != nil {
		if err := j.reasmr.Close(); err != nil {
			errs = errors.Join(errs, fmt.Errorf("close reassembler: %w", err))
		}
	}

	return errs
}

func (j *Journal) ReassemblyComplete(msgs []*auparse.AuditMessage) {
	// make event
	event, err := aucoalesce.CoalesceMessages(msgs)
	if err != nil {
		j.logger.Err(err).Msg("failed to coalesce messages")
		return
	}

	// clean event
	aucoalesce.ResolveIDs(event)
	event = CleanEvent(event)

	// print debug
	if j.logger.GetLevel() == zerolog.DebugLevel {
		data, err := json.Marshal(event)
		if err != nil {
			j.logger.Err(err).Msg("failed to marshal event")
			return
		}
		j.logger.Debug().RawJSON("event", data).Msg("new event")
	}

	// filter event
	if event.Type < 1420 || event.Type > 1425 || event.Data["exit"] != "EACCES" {
		return
	}

	// save event
	j.events = append(j.events, event)
	j.logger.Info().Msg(FormatEvent(event))
}

func (j *Journal) EventsLost(count int) {
	j.logger.Warn().Int("count", count).Msg("events lost")
}
