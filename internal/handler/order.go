package handler

import (
	"net/http"

	"github.com/arefev/gophermart/internal/repository"
	"github.com/arefev/gophermart/internal/repository/db"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type order struct {
	log *zap.Logger
}

func NewOrder(log *zap.Logger) *order {
	return &order{log: log}
}

func (o *order) Create(w http.ResponseWriter, r *http.Request) {
	rep := repository.NewOrder(o.log)
	
	db.Transaction(func(tx *sqlx.Tx) error {
		rep.Create(tx, 1, 1, "12345")
		return nil
	})
	
	o.log.Info("Create order handler called")
}

func (o *order) List(w http.ResponseWriter, r *http.Request) {
	o.log.Info("List orders handler called")
}
