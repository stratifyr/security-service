package jobprocessors

import (
	"fmt"
	"slices"

	"gofr.dev/pkg/gofr"

	dataProviders "github.com/stratifyr/security-service/daemon/stats-loader/internal/data-providers"
)

type ltpLoader struct {
	dataProvider dataProviders.Provider
}

func NewLtpLoader(dataProvider dataProviders.Provider) JobProcessor {
	return &ltpLoader{dataProvider: dataProvider}
}

func (l *ltpLoader) Process(ctx *gofr.Context) (logs *Logs, err error) {
	logs = initializeJobLogs(LoadLTP)
	defer func() { recordJobCompletionLogs(logs, err) }()

	symbolFilter := ctx.Param("symbol")

	securities, securityIDMap, err := getSecurityDetails(ctx, symbolFilter)
	if err != nil {
		return logs, err
	}

	if symbolFilter != "" && !slices.Contains(securities, symbolFilter) {
		return logs, fmt.Errorf("security not found with symbol %v", symbolFilter)
	}

	ltpMap, err := l.dataProvider.LTP(ctx, securities)
	if err != nil {
		return logs, fmt.Errorf("failed to get ltp data, err: %v", err)
	}

	for i := range securities {
		ltp, ok := ltpMap[securities[i]]
		if !ok {
			logs.Errors = append(logs.Errors, fmt.Sprintf("%s ltp data not found", securities[i]))
			continue
		}

		if err = updateLTP(ctx, securityIDMap[securities[i]], ltp); err != nil {
			logs.Errors = append(logs.Errors, fmt.Sprint(securities[i], err))
			continue
		}

		logs.Success = append(logs.Success, fmt.Sprintf("%s %0.2f", securities[i], ltp))
	}

	return logs, nil
}
