package handlers

import (
	"time"

	"gofr.dev/pkg/gofr"

	"github.com/stratifyr/security-service/internal/services"
	"github.com/stratifyr/security-service/proto"
)

type securityStatGRPCHandler struct {
	svc services.SecurityStatService
}

func NewSecurityStatGRPCHandler(svc services.SecurityStatService) *securityStatGRPCHandler {
	return &securityStatGRPCHandler{svc: svc}
}

func (h *securityStatGRPCHandler) Create(ctx *gofr.Context) (any, error) {
	var payload proto.CreateOrUpdateSecurityStatRequest

	if err := ctx.Bind(&payload); err != nil {
		return nil, err
	}

	date, _ := time.Parse(time.DateOnly, payload.Date)

	model := &services.SecurityStatCreate{
		UserID:     1,
		SecurityID: int(payload.SecurityId),
		Date:       date,
		Open:       payload.Open,
		Close:      payload.Close,
		High:       payload.High,
		Low:        payload.Low,
		Volume:     int(payload.Volume),
	}

	securityStat, err := h.svc.Create(ctx, model)
	if err != nil {
		return nil, err
	}

	return &proto.CreateOrUpdateSecurityStatResponse{
		SecurityStat: h.buildResponse(securityStat),
	}, nil
}

func (h *securityStatGRPCHandler) buildResponse(model *services.SecurityStat) *proto.SecurityStat {
	return &proto.SecurityStat{
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
