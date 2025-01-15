package helper

import (
	"errors"
)

func CheckLuhn(number string) error {
	const numForParity = 2
	const numSubtract = 9

	sum := 0
	numDigits := len(number)
	parity := numDigits % numForParity

	for i, digit := range number {
		digitInt := int(digit - '0')

		if i%numForParity == parity {
			digitInt *= numForParity
			if digitInt > numSubtract {
				digitInt -= numSubtract
			}
		}

		sum += digitInt
	}

	if sum%10 != 0 {
		return errors.New("number is not valid")
	}

	return nil
}
