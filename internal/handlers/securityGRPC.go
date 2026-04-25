package handlers

import (
	"time"

	"github.com/stratifyr/security-service/internal/services"
	"github.com/stratifyr/security-service/proto"
	"gofr.dev/pkg/gofr"
)

type securityGRPCHandler struct {
	svc services.SecurityService
}

func NewSecurityGRPCHandler(svc services.SecurityService) *securityGRPCHandler {
	return &securityGRPCHandler{svc: svc}
}

func (h *securityGRPCHandler) Index(ctx *gofr.Context) (any, error) {
	var payload proto.GetSecuritiesRequest

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

	return h.buildResponse(securities, count)
}

func (h *securityGRPCHandler) buildResponse(securities []*services.Security, count int) (*proto.GetSecuritiesResponse, error) {
	resp := &proto.GetSecuritiesResponse{
		Securities: make([]*proto.Security, len(securities)),
		Total:      int32(count),
	}

	for i := range resp.Securities {
		resp.Securities[i] = &proto.Security{
			Id:            int32(securities[i].ID),
			Isin:          securities[i].ISIN,
			Symbol:        securities[i].Symbol,
			Industry:      securities[i].Industry,
			Name:          securities[i].Name,
			Image:         securities[i].Image,
			Ltp:           securities[i].LTP,
			PreviousClose: securities[i].PreviousClose,
			CreatedAt:     securities[i].CreatedAt.Format(time.RFC3339),
			UpdatedAt:     securities[i].UpdatedAt.Format(time.RFC3339),
		}

		if securities[i].SecurityStat == nil {
			continue
		}

		resp.Securities[i].MarketData = &proto.MarketData{
			Date:    securities[i].SecurityStat.Date.Format(time.DateOnly),
			Open:    securities[i].SecurityStat.Open,
			Close:   securities[i].SecurityStat.Close,
			High:    securities[i].SecurityStat.High,
			Low:     securities[i].SecurityStat.Low,
			Volume:  int32(securities[i].SecurityStat.Volume),
			Metrics: make([]*proto.SecurityMetric, len(securities[i].SecurityMetrics)),
		}

		for j := range resp.Securities[i].MarketData.Metrics {
			resp.Securities[i].MarketData.Metrics[j] = &proto.SecurityMetric{
				Id:              int32(securities[i].SecurityMetrics[j].Metric.ID),
				Name:            securities[i].SecurityMetrics[j].Metric.Name,
				Type:            securities[i].SecurityMetrics[j].Metric.Type.String(),
				Period:          int32(securities[i].SecurityMetrics[j].Metric.Period),
				Indicator:       securities[i].SecurityMetrics[j].Metric.Indicator.String(),
				Value:           securities[i].SecurityMetrics[j].Value,
				NormalizedValue: securities[i].SecurityMetrics[j].ZValue,
			}
		}
	}

	return resp, nil
}
