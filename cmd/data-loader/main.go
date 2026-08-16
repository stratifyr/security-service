package main

import (
	_ "embed"

	"gofr.dev/pkg/gofr"

	marketholidays "github.com/stratifyr/security-service/cmd/data-loader/internal/market-holidays"
	"github.com/stratifyr/security-service/cmd/data-loader/internal/metrics"
	"github.com/stratifyr/security-service/cmd/data-loader/internal/securities"
)

func main() {
	app := gofr.NewCMD()

	app.AddHTTPService("security-service", app.Config.Get("SECURITY_SERVICE_HOST"))

	securitiesCMDHandler := securities.NewCMDHandler()
	metricsCMDHandler := metrics.NewCMDHandler()
	marketHolidaysCMDHandler := marketholidays.NewCMDHandler()

	app.SubCommand("load securities", securitiesCMDHandler.Load)
	app.SubCommand("load metrics", metricsCMDHandler.Load)
	app.SubCommand("load market-holidays", marketHolidaysCMDHandler.Load)

	app.Run()
}
