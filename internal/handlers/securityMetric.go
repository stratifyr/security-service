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
	ID         int    `json:"id"`
	SecurityID int    `json:"securityId"`
	MetricID   int    `json:"metricId"`
	Date       string `json:"date"`
	Value      string `json:"value"`
	CreatedAt  string `json:"createdAt"`
	UpdatedAt  string `json:"updatedAt"`
}

type SecurityMetricCreate struct {
	UserID     int     `json:"userId"`
	SecurityID int     `json:"securityId"`
	MetricID   int     `json:"metricId"`
	Date       string  `json:"date"`
	Value      float64 `json:"value"`
}

type SecurityMetricUpdate struct {
	UserID         int     `json:"userId"`
	Value          float64 `json:"value"`
	RecomputeValue bool    `json:"recomputeValue"`
}

type securityMetricHandler struct {
	svc services.SecurityMetricService
}

func NewSecurityMetricHandler(svc services.SecurityMetricService) *securityMetricHandler {
	return &securityMetricHandler{svc: svc}
}

func (h *securityMetricHandler) Index(ctx *gofr.Context) (interface{}, error) {
	var (
		filter services.SecurityMetricFilter
		err    error
	)

	if ctx.Param("securityId") != "" {
		filter.SecurityID, err = strconv.Atoi(ctx.Param("securityId"))
		if err != nil {
			return nil, http.ErrorInvalidParam{Params: []string{"securityId"}}
		}
	}

	if ctx.Param("metricId") != "" {
		filter.MetricID, err = strconv.Atoi(ctx.Param("metricId"))
		if err != nil {
			return nil, http.ErrorInvalidParam{Params: []string{"metricId"}}
		}
	}

	if ctx.Param("date") != "" {
		filter.Date, err = time.Parse(time.DateOnly, ctx.Param("date"))
		if err != nil {
			return nil, http.ErrorInvalidParam{Params: []string{"date"}}
		}
	}

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

	securityMetrics, count, err := h.svc.Index(ctx, &filter, page, perPage)
	if err != nil {
		return nil, err
	}

	var resp = make([]*SecurityMetric, len(securityMetrics))

	for i := range securityMetrics {
		resp[i] = h.buildResp(securityMetrics[i])
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

func (h *securityMetricHandler) Read(ctx *gofr.Context) (interface{}, error) {
	id, err := strconv.Atoi(ctx.PathParam("id"))
	if err != nil {
		return nil, http.ErrorInvalidParam{Params: []string{"id"}}
	}

	securityMetric, err := h.svc.Read(ctx, id)
	if err != nil {
		return nil, err
	}

	return response.Raw{Data: map[string]any{
		"data": h.buildResp(securityMetric),
	}}, nil
}

func (h *securityMetricHandler) Create(ctx *gofr.Context) (interface{}, error) {
	var payload SecurityMetricCreate

	if err := ctx.Bind(&payload); err != nil {
		return nil, http.ErrorInvalidParam{Params: []string{"request-body"}}
	}

	date, err := time.Parse(time.DateOnly, payload.Date)
	if err != nil {
		return nil, http.ErrorInvalidParam{Params: []string{"date"}}
	}

	model := &services.SecurityMetricCreate{
		UserID:     payload.UserID,
		SecurityID: payload.SecurityID,
		MetricID:   payload.MetricID,
		Date:       date,
		Value:      payload.Value,
	}

	securityMetric, err := h.svc.Create(ctx, model)
	if err != nil {
		return nil, err
	}

	return response.Raw{Data: map[string]any{
		"data": h.buildResp(securityMetric),
	}}, nil
}

func (h *securityMetricHandler) Patch(ctx *gofr.Context) (interface{}, error) {
	id, err := strconv.Atoi(ctx.PathParam("id"))
	if err != nil {
		return nil, http.ErrorInvalidParam{Params: []string{"id"}}
	}

	var payload SecurityMetricUpdate

	if err = ctx.Bind(&payload); err != nil {
		return nil, http.ErrorInvalidParam{Params: []string{"request-body"}}
	}

	model := &services.SecurityMetricUpdate{
		UserID:         payload.UserID,
		Value:          payload.Value,
		RecomputeValue: payload.RecomputeValue,
	}

	securityMetric, err := h.svc.Patch(ctx, id, model)
	if err != nil {
		return nil, err
	}

	return response.Raw{Data: map[string]any{
		"data": h.buildResp(securityMetric),
	}}, nil
}

func (h *securityMetricHandler) buildResp(model *services.SecurityMetric) *SecurityMetric {
	resp := &SecurityMetric{
		ID:         model.ID,
		SecurityID: model.SecurityID,
		MetricID:   model.MetricID,
		Date:       model.Date.Format(time.DateOnly),
		Value:      fmt.Sprintf("%0.2f", model.Value),
		CreatedAt:  model.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  model.UpdatedAt.Format(time.RFC3339),
	}

	return resp
}
