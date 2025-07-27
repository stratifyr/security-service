package handlers

import (
	"gofr.dev/pkg/gofr"
	"gofr.dev/pkg/gofr/http/response"

	"github.com/stratifyr/security-service/internal/services"
)

type industryHandler struct {
	svc services.IndustryService
}

func NewIndustryHandler(svc services.IndustryService) *industryHandler {
	return &industryHandler{svc: svc}
}

func (h *industryHandler) Index(ctx *gofr.Context) (interface{}, error) {
	industries := h.svc.Index(ctx)

	return response.Raw{Data: map[string]any{
		"data": industries,
	}}, nil
}
