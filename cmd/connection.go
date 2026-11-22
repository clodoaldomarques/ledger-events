package main

import (
	"context"
	"database/sql"

	"github.com/clodoaldomarques/core-sdk/pkg/zap/logger"
	"github.com/clodoaldomarques/ledger-events/config"
	_ "github.com/go-sql-driver/mysql"
)

// connectDB é chamado uma única vez pela raiz de composição.
func connectDB(ctx context.Context, cfg *config.Config) (*sql.DB, error) {
	dsn := cfg.GetMySQLConnectionString()
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		logger.Fatal(ctx, "error on connect database", logger.Fields{
			"error":             err.Error(),
			"connection_string": dsn,
		})
		return nil, err
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
