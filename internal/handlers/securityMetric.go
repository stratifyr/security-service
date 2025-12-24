package handlers

import (
	"fmt"
	"strconv"
	"time"

	"gofr.dev/pkg/gofr"
	"gofr.dev/pkg/gofr/http"
	"gofr.dev/pkg/gofr/http/response"

	"github.com/stratifyr/security-service/internal/services"
)

type SecurityMetric struct {
	MetricID  int    `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	Period    int    `json:"period"`
	Indicator string `json:"indicator"`
	Tier      int    `json:"tier"`
	Value     string `json:"value"`
}

type securityMetricHandler struct {
	svc services.SecurityMetricService
}

func NewSecurityMetricHandler(svc services.SecurityMetricService) *securityMetricHandler {
	return &securityMetricHandler{svc: svc}
}

func (h *securityMetricHandler) Get(ctx *gofr.Context) (interface{}, error) {
	securityID, err := strconv.Atoi(ctx.PathParam("id"))
	if err != nil {
		return nil, http.ErrorInvalidParam{Params: []string{"id"}}
	}

	date, err := time.Parse(time.DateOnly, ctx.Param("date"))
	if err != nil {
		return nil, http.ErrorInvalidParam{Params: []string{"date"}}
	}

	securityMetrics, err := h.svc.Get(ctx, securityID, date)
	if err != nil {
		return nil, err
	}

	var resp = make([]*SecurityMetric, len(securityMetrics))

	for i := range securityMetrics {
		resp[i] = h.buildResp(securityMetrics[i])
	}

	return response.Raw{Data: map[string]any{
		"data": resp,
	}}, nil
}

func (h *securityMetricHandler) buildResp(model *services.SecurityMetric) *SecurityMetric {
	resp := &SecurityMetric{
		MetricID:  model.Metric.ID,
		Name:      model.Metric.Name,
		Type:      model.Metric.Type,
		Period:    model.Metric.Period,
		Indicator: model.Metric.Indicator,
		Tier:      model.Metric.Tier,
		Value:     fmt.Sprintf("%0.2f", model.Value),
	}

	return resp
}
