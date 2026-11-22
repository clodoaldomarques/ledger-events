package application

import (
	"context"
	"testing"

	"github.com/clodoaldomarques/ledger-events/internal/core/ledger/domain"
	"github.com/clodoaldomarques/ledger-events/internal/core/ledger/ports"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

// =============================================================================
// Helpers de setup
// =============================================================================

// newBenchService monta o Service com mocks que sempre respondem com sucesso.
// Reaproveita os fakes (FakeEvent, FakeAmount, FakeFee, fakeScript) dos testes.
func newBenchService(b *testing.B, expr string) *Service {
	ctrl := gomock.NewController(b)
	// Não chamamos ctrl.Finish() aqui: o gomock registra um Cleanup
	// automático quando recebe *testing.B, validando as expectativas no final.

	a := ports.NewMockEventConfigProvider(ctrl)
	a.EXPECT().
		FindConfigByLevel(
			gomock.Any(),
			gomock.Any(),
			gomock.Any(),
			gomock.Any(),
			gomock.Any(),
		).
		Return(fakeScript(expr), nil).
		AnyTimes() // benchmark roda b.N vezes, então AnyTimes é obrigatório

	r := ports.NewMockEventRepository(ctrl)
	r.EXPECT().
		SaveEvent(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil).
		AnyTimes()

	t := ports.NewMockEventPublisher(ctrl)
	t.EXPECT().
		Emit(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil).
		AnyTimes()

	return New(a, r, t)
}

// newBenchServiceErrConfig monta um Service cujo ConfigProvider devolve erro.
func newBenchServiceErrConfig(b *testing.B) *Service {
	ctrl := gomock.NewController(b)

	a := ports.NewMockEventConfigProvider(ctrl)
	a.EXPECT().
		FindConfigByLevel(
			gomock.Any(),
			gomock.Any(),
			gomock.Any(),
			gomock.Any(),
			gomock.Any(),
		).
		Return(nil, domain.ErrScriptNotFound{}).
		AnyTimes()

	r := ports.NewMockEventRepository(ctrl)
	tp := ports.NewMockEventPublisher(ctrl)

	return New(a, r, tp)
}

// =============================================================================
// Benchmarks
// =============================================================================

// BenchmarkCreateEvent_Success mede o caminho feliz completo:
// busca config -> processa expressão -> valida -> salva -> emite.
func BenchmarkCreateEvent_Success(b *testing.B) {
	svc := newBenchService(b, "(Amounts.amount + Fees.iof) * (Fees.tax / Fees.iof) - Fees.tax")

	ctx := context.Background()
	e := FakeEvent()
	am := FakeAmount()
	f := FakeFee()
	cid := "cid-fixo" // fora do loop para não medir geração de UUID

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := svc.CreateEvent(ctx, cid, e, am, f)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkCreateEvent_Success_WithUUID é a mesma coisa, mas gerando UUID
// dentro do loop. Útil para comparar com o cenário real de produção.
func BenchmarkCreateEvent_Success_WithUUID(b *testing.B) {
	svc := newBenchService(b, "(Amounts.amount + Fees.iof) * (Fees.tax / Fees.iof) - Fees.tax")

	ctx := context.Background()
	e := FakeEvent()
	am := FakeAmount()
	f := FakeFee()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := svc.CreateEvent(ctx, uuid.NewString(), e, am, f)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkCreateEvent_InvalidExpression mede o custo de abortar cedo
// quando a expressão é inválida (falha no parsing).
func BenchmarkCreateEvent_InvalidExpression(b *testing.B) {
	svc := newBenchService(b, "jack sparrow is here")

	ctx := context.Background()
	e := FakeEvent()
	am := FakeAmount()
	f := FakeFee()
	cid := "cid-fixo"

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = svc.CreateEvent(ctx, cid, e, am, f)
	}
}

// BenchmarkCreateEvent_ScriptNotFound mede o custo de abortar ainda mais cedo,
// quando nem a config foi encontrada.
func BenchmarkCreateEvent_ScriptNotFound(b *testing.B) {
	svc := newBenchServiceErrConfig(b)

	ctx := context.Background()
	e := FakeEvent()
	am := FakeAmount()
	f := FakeFee()
	cid := "cid-fixo"

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = svc.CreateEvent(ctx, cid, e, am, f)
	}
}

// BenchmarkCreateEvent_Parallel mede o comportamento sob concorrência.
// Útil para detectar contenção (locks, alocações compartilhadas, etc).
func BenchmarkCreateEvent_Parallel(b *testing.B) {
	svc := newBenchService(b, "(Amounts.amount + Fees.iof) * (Fees.tax / Fees.iof) - Fees.tax")

	ctx := context.Background()
	cid := "cid-fixo"

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		// cada goroutine tem seu próprio conjunto de dados
		e := FakeEvent()
		am := FakeAmount()
		f := FakeFee()

		for pb.Next() {
			_, err := svc.CreateEvent(ctx, cid, e, am, f)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
