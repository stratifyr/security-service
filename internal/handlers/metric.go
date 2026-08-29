package handlers

import (
	"gofr.dev/pkg/gofr"
	"gofr.dev/pkg/gofr/http/response"

	"github.com/stratifyr/security-service/internal/services"
)

type Metric struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	Period    int    `json:"period"`
	Indicator string `json:"indicator"`
}

type metricHandler struct {
	svc services.MetricService
}

func NewMetricHandler(svc services.MetricService) *metricHandler {
	return &metricHandler{svc: svc}
}

func (h *metricHandler) Index(ctx *gofr.Context) (interface{}, error) {
	metrics := h.svc.Index(ctx)

	var resp = make([]*Metric, len(metrics))

	for i := range metrics {
		resp[i] = h.buildResp(metrics[i])
	}

	return response.Raw{Data: map[string]any{
		"data": resp,
	}}, nil
}

func (h *metricHandler) buildResp(model *services.Metric) *Metric {
	resp := &Metric{
		ID:        model.ID,
		Name:      model.Name,
		Type:      model.Type.String(),
		Period:    model.Period,
		Indicator: model.Indicator.String(),
	}

	return resp
}
