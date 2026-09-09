package stores

import (
	"gofr.dev/pkg/gofr"
	"gofr.dev/pkg/gofr/http"
)

type MarketDataJobTypeStore interface {
	List(ctx *gofr.Context) []MarketDataJobType
}

const (
	LoadSecurities     = iota
	LoadSecurityValues = iota
	LoadSecurityStats
	LoadSecurityShares
	BackfillSecurityStats
	LoadIndices
	LoadIndexValues
	LoadIndexStats
	BackfillIndexStats
)

type MarketDataJobType int

type marketDataJobTypeStore struct{}

func NewMarketDataJobTypeStore() *marketDataJobTypeStore {
	return &marketDataJobTypeStore{}
}

func (*marketDataJobTypeStore) List(_ *gofr.Context) []MarketDataJobType {
	return []MarketDataJobType{
		LoadSecurities,
		LoadSecurityValues,
		LoadSecurityStats,
		LoadSecurityShares,
		BackfillSecurityStats,
		LoadIndices,
		LoadIndexValues,
		LoadIndexStats,
		BackfillIndexStats,
	}
}

func (m MarketDataJobType) String() string {
	var conversionMap = map[MarketDataJobType]string{
		LoadSecurities:        "LOAD_SECURITIES",
		LoadSecurityValues:    "LOAD_SECURITY_VALUES",
		LoadSecurityStats:     "LOAD_SECURITY_STATS",
		LoadSecurityShares:    "LOAD_SECURITY_SHARES",
		BackfillSecurityStats: "BACKFILL_SECURITY_STATS",
		LoadIndices:           "LOAD_INDICES",
		LoadIndexValues:       "LOAD_INDEX_VALUES",
		LoadIndexStats:        "LOAD_INDEX_STATS",
		BackfillIndexStats:    "BACKFILL_INDEX_STATS",
	}

	return conversionMap[m]
}

func MarketDataJobTypeFromString(str string) (MarketDataJobType, error) {
	var conversionMap = map[string]MarketDataJobType{
		"LOAD_SECURITIES":         LoadSecurities,
		"LOAD_SECURITY_VALUES":    LoadSecurityValues,
		"LOAD_SECURITY_STATS":     LoadSecurityStats,
		"LOAD_SECURITY_SHARES":    LoadSecurityShares,
		"BACKFILL_SECURITY_STATS": BackfillSecurityStats,
		"LOAD_INDICES":            LoadIndices,
		"LOAD_INDEX_VALUES":       LoadIndexValues,
		"LOAD_INDEX_STATS":        LoadIndexStats,
		"BACKFILL_INDEX_STATS":    BackfillIndexStats,
	}

	marketDataJobType, ok := conversionMap[str]
	if !ok {
		return 0, http.ErrorEntityNotFound{Name: "market-data-job-types", Value: str}
	}

	return marketDataJobType, nil
}
