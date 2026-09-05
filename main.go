package main

import (
	gofrWrapper "github.com/stratifyr/security-service-proto/go/gofr-wrapper"
	"gofr.dev/pkg/gofr"

	"github.com/stratifyr/security-service/internal/handlers"
	"github.com/stratifyr/security-service/internal/middlewares"
	"github.com/stratifyr/security-service/internal/services"
	"github.com/stratifyr/security-service/internal/stores"
	"github.com/stratifyr/security-service/migrations"
)

//nolint:funlen // app setup requires many statements
func main() {
	app := gofr.New()

	app.Migrate(migrations.All())

	industryStore := stores.NewIndustryStore()
	metricStore := stores.NewMetricStore()
	securityStore := stores.NewSecurityStore()
	marketHolidayStore := stores.NewMarketHolidayStore()
	securityStatStore := stores.NewSecurityStatStore()
	securityMetricStore := stores.NewSecurityMetricStore()
	marketDataJobStore := stores.NewMarketDataJobStore()

	industryService := services.NewIndustryService(industryStore)
	metricService := services.NewMetricService(metricStore)
	marketHolidayService := services.NewMarketHolidayService(marketHolidayStore)
	marketDayService := services.NewMarketDayService(marketHolidayStore)
	securityStatService := services.NewSecurityStatService(marketDayService, securityStatStore)
	securityMetricService := services.NewSecurityMetricService(marketDayService, metricService, securityStatStore, securityMetricStore)
	securityService := services.NewSecurityService(marketDayService, securityMetricService, securityStatStore, securityStore)
	marketDataJobService := services.NewMarketDataJobService(marketDataJobStore)

	industryHandler := handlers.NewIndustryHandler(industryService)
	metricHandler := handlers.NewMetricHandler(metricService)
	marketHolidayHandler := handlers.NewMarketHolidayHandler(marketHolidayService)
	marketDayHandler := handlers.NewMarketDayHandler(marketDayService)
	securityHandler := handlers.NewSecurityHandler(securityService)
	securityStatHandler := handlers.NewSecurityStatHandler(securityStatService)
	marketDataJobHandler := handlers.NewMarketDataJobHandler(marketDataJobService)

	marketDayGRPCHandler := handlers.NewMarketDayGRPCHandler(marketDayService)
	metricGRPCHandler := handlers.NewMetricGRPCHandler(metricService)
	securityGRPCHandler := handlers.NewSecurityGRPCHandler(securityService)
	securityStatGRPCHandler := handlers.NewSecurityStatGRPCHandler(securityStatService)
	marketDataJobGRPCHandler := handlers.NewMarketDataJobGRPCHandler(marketDataJobService)
	securityServiceGRPCHandler := handlers.NewSecurityServiceGoFrGRPCHandler(marketDayGRPCHandler,
		metricGRPCHandler, securityGRPCHandler, securityStatGRPCHandler, marketDataJobGRPCHandler)

	gofrWrapper.RegisterSecurityServiceServerWithGofr(app, securityServiceGRPCHandler)

	app.UseMiddleware(middlewares.RBAC)

	app.GET("/industries", industryHandler.Index)

	app.GET("/metrics", metricHandler.Index)

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

	app.GET("/market-data-jobs", marketDataJobHandler.Index)
	app.POST("/market-data-jobs", marketDataJobHandler.Create)
	app.GET("/market-data-jobs/{id}", marketDataJobHandler.Read)
	app.PATCH("/market-data-jobs/{id}", marketDataJobHandler.Patch)
	app.DELETE("/market-data-jobs/{id}", marketDataJobHandler.Delete)

	app.Run()
}
