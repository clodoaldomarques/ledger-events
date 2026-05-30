package message

import (
	"context"
	"time"

	"github.com/clodoaldomarques/core-sdk/pkg/sns"
	"github.com/clodoaldomarques/ledger-events/config"
	"github.com/clodoaldomarques/ledger-events/internal/domain/events"
	"github.com/google/uuid"
)

type Topic struct {
	p *sns.Publisher
}

func New(ctx context.Context) *Topic {
	return &Topic{
		p: sns.NewPublisher(ctx, config.New()),
	}
}

func (t Topic) Emit(ctx context.Context, cid string, e events.Event) error {
	evt := sns.Event{
		EventID:   uuid.New(),
		EventType: "ledger",
		EventData: e,
		EventDate: time.Now(),
	}
	return t.p.Emit(ctx, evt)
}

func (t Topic) Close() {

}
