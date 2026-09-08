package handlers

import (
	"time"

	"github.com/stratifyr/security-service-proto/go/pb"
	"gofr.dev/pkg/gofr"

	"github.com/stratifyr/security-service/internal/services"
)

type indexGRPCHandler struct {
	svc services.IndexService
}

func NewIndexGRPCHandler(svc services.IndexService) *indexGRPCHandler {
	return &indexGRPCHandler{svc: svc}
}

func (h *indexGRPCHandler) List(ctx *gofr.Context) (any, error) {
	var payload pb.GetIndicesRequest

	if err := ctx.Bind(&payload); err != nil {
		return nil, err
	}

	var (
		filter services.IndexFilter
		err    error
	)

	if payload.Date != "" {
		filter.Date, err = time.Parse(time.DateOnly, payload.Date)
		if err != nil {
			return nil, err
		}
	}

	indices, err := h.svc.List(ctx, &filter)
	if err != nil {
		return nil, err
	}

	var resp = make([]*pb.Index, len(indices))

	for i := range indices {
		resp[i] = h.buildResponse(indices[i])
	}

	return &pb.GetIndicesResponse{
		Indices: resp,
		Total:   int32(len(indices)),
	}, nil
}

func (h *indexGRPCHandler) Upsert(ctx *gofr.Context) (any, error) {
	var payload pb.UpsertIndexRequest

	if err := ctx.Bind(&payload); err != nil {
		return nil, err
	}

	var securityIDs = make([]int, len(payload.SecurityIds))
	for i := range payload.SecurityIds {
		securityIDs[i] = int(payload.SecurityIds[i])
	}

	model := &services.IndexUpsert{
		Name:        payload.Name,
		Value:       payload.Value,
		SecurityIDs: securityIDs,
	}

	index, err := h.svc.Upsert(ctx, model)
	if err != nil {
		return nil, err
	}

	return &pb.UpsertIndexResponse{
		Index: h.buildResponse(index),
	}, nil
}

func (*indexGRPCHandler) buildResponse(index *services.Index) *pb.Index {
	var resp = &pb.Index{
		Id:            int32(index.ID),
		Name:          index.Name,
		Value:         index.Value,
		PreviousClose: index.PreviousClose,
		CreatedAt:     index.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     index.UpdatedAt.Format(time.RFC3339),
		MarketData:    nil,
		Constituents:  nil,
	}

	if index.Constituents != nil {
		resp.Constituents = make([]*pb.IndexConstituent, len(index.Constituents))

		for i, constituent := range index.Constituents {
			resp.Constituents[i] = &pb.IndexConstituent{
				Id:              int32(constituent.ID),
				IndexId:         int32(constituent.IndexID),
				SecurityId:      int32(constituent.SecurityID),
				Isin:            constituent.ISIN,
				Symbol:          constituent.Symbol,
				Industry:        constituent.Industry,
				Name:            constituent.Name,
				Image:           constituent.Image,
				Ltp:             constituent.LTP,
				Volume:          int64(constituent.Volume),
				FreeFloatShares: int64(constituent.FreeFloatShares),
				PreviousClose:   constituent.PreviousClose,
			}
		}
	}

	if index.IndexStat == nil {
		return resp
	}

	resp.MarketData = &pb.IndexMarketData{
		Date:  index.IndexStat.Date.Format(time.DateOnly),
		Open:  index.IndexStat.Open,
		Close: index.IndexStat.Close,
		High:  index.IndexStat.High,
		Low:   index.IndexStat.Low,
	}

	return resp
}
