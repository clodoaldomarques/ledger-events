package mysql

import (
	"context"
	"database/sql"

	"github.com/clodoaldomarques/core-sdk/pkg/logger"
	"github.com/clodoaldomarques/ledger-events/internal/domain/events"
)

var (
	INSERT_EVENT = `INSERT INTO event 
		(event_id, correlation_id, org_id, processing_code, program_id, account_id, description, producer, created_at)
	VALUES
		(?, ?, ?, ?, ?, ?, ?, ?, ?);`
	INSERT_ENTRIES = `INSERT INTO entry
		(event_id, entry_type_id, amount, debit_account, credit_account, description)
	VALUES
		(?, ?, ?, ?, ?, ?);`
)

type Repository struct {
	db *sql.DB
}

func NewRepository(ctx context.Context) *Repository {
	db, err := Connect()
	if err != nil {
		logger.Error(ctx, "error on connect to database", logger.Fields{"error": err.Error()})
		return &Repository{}
	}

	return &Repository{db: db}
}

func (r Repository) Close() {
	r.db.Close()
}

func (r Repository) SaveEvent(ctx context.Context, cid string, e events.Event) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		logger.Error(ctx, "error on start new transaction", logger.Fields{"event": e, "error": err.Error()})
		return err
	}
	defer tx.Rollback()

	evt, err := buildEventTable(cid, e)
	stat, err := tx.Prepare(INSERT_EVENT)
	if err != nil {
		logger.Error(ctx, "error on prepare statement", logger.Fields{"event": e, "error": err.Error(), "sql": INSERT_EVENT})
		return err
	}
	_, err = stat.Exec(
		evt.EventID,
		evt.CorrelationID,
		evt.OrgID,
		evt.ProcessingCode,
		evt.ProgramID,
		evt.AccountID,
		evt.Description,
		evt.Producer,
		evt.CreatedAt,
	)
	if err != nil {
		return err
	}

	for _, etr := range evt.Entries {
		stat, err = tx.Prepare(INSERT_ENTRIES)
		if err != nil {
			logger.Error(ctx, "error on save new entry", logger.Fields{"event": e, "entry": etr, "error": err.Error(), "sql": INSERT_ENTRIES})
			return err
		}
		defer stat.Close()

		_, err = stat.Exec(
			etr.EventID,
			etr.EntryTypeID,
			etr.Amount,
			etr.DebitAccount,
			etr.CreditAccount,
			etr.Description,
		)
		if err != nil {
			return err
		}
	}

	tx.Commit()
	return nil
}
