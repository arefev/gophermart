package db

import (
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
