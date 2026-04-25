package handlers

import (
	"github.com/stratifyr/security-service/proto"
	"gofr.dev/pkg/gofr"
)

func NewSecurityServiceGoFrGRPCHandler(marketDayGRPCHandler *marketDayGRPCHandler, metricGRPCHandler *metricGRPCHandler,
	securityGRPCHandler *securityGRPCHandler) *SecurityServiceGoFrGRPCHandler {
	return &SecurityServiceGoFrGRPCHandler{
		marketDayGRPCHandler: marketDayGRPCHandler,
		metricGRPCHandler:    metricGRPCHandler,
		securityGRPCHandler:  securityGRPCHandler,
	}
}

type SecurityServiceGoFrGRPCHandler struct {
	marketDayGRPCHandler *marketDayGRPCHandler
	metricGRPCHandler    *metricGRPCHandler
	securityGRPCHandler  *securityGRPCHandler

	proto.UnimplementedSecurityServiceServer
}

func (h *SecurityServiceGoFrGRPCHandler) GetMarketDays(ctx *gofr.Context) (any, error) {
	return h.marketDayGRPCHandler.Index(ctx)
}

func (h *SecurityServiceGoFrGRPCHandler) GetSecurities(ctx *gofr.Context) (any, error) {
	return h.securityGRPCHandler.Index(ctx)
}

func (h *SecurityServiceGoFrGRPCHandler) GetMetrics(ctx *gofr.Context) (any, error) {
	return h.metricGRPCHandler.Index(ctx)
}
