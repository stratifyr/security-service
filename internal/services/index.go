package services

import (
	"time"

	"gofr.dev/pkg/gofr"

	"github.com/stratifyr/security-service/internal/stores"
)

type IndexService interface {
	List(ctx *gofr.Context) ([]*Index, error)
	Upsert(ctx *gofr.Context, payload *IndexUpsert) (*Index, error)
}

type Index struct {
	ID           int
	Name         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Constituents []*IndexConstituent
}

type IndexConstituent struct {
	ID              int
	IndexID         int
	SecurityID      int
	ISIN            string
	Symbol          string
	Industry        string
	Name            string
	Image           string
	LTP             float64
	Volume          int
	FreeFloatShares int
	PreviousClose   float64
}

type IndexUpsert struct {
	Name        string
	SecurityIDs []int
}

type indexService struct {
	securityService SecurityService
	store           stores.IndexStore
}

func NewIndexService(securityService SecurityService, store stores.IndexStore) *indexService {
	return &indexService{
		securityService: securityService,
		store:           store,
	}
}

func (s *indexService) List(ctx *gofr.Context) ([]*Index, error) {
	indices, err := s.store.List(ctx, &stores.IndexFilter{}, 0, 0)
	if err != nil {
		return nil, err
	}

	if len(indices) == 0 {
		return nil, nil
	}

	securities, err := s.securityService.Index(ctx, &SecurityFilter{})
	if err != nil {
		return nil, err
	}

	resp := make([]*Index, len(indices))

	for i := range indices {
		resp[i] = s.buildResp(indices[i], securities)
	}

	return resp, nil
}

func (s *indexService) Upsert(ctx *gofr.Context, payload *IndexUpsert) (*Index, error) {
	indices, err := s.store.List(ctx, &stores.IndexFilter{Name: payload.Name}, 1, 0)
	if err != nil {
		return nil, err
	}

	if len(indices) > 0 {
		return s.patch(ctx, indices[0].ID, payload)
	}

	model := &stores.Index{
		Name:         payload.Name,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Constituents: make([]*stores.IndexConstituent, len(payload.SecurityIDs)),
	}

	for i := range payload.SecurityIDs {
		model.Constituents[i] = &stores.IndexConstituent{
			SecurityID: payload.SecurityIDs[i],
		}
	}

	index, err := s.store.Create(ctx, model)
	if err != nil {
		return nil, err
	}

	securities, err := s.securityService.Index(ctx, &SecurityFilter{})
	if err != nil {
		return nil, err
	}

	return s.buildResp(index, securities), nil
}

func (s *indexService) patch(ctx *gofr.Context, id int, payload *IndexUpsert) (*Index, error) {
	index, err := s.store.Retrieve(ctx, id)
	if err != nil {
		return nil, err
	}

	if payload.Name != "" {
		index.Name = payload.Name
		index.UpdatedAt = time.Now()
	}

	if payload.SecurityIDs != nil {
		index.Constituents = make([]*stores.IndexConstituent, len(payload.SecurityIDs))
		index.UpdatedAt = time.Now()

		for i := range payload.SecurityIDs {
			index.Constituents[i] = &stores.IndexConstituent{
				IndexID:    id,
				SecurityID: payload.SecurityIDs[i],
			}
		}
	}

	index, err = s.store.Update(ctx, index.ID, index)
	if err != nil {
		return nil, err
	}

	securities, err := s.securityService.Index(ctx, &SecurityFilter{})
	if err != nil {
		return nil, err
	}

	return s.buildResp(index, securities), nil
}

func (s *indexService) Delete(ctx *gofr.Context, id int) error {
	_, err := s.store.Retrieve(ctx, id)
	if err != nil {
		return err
	}

	return s.store.Delete(ctx, id)
}

func (*indexService) buildResp(model *stores.Index, securities []*Security) *Index {
	var securitiesByID = make(map[int]*Security)

	for _, security := range securities {
		securitiesByID[security.ID] = security
	}

	resp := &Index{
		ID:           model.ID,
		Name:         model.Name,
		CreatedAt:    model.CreatedAt,
		UpdatedAt:    model.UpdatedAt,
		Constituents: make([]*IndexConstituent, len(model.Constituents)),
	}

	for i := range model.Constituents {
		security, ok := securitiesByID[model.Constituents[i].SecurityID]
		if !ok {
			continue
		}

		resp.Constituents[i] = &IndexConstituent{
			ID:              model.Constituents[i].ID,
			IndexID:         model.Constituents[i].IndexID,
			SecurityID:      model.Constituents[i].SecurityID,
			ISIN:            security.ISIN,
			Symbol:          security.Symbol,
			Industry:        security.Industry,
			Name:            security.Name,
			Image:           security.Image,
			LTP:             security.LTP,
			Volume:          security.Volume,
			FreeFloatShares: security.FreeFloatShares,
			PreviousClose:   security.PreviousClose,
		}
	}

	return resp
}
