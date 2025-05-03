package handlers

import (
	"fmt"
	"strconv"
	"time"

	"gofr.dev/pkg/gofr"
	"gofr.dev/pkg/gofr/http"
	"gofr.dev/pkg/gofr/http/response"

	"github.com/stratifyr/security-service/services"
)

type SecurityStat struct {
	ID         int    `json:"id"`
	SecurityID int    `json:"securityId"`
	Date       string `json:"date"`
	Open       string `json:"open"`
	Close      string `json:"close"`
	High       string `json:"high"`
	Low        string `json:"low"`
	Volume     int    `json:"volume"`
	CreatedAt  string `json:"createdAt"`
	UpdatedAt  string `json:"updatedAt"`
}

type SecurityStatCreate struct {
	UserID     int     `json:"userId"`
	SecurityID int     `json:"securityId"`
	Date       string  `json:"date"`
	Open       float64 `json:"open"`
	Close      float64 `json:"close"`
	High       float64 `json:"high"`
	Low        float64 `json:"low"`
	Volume     int     `json:"volume"`
}

type SecurityStatUpdate struct {
	UserID int     `json:"userId"`
	Open   float64 `json:"open"`
	Close  float64 `json:"close"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Volume int     `json:"volume"`
}

type securityStatHandler struct {
	svc services.SecurityStatService
}

func NewSecurityStatHandler(svc services.SecurityStatService) *securityStatHandler {
	return &securityStatHandler{svc: svc}
}

func (h *securityStatHandler) Index(ctx *gofr.Context) (interface{}, error) {
	var (
		filter services.SecurityStatFilter
		err    error
	)

	if ctx.Param("date") != "" {
		filter.Date, err = time.Parse(time.DateOnly, ctx.Param("date"))
		if err != nil {
			return nil, http.ErrorInvalidParam{Params: []string{"date"}}
		}
	}

	if ctx.Param("securityId") != "" {
		filter.SecurityID, err = strconv.Atoi(ctx.Param("securityId"))
		if err != nil {
			return nil, http.ErrorInvalidParam{Params: []string{"securityId"}}
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

	securityStats, count, err := h.svc.Index(ctx, &filter, page, perPage)
	if err != nil {
		return nil, err
	}

	var resp = make([]*SecurityStat, len(securityStats))

	for i := range securityStats {
		resp[i] = h.buildResp(securityStats[i])
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

func (h *securityStatHandler) Read(ctx *gofr.Context) (interface{}, error) {
	id, err := strconv.Atoi(ctx.PathParam("id"))
	if err != nil {
		return nil, http.ErrorInvalidParam{Params: []string{"id"}}
	}

	securityStat, err := h.svc.Read(ctx, id)
	if err != nil {
		return nil, err
	}

	return response.Raw{Data: map[string]any{
		"data": h.buildResp(securityStat),
	}}, nil
}

func (h *securityStatHandler) Create(ctx *gofr.Context) (interface{}, error) {
	var payload SecurityStatCreate

	if err := ctx.Bind(&payload); err != nil {
		return nil, http.ErrorInvalidParam{Params: []string{"request-body"}}
	}

	date, err := time.Parse(time.DateOnly, payload.Date)
	if err != nil {
		return nil, http.ErrorInvalidParam{Params: []string{"date"}}
	}

	model := &services.SecurityStatCreate{
		UserID:     payload.UserID,
		SecurityID: payload.SecurityID,
		Date:       date,
		Open:       payload.Open,
		Close:      payload.Close,
		High:       payload.High,
		Low:        payload.Low,
		Volume:     payload.Volume,
	}

	securityStat, err := h.svc.Create(ctx, model)
	if err != nil {
		return nil, err
	}

	return response.Raw{Data: map[string]any{
		"data": h.buildResp(securityStat),
	}}, nil
}

func (h *securityStatHandler) Patch(ctx *gofr.Context) (interface{}, error) {
	id, err := strconv.Atoi(ctx.PathParam("id"))
	if err != nil {
		return nil, http.ErrorInvalidParam{Params: []string{"id"}}
	}

	var payload SecurityStatUpdate

	if err = ctx.Bind(&payload); err != nil {
		return nil, http.ErrorInvalidParam{Params: []string{"request-body"}}
	}

	model := &services.SecurityStatUpdate{
		UserID: payload.UserID,
		Open:   payload.Open,
		Close:  payload.Close,
		High:   payload.High,
		Low:    payload.Low,
		Volume: payload.Volume,
	}

	securityStat, err := h.svc.Patch(ctx, id, model)
	if err != nil {
		return nil, err
	}

	return response.Raw{Data: map[string]any{
		"data": h.buildResp(securityStat),
	}}, nil
}

func (h *securityStatHandler) buildResp(model *services.SecurityStat) *SecurityStat {
	resp := &SecurityStat{
		ID:         model.ID,
		SecurityID: model.SecurityID,
		Date:       model.Date.Format(time.DateOnly),
		Open:       fmt.Sprintf("%0.2f", model.Open),
		Close:      fmt.Sprintf("%0.2f", model.Close),
		High:       fmt.Sprintf("%0.2f", model.High),
		Low:        fmt.Sprintf("%0.2f", model.Low),
		Volume:     model.Volume,
		CreatedAt:  model.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  model.UpdatedAt.Format(time.RFC3339),
	}

	return resp
}
