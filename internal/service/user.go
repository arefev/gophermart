package service

import (
	"context"
	"errors"

	"github.com/arefev/gophermart/internal/model"
)

func UserWithContext(ctx context.Context) (*model.User, error) {
	user, ok := ctx.Value(model.User{}).(*model.User)

	if !ok {
		return nil, errors.New("user not found in context")
	}

	return user, nil
}