package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/arefev/gophermart/internal/model"
	"github.com/golang-jwt/jwt/v5"
)

type JWT struct {
	claims jwt.MapClaims
	err    error
	secret string
}

type Token struct {
	AccessToken string `json:"accessToken"`
	Exp         int64  `json:"exp"`
}

func NewToken(secret string) *JWT {
	return &JWT{
		secret: secret,
	}
}

func (j *JWT) GenerateToken(user *model.User, duration int) (*Token, error) {
	d := time.Minute * time.Duration(duration)
	exp := time.Now().Add(d).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"login": user.Login,
		"exp":   exp,
	})

	strToken, err := token.SignedString([]byte(j.secret))
	if err != nil {
		return nil, fmt.Errorf("generate token fail: %w", err)
	}

	return &Token{AccessToken: strToken, Exp: exp}, nil
}

func (j *JWT) Parse(tokenStr string) *JWT {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}

		return []byte(j.secret), nil
	})

	if err != nil {
		j.err = fmt.Errorf("token parse fail: %w", err)
		return j
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		j.err = errors.New("claims not found")
		return j
	}

	j.claims = claims

	return j
}

func (j *JWT) GetLogin() (string, error) {
	if err := j.checkErr(); err != nil {
		return "", fmt.Errorf("get login fail: %w", err)
	}

	errLoginNotFound := errors.New("login not found")
	value, ok := j.claims["login"]
	if !ok {
		return "", errLoginNotFound
	}

	login, ok := value.(string)
	if !ok {
		return "", errLoginNotFound
	}

	return login, nil
}

func (j *JWT) checkErr() error {
	if j.err != nil {
		return j.err
	}

	if j.claims == nil {
		return errors.New("claims not found")
	}

	return nil
}
