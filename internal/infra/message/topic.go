package message

import (
	"context"

	"github.com/clodoaldomarques/core-sdk/pkg/sns"
	"github.com/clodoaldomarques/ledger-events/config"
	"github.com/clodoaldomarques/ledger-events/internal/domain/events"
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
	return nil
}

func (t Topic) Close() {

}
