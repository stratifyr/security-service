package stores

import "gofr.dev/pkg/gofr"

type MetricStore interface {
	Index(ctx *gofr.Context) []*Metric
}

type Metric struct {
	ID        int
	Type      MetricType
	Period    int
	Indicator MetricIndicator
}

type metricStore struct{}

func NewMetricStore() *metricStore {
	return &metricStore{}
}

func (*metricStore) Index(_ *gofr.Context) []*Metric {
	metrics := []*Metric{
		{ID: 100, Type: SMA, Period: 5, Indicator: Trend},
		{ID: 101, Type: SMA, Period: 10, Indicator: Trend},
		{ID: 102, Type: SMA, Period: 20, Indicator: Trend},
		{ID: 103, Type: SMA, Period: 50, Indicator: Trend},
		{ID: 104, Type: SMA, Period: 100, Indicator: Trend},
		{ID: 105, Type: SMA, Period: 200, Indicator: Trend},

		{ID: 200, Type: EMA, Period: 5, Indicator: Trend},
		{ID: 201, Type: EMA, Period: 10, Indicator: Trend},
		{ID: 202, Type: EMA, Period: 20, Indicator: Trend},
		{ID: 203, Type: EMA, Period: 50, Indicator: Trend},
		{ID: 204, Type: EMA, Period: 100, Indicator: Trend},
		{ID: 205, Type: EMA, Period: 200, Indicator: Trend},

		{ID: 300, Type: ROC, Period: 5, Indicator: Momentum},
		{ID: 301, Type: ROC, Period: 10, Indicator: Momentum},
		{ID: 302, Type: ROC, Period: 20, Indicator: Momentum},
		{ID: 303, Type: ROC, Period: 50, Indicator: Momentum},
		{ID: 304, Type: ROC, Period: 100, Indicator: Momentum},
		{ID: 305, Type: ROC, Period: 200, Indicator: Momentum},

		{ID: 400, Type: ATR, Period: 5, Indicator: Volatility},
		{ID: 401, Type: ATR, Period: 10, Indicator: Volatility},
		{ID: 402, Type: ATR, Period: 20, Indicator: Volatility},
		{ID: 403, Type: ATR, Period: 50, Indicator: Volatility},
		{ID: 404, Type: ATR, Period: 100, Indicator: Volatility},
		{ID: 405, Type: ATR, Period: 200, Indicator: Volatility},

		{ID: 500, Type: VMA, Period: 5, Indicator: Volume},
		{ID: 501, Type: VMA, Period: 10, Indicator: Volume},
		{ID: 502, Type: VMA, Period: 20, Indicator: Volume},
		{ID: 503, Type: VMA, Period: 50, Indicator: Volume},
		{ID: 504, Type: VMA, Period: 100, Indicator: Volume},
		{ID: 505, Type: VMA, Period: 200, Indicator: Volume},
	}

	return metrics
}
