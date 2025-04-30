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

	industryStore := stores.NewIndustryStore()
	securityStore := stores.NewSecurityStore()
	universeSecurityStore := stores.NewUniverseSecurityStore()
	universeStore := stores.NewUniverseStore(universeSecurityStore)

	industryService := services.NewIndustryService(industryStore)
	securityService := services.NewSecurityService(securityStore)
	universeService := services.NewUniverseService(universeStore)
	universeSecurityService := services.NewUniverseSecurityService(securityService, universeService, universeSecurityStore)

	industryHandler := handlers.NewIndustryHandler(industryService)
	securityHandler := handlers.NewSecurityHandler(securityService)
	universeHandler := handlers.NewUniverseHandler(universeService)
	universeSecurityHandler := handlers.NewUniverseSecurityHandler(universeSecurityService)

	app.GET("/industries", industryHandler.Index)

	app.GET("/securities", securityHandler.Index)
	app.POST("/securities", securityHandler.Create)
	app.GET("/securities/{id}", securityHandler.Read)
	app.PATCH("/securities/{id}", securityHandler.Patch)

	app.GET("/universes", universeHandler.Index)
	app.POST("/universes", universeHandler.Create)
	app.GET("/universes/{id}", universeHandler.Read)
	app.PATCH("/universes/{id}", universeHandler.Patch)

	app.GET("/universe-securities", universeSecurityHandler.Index)
	app.POST("/universe-securities", universeSecurityHandler.Create)
	app.GET("/universe-securities/{id}", universeSecurityHandler.Read)
	app.PATCH("/universe-securities/{id}", universeSecurityHandler.Patch)
	app.DELETE("/universe-securities/{id}", universeSecurityHandler.Delete)

	app.Run()
}
