package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	snssdk "github.com/clodoaldomarques/core-sdk/pkg/aws/sns"
	"github.com/clodoaldomarques/core-sdk/pkg/otel"
	"github.com/clodoaldomarques/core-sdk/pkg/otel/httpclient"
	"github.com/clodoaldomarques/core-sdk/pkg/zap/logger"
	"github.com/clodoaldomarques/ledger-events/config"
	"github.com/clodoaldomarques/ledger-events/internal/adapters/inbound/http/ledger"
	"github.com/clodoaldomarques/ledger-events/internal/adapters/inbound/http/server"
	configadapter "github.com/clodoaldomarques/ledger-events/internal/adapters/outbound/config"
	"github.com/clodoaldomarques/ledger-events/internal/adapters/outbound/mysql"
	"github.com/clodoaldomarques/ledger-events/internal/adapters/outbound/sns"
	"github.com/clodoaldomarques/ledger-events/internal/core/ledger/application"
	"github.com/clodoaldomarques/ledger-events/internal/core/ledger/ports"
)

// healthCheckers preserva a ordem atual do /health/ready: MySQL e depois SNS.
type healthCheckers []ports.HealthChecker

var _ ports.HealthChecker = healthCheckers(nil)

func (h healthCheckers) Health(ctx context.Context) error {
	for _, checker := range h {
		if err := checker.Health(ctx); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	ctx := context.Background()
	otel.Start(ctx)

	cfg := config.New()

	db, err := connectDB(ctx, cfg)
	if err != nil {
		logger.Fatal(ctx, "error on connect to database", logger.Fields{"error": err.Error()})
	}
	repo := mysql.New(db)

	client := httpclient.Client()
	configAdapter := configadapter.New(ctx, cfg.LedgerConfigApiUrl, client)

	publisher := snssdk.NewPublisher(ctx, cfg)
	topic := sns.New(publisher, cfg)

	uc := application.New(configAdapter, repo, topic)
	handler := ledger.NewHandler(uc)
	health := healthCheckers{repo, topic}
	s := server.New(handler, health)

	go func() {
		err := s.Start(cfg.AppPort)
		if err != http.ErrServerClosed {
			logger.Fatal(ctx, err.Error(), logger.Fields{})
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("Recebido sinal de desligamento, iniciando graceful shutdown...")

	ctx, cancel := context.WithTimeout(ctx, 25*time.Second)
	defer cancel()

	if err := s.Shutdown(ctx); err != nil {
		fmt.Printf("Erro no shutdown do servidor HTTP: %v\n", err)
	} else {
		fmt.Println("Servidor HTTP finalizado com sucesso")
	}

	configAdapter.Close()
	repo.Close()
	topic.Close(ctx)

	if err := otel.Shutdown(ctx); err != nil {
		fmt.Printf("erro no shutdown do opentelemetry: %v\n", err)
	}

	fmt.Println("Graceful shutdown concluído")
}
