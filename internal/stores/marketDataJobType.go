package stores

import (
	"gofr.dev/pkg/gofr/http"
)

const (
	LoadLTP = iota
	LoadSecurityStats
	BackfillSecurityStats
	LoadIndices
	LoadVolume
	LoadFreeFloatShares
)

type MarketDataJobType int

func (m MarketDataJobType) String() string {
	var conversionMap = map[MarketDataJobType]string{
		LoadLTP:               "LOAD_LTP",
		LoadSecurityStats:     "LOAD_SECURITY_STATS",
		BackfillSecurityStats: "BACKFILL_SECURITY_STATS",
		LoadIndices:           "LOAD_INDICES",
		LoadVolume:            "LOAD_VOLUME",
		LoadFreeFloatShares:   "LOAD_FREE_FLOAT_SHARES",
	}

	return conversionMap[m]
}

func MarketDataJobTypeFromString(str string) (MarketDataJobType, error) {
	var conversionMap = map[string]MarketDataJobType{
		"LOAD_LTP":                LoadLTP,
		"LOAD_SECURITY_STATS":     LoadSecurityStats,
		"BACKFILL_SECURITY_STATS": BackfillSecurityStats,
		"LOAD_INDICES":            LoadIndices,
		"LOAD_VOLUME":             LoadVolume,
		"LOAD_FREE_FLOAT_SHARES":  LoadFreeFloatShares,
	}

	marketDataJobType, ok := conversionMap[str]
	if !ok {
		return 0, http.ErrorEntityNotFound{Name: "market-data-job-types", Value: str}
	}

	return marketDataJobType, nil
}
