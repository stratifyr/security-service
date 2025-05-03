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

type Universe struct {
	ID                 int    `json:"id"`
	UserID             int    `json:"userId"`
	Name               string `json:"name"`
	CreatedAt          string `json:"createdAt"`
	UpdatedAt          string `json:"updatedAt"`
	UniverseSecurities []*struct {
		ID         int    `json:"id"`
		SecurityID int    `json:"securityId"`
		Status     string `json:"status"`
		ISIN       string `json:"isin"`
		Symbol     string `json:"symbol"`
		Industry   string `json:"industry"`
		Name       string `json:"name"`
		Image      string `json:"image"`
		LTP        string `json:"ltp"`
		MarketData *struct {
			Date    string `json:"date"`
			Open    string `json:"open"`
			Close   string `json:"close"`
			High    string `json:"high"`
			Low     string `json:"low"`
			Volume  int    `json:"volume"`
			Metrics []*struct {
				Name  string `json:"name"`
				Type  string `json:"type"`
				Value string `json:"value"`
			} `json:"metrics"`
		} `json:"marketData"`
	} `json:"universeSecurities"`
}

type UniverseCreate struct {
	UserID      int    `json:"userId"`
	Name        string `json:"name"`
	SecurityIDs []int  `json:"securityIds"`
}

type UniverseUpdate struct {
	UserID int    `json:"userId"`
	Name   string `json:"name"`
}

type universeHandler struct {
	svc services.UniverseService
}

func NewUniverseHandler(svc services.UniverseService) *universeHandler {
	return &universeHandler{svc: svc}
}

func (h *universeHandler) Index(ctx *gofr.Context) (interface{}, error) {
	var (
		filter services.UniverseFilter
		err    error
	)

	if ctx.Param("userId") != "" {
		filter.UserID, err = strconv.Atoi(ctx.Param("userId"))
		if err != nil || filter.UserID < 1 {
			return nil, http.ErrorInvalidParam{Params: []string{"userId"}}
		}
	}

	if ctx.Param("includeDefault") != "" {
		filter.IncludeDefault = ctx.Param("includeDefault") == "true"
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

	universes, count, err := h.svc.Index(ctx, &filter, page, perPage)
	if err != nil {
		return nil, err
	}

	var resp = make([]*Universe, len(universes))

	for i := range universes {
		resp[i] = h.buildResp(universes[i])
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

func (h *universeHandler) Read(ctx *gofr.Context) (interface{}, error) {
	id, err := strconv.Atoi(ctx.PathParam("id"))
	if err != nil {
		return nil, http.ErrorInvalidParam{Params: []string{"id"}}
	}

	userID, err := strconv.Atoi(ctx.Param("userId"))
	if err != nil {
		return nil, http.ErrorInvalidParam{Params: []string{"userId"}}
	}

	universe, err := h.svc.Read(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	return response.Raw{Data: map[string]any{
		"data": h.buildResp(universe),
	}}, nil
}

func (h *universeHandler) Create(ctx *gofr.Context) (interface{}, error) {
	var payload UniverseCreate

	if err := ctx.Bind(&payload); err != nil {
		return nil, http.ErrorInvalidParam{Params: []string{"request-body"}}
	}

	model := &services.UniverseCreate{
		UserID:      payload.UserID,
		Name:        payload.Name,
		SecurityIDs: payload.SecurityIDs,
	}

	universe, err := h.svc.Create(ctx, model)
	if err != nil {
		return nil, err
	}

	return response.Raw{Data: map[string]any{
		"data": h.buildResp(universe),
	}}, nil
}

func (h *universeHandler) Patch(ctx *gofr.Context) (interface{}, error) {
	id, err := strconv.Atoi(ctx.PathParam("id"))
	if err != nil {
		return nil, err
	}

	var payload UniverseUpdate

	if err := ctx.Bind(&payload); err != nil {
		return nil, http.ErrorInvalidParam{Params: []string{"request-body"}}
	}

	model := &services.UniverseUpdate{
		UserID: payload.UserID,
		Name:   payload.Name,
	}

	universe, err := h.svc.Patch(ctx, id, model)
	if err != nil {
		return nil, err
	}

	return response.Raw{Data: map[string]any{
		"data": h.buildResp(universe),
	}}, nil
}

func (h *universeHandler) buildResp(model *services.Universe) *Universe {
	resp := &Universe{
		ID:        model.ID,
		UserID:    model.UserID,
		Name:      model.Name,
		CreatedAt: model.CreatedAt.Format(time.RFC3339),
		UpdatedAt: model.UpdatedAt.Format(time.RFC3339),
		UniverseSecurities: make([]*struct {
			ID         int    `json:"id"`
			SecurityID int    `json:"securityId"`
			Status     string `json:"status"`
			ISIN       string `json:"isin"`
			Symbol     string `json:"symbol"`
			Industry   string `json:"industry"`
			Name       string `json:"name"`
			Image      string `json:"image"`
			LTP        string `json:"ltp"`
			MarketData *struct {
				Date    string `json:"date"`
				Open    string `json:"open"`
				Close   string `json:"close"`
				High    string `json:"high"`
				Low     string `json:"low"`
				Volume  int    `json:"volume"`
				Metrics []*struct {
					Name  string `json:"name"`
					Type  string `json:"type"`
					Value string `json:"value"`
				} `json:"metrics"`
			} `json:"marketData"`
		}, len(model.UniverseSecurities)),
	}

	for i := range model.UniverseSecurities {
		resp.UniverseSecurities[i] = &struct {
			ID         int    `json:"id"`
			SecurityID int    `json:"securityId"`
			Status     string `json:"status"`
			ISIN       string `json:"isin"`
			Symbol     string `json:"symbol"`
			Industry   string `json:"industry"`
			Name       string `json:"name"`
			Image      string `json:"image"`
			LTP        string `json:"ltp"`
			MarketData *struct {
				Date    string `json:"date"`
				Open    string `json:"open"`
				Close   string `json:"close"`
				High    string `json:"high"`
				Low     string `json:"low"`
				Volume  int    `json:"volume"`
				Metrics []*struct {
					Name  string `json:"name"`
					Type  string `json:"type"`
					Value string `json:"value"`
				} `json:"metrics"`
			} `json:"marketData"`
		}{}

		resp.UniverseSecurities[i].ID = model.UniverseSecurities[i].ID
		resp.UniverseSecurities[i].SecurityID = model.UniverseSecurities[i].SecurityID
		resp.UniverseSecurities[i].Status = model.UniverseSecurities[i].Status
		resp.UniverseSecurities[i].ISIN = model.UniverseSecurities[i].Security.ISIN
		resp.UniverseSecurities[i].Symbol = model.UniverseSecurities[i].Security.Symbol
		resp.UniverseSecurities[i].Industry = model.UniverseSecurities[i].Security.Industry
		resp.UniverseSecurities[i].Name = model.UniverseSecurities[i].Security.Name
		resp.UniverseSecurities[i].Image = model.UniverseSecurities[i].Security.Image
		resp.UniverseSecurities[i].LTP = fmt.Sprintf("%0.2f", model.UniverseSecurities[i].Security.LTP)

		if model.UniverseSecurities[i].Security.SecurityStat == nil {
			return resp
		}

		resp.UniverseSecurities[i].MarketData = &struct {
			Date    string `json:"date"`
			Open    string `json:"open"`
			Close   string `json:"close"`
			High    string `json:"high"`
			Low     string `json:"low"`
			Volume  int    `json:"volume"`
			Metrics []*struct {
				Name  string `json:"name"`
				Type  string `json:"type"`
				Value string `json:"value"`
			} `json:"metrics"`
		}{
			Date:   model.UniverseSecurities[i].Security.SecurityStat.Date.Format(time.DateOnly),
			Open:   fmt.Sprintf("%0.2f", model.UniverseSecurities[i].Security.SecurityStat.Open),
			Close:  fmt.Sprintf("%0.2f", model.UniverseSecurities[i].Security.SecurityStat.Close),
			High:   fmt.Sprintf("%0.2f", model.UniverseSecurities[i].Security.SecurityStat.High),
			Low:    fmt.Sprintf("%0.2f", model.UniverseSecurities[i].Security.SecurityStat.Low),
			Volume: model.UniverseSecurities[i].Security.SecurityStat.Volume,
			Metrics: make([]*struct {
				Name  string `json:"name"`
				Type  string `json:"type"`
				Value string `json:"value"`
			}, len(model.UniverseSecurities[i].Security.SecurityMetrics)),
		}

		for j := range model.UniverseSecurities[i].Security.SecurityMetrics {
			resp.UniverseSecurities[i].MarketData.Metrics[j] = &struct {
				Name  string `json:"name"`
				Type  string `json:"type"`
				Value string `json:"value"`
			}{
				Name:  model.UniverseSecurities[i].Security.SecurityMetrics[j].Metric.Name,
				Type:  model.UniverseSecurities[i].Security.SecurityMetrics[j].Metric.Type,
				Value: fmt.Sprintf("%0.2f", model.UniverseSecurities[i].Security.SecurityMetrics[j].Value),
			}
		}
	}

	return resp
}
