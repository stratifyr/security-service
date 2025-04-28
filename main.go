package main

import (
	"gofr.dev/pkg/gofr"

	"github.com/stratifyr/security-service/handlers"
	"github.com/stratifyr/security-service/migrations"
	"github.com/stratifyr/security-service/services"
	"github.com/stratifyr/security-service/stores"
)

func main() {
	app := gofr.New()

	app.Migrate(migrations.All())

	securityStore := stores.NewSecurityStore()
	securityService := services.NewSecurityService(securityStore)
	securityHandler := handlers.NewSecurityHandler(securityService)

	app.GET("/securities", securityHandler.Index)
	app.GET("/securities/{id}", securityHandler.Read)
	app.POST("/securities", securityHandler.Create)

	app.Run()
}
