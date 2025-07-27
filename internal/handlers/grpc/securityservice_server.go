package grpc

import (
	"time"

	"gofr.dev/pkg/gofr"

	"github.com/stratifyr/security-service/internal/services"
)

type SecurityServiceGoFrServer struct {
	svc    services.SecurityService
	health *healthServer

	UnimplementedSecurityServiceServer
}

func (s *SecurityServiceGoFrServer) Index(ctx *gofr.Context) (any, error) {
	var payload SecurityIndexRequest

	if err := ctx.Bind(&payload); err != nil {
		return nil, err
	}

	var ids []int

	if len(payload.Ids) > 0 {
		for i := range payload.Ids {
			ids = append(ids, int(payload.Ids[i]))
		}
	}

	filter := &services.SecurityFilter{
		IDs:    ids,
		ISIN:   payload.Isin,
		Symbol: payload.Symbol,
	}

	securities, count, err := s.svc.Index(ctx, filter, 0, 0)
	if err != nil {
		return nil, err
	}

	return s.buildResponse(securities, count)
}

func (s *SecurityServiceGoFrServer) buildResponse(securities []*services.Security, count int) (*SecurityIndexResponse, error) {
	resp := &SecurityIndexResponse{
		Securities: make([]*Security, len(securities)),
		Total:      int32(count),
	}

	for i := range resp.Securities {
		resp.Securities[i] = &Security{
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

		resp.Securities[i].MarketData = &MarketData{
			Date:    securities[i].SecurityStat.Date.Format(time.DateOnly),
			Open:    securities[i].SecurityStat.Open,
			Close:   securities[i].SecurityStat.Close,
			High:    securities[i].SecurityStat.High,
			Low:     securities[i].SecurityStat.Low,
			Volume:  int32(securities[i].SecurityStat.Volume),
			Metrics: make([]*Metric, len(securities[i].SecurityMetrics)),
		}

		for j := range resp.Securities[i].MarketData.Metrics {
			resp.Securities[i].MarketData.Metrics[j] = &Metric{
				Id:              int32(securities[i].SecurityMetrics[j].Metric.ID),
				Name:            securities[i].SecurityMetrics[j].Metric.Name,
				Type:            securities[i].SecurityMetrics[j].Metric.Type,
				Period:          int32(securities[i].SecurityMetrics[j].Metric.Period),
				Indicator:       securities[i].SecurityMetrics[j].Metric.Indicator,
				Value:           securities[i].SecurityMetrics[j].Value,
				NormalizedValue: securities[i].SecurityMetrics[j].NormalizedValue,
			}
		}
	}

	return resp, nil
}
