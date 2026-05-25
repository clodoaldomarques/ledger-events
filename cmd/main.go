package main

import (
	"context"
	"net/http"

	"github.com/clodoaldomarques/core-sdk/pkg/logger"
	"github.com/clodoaldomarques/ledger-events/config"
	"github.com/clodoaldomarques/ledger-events/internal/infra/rest/server"
)

func main() {
	c := config.New(config.WithAppPort(5001))
	err := server.New().Start(c.AppPort)
	if err != http.ErrServerClosed {
		logger.Fatal(context.Background(), err.Error(), logger.Fields{})
	}
}
