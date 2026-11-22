package sns

import (
	"context"
	"time"

	snssdk "github.com/clodoaldomarques/core-sdk/pkg/aws/sns"
	"github.com/clodoaldomarques/core-sdk/pkg/otel/tracer"
	"github.com/clodoaldomarques/core-sdk/pkg/zap/logger"
	"github.com/clodoaldomarques/ledger-events/internal/core/ledger/domain"
	"github.com/clodoaldomarques/ledger-events/internal/core/ledger/ports"
	"github.com/google/uuid"
)

type Topic struct {
	p   *snssdk.Publisher
	cfg snssdk.Config
}

var (
	_ ports.EventPublisher = (*Topic)(nil)
	_ ports.HealthChecker  = (*Topic)(nil)
)

// New recebe o publicador compartilhado e a configuração AWS usada no health check.
func New(p *snssdk.Publisher, cfg snssdk.Config) *Topic {
	return &Topic{p: p, cfg: cfg}
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

// Health usa o ctx da requisição para preservar o prazo de 500ms do /health/ready:
// snssdk.Publisher.Health consulta o tópico com o ctx recebido no construtor.
func (t Topic) Health(ctx context.Context) error {
	return snssdk.NewPublisher(ctx, t.cfg).Health()
}
