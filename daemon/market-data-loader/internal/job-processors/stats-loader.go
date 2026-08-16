package jobprocessors

import (
	"fmt"
	"slices"
	"time"

	"gofr.dev/pkg/gofr"

	dataProviders "github.com/stratifyr/security-service/daemon/stats-loader/internal/data-providers"
)

type statsLoader struct {
	dataProvider dataProviders.Provider
}

func NewStatsLoader(dataProvider dataProviders.Provider) JobProcessor {
	return &statsLoader{dataProvider}
}

func (s *statsLoader) Process(ctx *gofr.Context) (logs *Logs, err error) {
	logs = initializeJobLogs(LoadSecurityStats)
	defer func() { recordJobCompletionLogs(logs, err) }()

	today := time.Now()

	marketDays, err := getMarketDays(ctx, today, today)
	if err != nil {
		return logs, err
	}

	if len(marketDays) != 1 || marketDays[0].Format(time.DateOnly) != today.Format(time.DateOnly) {
		return logs, fmt.Errorf("cannot load stats on market holiday %v", today.Format(time.DateOnly))
	}

	symbolFilter := ctx.Param("symbol")

	securities, securityIDMap, err := getSecurityDetails(ctx, symbolFilter)
	if err != nil {
		return logs, err
	}

	if symbolFilter != "" && !slices.Contains(securities, symbolFilter) {
		return logs, fmt.Errorf("security not found with symbol %s", symbolFilter)
	}

	ohlcData, err := s.dataProvider.OHLC(ctx, securities)
	if err != nil {
		return logs, fmt.Errorf("failed to get ohlc data, err: %v", err)
	}

	for i := range securities {
		ohlc, ok := ohlcData[securities[i]]
		if !ok {
			logs.Errors = append(logs.Errors, fmt.Sprintf("%s ohlc data not found", securities[i]))
			continue
		}

		if err = createSecurityStat(ctx, securityIDMap[securities[i]], today, ohlc); err != nil {
			logs.Errors = append(logs.Errors, fmt.Sprintf("%s %v", securities[i], err))
			continue
		}

		logs.Success = append(logs.Success, fmt.Sprintf("%s %s", securities[i], ohlc))
	}

	return logs, nil
}
