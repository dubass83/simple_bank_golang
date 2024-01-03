package util

const (
	USD = "USD"
	EUR = "EUR"
	UAH = "UAH"
)

func IfSupportedCurrency(currency string) bool {
	switch currency {
	case USD, EUR, UAH:
		return true
	}
	return false
}
