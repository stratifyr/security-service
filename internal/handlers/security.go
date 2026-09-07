package handlers

import (
	"strconv"
	"time"

	"gofr.dev/pkg/gofr"
	"gofr.dev/pkg/gofr/http"
	"gofr.dev/pkg/gofr/http/response"

	"github.com/stratifyr/security-service/internal/services"
)

type Security struct {
	ID              int         `json:"id"`
	ISIN            string      `json:"isin"`
	Symbol          string      `json:"symbol"`
	Industry        string      `json:"industry"`
	Name            string      `json:"name"`
	Image           string      `json:"image"`
	LTP             float64     `json:"ltp"`
	Volume          int         `json:"volume"`
	FreeFloatShares int         `json:"freeFloatShares"`
	PreviousClose   float64     `json:"previousClose"`
	CreatedAt       string      `json:"createdAt"`
	UpdatedAt       string      `json:"updatedAt"`
	MarketData      *MarketData `json:"marketData"`
}

type MarketData struct {
	Date    string              `json:"date"`
	Open    float64             `json:"open"`
	Close   float64             `json:"close"`
	High    float64             `json:"high"`
	Low     float64             `json:"low"`
	Volume  int                 `json:"volume"`
	Metrics []*MarketDataMetric `json:"metrics"`
}

type MarketDataMetric struct {
	ID              int     `json:"id"`
	Name            string  `json:"name"`
	Type            string  `json:"type"`
	Period          int     `json:"period"`
	Indicator       string  `json:"indicator"`
	Value           float64 `json:"value"`
	NormalizedValue float64 `json:"normalizedValue"`
}

type SecurityCreate struct {
	ISIN            string  `json:"isin"`
	Symbol          string  `json:"symbol"`
	Industry        string  `json:"industry"`
	Name            string  `json:"name"`
	Image           string  `json:"image"`
	LTP             float64 `json:"ltp"`
	Volume          int     `json:"volume"`
	FreeFloatShares int     `json:"freeFloatShares"`
}

type SecurityUpdate struct {
	Symbol          string  `json:"symbol"`
	Industry        string  `json:"industry"`
	Name            string  `json:"name"`
	Image           string  `json:"image"`
	LTP             float64 `json:"ltp"`
	Volume          int     `json:"volume"`
	FreeFloatShares int     `json:"freeFloatShares"`
}

type securityHandler struct {
	svc services.SecurityService
}

func NewSecurityHandler(svc services.SecurityService) *securityHandler {
	return &securityHandler{svc: svc}
}

func (h *securityHandler) Index(ctx *gofr.Context) (any, error) {
	var (
		filter services.SecurityFilter
		err    error
	)

	if ctx.Param("date") != "" {
		filter.Date, err = time.Parse(time.DateOnly, ctx.Param("date"))
		if err != nil {
			return nil, http.ErrorInvalidParam{Params: []string{"date"}}
		}
	}

	securities, err := h.svc.Index(ctx, &filter)
	if err != nil {
		return nil, err
	}

	var resp = make([]*Security, len(securities))

	for i := range securities {
		resp[i] = h.buildResp(securities[i])
	}

	return response.Raw{Data: map[string]any{
		"data": resp,
		"meta": map[string]any{
			"total": len(securities),
		},
	}}, nil
}

func (h *securityHandler) Read(ctx *gofr.Context) (any, error) {
	id, err := strconv.Atoi(ctx.PathParam("id"))
	if err != nil {
		return nil, http.ErrorInvalidParam{Params: []string{"id"}}
	}

	security, err := h.svc.Read(ctx, id)
	if err != nil {
		return nil, err
	}

	return response.Raw{Data: map[string]any{
		"data": h.buildResp(security),
	}}, nil
}

func (h *securityHandler) Create(ctx *gofr.Context) (any, error) {
	var payload SecurityCreate

	if err := ctx.Bind(&payload); err != nil {
		return nil, http.ErrorInvalidParam{Params: []string{"request-body"}}
	}

	model := &services.SecurityCreate{
		ISIN:            payload.ISIN,
		Symbol:          payload.Symbol,
		Industry:        payload.Industry,
		Name:            payload.Name,
		Image:           payload.Image,
		LTP:             payload.LTP,
		Volume:          payload.Volume,
		FreeFloatShares: payload.FreeFloatShares,
	}

	security, err := h.svc.Create(ctx, model)
	if err != nil {
		return nil, err
	}

	return response.Raw{Data: map[string]any{
		"data": h.buildResp(security),
	}}, nil
}

func (h *securityHandler) Patch(ctx *gofr.Context) (any, error) {
	id, err := strconv.Atoi(ctx.PathParam("id"))
	if err != nil {
		return nil, http.ErrorInvalidParam{Params: []string{"id"}}
	}

	var payload SecurityUpdate

	if err := ctx.Bind(&payload); err != nil {
		return nil, http.ErrorInvalidParam{Params: []string{"request-body"}}
	}

	model := &services.SecurityUpdate{
		Symbol:          payload.Symbol,
		Industry:        payload.Industry,
		Name:            payload.Name,
		Image:           payload.Image,
		LTP:             payload.LTP,
		Volume:          payload.Volume,
		FreeFloatShares: payload.FreeFloatShares,
	}

	security, err := h.svc.Patch(ctx, id, model)
	if err != nil {
		return nil, err
	}

	return response.Raw{Data: map[string]any{
		"data": h.buildResp(security),
	}}, nil
}

func (*securityHandler) buildResp(model *services.Security) *Security {
	resp := &Security{
		ID:              model.ID,
		ISIN:            model.ISIN,
		Symbol:          model.Symbol,
		Industry:        model.Industry,
		Name:            model.Name,
		Image:           model.Image,
		LTP:             model.LTP,
		Volume:          model.Volume,
		FreeFloatShares: model.FreeFloatShares,
		PreviousClose:   model.PreviousClose,
		CreatedAt:       model.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       model.UpdatedAt.Format(time.RFC3339),
		MarketData:      nil,
	}

	if model.SecurityStat == nil {
		return resp
	}

	resp.MarketData = &MarketData{
		Date:    model.SecurityStat.Date.Format(time.DateOnly),
		Open:    model.SecurityStat.Open,
		Close:   model.SecurityStat.Close,
		High:    model.SecurityStat.High,
		Low:     model.SecurityStat.Low,
		Volume:  model.SecurityStat.Volume,
		Metrics: make([]*MarketDataMetric, len(model.SecurityMetrics)),
	}

	for i := range model.SecurityMetrics {
		resp.MarketData.Metrics[i] = &MarketDataMetric{
			ID:              model.SecurityMetrics[i].Metric.ID,
			Name:            model.SecurityMetrics[i].Metric.Name,
			Type:            model.SecurityMetrics[i].Metric.Type.String(),
			Period:          model.SecurityMetrics[i].Metric.Period,
			Indicator:       model.SecurityMetrics[i].Metric.Indicator.String(),
			Value:           model.SecurityMetrics[i].Value,
			NormalizedValue: model.SecurityMetrics[i].ZValue,
		}
	}

	return resp
}
