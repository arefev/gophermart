package password

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func Encrypt(pwd string) (string, error) {
	passHash, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("register encrypt password fail: %w", err)
	}

	return string(passHash), nil
}

func Check(userPwd, pwd string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(userPwd), []byte(pwd))
	return err == nil
}
