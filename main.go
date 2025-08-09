package main

import (
	"gofr.dev/pkg/gofr"

	"github.com/stratifyr/security-service/internal/handlers"
	"github.com/stratifyr/security-service/internal/handlers/grpc"
	"github.com/stratifyr/security-service/internal/services"
	"github.com/stratifyr/security-service/internal/stores"
	"github.com/stratifyr/security-service/migrations"
)

func main() {
	app := gofr.New()

	app.AddHTTPService("account-service", app.Config.Get("ACCOUNT_SERVICE_HOST"))

	app.Migrate(migrations.All())

	industryStore := stores.NewIndustryStore()
	metricStore := stores.NewMetricStore()
	securityStore := stores.NewSecurityStore()
	marketHolidayStore := stores.NewMarketHolidayStore()
	securityStatStore := stores.NewSecurityStatStore()
	securityMetricStore := stores.NewSecurityMetricStore()

	industryService := services.NewIndustryService(industryStore)
	metricService := services.NewMetricService(metricStore)
	marketHolidayService := services.NewMarketHolidayService(marketHolidayStore)
	marketDayService := services.NewMarketDayService(marketHolidayStore)
	securityStatService := services.NewSecurityStatService(marketDayService, securityStatStore)
	securityMetricService := services.NewSecurityMetricService(marketDayService, metricStore, securityStatStore, securityMetricStore)
	securityService := services.NewSecurityService(marketDayService, metricStore, securityMetricStore, securityStatStore, securityStore)

	industryHandler := handlers.NewIndustryHandler(industryService)
	metricHandler := handlers.NewMetricHandler(metricService)
	marketHolidayHandler := handlers.NewMarketHolidayHandler(marketHolidayService)
	marketDayHandler := handlers.NewMarketDayHandler(marketDayService)
	securityHandler := handlers.NewSecurityHandler(securityService)
	securityStatHandler := handlers.NewSecurityStatHandler(securityStatService)
	securityMetricHandler := handlers.NewSecurityMetricHandler(securityMetricService)

	grpc.RegisterSecurityServiceServerWithGofr(app, grpc.NewSecurityServiceGoFrServer(securityService))

	app.GET("/industries", industryHandler.Index)

	app.GET("/metrics", metricHandler.Index)
	app.POST("/metrics", metricHandler.Create)
	app.GET("/metrics/{id}", metricHandler.Read)
	app.PATCH("/metrics/{id}", metricHandler.Patch)

	app.GET("/market-holidays", marketHolidayHandler.Index)
	app.POST("/market-holidays", marketHolidayHandler.Create)
	app.GET("/market-holidays/{id}", marketHolidayHandler.Read)
	app.PATCH("/market-holidays/{id}", marketHolidayHandler.Patch)
	app.DELETE("/market-holidays/{id}", marketHolidayHandler.Delete)

	app.GET("/market-days", marketDayHandler.Index)

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

	app.Run()
}
