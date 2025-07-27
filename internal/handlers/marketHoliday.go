package handlers

import (
	"strconv"
	"strings"
	"time"

	"gofr.dev/pkg/gofr"
	"gofr.dev/pkg/gofr/http"
	"gofr.dev/pkg/gofr/http/response"

	"github.com/stratifyr/security-service/internal/services"
)

type MarketHoliday struct {
	ID          int    `json:"id"`
	Date        string `json:"date"`
	Description string `json:"description"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

type MarketHolidayCreate struct {
	UserID      int    `json:"userId"`
	Date        string `json:"date"`
	Description string `json:"description"`
}

type MarketHolidayUpdate struct {
	UserID      int    `json:"userId"`
	Date        string `json:"date"`
	Description string `json:"description"`
}

type marketHolidayHandler struct {
	svc services.MarketHolidayService
}

func NewMarketHolidayHandler(svc services.MarketHolidayService) *marketHolidayHandler {
	return &marketHolidayHandler{svc: svc}
}

func (h *marketHolidayHandler) Index(ctx *gofr.Context) (interface{}, error) {
	var (
		filter services.MarketHolidayFilter
		err    error
	)

	if ctx.Param("date") != "" {
		filter.Date, err = time.Parse(time.DateOnly, ctx.Param("date"))
		if err != nil {
			return nil, http.ErrorInvalidParam{Params: []string{"date"}}
		}
	}

	if ctx.Param("dateBetween") != "" {
		dates := strings.Split(ctx.Param("dateBetween"), ",")
		if len(dates) != 2 {
			return nil, http.ErrorInvalidParam{Params: []string{"dateBetween"}}
		}

		filter.DateBetween = &struct {
			StartDate time.Time
			EndDate   time.Time
		}{}

		filter.DateBetween.StartDate, err = time.Parse(time.DateOnly, dates[0])
		if err != nil {
			return nil, http.ErrorInvalidParam{Params: []string{"dateBetween"}}
		}

		filter.DateBetween.EndDate, err = time.Parse(time.DateOnly, dates[1])
		if err != nil {
			return nil, http.ErrorInvalidParam{Params: []string{"dateBetween"}}
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

	marketHolidays, count, err := h.svc.Index(ctx, &filter, page, perPage)
	if err != nil {
		return nil, err
	}

	var resp = make([]*MarketHoliday, len(marketHolidays))

	for i := range marketHolidays {
		resp[i] = h.buildResp(marketHolidays[i])
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

func (h *marketHolidayHandler) Read(ctx *gofr.Context) (interface{}, error) {
	id, err := strconv.Atoi(ctx.PathParam("id"))
	if err != nil {
		return nil, http.ErrorInvalidParam{Params: []string{"id"}}
	}

	marketHoliday, err := h.svc.Read(ctx, id)
	if err != nil {
		return nil, err
	}

	return response.Raw{Data: map[string]any{
		"data": h.buildResp(marketHoliday),
	}}, nil
}

func (h *marketHolidayHandler) Create(ctx *gofr.Context) (interface{}, error) {
	var payload MarketHolidayCreate

	if err := ctx.Bind(&payload); err != nil {
		return nil, http.ErrorInvalidParam{Params: []string{"request-body"}}
	}

	date, err := time.Parse(time.DateOnly, payload.Date)
	if err != nil {
		return nil, http.ErrorInvalidParam{Params: []string{"date"}}
	}

	model := &services.MarketHolidayCreate{
		UserID:      payload.UserID,
		Date:        date,
		Description: payload.Description,
	}

	marketHoliday, err := h.svc.Create(ctx, model)
	if err != nil {
		return nil, err
	}

	return response.Raw{Data: map[string]any{
		"data": h.buildResp(marketHoliday),
	}}, nil
}

func (h *marketHolidayHandler) Patch(ctx *gofr.Context) (interface{}, error) {
	id, err := strconv.Atoi(ctx.PathParam("id"))
	if err != nil {
		return nil, http.ErrorInvalidParam{Params: []string{"id"}}
	}

	var payload MarketHolidayUpdate

	if err = ctx.Bind(&payload); err != nil {
		return nil, http.ErrorInvalidParam{Params: []string{"request-body"}}
	}

	var date time.Time

	if payload.Date != "" {
		date, err = time.Parse(time.DateOnly, payload.Date)
		if err != nil {
			return nil, http.ErrorInvalidParam{Params: []string{"date"}}
		}
	}

	model := &services.MarketHolidayUpdate{
		UserID:      payload.UserID,
		Date:        date,
		Description: payload.Description,
	}

	marketHoliday, err := h.svc.Patch(ctx, id, model)
	if err != nil {
		return nil, err
	}

	return response.Raw{Data: map[string]any{
		"data": h.buildResp(marketHoliday),
	}}, nil
}

func (h *marketHolidayHandler) Delete(ctx *gofr.Context) (interface{}, error) {
	id, err := strconv.Atoi(ctx.PathParam("id"))
	if err != nil {
		return nil, http.ErrorInvalidParam{Params: []string{"id"}}
	}

	userID, err := strconv.Atoi(ctx.Param("userId"))
	if err != nil {
		return nil, http.ErrorInvalidParam{Params: []string{"userId"}}
	}

	err = h.svc.Delete(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (h *marketHolidayHandler) buildResp(model *services.MarketHoliday) *MarketHoliday {
	resp := &MarketHoliday{
		ID:          model.ID,
		Date:        model.Date.Format(time.DateOnly),
		Description: model.Description,
		CreatedAt:   model.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   model.UpdatedAt.Format(time.RFC3339),
	}

	return resp
}
