package handlers

import (
	"strconv"
	"time"

	"gofr.dev/pkg/gofr"
	"gofr.dev/pkg/gofr/http"
	"gofr.dev/pkg/gofr/http/response"

	"github.com/stratifyr/security-service/services"
)

type Security struct {
	ID         int     `json:"id"`
	ISIN       string  `json:"isin"`
	Symbol     string  `json:"symbol"`
	Industry   string  `json:"industry"`
	Name       string  `json:"name"`
	Image      string  `json:"image"`
	LTP        float64 `json:"ltp"`
	CreatedAt  string  `json:"createdAt"`
	UpdatedAt  string  `json:"updatedAt"`
	MarketData *struct {
		Date    string  `json:"date"`
		Open    float64 `json:"open"`
		Close   float64 `json:"close"`
		High    float64 `json:"high"`
		Low     float64 `json:"low"`
		Volume  int     `json:"volume"`
		Metrics []*struct {
			ID              int     `json:"id"`
			Name            string  `json:"name"`
			Type            string  `json:"type"`
			Period          int     `json:"period"`
			Indicator       string  `json:"indicator"`
			Value           float64 `json:"value"`
			NormalizedValue float64 `json:"normalizedValue"`
		} `json:"metrics"`
	} `json:"marketData"`
}

type SecurityCreate struct {
	UserID   int     `json:"userId"`
	ISIN     string  `json:"isin"`
	Symbol   string  `json:"symbol"`
	Industry string  `json:"industry"`
	Name     string  `json:"name"`
	Image    string  `json:"image"`
	LTP      float64 `json:"ltp"`
}

type SecurityUpdate struct {
	UserID   int     `json:"userId"`
	Symbol   string  `json:"symbol"`
	Industry string  `json:"industry"`
	Name     string  `json:"name"`
	Image    string  `json:"image"`
	LTP      float64 `json:"ltp"`
}

type securityHandler struct {
	svc services.SecurityService
}

func NewSecurityHandler(svc services.SecurityService) *securityHandler {
	return &securityHandler{svc: svc}
}

func (h *securityHandler) Index(ctx *gofr.Context) (interface{}, error) {
	var filter services.SecurityFilter

	if ctx.Param("symbol") != "" {
		filter.Symbol = ctx.Param("symbol")
	}

	var err error

	page := 1
	if ctx.Param("page") != "" {
		page, err = strconv.Atoi(ctx.Param("page"))
		if err != nil || page < 1 {
			return nil, http.ErrorInvalidParam{Params: []string{"page"}}
		}
	}

	perPage := 20
	if ctx.Param("perPage") != "" {
		perPage, err = strconv.Atoi(ctx.Param("perPage"))
		if err != nil || perPage < 1 {
			return nil, http.ErrorInvalidParam{Params: []string{"perPage"}}
		}
	}

	securities, count, err := h.svc.Index(ctx, &filter, page, perPage)
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
			"page":    page,
			"perPage": perPage,
			"total":   count,
		},
	}}, nil
}

func (h *securityHandler) Read(ctx *gofr.Context) (interface{}, error) {
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

func (h *securityHandler) Create(ctx *gofr.Context) (interface{}, error) {
	var payload SecurityCreate

	if err := ctx.Bind(&payload); err != nil {
		return nil, http.ErrorInvalidParam{Params: []string{"request-body"}}
	}

	model := &services.SecurityCreate{
		UserID:   payload.UserID,
		ISIN:     payload.ISIN,
		Symbol:   payload.Symbol,
		Industry: payload.Industry,
		Name:     payload.Name,
		Image:    payload.Image,
		LTP:      payload.LTP,
	}

	security, err := h.svc.Create(ctx, model)
	if err != nil {
		return nil, err
	}

	return response.Raw{Data: map[string]any{
		"data": h.buildResp(security),
	}}, nil
}

func (h *securityHandler) Patch(ctx *gofr.Context) (interface{}, error) {
	id, err := strconv.Atoi(ctx.PathParam("id"))
	if err != nil {
		return nil, http.ErrorInvalidParam{Params: []string{"id"}}
	}

	var payload SecurityUpdate

	if err := ctx.Bind(&payload); err != nil {
		return nil, http.ErrorInvalidParam{Params: []string{"request-body"}}
	}

	model := &services.SecurityUpdate{
		UserID:   payload.UserID,
		Symbol:   payload.Symbol,
		Industry: payload.Industry,
		Name:     payload.Name,
		Image:    payload.Image,
		LTP:      payload.LTP,
	}

	security, err := h.svc.Patch(ctx, id, model)
	if err != nil {
		return nil, err
	}

	return response.Raw{Data: map[string]any{
		"data": h.buildResp(security),
	}}, nil
}

func (h *securityHandler) buildResp(model *services.Security) *Security {
	resp := &Security{
		ID:         model.ID,
		ISIN:       model.ISIN,
		Symbol:     model.Symbol,
		Industry:   model.Industry,
		Name:       model.Name,
		Image:      model.Image,
		LTP:        model.LTP,
		CreatedAt:  model.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  model.UpdatedAt.Format(time.RFC3339),
		MarketData: nil,
	}

	if model.SecurityStat == nil {
		return resp
	}

	resp.MarketData = &struct {
		Date    string  `json:"date"`
		Open    float64 `json:"open"`
		Close   float64 `json:"close"`
		High    float64 `json:"high"`
		Low     float64 `json:"low"`
		Volume  int     `json:"volume"`
		Metrics []*struct {
			ID              int     `json:"id"`
			Name            string  `json:"name"`
			Type            string  `json:"type"`
			Period          int     `json:"period"`
			Indicator       string  `json:"indicator"`
			Value           float64 `json:"value"`
			NormalizedValue float64 `json:"normalizedValue"`
		} `json:"metrics"`
	}{
		Date:   model.SecurityStat.Date.Format(time.DateOnly),
		Open:   model.SecurityStat.Open,
		Close:  model.SecurityStat.Close,
		High:   model.SecurityStat.High,
		Low:    model.SecurityStat.Low,
		Volume: model.SecurityStat.Volume,
		Metrics: make([]*struct {
			ID              int     `json:"id"`
			Name            string  `json:"name"`
			Type            string  `json:"type"`
			Period          int     `json:"period"`
			Indicator       string  `json:"indicator"`
			Value           float64 `json:"value"`
			NormalizedValue float64 `json:"normalizedValue"`
		}, len(model.SecurityMetrics)),
	}

	for i := range model.SecurityMetrics {
		resp.MarketData.Metrics[i] = &struct {
			ID              int     `json:"id"`
			Name            string  `json:"name"`
			Type            string  `json:"type"`
			Period          int     `json:"period"`
			Indicator       string  `json:"indicator"`
			Value           float64 `json:"value"`
			NormalizedValue float64 `json:"normalizedValue"`
		}{
			ID:              model.SecurityMetrics[i].Metric.ID,
			Name:            model.SecurityMetrics[i].Metric.Name,
			Type:            model.SecurityMetrics[i].Metric.Type,
			Period:          model.SecurityMetrics[i].Metric.Period,
			Indicator:       model.SecurityMetrics[i].Metric.Indicator,
			Value:           model.SecurityMetrics[i].Value,
			NormalizedValue: model.SecurityMetrics[i].NormalizedValue,
		}
	}

	return resp
}
