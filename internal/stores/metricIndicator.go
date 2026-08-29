package stores

const (
	Trend MetricIndicator = iota
	Momentum
	Volatility
	Volume
)

type MetricIndicator int

func (m MetricIndicator) String() string {
	var conversionMap = map[MetricIndicator]string{
		Trend:      "Trend",
		Momentum:   "Momentum",
		Volatility: "Volatility",
		Volume:     "Volume",
	}

	return conversionMap[m]
}
