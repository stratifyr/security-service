package dataProviders

import (
	"time"

	"gofr.dev/pkg/gofr"
)

const (
	DhanHQ Provider = iota
	SmartAPI
	KiteConnect
)

type DataProvider interface {
	GetLTP(ctx *gofr.Context, isin string) (*LTPData, error)
	GetLTPBulk(ctx *gofr.Context, isins []string) ([]*LTPData, error)
	GetMarketData(ctx *gofr.Context, isin string, date time.Time) (*MarketData, error)
	GetMarketDataBulk(ctx *gofr.Context, isins []string, date time.Time) ([]*MarketData, error)
}

type LTPData struct {
	ISIN string
	LTP  float64
}

type MarketData struct {
	ISIN   string
	Open   float64
	Close  float64
	High   float64
	Low    float64
	Volume int
}

func New(app *gofr.App) DataProvider {
	switch app.Config.Get("MARKET_DATA_PROVIDER") {
	case DhanHQ.String():
		return NewDhanHQClient(app)
	default:
		app.Logger().Fatalf("invalid MARKET_DATA_PROVIDER")
	}

	return nil
}

type Provider int

func (p Provider) String() string {
	conversionMap := map[Provider]string{
		DhanHQ:      "DHAN_MARKET_API",
		SmartAPI:    "ANGELONE_SMART_API",
		KiteConnect: "ZERODHA_KITE_CONNECT_API",
	}

	return conversionMap[p]
}
