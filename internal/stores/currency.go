package stores

import "gofr.dev/pkg/gofr/http"

const (
	INR Currency = iota
	USD
)

type Currency int

var currencyNames = map[Currency]string{ //nolint:gochecknoglobals // static display-name lookup
	INR: "INR",
	USD: "USD",
}

func (c Currency) String() string {
	return currencyNames[c]
}

var currencyValues = map[string]Currency{ //nolint:gochecknoglobals // static reverse lookup
	"INR": INR,
	"USD": USD,
}

func CurrencyFromString(str string) (Currency, error) {
	currency, ok := currencyValues[str]
	if !ok {
		return 0, http.ErrorEntityNotFound{Name: "currencies", Value: str}
	}

	return currency, nil
}
