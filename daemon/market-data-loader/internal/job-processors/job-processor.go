package jobprocessors

import (
	"fmt"

	"gofr.dev/pkg/gofr"

	dataProviders "github.com/stratifyr/security-service/daemon/stats-loader/internal/data-providers"
)

const (
	LoadLTP               = "LOAD_LTP"
	LoadSecurityStats     = "LOAD_SECURITY_STATS"
	BackfillSecurityStats = "BACKFILL_SECURITY_STATS"
)

type JobProcessor interface {
	Process(ctx *gofr.Context) (logs *Logs, err error)
}

type Logs struct {
	Job     string            `json:"job"`
	Meta    map[string]string `json:"meta"`
	Success []string          `json:"success"`
	Errors  []string          `json:"errors"`
}

func GetJobProcessor(marketDataJob string, dataProvider dataProviders.Provider) (JobProcessor, error) {
	switch marketDataJob {
	case LoadLTP:
		return NewLtpLoader(dataProvider), nil
	case LoadSecurityStats:
		return NewStatsLoader(dataProvider), nil
	case BackfillSecurityStats:
		return NewStatsBackfiller(dataProvider), nil
	default:
		return nil, fmt.Errorf("invalid market data job type: %s", marketDataJob)
	}
}
