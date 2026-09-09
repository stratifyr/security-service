package handlers

import (
	"gofr.dev/pkg/gofr"
	"gofr.dev/pkg/gofr/http/response"

	"github.com/stratifyr/security-service/internal/services"
)

type marketDataJobTypeHandler struct {
	svc services.MarketDataJobTypeService
}

func NewMarketDataJobTypeHandler(svc services.MarketDataJobTypeService) *marketDataJobTypeHandler {
	return &marketDataJobTypeHandler{svc: svc}
}

func (h *marketDataJobTypeHandler) List(ctx *gofr.Context) (any, error) {
	marketDataJobTypes := h.svc.List(ctx)

	resp := make([]string, len(marketDataJobTypes))

	for i := range marketDataJobTypes {
		resp[i] = marketDataJobTypes[i].String()
	}

	return response.Raw{Data: map[string]any{
		"data": resp,
	}}, nil
}
