package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/arefev/gophermart/internal/model"
	"github.com/golang-jwt/jwt/v5"
)

type Token struct {
	claims jwt.MapClaims
	err    error
	secret string
}

func NewToken(secret string) *Token {
	return &Token{
		secret: secret,
	}
}

func (t *Token) GenerateToken(user *model.User, duration int) (string, error) {
	d := time.Minute * time.Duration(duration)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"login": user.Login,
		"exp":   time.Now().Add(d).Unix(),
	})

	strToken, err := token.SignedString([]byte(t.secret))
	if err != nil {
		return "", fmt.Errorf("generate token fail: %w", err)
	}

	return strToken, nil
}

func (t *Token) Parse(tokenStr string) *Token {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}

		return []byte(t.secret), nil
	})

	if err != nil {
		t.err = fmt.Errorf("token parse fail: %w", err)
		return t
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		t.err = errors.New("claims not found")
		return t
	}

	t.claims = claims

	return t
}

func (t *Token) GetLogin() (string, error) {
	if err := t.checkErr(); err != nil {
		return "", fmt.Errorf("get login fail: %w", err)
	}

	errLoginNotFound := errors.New("login not found")
	value, ok := t.claims["login"]
	if !ok {
		return "", errLoginNotFound
	}

	login, ok := value.(string)
	if !ok {
		return "", errLoginNotFound
	}

	return login, nil
}

func (t *Token) checkErr() error {
	if t.err != nil {
		return t.err
	}

	if t.claims == nil {
		return errors.New("claims not found")
	}

	return nil
}
