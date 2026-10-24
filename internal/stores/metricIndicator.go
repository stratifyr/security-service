package stores

const (
	Trend MetricIndicator = iota
	Momentum
	Volatility
	Volume
)

type MetricIndicator int

var metricIndicatorNames = map[MetricIndicator]string{ //nolint:gochecknoglobals // static display-name lookup
	Trend:      "Trend",
	Momentum:   "Momentum",
	Volatility: "Volatility",
	Volume:     "Volume",
}

func (m MetricIndicator) String() string {
	return metricIndicatorNames[m]
}
