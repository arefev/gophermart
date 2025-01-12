package db

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

var ErrTransactionFail = errors.New("transaction fail")

type db struct {
	db  *sqlx.DB
	log *zap.Logger
}

var connection *db

func Connect(dsn string, log *zap.Logger) error {
	dbConn, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		return fmt.Errorf("db connect fail: %w", err)
	}

	connection = &db{
		db:  dbConn,
		log: log,
	}

	return nil
}

func Close() error {
	if err := Connection().Close(); err != nil {
		return fmt.Errorf("db close fail: %w", err)
	}

	connection.log.Info("db connection closed")
	return nil
}

func Connection() *sqlx.DB {
	return connection.db
}

func Transaction(action func(tx *sqlx.Tx) error) error {
	tx, err := Connection().Beginx()
	if err != nil {
		return fmt.Errorf("db transaction begin fail: %w", err)
	}

	defer func() {
		if err := tx.Rollback(); err != nil {
			if !errors.Is(err, sql.ErrTxDone) {
				connection.log.Error("db transaction rollback fail", zap.Error(err))
			}
		}
	}()

	if err := action(tx); err != nil {
		return fmt.Errorf("db %w: %w", ErrTransactionFail, err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("db transaction commit fail: %w", err)
	}

	return nil
}
