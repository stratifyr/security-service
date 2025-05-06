package main

import (
	"gofr.dev/pkg/gofr"

	dataProviders "github.com/stratifyr/security-service/crons/marketdata-populator/data-providers"
)

func main() {
	app := gofr.NewCMD()

	// todo: add logic for populating daily market data
	dataProviders.New(app)

	app.Run()
}
