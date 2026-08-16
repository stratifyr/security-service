package handlers

import (
	"github.com/stratifyr/security-service-proto/go/pb"
	"gofr.dev/pkg/gofr"
)

func NewSecurityServiceGoFrGRPCHandler(marketDayGRPCHandler *marketDayGRPCHandler, metricGRPCHandler *metricGRPCHandler,
	securityGRPCHandler *securityGRPCHandler, securityStatGRPCHandler *securityStatGRPCHandler,
	marketDataJobGRPCHandler *marketDataJobGRPCHandler) *SecurityServiceGoFrGRPCHandler {
	return &SecurityServiceGoFrGRPCHandler{
		marketDayGRPCHandler:     marketDayGRPCHandler,
		metricGRPCHandler:        metricGRPCHandler,
		securityGRPCHandler:      securityGRPCHandler,
		securityStatGRPCHandler:  securityStatGRPCHandler,
		marketDataJobGRPCHandler: marketDataJobGRPCHandler,
	}
}

type SecurityServiceGoFrGRPCHandler struct {
	marketDayGRPCHandler     *marketDayGRPCHandler
	metricGRPCHandler        *metricGRPCHandler
	securityGRPCHandler      *securityGRPCHandler
	securityStatGRPCHandler  *securityStatGRPCHandler
	marketDataJobGRPCHandler *marketDataJobGRPCHandler

	pb.UnimplementedSecurityServiceServer
}

func (h *SecurityServiceGoFrGRPCHandler) GetMarketDays(ctx *gofr.Context) (any, error) {
	return h.marketDayGRPCHandler.Index(ctx)
}

func (h *SecurityServiceGoFrGRPCHandler) GetSecurities(ctx *gofr.Context) (any, error) {
	return h.securityGRPCHandler.Index(ctx)
}

func (h *SecurityServiceGoFrGRPCHandler) UpdateSecurity(ctx *gofr.Context) (any, error) {
	return h.securityGRPCHandler.Patch(ctx)
}

func (h *SecurityServiceGoFrGRPCHandler) CreateOrUpdateSecurityStat(ctx *gofr.Context) (any, error) {
	return h.securityStatGRPCHandler.Create(ctx)
}

func (h *SecurityServiceGoFrGRPCHandler) GetMetrics(ctx *gofr.Context) (any, error) {
	return h.metricGRPCHandler.Index(ctx)
}

func (h *SecurityServiceGoFrGRPCHandler) GetMarketDataJobs(ctx *gofr.Context) (any, error) {
	return h.marketDataJobGRPCHandler.Index(ctx)
}

func (h *SecurityServiceGoFrGRPCHandler) UpdateMarketDataJob(ctx *gofr.Context) (any, error) {
	return h.marketDataJobGRPCHandler.Patch(ctx)
}
