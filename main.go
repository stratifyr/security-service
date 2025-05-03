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
	securityStatStore := stores.NewSecurityStatStore()
	metricStore := stores.NewMetricStore()
	securityMetricStore := stores.NewSecurityMetricStore()
	universeSecurityStore := stores.NewUniverseSecurityStore()
	universeStore := stores.NewUniverseStore(universeSecurityStore)

	industryService := services.NewIndustryService(industryStore)
	securityStatService := services.NewSecurityStatService(securityStatStore)
	metricService := services.NewMetricService(metricStore)
	securityMetricService := services.NewSecurityMetricService(securityMetricStore)
	securityService := services.NewSecurityService(securityStatService, metricService, securityMetricService, securityStore)
	universeService := services.NewUniverseService(securityService, universeStore)
	universeSecurityService := services.NewUniverseSecurityService(securityService, universeService, universeSecurityStore)

	industryHandler := handlers.NewIndustryHandler(industryService)
	metricHandler := handlers.NewMetricHandler(metricService)
	securityHandler := handlers.NewSecurityHandler(securityService)
	securityStatHandler := handlers.NewSecurityStatHandler(securityStatService)
	securityMetricHandler := handlers.NewSecurityMetricHandler(securityMetricService)
	universeHandler := handlers.NewUniverseHandler(universeService)
	universeSecurityHandler := handlers.NewUniverseSecurityHandler(universeSecurityService)

	app.GET("/industries", industryHandler.Index)

	app.GET("/metrics", metricHandler.Index)
	app.POST("/metrics", metricHandler.Create)
	app.GET("/metrics/{id}", metricHandler.Read)
	app.PATCH("/metrics/{id}", metricHandler.Patch)

	app.GET("/securities", securityHandler.Index)
	app.POST("/securities", securityHandler.Create)
	app.GET("/securities/{id}", securityHandler.Read)
	app.PATCH("/securities/{id}", securityHandler.Patch)

	app.GET("/security-stats", securityStatHandler.Index)
	app.POST("/security-stats", securityStatHandler.Create)
	app.GET("/security-stats/{id}", securityStatHandler.Read)
	app.PATCH("/security-stats/{id}", securityStatHandler.Patch)

	app.GET("/security-metrics", securityMetricHandler.Index)
	app.POST("/security-metrics", securityMetricHandler.Create)
	app.GET("/security-metrics/{id}", securityMetricHandler.Read)
	app.PATCH("/security-metrics/{id}", securityMetricHandler.Patch)

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
