package db

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

var connection *sqlx.DB

func Connect(dsn string) error {
	db, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		return fmt.Errorf("db connect fail: %w", err)
	}

	connection = db
	return nil
}

func Close() error {
	if err := Connection().Close(); err != nil {
		return fmt.Errorf("db close fail: %w", err)
	}
	return nil
}

func Connection() *sqlx.DB {
	return connection
}

func Transaction(action func(tx *sqlx.Tx) error) error {
	tx, err := Connection().Beginx()
	if err != nil {
		return fmt.Errorf("db transaction failed: %w", err)
	}

	defer func() {
		if err := tx.Rollback(); err != nil {
			if !errors.Is(err, sql.ErrTxDone) {
				
			}
		}
	}()

	if err := action(tx); err != nil {
        return fmt.Errorf("db transaction fail: %w", err)
    }

	return tx.Commit()
}
