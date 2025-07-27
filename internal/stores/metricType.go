package stores

import (
	"gofr.dev/pkg/gofr"
	"gofr.dev/pkg/gofr/http"
)

type MetricTypeStore interface {
	Index(ctx *gofr.Context) []MetricType
}

const (
	SMA MetricType = iota
	EMA
	RSI
	ROC
	ATR
	VMA
)

type MetricType int

type metricTypeStore struct{}

func NewMetricTypeStore() *metricTypeStore {
	return &metricTypeStore{}
}

func (s *metricTypeStore) Index(ctx *gofr.Context) []MetricType {
	return []MetricType{
		SMA,
		EMA,
		RSI,
		ROC,
		ATR,
		VMA,
	}
}

func (m MetricType) String() string {
	var conversionMap = map[MetricType]string{
		SMA: "SMA",
		EMA: "EMA",
		RSI: "RSI",
		ROC: "ROC",
		ATR: "ATR",
		VMA: "VMA",
	}

	return conversionMap[m]
}

func MetricTypeFromString(str string) (MetricType, error) {
	var conversionMap = map[string]MetricType{
		"SMA": SMA,
		"EMA": EMA,
		"RSI": RSI,
		"ROC": ROC,
		"ATR": ATR,
		"VMA": VMA,
	}

	metricType, ok := conversionMap[str]
	if !ok {
		return 0, http.ErrorEntityNotFound{Name: "metric-types", Value: str}
	}

	return metricType, nil
}
