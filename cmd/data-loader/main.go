package main

import (
	_ "embed"
	"fmt"

	"gofr.dev/pkg/gofr"

	"github.com/stratifyr/security-service/cmd/data-loader/internal/data-providers"
	marketholidays "github.com/stratifyr/security-service/cmd/data-loader/internal/market-holidays"
	"github.com/stratifyr/security-service/cmd/data-loader/internal/metrics"
	"github.com/stratifyr/security-service/cmd/data-loader/internal/securities"
	"github.com/stratifyr/security-service/cmd/data-loader/internal/stats"
)

func main() {
	app := gofr.NewCMD()

	app.AddHTTPService("security-service", app.Config.Get("SECURITY_SERVICE_HOST"))

	client, err := dataProviders.New(app)
	if err != nil {
		fmt.Println("failed to get data provider, err: " + err.Error())
		return
	}

	securitiesCMDHandler := securities.NewCMDHandler()
	metricsCMDHandler := metrics.NewCMDHandler()
	marketHolidaysCMDHandler := marketholidays.NewCMDHandler()
	statsCMDHandler := stats.NewCMDHandler(client)

	app.SubCommand("load securities", securitiesCMDHandler.Load)
	app.SubCommand("load metrics", metricsCMDHandler.Load)
	app.SubCommand("load market-holidays", marketHolidaysCMDHandler.Load)
	app.SubCommand("load ltp", statsCMDHandler.LoadLTP)
	app.SubCommand("load security-stats", statsCMDHandler.LoadSecurityStats)
	app.SubCommand("backfill security-stats", statsCMDHandler.BackFillSecurityStats)

	app.Run()
}
