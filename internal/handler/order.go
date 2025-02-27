package handler

import (
	"errors"
	"net/http"

	action "github.com/arefev/gophermart/internal/action/order"
	"github.com/arefev/gophermart/internal/application"
	"github.com/arefev/gophermart/internal/response"
	"github.com/arefev/gophermart/internal/service"
	"go.uber.org/zap"
)

type order struct {
	app *application.App
}

func NewOrder(app *application.App) *order {
	return &order{app: app}
}

func (o *order) Create(w http.ResponseWriter, r *http.Request) {
	err := action.NewCreateAction(o.app).Handle(r)

	switch {
	case errors.Is(err, action.ErrOrderCreateValidateFail):
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	case errors.Is(err, action.ErrOrderCreateUploadedByCurrentUser):
		w.WriteHeader(http.StatusOK)
		return
	case errors.Is(err, action.ErrOrderCreateUploadedByOtherUser):
		w.WriteHeader(http.StatusConflict)
		return
	case err != nil:
		o.app.Log.Error("Create order handler", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (o *order) List(w http.ResponseWriter, r *http.Request) {
	orders, err := action.NewListAction(o.app).Handle(r)

	if err != nil {
		o.app.Log.Error("List orders handler", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if err := service.JSONResponse(w, response.NewOrders(orders)); err != nil {
		o.app.Log.Error("List orders handler", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
