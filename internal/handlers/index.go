package handlers

import (
	"time"

	"gofr.dev/pkg/gofr"
	"gofr.dev/pkg/gofr/http"
	"gofr.dev/pkg/gofr/http/response"

	"github.com/stratifyr/security-service/internal/services"
)

type Index struct {
	ID            int                 `json:"id"`
	Name          string              `json:"name"`
	Value         float64             `json:"value"`
	PreviousClose float64             `json:"previousClose"`
	CreatedAt     string              `json:"createdAt"`
	UpdatedAt     string              `json:"updatedAt"`
	MarketData    *IndexMarketData    `json:"marketData"`
	Constituents  []*IndexConstituent `json:"constituents"`
}

type IndexMarketData struct {
	Date  string  `json:"date"`
	Open  float64 `json:"open"`
	Close float64 `json:"close"`
	High  float64 `json:"high"`
	Low   float64 `json:"low"`
}

type IndexConstituent struct {
	ID              int     `json:"id"`
	IndexID         int     `json:"indexId"`
	SecurityID      int     `json:"securityId"`
	ISIN            string  `json:"isin"`
	Symbol          string  `json:"symbol"`
	Industry        string  `json:"industry"`
	Name            string  `json:"name"`
	Image           string  `json:"image"`
	LTP             float64 `json:"ltp"`
	Volume          int     `json:"volume"`
	FreeFloatShares int     `json:"freeFloatShares"`
	PreviousClose   float64 `json:"previousClose"`
}

type indexHandler struct {
	svc services.IndexService
}

func NewIndexHandler(svc services.IndexService) *indexHandler {
	return &indexHandler{svc: svc}
}

func (h *indexHandler) List(ctx *gofr.Context) (any, error) {
	var (
		filter services.IndexFilter
		err    error
	)

	if ctx.Param("date") != "" {
		filter.Date, err = time.Parse(time.DateOnly, ctx.Param("date"))
		if err != nil {
			return nil, http.ErrorInvalidParam{Params: []string{"date"}}
		}
	}

	indices, err := h.svc.List(ctx, &filter)
	if err != nil {
		return nil, err
	}

	var resp = make([]*Index, len(indices))

	for i := range indices {
		resp[i] = h.buildResp(indices[i])
	}

	return response.Raw{Data: map[string]any{
		"data": resp,
		"meta": map[string]any{
			"total": len(indices),
		},
	}}, nil
}

func (*indexHandler) buildResp(model *services.Index) *Index {
	resp := &Index{
		ID:            model.ID,
		Name:          model.Name,
		Value:         model.Value,
		PreviousClose: model.PreviousClose,
		CreatedAt:     model.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     model.UpdatedAt.Format(time.RFC3339),
		MarketData:    nil,
		Constituents:  make([]*IndexConstituent, len(model.Constituents)),
	}

	for i, c := range model.Constituents {
		resp.Constituents[i] = &IndexConstituent{
			ID:              c.ID,
			IndexID:         c.IndexID,
			SecurityID:      c.SecurityID,
			ISIN:            c.ISIN,
			Symbol:          c.Symbol,
			Industry:        c.Industry,
			Name:            c.Name,
			Image:           c.Image,
			LTP:             c.LTP,
			Volume:          c.Volume,
			FreeFloatShares: c.FreeFloatShares,
			PreviousClose:   c.PreviousClose,
		}
	}

	if model.IndexStat == nil {
		return resp
	}

	resp.MarketData = &IndexMarketData{
		Date:  model.IndexStat.Date.Format(time.DateOnly),
		Open:  model.IndexStat.Open,
		Close: model.IndexStat.Close,
		High:  model.IndexStat.High,
		Low:   model.IndexStat.Low,
	}

	return resp
}
