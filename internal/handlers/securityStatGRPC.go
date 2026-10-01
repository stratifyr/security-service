package handlers

import (
	"time"

	"github.com/stratifyr/security-service-proto/go/pb"
	"gofr.dev/pkg/gofr"

	"github.com/stratifyr/security-service/internal/services"
)

type securityStatGRPCHandler struct {
	svc services.SecurityStatService
}

func NewSecurityStatGRPCHandler(svc services.SecurityStatService) *securityStatGRPCHandler {
	return &securityStatGRPCHandler{svc: svc}
}

func (h *securityStatGRPCHandler) Upsert(ctx *gofr.Context) (any, error) {
	var payload pb.UpsertSecurityStatRequest

	if err := ctx.Bind(&payload); err != nil {
		return nil, err
	}

	date, _ := time.Parse(time.DateOnly, payload.Date)

	model := &services.SecurityStatUpsert{
		SecurityID: int(payload.SecurityId),
		Date:       date,
		Open:       payload.Open,
		Close:      payload.Close,
		High:       payload.High,
		Low:        payload.Low,
		Volume:     int(payload.Volume),
	}

	securityStat, err := h.svc.Upsert(ctx, model)
	if err != nil {
		return nil, err
	}

	return &pb.UpsertSecurityStatResponse{
		SecurityStat: h.buildResponse(securityStat),
	}, nil
}

func (*securityStatGRPCHandler) buildResponse(model *services.SecurityStat) *pb.SecurityStat {
	return &pb.SecurityStat{
		Id:         int32(model.ID),
		SecurityId: int32(model.SecurityID),
		Date:       model.Date.Format(time.DateOnly),
		Open:       model.Open,
		Close:      model.Close,
		High:       model.High,
		Low:        model.Low,
		Volume:     int32(model.Volume),
		CreatedAt:  model.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  model.UpdatedAt.Format(time.RFC3339),
	}
}
