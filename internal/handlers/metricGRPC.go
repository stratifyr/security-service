package handlers

import (
	"time"

	"github.com/stratifyr/security-service/internal/services"
	"github.com/stratifyr/security-service/proto"
	"gofr.dev/pkg/gofr"
)

type metricGRPCHandler struct {
	svc services.MetricService
}

func NewMetricGRPCHandler(svc services.MetricService) *metricGRPCHandler {
	return &metricGRPCHandler{svc: svc}
}

func (h *metricGRPCHandler) Index(ctx *gofr.Context) (any, error) {
	var payload proto.GetMetricsRequest

	if err := ctx.Bind(&payload); err != nil {
		return nil, err
	}

	metrics, err := h.svc.Index(ctx)
	if err != nil {
		return nil, err
	}

	return h.buildResponse(metrics)
}

func (h *metricGRPCHandler) buildResponse(metrics []*services.Metric) (*proto.GetMetricsResponse, error) {
	resp := &proto.GetMetricsResponse{
		Metrics: make([]*proto.Metric, len(metrics)),
		Total:   int32(len(metrics)),
	}

	for i := range resp.Metrics {
		resp.Metrics[i] = &proto.Metric{
			Id:        int32(metrics[i].ID),
			Name:      metrics[i].Name,
			Type:      metrics[i].Type.String(),
			Period:    int32(metrics[i].Period),
			Indicator: metrics[i].Indicator.String(),
			CreatedAt: metrics[i].CreatedAt.Format(time.RFC3339),
			UpdatedAt: metrics[i].UpdatedAt.Format(time.RFC3339),
		}
	}

	return resp, nil
}
