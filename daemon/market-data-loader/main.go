package main

import (
	"gofr.dev/pkg/gofr"

	dataProviders "github.com/stratifyr/security-service/daemon/stats-loader/internal/data-providers"
	"github.com/stratifyr/security-service/daemon/stats-loader/internal/job-processors"
)

func main() {
	app := gofr.New()

	app.AddHTTPService("security-service", app.Config.Get("SECURITY_SERVICE_HOST"))

	dataProvider, err := dataProviders.New(app)
	if err != nil {
		app.Logger().Fatalf("failed to get data provider, err: " + err.Error())
	}

	daemon := jobprocessors.NewDaemon(dataProvider)

	app.OnStart(daemon.Start)

	app.Run()
}
