package message

import (
	"context"

	"github.com/clodoaldomarques/ledger-events/internal/domain/events"
)

type Topic struct {
}

func New(ctx context.Context) *Topic {
	return &Topic{}
}

func (t Topic) Emit(ctx context.Context, cid string, e events.Event) error {
	return nil
}

func (t Topic) Close() {

}
