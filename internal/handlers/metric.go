package handlers

import (
	"strconv"
	"time"

	"gofr.dev/pkg/gofr"
	"gofr.dev/pkg/gofr/http"
	"gofr.dev/pkg/gofr/http/response"

	"github.com/stratifyr/security-service/internal/services"
)

type Metric struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	Period    int    `json:"period"`
	Indicator string `json:"indicator"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type MetricCreate struct {
	UserID int    `json:"userId"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	Period int    `json:"period"`
}

type MetricUpdate struct {
	UserID int    `json:"userId"`
	Name   string `json:"name"`
}

type metricHandler struct {
	svc services.MetricService
}

func NewMetricHandler(svc services.MetricService) *metricHandler {
	return &metricHandler{svc: svc}
}

func (h *metricHandler) Index(ctx *gofr.Context) (interface{}, error) {
	var filter services.MetricFilter

	if ctx.Param("type") != "" {
		filter.Type = ctx.Param("type")
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

	metrics, count, err := h.svc.Index(ctx, &filter, page, perPage)
	if err != nil {
		return nil, err
	}

	var resp = make([]*Metric, len(metrics))

	for i := range metrics {
		resp[i] = h.buildResp(metrics[i])
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

func (h *metricHandler) Read(ctx *gofr.Context) (interface{}, error) {
	id, err := strconv.Atoi(ctx.PathParam("id"))
	if err != nil {
		return nil, http.ErrorInvalidParam{Params: []string{"id"}}
	}

	metric, err := h.svc.Read(ctx, id)
	if err != nil {
		return nil, err
	}

	return response.Raw{Data: map[string]any{
		"data": h.buildResp(metric),
	}}, nil
}

func (h *metricHandler) Create(ctx *gofr.Context) (interface{}, error) {
	var payload MetricCreate

	if err := ctx.Bind(&payload); err != nil {
		return nil, http.ErrorInvalidParam{Params: []string{"request-body"}}
	}

	model := &services.MetricCreate{
		UserID: payload.UserID,
		Name:   payload.Name,
		Type:   payload.Type,
		Period: payload.Period,
	}

	metric, err := h.svc.Create(ctx, model)
	if err != nil {
		return nil, err
	}

	return response.Raw{Data: map[string]any{
		"data": h.buildResp(metric),
	}}, nil
}

func (h *metricHandler) Patch(ctx *gofr.Context) (interface{}, error) {
	id, err := strconv.Atoi(ctx.PathParam("id"))
	if err != nil {
		return nil, http.ErrorInvalidParam{Params: []string{"id"}}
	}

	var payload MetricUpdate

	if err := ctx.Bind(&payload); err != nil {
		return nil, http.ErrorInvalidParam{Params: []string{"request-body"}}
	}

	model := &services.MetricUpdate{
		UserID: payload.UserID,
		Name:   payload.Name,
	}

	metric, err := h.svc.Patch(ctx, id, model)
	if err != nil {
		return nil, err
	}

	return response.Raw{Data: map[string]any{
		"data": h.buildResp(metric),
	}}, nil
}

func (h *metricHandler) buildResp(model *services.Metric) *Metric {
	resp := &Metric{
		ID:        model.ID,
		Name:      model.Name,
		Type:      model.Type,
		Period:    model.Period,
		Indicator: model.Indicator,
		CreatedAt: model.CreatedAt.Format(time.RFC3339),
		UpdatedAt: model.UpdatedAt.Format(time.RFC3339),
	}

	return resp
}
