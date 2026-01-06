package consumers

import (
	"context"

	"github.com/google/uuid"
	"github.com/netbill/kafkakit/box"
	"github.com/netbill/kafkakit/subscriber"
	"github.com/netbill/logium"
	"github.com/netbill/replicas/kafka/contracts"
	"github.com/segmentio/kafka-go"
	"golang.org/x/sync/errgroup"
)

type Inbox interface {
	CreateInboxEvent(
		ctx context.Context,
		message kafka.Message,
	) (box.InboxEvent, error)

	UpdateInboxEventStatus(
		ctx context.Context,
		id uuid.UUID,
		status string,
	) (box.InboxEvent, error)

	Transaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type processor interface {
	ProfileUpdated(
		ctx context.Context,
		event box.InboxEvent,
	) string
}

type Profile struct {
	log       logium.Logger
	inbox     Inbox
	processor processor
}

func NewProfile(log logium.Logger, inbox Inbox, processor processor) *Profile {
	return &Profile{
		processor: processor,
		inbox:     inbox,
		log:       log,
	}
}

func (c Profile) Run(ctx context.Context, g *errgroup.Group, group string, addr ...string) {
	c.log.Info("starting events consumer", "addr", addr)

	g.Go(func() error {
		profileSub := subscriber.New(addr, contracts.ProfilesTopicV1, group)
		err := profileSub.Consume(ctx, func(m kafka.Message) (subscriber.HandlerFunc, bool) {
			et, ok := subscriber.Header(m, "event_type")
			if !ok {
				return nil, false
			}
			switch et {
			case contracts.ProfileUpdatedEvent:
				return c.ProfileUpdated, true
			default:
				return nil, false
			}
		})
		if err != nil {
			c.log.Warnf("profiles consumer stopped: %v", err)
		}
		return err
	})

	_ = g.Wait()
}

func (c Profile) ProfileUpdated(ctx context.Context, event kafka.Message) error {
	return c.inbox.Transaction(ctx, func(ctx context.Context) error {
		eventInBox, err := c.inbox.CreateInboxEvent(ctx, event)
		if err != nil {
			c.log.Errorf("failed to upsert inbox event for account %s: %v", string(event.Key), err)
			return err
		}

		if _, err = c.inbox.UpdateInboxEventStatus(ctx, eventInBox.ID, c.processor.ProfileUpdated(ctx, eventInBox)); err != nil {
			c.log.Errorf(
				"failed to update inbox event status for key %s, id: %s, error: %v", eventInBox.Key, eventInBox.ID, err,
			)
		}

		return nil
	})
}
