package main

import (
	"gofr.dev/pkg/gofr"

	client "github.com/stratifyr/security-service-client"

	dataProviders "github.com/stratifyr/security-service/daemon/stats-loader/internal/data-providers"
	"github.com/stratifyr/security-service/daemon/stats-loader/internal/job-processors"
)

func main() {
	app := gofr.New()

	dataProvider, err := dataProviders.New(app)
	if err != nil {
		app.Logger().Fatalf("failed to get data provider, err: " + err.Error())
	}

	securityServiceClient, err := client.NewSecurityServiceClient(app.Config, app.Metrics())
	if err != nil {
		app.Logger().Fatalf("failed to create security service client: %s", err)
	}

	daemon := jobprocessors.NewDaemon(dataProvider, securityServiceClient)

	app.OnStart(daemon.Start)

	app.Run()
}
