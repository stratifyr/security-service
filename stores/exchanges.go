package stores

import (
	"gofr.dev/pkg/gofr"
	"gofr.dev/pkg/gofr/http"
)

const (
	NSE Exchange = iota
	BSE
)

type Exchange int

type ExchangeStore interface {
	Index(ctx *gofr.Context) ([]Exchange, error)
}

type exchangeStore struct{}

func NewExchangeStore() ExchangeStore {
	return &exchangeStore{}
}

func (s *exchangeStore) Index(ctx *gofr.Context) ([]Exchange, error) {
	return []Exchange{
		NSE,
		BSE,
	}, nil
}

func (ex Exchange) String() string {
	var conversionMap = map[Exchange]string{
		NSE: "NSE",
		BSE: "BSE",
	}

	return conversionMap[ex]
}

func ExchangeFromString(str string) (Exchange, error) {
	var conversionMap = map[string]Exchange{
		"NSE": NSE,
		"BSE": BSE,
	}

	exchange, ok := conversionMap[str]
	if !ok {
		return 0, http.ErrorEntityNotFound{Name: "exchanges", Value: str}
	}

	return exchange, nil
}
