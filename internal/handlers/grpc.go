package handlers

import (
	"github.com/stratifyr/security-service-proto/go/pb"
	"gofr.dev/pkg/gofr"
)

func NewSecurityServiceGoFrGRPCHandler(marketDayGRPCHandler *marketDayGRPCHandler, metricGRPCHandler *metricGRPCHandler,
	securityGRPCHandler *securityGRPCHandler, securityStatGRPCHandler *securityStatGRPCHandler, indexGRPCHandler *indexGRPCHandler,
	indexStatGRPCHandler *indexStatGRPCHandler, marketDataJobGRPCHandler *marketDataJobGRPCHandler) *SecurityServiceGoFrGRPCHandler {
	return &SecurityServiceGoFrGRPCHandler{
		marketDayGRPCHandler:     marketDayGRPCHandler,
		metricGRPCHandler:        metricGRPCHandler,
		securityGRPCHandler:      securityGRPCHandler,
		securityStatGRPCHandler:  securityStatGRPCHandler,
		indexGRPCHandler:         indexGRPCHandler,
		indexStatGRPCHandler:     indexStatGRPCHandler,
		marketDataJobGRPCHandler: marketDataJobGRPCHandler,
	}
}

type SecurityServiceGoFrGRPCHandler struct {
	marketDayGRPCHandler     *marketDayGRPCHandler
	metricGRPCHandler        *metricGRPCHandler
	securityGRPCHandler      *securityGRPCHandler
	securityStatGRPCHandler  *securityStatGRPCHandler
	indexGRPCHandler         *indexGRPCHandler
	indexStatGRPCHandler     *indexStatGRPCHandler
	marketDataJobGRPCHandler *marketDataJobGRPCHandler

	pb.UnimplementedSecurityServiceServer
}

func (h *SecurityServiceGoFrGRPCHandler) GetMarketDays(ctx *gofr.Context) (any, error) {
	return h.marketDayGRPCHandler.List(ctx)
}

func (h *SecurityServiceGoFrGRPCHandler) GetSecurities(ctx *gofr.Context) (any, error) {
	return h.securityGRPCHandler.List(ctx)
}

func (h *SecurityServiceGoFrGRPCHandler) UpsertSecurity(ctx *gofr.Context) (any, error) {
	return h.securityGRPCHandler.Upsert(ctx)
}

func (h *SecurityServiceGoFrGRPCHandler) UpsertSecurityStat(ctx *gofr.Context) (any, error) {
	return h.securityStatGRPCHandler.Upsert(ctx)
}

func (h *SecurityServiceGoFrGRPCHandler) GetMetrics(ctx *gofr.Context) (any, error) {
	return h.metricGRPCHandler.List(ctx)
}

func (h *SecurityServiceGoFrGRPCHandler) GetIndices(ctx *gofr.Context) (any, error) {
	return h.indexGRPCHandler.List(ctx)
}

func (h *SecurityServiceGoFrGRPCHandler) UpsertIndex(ctx *gofr.Context) (any, error) {
	return h.indexGRPCHandler.Upsert(ctx)
}

func (h *SecurityServiceGoFrGRPCHandler) UpsertIndexStat(ctx *gofr.Context) (any, error) {
	return h.indexStatGRPCHandler.Upsert(ctx)
}

func (h *SecurityServiceGoFrGRPCHandler) GetMarketDataJobs(ctx *gofr.Context) (any, error) {
	return h.marketDataJobGRPCHandler.List(ctx)
}

func (h *SecurityServiceGoFrGRPCHandler) UpdateMarketDataJob(ctx *gofr.Context) (any, error) {
	return h.marketDataJobGRPCHandler.Patch(ctx)
}
