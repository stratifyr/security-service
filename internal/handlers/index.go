package handlers

import (
	"time"

	"gofr.dev/pkg/gofr"
	"gofr.dev/pkg/gofr/http/response"

	"github.com/stratifyr/security-service/internal/services"
)

type Index struct {
	ID           int                 `json:"id"`
	Name         string              `json:"name"`
	CreatedAt    string              `json:"createdAt"`
	UpdatedAt    string              `json:"updatedAt"`
	Constituents []*IndexConstituent `json:"constituents"`
}

type IndexConstituent struct {
	ID            int     `json:"id"`
	IndexID       int     `json:"indexId"`
	SecurityID    int     `json:"securityId"`
	ISIN          string  `json:"isin"`
	Symbol        string  `json:"symbol"`
	Industry      string  `json:"industry"`
	Name          string  `json:"name"`
	Image         string  `json:"image"`
	LTP           float64 `json:"ltp"`
	PreviousClose float64 `json:"previousClose"`
}

type indexHandler struct {
	svc services.IndexService
}

func NewIndexHandler(svc services.IndexService) *indexHandler {
	return &indexHandler{svc: svc}
}

func (h *indexHandler) List(ctx *gofr.Context) (any, error) {
	indices, err := h.svc.List(ctx)
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
		ID:           model.ID,
		Name:         model.Name,
		CreatedAt:    model.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    model.UpdatedAt.Format(time.RFC3339),
		Constituents: make([]*IndexConstituent, len(model.Constituents)),
	}

	for i, c := range model.Constituents {
		resp.Constituents[i] = &IndexConstituent{
			ID:            c.ID,
			IndexID:       c.IndexID,
			SecurityID:    c.SecurityID,
			ISIN:          c.ISIN,
			Symbol:        c.Symbol,
			Industry:      c.Industry,
			Name:          c.Name,
			Image:         c.Image,
			LTP:           c.LTP,
			PreviousClose: c.PreviousClose,
		}
	}

	return resp
}
