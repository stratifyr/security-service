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

func (h *securityGRPCHandler) List(ctx *gofr.Context) (any, error) {
	var payload pb.GetSecuritiesRequest

	if err := ctx.Bind(&payload); err != nil {
		return nil, err
	}

	securities, err := h.svc.List(ctx)
	if err != nil {
		return nil, err
	}

	var resp = make([]*pb.Security, len(securities))

	for i := range securities {
		resp[i] = h.buildResponse(securities[i])
	}

	return &pb.GetSecuritiesResponse{
		Securities: resp,
		Total:      int32(len(securities)),
	}, nil
}

func (h *securityGRPCHandler) Upsert(ctx *gofr.Context) (any, error) {
	var payload pb.UpsertSecurityRequest

	if err := ctx.Bind(&payload); err != nil {
		return nil, err
	}

	model := &services.SecurityUpsert{
		ISIN:            payload.Isin,
		Symbol:          payload.Symbol,
		Industry:        payload.Industry,
		Name:            payload.Name,
		Image:           payload.Image,
		LTP:             payload.Ltp,
		Volume:          int(payload.Volume),
		FreeFloatShares: int(payload.FreeFloatShares),
	}

	security, err := h.svc.Upsert(ctx, model)
	if err != nil {
		return nil, err
	}

	return &pb.UpsertSecurityResponse{
		Security: h.buildResponse(security),
	}, nil
}

func (h *securityGRPCHandler) GetHistoricalData(ctx *gofr.Context) (any, error) {
	var payload pb.GetHistoricalDataRequest

	if err := ctx.Bind(&payload); err != nil {
		return nil, err
	}

	marketData, err := h.svc.GetHistoricalData(ctx, int(payload.SecurityId))
	if err != nil {
		return nil, err
	}

	var resp = make([]*pb.MarketData, len(marketData))

	for i := range marketData {
		resp[i] = h.buildMarketDateResp(marketData[i])
	}

	return &pb.GetHistoricalDataResponse{
		MarketData: resp,
		Total:      int32(len(marketData)),
	}, nil
}

func (h *securityGRPCHandler) buildResponse(security *services.Security) *pb.Security {
	var resp = &pb.Security{
		Id:              int32(security.ID),
		Isin:            security.ISIN,
		Symbol:          security.Symbol,
		Industry:        security.Industry,
		Name:            security.Name,
		Image:           security.Image,
		Ltp:             security.LTP,
		Volume:          int64(security.Volume),
		FreeFloatShares: int64(security.FreeFloatShares),
		PreviousClose:   security.PreviousClose,
		CreatedAt:       security.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       security.UpdatedAt.Format(time.RFC3339),
	}

	if security.SecurityStat == nil {
		return resp
	}

	resp.MarketData = h.buildMarketDateResp(security.MarketData)

	return resp
}

func (*securityGRPCHandler) buildMarketDateResp(model *services.MarketData) *pb.MarketData {
	resp := &pb.MarketData{
		Date:    model.SecurityStat.Date.Format(time.DateOnly),
		Open:    model.SecurityStat.Open,
		Close:   model.SecurityStat.Close,
		High:    model.SecurityStat.High,
		Low:     model.SecurityStat.Low,
		Volume:  int32(model.SecurityStat.Volume),
		Metrics: make([]*pb.SecurityMetric, len(model.SecurityMetrics)),
	}

	for i := range model.SecurityMetrics {
		resp.Metrics[i] = &pb.SecurityMetric{
			Id:              int32(model.SecurityMetrics[i].Metric.ID),
			Name:            model.SecurityMetrics[i].Metric.Name,
			Type:            model.SecurityMetrics[i].Metric.Type.String(),
			Period:          int32(model.SecurityMetrics[i].Metric.Period),
			Indicator:       model.SecurityMetrics[i].Metric.Indicator.String(),
			Value:           model.SecurityMetrics[i].Value,
			NormalizedValue: model.SecurityMetrics[i].ZValue,
		}
	}

	return resp
}
