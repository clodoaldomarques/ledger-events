package mysql

import (
	"context"
	"database/sql"

	"github.com/clodoaldomarques/ledger-events/configs"
	"github.com/clodoaldomarques/ledger-events/pkg/logger"
	_ "github.com/go-sql-driver/mysql"
)

func Connect() (*sql.DB, error) {
	db, err := sql.Open("mysql", configs.New().GetMySQLConnectionString())
	if err != nil {
		logger.Fatal(context.Background(), "error on connect database", logger.Fields{
			"error":             err.Error(),
			"connection_string": configs.New().GetMySQLConnectionString(),
		})
		return nil, err
	}
	if err = db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}
