package handlers

import (
	"encoding/json"
	"time"

	"github.com/stratifyr/security-service-proto/go/pb"
	"gofr.dev/pkg/gofr"

	"github.com/stratifyr/security-service/internal/services"
)

type marketDataJobGRPCHandler struct {
	svc services.MarketDataJobService
}

func NewMarketDataJobGRPCHandler(svc services.MarketDataJobService) *marketDataJobGRPCHandler {
	return &marketDataJobGRPCHandler{svc: svc}
}

func (h *marketDataJobGRPCHandler) Index(ctx *gofr.Context) (any, error) {
	var payload pb.GetMarketDataJobsRequest

	if err := ctx.Bind(&payload); err != nil {
		return nil, err
	}

	filter := &services.MarketDataJobFilter{
		Status: payload.Status,
	}

	marketDataJobs, count, err := h.svc.Index(ctx, filter, 0, 0)
	if err != nil {
		return nil, err
	}

	var resp = make([]*pb.MarketDataJob, len(marketDataJobs))

	for i := range marketDataJobs {
		resp[i] = h.buildResponse(marketDataJobs[i])
	}

	return &pb.GetMarketDataJobsResponse{
		MarketDataJobs: resp,
		Total:          int32(count),
	}, nil
}

func (h *marketDataJobGRPCHandler) Patch(ctx *gofr.Context) (any, error) {
	var payload pb.UpdateMarketDataJobRequest

	if err := ctx.Bind(&payload); err != nil {
		return nil, err
	}

	var logs *json.RawMessage

	if payload.Logs != nil {
		raw := json.RawMessage(payload.Logs)
		logs = &raw
	}

	model := &services.MarketDataJobUpdate{
		Status: payload.Status,
		Logs:   logs,
	}

	marketDataJob, err := h.svc.Patch(ctx, int(payload.Id), model)
	if err != nil {
		return nil, err
	}

	return &pb.UpdateMarketDataJobResponse{
		MarketDataJob: h.buildResponse(marketDataJob),
	}, nil
}

func (*marketDataJobGRPCHandler) buildResponse(model *services.MarketDataJob) *pb.MarketDataJob {
	var logs []byte

	if model.Logs != nil {
		logs = *model.Logs
	}

	return &pb.MarketDataJob{
		Id:        int32(model.ID),
		Type:      model.Type.String(),
		Status:    model.Status,
		Logs:      logs,
		CreatedAt: model.CreatedAt.Format(time.RFC3339),
		UpdatedAt: model.UpdatedAt.Format(time.RFC3339),
	}
}
