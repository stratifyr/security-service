package handlers

import (
	"time"

	"github.com/stratifyr/security-service-proto/go/pb"
	"gofr.dev/pkg/gofr"

	"github.com/stratifyr/security-service/internal/services"
)

type marketDayGRPCHandler struct {
	svc services.MarketDayService
}

func NewMarketDayGRPCHandler(svc services.MarketDayService) *marketDayGRPCHandler {
	return &marketDayGRPCHandler{svc: svc}
}

func (h *marketDayGRPCHandler) Index(ctx *gofr.Context) (any, error) {
	var payload pb.GetMarketDaysRequest

	if err := ctx.Bind(&payload); err != nil {
		return nil, err
	}

	startDate, err := time.Parse(time.DateOnly, payload.StartDate)
	if err != nil {
		return nil, err
	}

	endDate, err := time.Parse(time.DateOnly, payload.EndDate)
	if err != nil {
		return nil, err
	}

	filter := &services.MarketDayFilter{
		DateBetween: &struct {
			StartDate time.Time
			EndDate   time.Time
		}{StartDate: startDate, EndDate: endDate},
	}

	marketDays, _, err := h.svc.Index(ctx, filter)
	if err != nil {
		return nil, err
	}

	resp := &pb.GetMarketDaysResponse{
		Days: make([]string, len(marketDays)),
	}

	for i := range marketDays {
		resp.Days[i] = marketDays[i].Format(time.DateOnly)
	}

	return resp, nil
}
