package handlers

import (
	"time"

	"github.com/stratifyr/security-service-proto/go/pb"
	"gofr.dev/pkg/gofr"

	"github.com/stratifyr/security-service/internal/services"
)

type exchangeGRPCHandler struct {
	svc services.ExchangeService
}

func NewExchangeGRPCHandler(svc services.ExchangeService) *exchangeGRPCHandler {
	return &exchangeGRPCHandler{svc: svc}
}

func (h *exchangeGRPCHandler) Upsert(ctx *gofr.Context) (any, error) {
	var payload pb.UpsertExchangeRequest

	if err := ctx.Bind(&payload); err != nil {
		return nil, err
	}

	model := &services.ExchangeUpsert{
		Name:      payload.Name,
		Currency:  payload.Currency,
		ForexRate: payload.ForexRate,
	}

	exchange, err := h.svc.Upsert(ctx, model)
	if err != nil {
		return nil, err
	}

	return &pb.UpsertExchangeResponse{
		Exchange: h.buildResponse(exchange),
	}, nil
}

func (*exchangeGRPCHandler) buildResponse(model *services.Exchange) *pb.Exchange {
	return &pb.Exchange{
		Id:        int32(model.ID),
		Name:      model.Name,
		Currency:  model.Currency.String(),
		ForexRate: model.ForexRate,
		CreatedAt: model.CreatedAt.Format(time.RFC3339),
		UpdatedAt: model.UpdatedAt.Format(time.RFC3339),
	}
}
