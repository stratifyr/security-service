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
	indices, err := h.svc.List(ctx)
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
		Id:        int32(index.ID),
		Name:      index.Name,
		CreatedAt: index.CreatedAt.Format(time.RFC3339),
		UpdatedAt: index.UpdatedAt.Format(time.RFC3339),
	}

	if index.Constituents == nil {
		return resp
	}

	resp.Constituents = make([]*pb.IndexConstituent, len(index.Constituents))

	for i, constituent := range index.Constituents {
		resp.Constituents[i] = &pb.IndexConstituent{
			Id:            int32(constituent.ID),
			IndexId:       int32(constituent.IndexID),
			SecurityId:    int32(constituent.SecurityID),
			Isin:          constituent.ISIN,
			Symbol:        constituent.Symbol,
			Industry:      constituent.Industry,
			Name:          constituent.Name,
			Image:         constituent.Image,
			Ltp:           constituent.LTP,
			PreviousClose: constituent.PreviousClose,
		}
	}

	return resp
}
