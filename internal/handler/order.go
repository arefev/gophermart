package handler

import (
	"errors"
	"net/http"

	"github.com/arefev/gophermart/internal/repository"
	"github.com/arefev/gophermart/internal/response"
	"github.com/arefev/gophermart/internal/service"
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
	s := service.NewOrderCreate(rep)

	err := s.FromRequest(r)

	switch {
	case errors.Is(err, service.ErrOrderCreateValidateFail):
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	case errors.Is(err, service.ErrOrderCreateUploadedByCurrentUser):
		w.WriteHeader(http.StatusOK)
		return
	case errors.Is(err, service.ErrOrderCreateUploadedByOtherUser):
		w.WriteHeader(http.StatusConflict)
		return
	case err != nil:
		o.log.Error("Create order handler", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)

	o.log.Info("Create order handler called")
}

func (o *order) List(w http.ResponseWriter, r *http.Request) {
	rep := repository.NewOrder(o.log)
	s := service.NewOrderList(rep)
	orders, err := s.FromRequest(r)

	if err != nil {
		o.log.Error("List orders handler", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if err := service.JSONResponse(w, response.NewOrders(orders)); err != nil {
		o.log.Error("List orders handler", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	o.log.Info("List orders handler called")
}
