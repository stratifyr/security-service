package stores

import (
	"gofr.dev/pkg/gofr/http"
)

const (
	LoadLTP = iota
	LoadSecurityStats
	BackfillSecurityStats
)

type MarketDataJobType int

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
