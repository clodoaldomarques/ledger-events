package sns

import (
	"context"
	"time"

	snssdk "github.com/clodoaldomarques/core-sdk/pkg/aws/sns"
	"github.com/clodoaldomarques/core-sdk/pkg/otel/tracer"
	"github.com/clodoaldomarques/core-sdk/pkg/zap/logger"
	"github.com/clodoaldomarques/ledger-events/config"
	"github.com/clodoaldomarques/ledger-events/internal/core/ledger/domain"
	"github.com/google/uuid"
)

type Topic struct {
	p *snssdk.Publisher
}

func New(ctx context.Context) *Topic {
	return &Topic{
		p: snssdk.NewPublisher(ctx, config.New()),
	}
}

func (t Topic) Emit(ctx context.Context, cid string, e domain.Event) error {
	span, ctx := tracer.NewSpanFromContext(ctx, "Topic::Emit", map[string]any{
		"cid":   cid,
		"event": e,
	})
	defer span.End()
	evt := snssdk.Event{
		EventID:   uuid.New(),
		EventType: "ledger",
		EventData: ToEventMessage(e),
		EventDate: time.Now(),
	}
	return t.p.Emit(ctx, evt)
}

func (t Topic) Close(ctx context.Context) {
	logger.Info(ctx, "ending topic connection", logger.Fields{})
}

func (t Topic) Health() error {
	return t.p.Health()
}
