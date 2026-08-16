package handlers

import (
	"time"

	"github.com/stratifyr/security-service-proto/go/pb"
	"gofr.dev/pkg/gofr"

	"github.com/stratifyr/security-service/internal/services"
)

type securityGRPCHandler struct {
	svc services.SecurityService
}

func NewSecurityGRPCHandler(svc services.SecurityService) *securityGRPCHandler {
	return &securityGRPCHandler{svc: svc}
}

func (h *securityGRPCHandler) Index(ctx *gofr.Context) (any, error) {
	var payload pb.GetSecuritiesRequest

	if err := ctx.Bind(&payload); err != nil {
		return nil, err
	}

	var (
		filter services.SecurityFilter
		err    error
	)

	if payload.Date != "" {
		filter.Date, err = time.Parse(time.DateOnly, payload.Date)
		if err != nil {
			return nil, err
		}
	}

	securities, count, err := h.svc.Index(ctx, &filter, 0, 0)
	if err != nil {
		return nil, err
	}

	var resp = make([]*pb.Security, len(securities))

	for i := range securities {
		resp[i] = h.buildResponse(securities[i])
	}

	return &pb.GetSecuritiesResponse{
		Securities: resp,
		Total:      int32(count),
	}, nil
}

func (h *securityGRPCHandler) Patch(ctx *gofr.Context) (any, error) {
	var payload pb.UpdateSecurityRequest

	if err := ctx.Bind(&payload); err != nil {
		return nil, err
	}

	model := &services.SecurityUpdate{
		UserID:   1,
		Symbol:   payload.Symbol,
		Industry: payload.Industry,
		Name:     payload.Name,
		Image:    payload.Image,
		LTP:      payload.Ltp,
	}

	security, err := h.svc.Patch(ctx, int(payload.Id), model)
	if err != nil {
		return nil, err
	}

	return &pb.UpdateSecurityResponse{
		Security: h.buildResponse(security),
	}, nil
}

func (h *securityGRPCHandler) buildResponse(securitiy *services.Security) *pb.Security {
	var resp = &pb.Security{
		Id:            int32(securitiy.ID),
		Isin:          securitiy.ISIN,
		Symbol:        securitiy.Symbol,
		Industry:      securitiy.Industry,
		Name:          securitiy.Name,
		Image:         securitiy.Image,
		Ltp:           securitiy.LTP,
		PreviousClose: securitiy.PreviousClose,
		CreatedAt:     securitiy.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     securitiy.UpdatedAt.Format(time.RFC3339),
	}

	if securitiy.SecurityStat == nil {
		return resp
	}

	resp.MarketData = &pb.MarketData{
		Date:    securitiy.SecurityStat.Date.Format(time.DateOnly),
		Open:    securitiy.SecurityStat.Open,
		Close:   securitiy.SecurityStat.Close,
		High:    securitiy.SecurityStat.High,
		Low:     securitiy.SecurityStat.Low,
		Volume:  int32(securitiy.SecurityStat.Volume),
		Metrics: make([]*pb.SecurityMetric, len(securitiy.SecurityMetrics)),
	}

	for j := range resp.MarketData.Metrics {
		resp.MarketData.Metrics[j] = &pb.SecurityMetric{
			Id:              int32(securitiy.SecurityMetrics[j].Metric.ID),
			Name:            securitiy.SecurityMetrics[j].Metric.Name,
			Type:            securitiy.SecurityMetrics[j].Metric.Type.String(),
			Period:          int32(securitiy.SecurityMetrics[j].Metric.Period),
			Indicator:       securitiy.SecurityMetrics[j].Metric.Indicator.String(),
			Value:           securitiy.SecurityMetrics[j].Value,
			NormalizedValue: securitiy.SecurityMetrics[j].ZValue,
		}
	}

	return resp
}
