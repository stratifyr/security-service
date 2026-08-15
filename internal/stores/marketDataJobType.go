package stores

import (
	"gofr.dev/pkg/gofr"
	"gofr.dev/pkg/gofr/http"
)

type MarketDataJobTypeStore interface {
	Index(ctx *gofr.Context) []MetricType
}

const (
	LoadLTP = iota
	LoadSecurityStats
	BackfillSecurityStats
)

type MarketDataJobType int

type marketDataJobTypeStore struct{}

func NewMarketDataJobTypeStore() *marketDataJobTypeStore {
	return &marketDataJobTypeStore{}
}

func (s *marketDataJobTypeStore) Index(ctx *gofr.Context) []MarketDataJobType {
	return []MarketDataJobType{
		LoadLTP,
		LoadSecurityStats,
		BackfillSecurityStats,
	}
}

func (m MarketDataJobType) String() string {
	var conversionMap = map[MarketDataJobType]string{
		LoadLTP:               "LOAD_LTP",
		LoadSecurityStats:     "LOAD_SECURITY_STATS",
		BackfillSecurityStats: "BACKFILL_SECURITY_STATS",
	}

	return conversionMap[m]
}

func MarketDataJobTypeFromString(str string) (MarketDataJobType, error) {
	var conversionMap = map[string]MarketDataJobType{
		"LOAD_LTP":                LoadLTP,
		"LOAD_SECURITY_STATS":     LoadSecurityStats,
		"BACKFILL_SECURITY_STATS": BackfillSecurityStats,
	}

	marketDataJobType, ok := conversionMap[str]
	if !ok {
		return 0, http.ErrorEntityNotFound{Name: "market-data-job-types", Value: str}
	}

	return marketDataJobType, nil
}
