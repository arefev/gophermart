package handler

import (
	"net/http"

	"go.uber.org/zap"
)

type order struct {
	log *zap.Logger
}

func NewOrder(log *zap.Logger) *order {
	return &order{log: log}
}

func (o *order) Save(w http.ResponseWriter, r *http.Request) {
	o.log.Info("Save order handler called")
}

func (o *order) List(w http.ResponseWriter, r *http.Request) {
	o.log.Info("List orders handler called")
}
