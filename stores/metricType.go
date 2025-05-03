package stores

import (
	"gofr.dev/pkg/gofr"
	"gofr.dev/pkg/gofr/http"
)

type MetricTypeStore interface {
	Index(ctx *gofr.Context) []MetricType
}

const (
	Trend MetricType = iota
	Momentum
	Volatility
	Volume
)

type MetricType int

type metricTypeStore struct{}

func NewMetricTypeStore() *metricTypeStore {
	return &metricTypeStore{}
}

func (s *metricTypeStore) Index(ctx *gofr.Context) []MetricType {
	return []MetricType{
		Trend,
		Momentum,
		Volatility,
		Volume,
	}
}

func (m MetricType) String() string {
	var conversionMap = map[MetricType]string{
		Trend:      "Trend",
		Momentum:   "Momentum",
		Volatility: "Volatility",
		Volume:     "Volume",
	}

	return conversionMap[m]
}

func MetricTypeFromString(str string) (MetricType, error) {
	var conversionMap = map[string]MetricType{
		"Trend":      Trend,
		"Momentum":   Momentum,
		"Volatility": Volatility,
		"Volume":     Volume,
	}

	metricType, ok := conversionMap[str]
	if !ok {
		return 0, http.ErrorEntityNotFound{Name: "metric-types", Value: str}
	}

	return metricType, nil
}
