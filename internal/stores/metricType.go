package stores

const (
	SMA MetricType = iota
	EMA
	ROC
	ATR
	VMA
)

type MetricType int

var metricTypeNames = map[MetricType]string{ //nolint:gochecknoglobals // static display-name lookup
	SMA: "SMA",
	EMA: "EMA",
	ROC: "ROC",
	ATR: "ATR",
	VMA: "VMA",
}

func (m MetricType) String() string {
	return metricTypeNames[m]
}
