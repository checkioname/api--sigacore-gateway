package util

const (
	EUR = "EUR"
	USD = "USD"
	CAD = "CAD"
)

func IsValidCurrency(currency string) bool {
	switch currency {
	case EUR, USD, CAD:
		return true
	}
	return false
}
