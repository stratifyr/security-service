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

	var resp = make([]string, len(industries))

	for i := range industries {
		resp[i] = industries[i].String()
	}

	return response.Raw{Data: map[string]any{
		"data": resp,
	}}, nil
}
