package service

import "strconv"

type Logger interface {
	Error(message string, err error)
}

func isOrderNumberValid(number string) bool {
	if _, err := strconv.ParseUint(number, 10, 64); err != nil {
		return false
	}

	// проверка по алгоритму Луна
	sum := 0
	evenPos := false
	for i := len(number) - 1; i >= 0; i-- {
		digit := int(number[i] - '0')
		if evenPos {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
		evenPos = !evenPos
	}

	return sum%10 == 0
}
