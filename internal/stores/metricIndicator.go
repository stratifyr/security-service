package stores

import (
	"gofr.dev/pkg/gofr"
	"gofr.dev/pkg/gofr/http"
)

type MetricIndicatorStore interface {
	Index(ctx *gofr.Context) []MetricIndicator
}

const (
	Trend MetricIndicator = iota
	Momentum
	Volatility
	Volume
)

type MetricIndicator int

type metricIndicatorStore struct{}

func NewMetricIndicatorStore() *metricIndicatorStore {
	return &metricIndicatorStore{}
}

func (s *metricIndicatorStore) Index(ctx *gofr.Context) []MetricIndicator {
	return []MetricIndicator{
		Trend,
		Momentum,
		Volatility,
		Volume,
	}
}

func (m MetricIndicator) String() string {
	var conversionMap = map[MetricIndicator]string{
		Trend:      "Trend",
		Momentum:   "Momentum",
		Volatility: "Volatility",
		Volume:     "Volume",
	}

	return conversionMap[m]
}

func MetricIndicatorFromString(str string) (MetricIndicator, error) {
	var conversionMap = map[string]MetricIndicator{
		"Trend":      Trend,
		"Momentum":   Momentum,
		"Volatility": Volatility,
		"Volume":     Volume,
	}

	metricIndicator, ok := conversionMap[str]
	if !ok {
		return 0, http.ErrorEntityNotFound{Name: "metric-indicator", Value: str}
	}

	return metricIndicator, nil
}
