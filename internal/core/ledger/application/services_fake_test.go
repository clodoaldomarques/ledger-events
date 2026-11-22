package application

import (
	"context"
	"testing"

	"github.com/clodoaldomarques/ledger-events/internal/core/ledger/domain"
	"github.com/clodoaldomarques/ledger-events/internal/core/ledger/ports"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// Fakes in-memory das portas de saída: a mesma suíte de CreateEvent roda sem
// MySQL, SNS, Ledger Config ou mocks gerados (US1).
type memoryRepository struct {
	saved []domain.Event
}

func (r *memoryRepository) SaveEvent(_ context.Context, _ string, e domain.Event) error {
	r.saved = append(r.saved, e)
	return nil
}

type memoryPublisher struct {
	emitted []domain.Event
}

func (p *memoryPublisher) Emit(_ context.Context, _ string, e domain.Event) error {
	p.emitted = append(p.emitted, e)
	return nil
}

type memoryConfigProvider struct {
	cfg domain.Config
	err error
}

func (p memoryConfigProvider) FindConfigByLevel(_ context.Context, _ string, _ string, _ string, _ int64) (domain.Config, error) {
	return p.cfg, p.err
}

func TestService_CreateEvent_WithInMemoryAdapters(t *testing.T) {
	tests := []struct {
		name        string
		expression  string
		configErr   error
		wantErr     string
		wantSaved   int
		wantEmitted int
		wantAmount  string
	}{
		{
			name:        "calcula varias entradas e publica",
			expression:  "(Amounts.amount + Fees.iof) * (Fees.tax / Fees.iof) - Fees.tax",
			wantSaved:   1,
			wantEmitted: 1,
			wantAmount:  "150",
		},
		{
			name:        "calcula entrada unica",
			expression:  "Amounts.amount",
			wantSaved:   1,
			wantEmitted: 1,
			wantAmount:  "150",
		},
		{
			name:       "expressao invalida aborta antes de salvar",
			expression: "jack sparrow is here",
			wantErr:    "invalid expression: jack sparrow is here",
		},
		{
			name:      "config nao encontrada aborta antes de calcular",
			configErr: domain.ErrScriptNotFound{},
			wantErr:   "ledger config not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &memoryRepository{}
			pub := &memoryPublisher{}
			provider := memoryConfigProvider{cfg: fakeScript(tt.expression), err: tt.configErr}
			svc := New(provider, repo, pub)

			e, a, f := FakeEvent(), FakeAmount(), FakeFee()
			got, err := svc.CreateEvent(context.Background(), uuid.NewString(), e, a, f)

			if tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
				assert.Equal(t, domain.Event{}, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantAmount, got.Entries[0].Amount.String())
			}
			assert.Len(t, repo.saved, tt.wantSaved)
			assert.Len(t, pub.emitted, tt.wantEmitted)
			if tt.wantSaved > 0 {
				assert.Equal(t, repo.saved[0], pub.emitted[0])
			}
		})
	}
}

var (
	_ ports.EventRepository     = (*memoryRepository)(nil)
	_ ports.EventPublisher      = (*memoryPublisher)(nil)
	_ ports.EventConfigProvider = memoryConfigProvider{}
)
