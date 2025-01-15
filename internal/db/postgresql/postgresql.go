package postgresql

import (
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type db struct {
	conn *sqlx.DB
	log  *zap.Logger
}

func NewDB(log *zap.Logger) *db {
	return &db{
		log: log,
	}
}

func (db *db) Connect(dsn string) (*db, error) {
	conn, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("db connect fail: %w", err)
	}

	db.conn = conn

	return db, nil
}

func (db *db) Connection() *sqlx.DB {
	return db.conn
}

func (db *db) Close() error {
	if err := db.conn.Close(); err != nil {
		return fmt.Errorf("db close fail: %w", err)
	}

	db.log.Info("db connection closed")
	return nil
}
