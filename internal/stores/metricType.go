package stores

const (
	SMA MetricType = iota
	EMA
	ROC
	ATR
	VMA
)

type MetricType int

func (m MetricType) String() string {
	var conversionMap = map[MetricType]string{
		SMA: "SMA",
		EMA: "EMA",
		ROC: "ROC",
		ATR: "ATR",
		VMA: "VMA",
	}

	return conversionMap[m]
}
