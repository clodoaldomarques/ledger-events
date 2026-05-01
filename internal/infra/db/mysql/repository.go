package mysql

import (
	"context"
	"database/sql"

	"github.com/clodoaldomarques/ledger-events/internal/domain/events"
	"github.com/clodoaldomarques/ledger-events/pkg/logger"
)

var (
	INSERT_EVENT = `INSERT INTO event 
		(event_id, correlation_id, org_id, event_type_id, transaction_id, authorization_id, program_id, customer_id, account_id, days_due, description, transaction_type, process_id, producer, processing_code, processing_description, clearing_date, chunk_part, created_at)
	VALUES
		(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`
	INSERT_ENTRIES = `INSERT INTO entry
		(event_id, event_type_id, amount, debit_account, credit_account, cost_center_debit_org, cost_center_debit_cost, cost_center_credit_org, cost_center_credit_cost, company_code, company_type, description)
	VALUES
		(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`
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
		evt.EventTypeID,
		evt.TransactionID,
		evt.AuthorizationID,
		evt.ProcessID,
		evt.CustomerID,
		evt.AccountID,
		evt.DaysDue,
		evt.Description,
		evt.TransactionType,
		evt.ProcessID,
		evt.Producer,
		evt.ProcessingCode,
		evt.ProcessingDescription,
		evt.ClearingDate,
		evt.ChunkPart,
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
			etr.CostCenterDebitOrg,
			etr.CostCenterDebitCost,
			etr.CostCenterCreditOrg,
			etr.CostCenterCreditCost,
			etr.CompanyCode,
			etr.CompanyType,
			etr.Description,
		)
		if err != nil {
			return err
		}
	}

	tx.Commit()
	return nil
}
