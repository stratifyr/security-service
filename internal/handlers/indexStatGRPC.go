package handlers

import (
	"time"

	"github.com/stratifyr/security-service-proto/go/pb"
	"gofr.dev/pkg/gofr"

	"github.com/stratifyr/security-service/internal/services"
)

type indexStatGRPCHandler struct {
	svc services.IndexStatService
}

func NewIndexStatGRPCHandler(svc services.IndexStatService) *indexStatGRPCHandler {
	return &indexStatGRPCHandler{svc: svc}
}

func (h *indexStatGRPCHandler) Upsert(ctx *gofr.Context) (any, error) {
	var payload pb.UpsertIndexStatRequest

	if err := ctx.Bind(&payload); err != nil {
		return nil, err
	}

	date, err := time.Parse(time.DateOnly, payload.Date)
	if err != nil {
		return nil, err
	}

	model := &services.IndexStatUpsert{
		IndexID: int(payload.IndexId),
		Date:    date,
		Open:    payload.Open,
		Close:   payload.Close,
		High:    payload.High,
		Low:     payload.Low,
	}

	index, err := h.svc.Upsert(ctx, model)
	if err != nil {
		return nil, err
	}

	return &pb.UpsertIndexStatResponse{
		IndexStat: h.buildResponse(index),
	}, nil
}

func (*indexStatGRPCHandler) buildResponse(model *services.IndexStat) *pb.IndexStat {
	return &pb.IndexStat{
		Id:        int32(model.ID),
		IndexId:   int32(model.IndexID),
		Date:      model.Date.Format(time.DateOnly),
		Open:      model.Open,
		Close:     model.Close,
		High:      model.High,
		Low:       model.Low,
		CreatedAt: model.CreatedAt.Format(time.RFC3339),
		UpdatedAt: model.UpdatedAt.Format(time.RFC3339),
	}
}
