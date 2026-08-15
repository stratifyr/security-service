package handlers

import (
	"encoding/json"
	"slices"
	"strconv"
	"time"

	"gofr.dev/pkg/gofr"
	"gofr.dev/pkg/gofr/http"
	"gofr.dev/pkg/gofr/http/response"

	"github.com/stratifyr/security-service/internal/services"
)

type MarketDataJob struct {
	ID        int              `json:"id"`
	Type      string           `json:"type"`
	Status    string           `json:"status"`
	Logs      *json.RawMessage `json:"logs"`
	CreatedAt string           `json:"createdAt"`
	UpdatedAt string           `json:"updatedAt"`
}

type MarketDataJobCreate struct {
	UserID int    `json:"userId"`
	Type   string `json:"type"`
}

type MarketDataJobUpdate struct {
	UserID int              `json:"userId"`
	Status string           `json:"status"`
	Logs   *json.RawMessage `json:"logs"`
}

type marketDataJobHandler struct {
	svc services.MarketDataJobService
}

func NewMarketDataJobHandler(svc services.MarketDataJobService) *marketDataJobHandler {
	return &marketDataJobHandler{svc: svc}
}

func (h *marketDataJobHandler) Index(ctx *gofr.Context) (interface{}, error) {
	var (
		filter services.MarketDataJobFilter
		err    error
	)

	filter.UserID, err = strconv.Atoi(ctx.Param("userId"))
	if err != nil {
		return nil, http.ErrorInvalidParam{Params: []string{"userId"}}
	}

	if ctx.Param("status") != "" {
		filter.Status = ctx.Param("status")

		if !slices.Contains([]string{"CREATED", "COMPLETED", "FAILED"}, filter.Status) {
			return nil, http.ErrorInvalidParam{Params: []string{"status"}}
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

	marketDataJobs, count, err := h.svc.Index(ctx, &filter, page, perPage)
	if err != nil {
		return nil, err
	}

	var resp = make([]*MarketDataJob, len(marketDataJobs))

	for i := range marketDataJobs {
		resp[i] = h.buildResp(marketDataJobs[i])
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

func (h *marketDataJobHandler) Read(ctx *gofr.Context) (interface{}, error) {
	id, err := strconv.Atoi(ctx.PathParam("id"))
	if err != nil {
		return nil, http.ErrorInvalidParam{Params: []string{"id"}}
	}

	marketDataJob, err := h.svc.Read(ctx, id)
	if err != nil {
		return nil, err
	}

	return response.Raw{Data: map[string]any{
		"data": h.buildResp(marketDataJob),
	}}, nil
}

func (h *marketDataJobHandler) Create(ctx *gofr.Context) (interface{}, error) {
	var payload MarketDataJobCreate

	if err := ctx.Bind(&payload); err != nil {
		return nil, http.ErrorInvalidParam{Params: []string{"request-body"}}
	}

	model := &services.MarketDataJobCreate{
		UserID: payload.UserID,
		Type:   payload.Type,
	}

	marketDataJob, err := h.svc.Create(ctx, model)
	if err != nil {
		return nil, err
	}

	return response.Raw{Data: map[string]any{
		"data": h.buildResp(marketDataJob),
	}}, nil
}

func (h *marketDataJobHandler) Patch(ctx *gofr.Context) (interface{}, error) {
	id, err := strconv.Atoi(ctx.PathParam("id"))
	if err != nil {
		return nil, http.ErrorInvalidParam{Params: []string{"id"}}
	}

	var payload MarketDataJobUpdate

	if err = ctx.Bind(&payload); err != nil {
		return nil, http.ErrorInvalidParam{Params: []string{"request-body"}}
	}

	model := &services.MarketDataJobUpdate{
		UserID: payload.UserID,
		Status: payload.Status,
		Logs:   payload.Logs,
	}

	marketDataJob, err := h.svc.Patch(ctx, id, model)
	if err != nil {
		return nil, err
	}

	return response.Raw{Data: map[string]any{
		"data": h.buildResp(marketDataJob),
	}}, nil
}

func (h *marketDataJobHandler) buildResp(model *services.MarketDataJob) *MarketDataJob {
	resp := &MarketDataJob{
		ID:        model.ID,
		Type:      model.Type.String(),
		Status:    model.Status,
		Logs:      model.Logs,
		CreatedAt: model.CreatedAt.Format(time.RFC3339),
		UpdatedAt: model.UpdatedAt.Format(time.RFC3339),
	}

	return resp
}
