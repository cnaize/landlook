package journal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"os"
	"slices"
	"strconv"
	"time"

	"github.com/elastic/go-libaudit/v2"
	"github.com/elastic/go-libaudit/v2/aucoalesce"
	"github.com/elastic/go-libaudit/v2/auparse"
	"github.com/rs/zerolog"

	"github.com/cnaize/landlook/app/helper"
)

var _ libaudit.Stream = (*Journal)(nil)

type Journal struct {
	logger zerolog.Logger
	client *libaudit.AuditClient
	reasmr *libaudit.Reassembler
	domain string
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

		if err := client.SetEnabled(true, libaudit.WaitForReply); err != nil {
			return fmt.Errorf("set enabled: %w", err)
		}

		if err := client.SetPID(libaudit.WaitForReply); err != nil {
			return fmt.Errorf("set pid: %w", err)
		}
	}
	j.client = client

	// create reassembler
	j.reasmr, err = libaudit.NewReassembler(32, 3*time.Second, j)
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

func (j *Journal) GetEvents() []*aucoalesce.Event {
	events := make(map[string]*aucoalesce.Event, len(j.events))
	for _, event := range j.events {
		action, target := helper.GetEventActionTarget(event)
		evtkey := string(action) + target
		if events[evtkey] == nil {
			events[evtkey] = event
		}
	}

	return slices.Collect(maps.Values(events))
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
	helper.CleanEvent(event)

	// find domain
	ppid := strconv.Itoa(os.Getpid())
	if j.domain == "" && event.Process.PPID == ppid {
		j.domain = event.Data["domain"]
	}

	// filter event
	if event.Type != 1423 || j.domain == "" || event.Data["domain"] != j.domain || event.Process.PPID != ppid {
		return
	}

	// print debug
	if j.logger.GetLevel() == zerolog.DebugLevel {
		data, err := json.Marshal(event)
		if err != nil {
			j.logger.Err(err).Msg("failed to marshal event")
		} else {
			j.logger.Debug().RawJSON("event", data).Msg("new event")
		}
	}

	// save event
	j.events = append(j.events, event)
	j.logger.Info().Msg(helper.FormatEventLog(event))
}

func (j *Journal) EventsLost(count int) {
	j.logger.Warn().Int("count", count).Msg("events lost")
}
