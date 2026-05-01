package main

import (
	"context"
	"net/http"

	"github.com/clodoaldomarques/ledger-events/configs"
	"github.com/clodoaldomarques/ledger-events/internal/infra/rest/server"
	"github.com/clodoaldomarques/ledger-events/pkg/logger"
)

func main() {
	c := configs.New(configs.WithAppPort(5001))
	err := server.New().Start(c.AppPort)
	if err != http.ErrServerClosed {
		logger.Fatal(context.Background(), err.Error(), logger.Fields{})
	}
}
