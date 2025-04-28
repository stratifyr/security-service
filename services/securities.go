package services

import (
	"time"

	"gofr.dev/pkg/gofr"

	"github.com/stratifyr/security-service/stores"
)

type SecurityService interface {
	Index(ctx *gofr.Context, f *SecurityFilter, page, perPage int) ([]*Security, int, error)
	Read(ctx *gofr.Context, id int) (*Security, error)
	Create(ctx *gofr.Context, payload *SecurityCreate) (*Security, error)
}

type SecurityCreate struct {
	ISIN     string
	Symbol   string
	Exchange string
	Industry string
	Name     string
	Image    string
	LTP      float64
}

type Security struct {
	ID        int
	ISIN      string
	Symbol    string
	Exchange  string
	Industry  string
	Name      string
	Image     string
	LTP       float64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type SecurityFilter struct {
	ID       []int
	Symbol   string
	Exchange string
}

type securityService struct {
	store stores.SecurityStore
}

func NewSecurityService(store stores.SecurityStore) *securityService {
	return &securityService{store: store}
}

func (s *securityService) Index(ctx *gofr.Context, f *SecurityFilter, page, perPage int) ([]*Security, int, error) {
	limit := perPage
	offset := limit * (page - 1)

	filter := &stores.SecurityFilter{
		ID:       f.ID,
		Symbol:   f.Symbol,
		Exchange: nil,
	}

	if f.Exchange != "" {
		exchange, err := stores.ExchangeFromString(f.Exchange)
		if err != nil {
			return nil, 0, err
		}

		filter.Exchange = &exchange
	}

	securities, err := s.store.Index(ctx, filter, limit, offset)
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

	var resp = make([]*Security, len(securities))

	for i := range securities {
		resp[i] = s.buildResp(securities[i])
	}

	return resp, count, nil
}

func (s *securityService) Read(ctx *gofr.Context, id int) (*Security, error) {
	security, err := s.store.Retrieve(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.buildResp(security), nil
}

func (s *securityService) Create(ctx *gofr.Context, payload *SecurityCreate) (*Security, error) {
	exchange, err := stores.ExchangeFromString(payload.Exchange)
	if err != nil {
		return nil, err
	}

	industry, err := stores.IndustryFromString(payload.Industry)
	if err != nil {
		return nil, err
	}

	model := &stores.Security{
		ISIN:      payload.ISIN,
		Symbol:    payload.Symbol,
		Exchange:  exchange,
		Industry:  industry,
		Name:      payload.Name,
		Image:     payload.Image,
		LTP:       payload.LTP,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	security, err := s.store.Create(ctx, model)
	if err != nil {
		return nil, err
	}

	return s.buildResp(security), nil
}

func (s *securityService) buildResp(model *stores.Security) *Security {
	security := &Security{
		ID:        model.ID,
		ISIN:      model.ISIN,
		Symbol:    model.Symbol,
		Exchange:  model.Exchange.String(),
		Industry:  model.Industry.String(),
		Name:      model.Name,
		Image:     model.Image,
		LTP:       model.LTP,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.CreatedAt,
	}

	return security
}
