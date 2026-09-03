package handlers

import (
	"github.com/stratifyr/security-service-proto/go/pb"
	"gofr.dev/pkg/gofr"

	"github.com/stratifyr/security-service/internal/services"
)

type metricGRPCHandler struct {
	svc services.MetricService
}

func NewMetricGRPCHandler(svc services.MetricService) *metricGRPCHandler {
	return &metricGRPCHandler{svc: svc}
}

func (h *metricGRPCHandler) Index(ctx *gofr.Context) (any, error) {
	var payload pb.GetMetricsRequest

	if err := ctx.Bind(&payload); err != nil {
		return nil, err
	}

	metrics := h.svc.Index(ctx)

	return h.buildResponse(metrics), nil
}

func (*metricGRPCHandler) buildResponse(metrics []*services.Metric) *pb.GetMetricsResponse {
	resp := &pb.GetMetricsResponse{
		Metrics: make([]*pb.Metric, len(metrics)),
		Total:   int32(len(metrics)),
	}

	for i := range resp.Metrics {
		resp.Metrics[i] = &pb.Metric{
			Id:        int32(metrics[i].ID),
			Name:      metrics[i].Name,
			Type:      metrics[i].Type.String(),
			Period:    int32(metrics[i].Period),
			Indicator: metrics[i].Indicator.String(),
		}
	}

	return resp
}
