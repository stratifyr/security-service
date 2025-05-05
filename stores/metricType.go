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
		SMA: "Simple Moving Average",
		EMA: "Exponential Moving Average",
		RSI: "Relative Strength Index",
		ROC: "Rate of Change of Price",
		ATR: "Average True Range",
		VMA: "Volume Moving Average",
	}

	return conversionMap[m]
}

func MetricTypeFromString(str string) (MetricType, error) {
	var conversionMap = map[string]MetricType{
		"Simple Moving Average":      SMA,
		"Exponential Moving Average": EMA,
		"Relative Strength Index":    RSI,
		"Rate of Change of Price":    ROC,
		"Average True Range":         ATR,
		"Volume Moving Average":      VMA,
	}

	metricType, ok := conversionMap[str]
	if !ok {
		return 0, http.ErrorEntityNotFound{Name: "metric-types", Value: str}
	}

	return metricType, nil
}
