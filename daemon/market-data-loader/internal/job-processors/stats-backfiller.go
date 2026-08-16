package jobprocessors

import (
	"fmt"
	"slices"
	"time"

	"gofr.dev/pkg/gofr"

	dataProviders "github.com/stratifyr/security-service/daemon/stats-loader/internal/data-providers"
)

type statsBackfiller struct {
	dataProvider dataProviders.Provider
}

func NewStatsBackfiller(dataProvider dataProviders.Provider) JobProcessor {
	return &statsBackfiller{dataProvider: dataProvider}
}

func (s *statsBackfiller) Process(ctx *gofr.Context) (logs *Logs, err error) {
	logs = initializeJobLogs(BackfillSecurityStats)
	defer func() { recordJobCompletionLogs(logs, err) }()

	today := time.Now()
	fourYearsEarlier := today.AddDate(-4, 0, 0)
	startDate, endDate := fourYearsEarlier, today.AddDate(0, 0, -1)
	logs.Meta["backfill_period"] = fmt.Sprintf("%s - %s", startDate.Format(time.DateOnly), endDate.Format(time.DateOnly))

	symbolFilter := ctx.Param("symbol")

	securities, securityIDMap, err := getSecurityDetails(ctx, ctx.Param("symbol"))
	if err != nil {
		return logs, err
	}

	if symbolFilter != "" && !slices.Contains(securities, symbolFilter) {
		return logs, fmt.Errorf("security not found with symbol %s", symbolFilter)
	}

	marketDays, err := getMarketDays(ctx, startDate, endDate)
	if err != nil {
		return logs, err
	}

	for i := range securities {
		historicalData, err := s.dataProvider.HistoricalOHLC(ctx, securities[i], startDate, endDate)
		if err != nil {
			logs.Errors = append(logs.Errors, fmt.Sprintf("%s %v", securities[i], err))
			continue
		}

		slices.SortFunc(marketDays, func(a, b time.Time) int {
			if a.After(b) {
				return -1
			}

			return 1
		})

		for j, date := range marketDays {
			if j == len(marketDays)-1 {
				logs.Success = append(logs.Success, fmt.Sprintf("%s %s - %s",
					securities[i], marketDays[0].Format(time.DateOnly), date.Format(time.DateOnly)))
			}

			idx := slices.IndexFunc(historicalData, func(ohlc *dataProviders.HistoricalOHLC) bool {
				return ohlc.Date.Format(time.DateOnly) == date.Format(time.DateOnly)
			})

			if idx == -1 {
				if j != len(marketDays)-1 {
					logs.Success = append(logs.Success, fmt.Sprintf("%s %s - %s",
						securities[i], marketDays[0].Format(time.DateOnly), date.Format(time.DateOnly)))
				}

				break
			}

			if err = createSecurityStat(ctx, securityIDMap[securities[i]], date, historicalData[idx].OHLCData); err != nil {
				logs.Errors = append(logs.Errors, fmt.Sprintf("%s %s %v", securities[i], date.Format(time.DateOnly), err))
				continue
			}
		}

	}

	return logs, nil
}
