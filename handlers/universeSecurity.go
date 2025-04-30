package handlers

import (
	"github.com/stratifyr/security-service/services"
	"gofr.dev/pkg/gofr"
	"gofr.dev/pkg/gofr/http"
	"gofr.dev/pkg/gofr/http/response"
	"strconv"
	"time"
)

type UniverseSecurity struct {
	ID         int    `json:"id"`
	UniverseID int    `json:"universeId"`
	SecurityID int    `json:"securityId"`
	Status     string `json:"status"`
	CreatedAt  string `json:"createdAt"`
	UpdatedAt  string `json:"updatedAt"`
}

type UniverseSecurityCreate struct {
	UserID     int `json:"userId"`
	UniverseID int `json:"universeId"`
	SecurityID int `json:"securityId"`
}

type UniverseSecurityUpdate struct {
	UserID int    `json:"userId"`
	Status string `json:"status"`
}

type universeSecurityHandler struct {
	svc services.UniverseSecurityService
}

func NewUniverseSecurityHandler(svc services.UniverseSecurityService) *universeSecurityHandler {
	return &universeSecurityHandler{svc: svc}
}

func (h *universeSecurityHandler) Index(ctx *gofr.Context) (interface{}, error) {
	var (
		filter services.UniverseSecurityFilter
		err    error
	)

	if ctx.Param("userId") != "" {
		filter.UserID, err = strconv.Atoi(ctx.Param("symbol"))
		if err != nil || filter.UserID < 1 {
			return nil, http.ErrorInvalidParam{Params: []string{"userId"}}
		}
	}

	if ctx.Param("status") != "" {
		filter.Status = ctx.Param("status")
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

	universeSecurities, count, err := h.svc.Index(ctx, &filter, page, perPage)
	if err != nil {
		return nil, err
	}

	var resp = make([]*UniverseSecurity, len(universeSecurities))

	for i := range universeSecurities {
		resp[i] = h.buildResp(universeSecurities[i])
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

func (h *universeSecurityHandler) Read(ctx *gofr.Context) (interface{}, error) {
	id, err := strconv.Atoi(ctx.PathParam("id"))
	if err != nil {
		return nil, http.ErrorInvalidParam{Params: []string{"id"}}
	}

	userID, err := strconv.Atoi(ctx.Param("userId"))
	if err != nil {
		return nil, http.ErrorInvalidParam{Params: []string{"userId"}}
	}

	universeSecurity, err := h.svc.Read(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	return response.Raw{Data: map[string]any{
		"data": h.buildResp(universeSecurity),
	}}, nil
}

func (h *universeSecurityHandler) Create(ctx *gofr.Context) (interface{}, error) {
	var payload UniverseSecurityCreate

	if err := ctx.Bind(&payload); err != nil {
		return nil, http.ErrorInvalidParam{Params: []string{"request-body"}}
	}

	model := &services.UniverseSecurityCreate{
		UserID:     payload.UserID,
		UniverseID: payload.UniverseID,
		SecurityID: payload.SecurityID,
	}

	universeSecurity, err := h.svc.Create(ctx, model)
	if err != nil {
		return nil, err
	}

	return response.Raw{Data: map[string]any{
		"data": h.buildResp(universeSecurity),
	}}, nil
}

func (h *universeSecurityHandler) Patch(ctx *gofr.Context) (interface{}, error) {
	id, err := strconv.Atoi(ctx.PathParam("id"))
	if err != nil {
		return nil, http.ErrorInvalidParam{Params: []string{"id"}}
	}

	var payload UniverseSecurityUpdate

	if err := ctx.Bind(&payload); err != nil {
		return nil, http.ErrorInvalidParam{Params: []string{"request-body"}}
	}

	model := &services.UniverseSecurityUpdate{
		UserID: payload.UserID,
		Status: payload.Status,
	}

	universeSecurity, err := h.svc.Patch(ctx, id, model)
	if err != nil {
		return nil, err
	}

	return response.Raw{Data: map[string]any{
		"data": h.buildResp(universeSecurity),
	}}, nil
}

func (h *universeSecurityHandler) Delete(ctx *gofr.Context) (interface{}, error) {
	id, err := strconv.Atoi(ctx.PathParam("id"))
	if err != nil {
		return nil, http.ErrorInvalidParam{Params: []string{"id"}}
	}

	userId, err := strconv.Atoi(ctx.Param("userId"))
	if err != nil {
		return nil, http.ErrorInvalidParam{Params: []string{"userId"}}
	}

	err = h.svc.Delete(ctx, id, userId)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (h *universeSecurityHandler) buildResp(model *services.UniverseSecurity) *UniverseSecurity {
	resp := &UniverseSecurity{
		ID:         model.ID,
		UniverseID: model.UniverseID,
		SecurityID: model.SecurityID,
		Status:     model.Status,
		CreatedAt:  model.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  model.UpdatedAt.Format(time.RFC3339),
	}

	return resp
}
