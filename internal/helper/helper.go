package helper

import (
	"context"
	"errors"

	"github.com/arefev/gophermart/internal/model"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

func UserWithContext(ctx context.Context) (*model.User, error) {
	user, ok := ctx.Value(model.User{}).(*model.User)

	if !ok {
		return nil, errors.New("user not found in context")
	}

	return user, nil
}

func CheckLuhn(number string) error {
	sum := 0
	numDigits := len(number)
	parity := numDigits % 2

	for i, digit := range number {
		digitInt := int(digit - '0')

		if i%2 == parity {
			digitInt *= 2
			if digitInt > 9 {
				digitInt -= 9
			}
		}

		sum += digitInt
	}

	if sum%10 != 0 {
		return errors.New("number is not valid")
	}

	return nil
}