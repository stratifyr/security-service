package handlers

import (
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

type securityHandler struct {
	svc services.SecurityService
}

func NewSecurityHandler(svc services.SecurityService) *securityHandler {
	return &securityHandler{svc: svc}
}

func (h *securityHandler) List(ctx *gofr.Context) (any, error) {
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

	securities, err := h.svc.List(ctx, &filter)
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
