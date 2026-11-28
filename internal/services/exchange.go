package services

import (
	"time"

	"gofr.dev/pkg/gofr"

	"github.com/stratifyr/security-service/internal/stores"
)

type ExchangeService interface {
	List(ctx *gofr.Context) ([]*Exchange, error)
	GetByID(ctx *gofr.Context, id int) (*Exchange, error)
	GetByName(ctx *gofr.Context, name string) (*Exchange, error)
	Upsert(ctx *gofr.Context, payload *ExchangeUpsert) (*Exchange, error)
}

type Exchange struct {
	ID        int
	Name      string
	Currency  stores.Currency
	ForexRate float64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ExchangeUpsert struct {
	Name      string
	Currency  string
	ForexRate float64
}

type exchangeService struct {
	store stores.ExchangeStore
}

func NewExchangeService(exchangeStore stores.ExchangeStore) *exchangeService {
	return &exchangeService{store: exchangeStore}
}

func (s *exchangeService) List(ctx *gofr.Context) ([]*Exchange, error) {
	exchanges, err := s.store.List(ctx, &stores.ExchangeFilter{}, 0, 0)
	if err != nil {
		return nil, err
	}

	if len(exchanges) == 0 {
		return nil, nil
	}

	var resp = make([]*Exchange, len(exchanges))

	for i := range exchanges {
		resp[i] = s.buildResponse(exchanges[i])
	}

	return resp, nil
}

func (s *exchangeService) GetByID(ctx *gofr.Context, id int) (*Exchange, error) {
	exchange, err := s.store.Retrieve(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.buildResponse(exchange), nil
}

func (s *exchangeService) GetByName(ctx *gofr.Context, name string) (*Exchange, error) {
	if name == "" {
		return nil, ErrInvalidExchange
	}

	exchanges, err := s.store.List(ctx, &stores.ExchangeFilter{Name: name}, 1, 0)
	if err != nil {
		return nil, err
	}

	if len(exchanges) == 0 {
		return nil, ErrInvalidExchange
	}

	return s.buildResponse(exchanges[0]), nil
}

func (s *exchangeService) Upsert(ctx *gofr.Context, payload *ExchangeUpsert) (*Exchange, error) {
	if payload.Name == "" {
		return nil, ErrInvalidExchange
	}

	exchanges, err := s.store.List(ctx, &stores.ExchangeFilter{Name: payload.Name}, 1, 0)
	if err != nil {
		return nil, err
	}

	if len(exchanges) > 0 {
		return s.patch(ctx, exchanges[0].ID, payload)
	}

	currency, err := stores.CurrencyFromString(payload.Currency)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()

	model := &stores.Exchange{
		Name:      payload.Name,
		Currency:  currency,
		ForexRate: payload.ForexRate,
		CreatedAt: now,
		UpdatedAt: now,
	}

	exchange, err := s.store.Create(ctx, model)
	if err != nil {
		return nil, err
	}

	return s.buildResponse(exchange), nil
}

func (s *exchangeService) patch(ctx *gofr.Context, id int, payload *ExchangeUpsert) (*Exchange, error) {
	exchange, err := s.store.Retrieve(ctx, id)
	if err != nil {
		return nil, err
	}

	if payload.Currency != "" {
		currency, err := stores.CurrencyFromString(payload.Currency)
		if err != nil {
			return nil, err
		}

		exchange.Currency = currency
		exchange.UpdatedAt = time.Now().UTC()
	}

	if payload.ForexRate != 0 {
		exchange.ForexRate = payload.ForexRate
		exchange.UpdatedAt = time.Now().UTC()
	}

	exchange, err = s.store.Update(ctx, id, exchange)
	if err != nil {
		return nil, err
	}

	return s.buildResponse(exchange), nil
}

func (*exchangeService) buildResponse(model *stores.Exchange) *Exchange {
	return &Exchange{
		ID:        model.ID,
		Name:      model.Name,
		Currency:  model.Currency,
		ForexRate: model.ForexRate,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}
}
