package dataProviders

import (
	"errors"
	"time"

	"gofr.dev/pkg/gofr"
)

const (
	DhanHQ Provider = iota
	SmartAPI
	KiteConnect
)

type DataProvider interface {
	LTP(ctx *gofr.Context, isin string) (*LTPData, error)
	LTPBulk(ctx *gofr.Context, isins []string) ([]*LTPData, error)
	OHLC(ctx *gofr.Context, isin string) (*OHLCData, error)
	OHLCBulk(ctx *gofr.Context, isins []string) ([]*OHLCData, error)
	HistoricalOHLC(ctx *gofr.Context, isin string, startDate, endDate time.Time) ([]*HistoricalOHLC, error)
}

type LTPData struct {
	ISIN string
	LTP  float64
}

type OHLCData struct {
	ISIN   string
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume int
}

type HistoricalOHLC struct {
	Date time.Time
	*OHLCData
}

func New(app *gofr.App) (DataProvider, error) {
	switch app.Config.Get("MARKET_DATA_PROVIDER") {
	case DhanHQ.String():
		return NewDhanHQClient(app)
	default:
		return nil, errors.New("invalid MARKET_DATA_PROVIDER")
	}
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
