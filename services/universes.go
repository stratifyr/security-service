package services

import (
	"database/sql"
	"fmt"
	"time"

	"gofr.dev/pkg/gofr"

	"github.com/stratifyr/security-service/stores"
)

type UniverseService interface {
	Index(ctx *gofr.Context, f *SecurityFilter, page, perPage int) ([]*Security, int, error)
	Read(ctx *gofr.Context, id int) (*Security, error)
	Create(ctx *gofr.Context, payload *UniverseCreatePayload) (*Security, error)
}

type Universe struct {
	ID        int    `json:"id"`
	UserID    *int   `json:"userId"`
	Name      string `json:"name"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`

	Securities []*struct {
		ID              int    `json:"id"`
		Symbol          string `json:"symbol"`
		Exchange        string `json:"exchange"`
		Industry        string `json:"industry"`
		Name            string `json:"name"`
		Image           string `json:"image"`
		LastTradedPrice string `json:"lastTradedPrice"`
	} `json:"securities"`
}

type UniverseCreatePayload struct {
	UserID      int    `json:"userId"`
	Name        string `json:"name"`
	SecurityIDs []int  `json:"securityIds"`
}

type UniverseFilter struct {
	UserID int
}

type universeService struct {
	securityService SecurityService
	store           stores.UniverseStore
}

func NewUniverseService(store stores.UniverseStore) *universeService {
	return &universeService{store: store}
}

func (s *universeService) Index(ctx *gofr.Context, f *UniverseFilter, page, perPage int) ([]*Universe, int, error) {
	limit := perPage
	offset := limit * (page - 1)

	filter := &stores.UniverseFilter{
		UserID: f.UserID,
	}

	universes, err := s.store.Index(ctx, filter, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	count, err := s.store.Count(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	if count == 0 {
		return nil, 0, nil
	}

	var resp = make([]*Universe, len(universes))

	for i := range universes {
		resp[i], err = s.buildResp(ctx, universes[i])
		if err != nil {
			return nil, 0, err
		}
	}

	return resp, count, nil
}

func (s *universeService) Read(ctx *gofr.Context, id int) (*Universe, error) {
	universe, err := s.store.Retrieve(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.buildResp(ctx, universe)
}

func (s *universeService) Create(ctx *gofr.Context, payload *UniverseCreatePayload) (*Universe, error) {
	model := &stores.Universe{
		UserID:      sql.NullInt64{Int64: int64(payload.UserID), Valid: true},
		Name:        payload.Name,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
		SecurityIDs: payload.SecurityIDs,
	}

	universe, err := s.store.Create(ctx, model)
	if err != nil {
		return nil, err
	}

	return s.buildResp(ctx, universe)
}

func (s *universeService) buildResp(ctx *gofr.Context, model *stores.Universe) (*Universe, error) {
	universe := &Universe{
		ID:         model.ID,
		UserID:     nil,
		Name:       model.Name,
		CreatedAt:  model.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  model.UpdatedAt.Format(time.RFC3339),
		Securities: nil,
	}

	if model.UserID.Valid {
		userId := int(model.UserID.Int64)
		universe.UserID = &userId
	}

	if len(model.SecurityIDs) == 0 {
		return universe, nil
	}

	universe.Securities = make([]*struct {
		ID              int    `json:"id"`
		Symbol          string `json:"symbol"`
		Exchange        string `json:"exchange"`
		Industry        string `json:"industry"`
		Name            string `json:"name"`
		Image           string `json:"image"`
		LastTradedPrice string `json:"lastTradedPrice"`
	}, len(model.SecurityIDs))

	for i := range model.SecurityIDs {
		security, err := s.securityService.Read(ctx, model.SecurityIDs[i])
		if err != nil {
			return nil, err
		}

		universe.Securities[i].ID = security.ID
		universe.Securities[i].Symbol = security.Symbol
		universe.Securities[i].Exchange = security.Exchange
		universe.Securities[i].Industry = security.Industry
		universe.Securities[i].Name = security.Name
		universe.Securities[i].Image = security.Image
		universe.Securities[i].LastTradedPrice = fmt.Sprintf("%0.2f", security.LTP)
	}

	return universe, nil
}
