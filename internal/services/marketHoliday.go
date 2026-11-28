package services

import (
	"time"

	"gofr.dev/pkg/gofr"

	"github.com/stratifyr/security-service/internal/stores"
)

type MarketHolidayService interface {
	List(ctx *gofr.Context, f *MarketHolidayFilter, page, perPage int) ([]*MarketHoliday, int, error)
	Create(ctx *gofr.Context, payload *MarketHolidayCreate) (*MarketHoliday, error)
	Patch(ctx *gofr.Context, id int, payload *MarketHolidayUpdate) (*MarketHoliday, error)
	Delete(ctx *gofr.Context, id int) error
}

type MarketHolidayFilter struct {
	Exchange    string
	Date        time.Time
	DateBetween *struct {
		StartDate time.Time
		EndDate   time.Time
	}
}

type MarketHoliday struct {
	ID          int
	Exchange    string
	Date        time.Time
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type MarketHolidayCreate struct {
	Exchange    string
	Date        time.Time
	Description string
}

type MarketHolidayUpdate struct {
	Exchange    string
	Date        time.Time
	Description string
}

type marketHolidayService struct {
	exchangeService ExchangeService
	store           stores.MarketHolidayStore
}

func NewMarketHolidayService(exchangeService ExchangeService, store stores.MarketHolidayStore) *marketHolidayService {
	return &marketHolidayService{store: store, exchangeService: exchangeService}
}

func (s *marketHolidayService) List(ctx *gofr.Context, f *MarketHolidayFilter, page, perPage int) ([]*MarketHoliday, int, error) {
	limit := perPage
	offset := limit * (page - 1)

	exchange, err := s.exchangeService.GetByName(ctx, f.Exchange)
	if err != nil {
		return nil, 0, err
	}

	filter := &stores.MarketHolidayFilter{
		ExchangeID:  exchange.ID,
		Date:        f.Date,
		DateBetween: f.DateBetween,
	}

	marketHolidays, err := s.store.List(ctx, filter, limit, offset)
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

	var resp = make([]*MarketHoliday, len(marketHolidays))

	for i := range marketHolidays {
		resp[i] = s.buildResp(marketHolidays[i], exchange)
	}

	return resp, count, nil
}

func (s *marketHolidayService) Create(ctx *gofr.Context, payload *MarketHolidayCreate) (*MarketHoliday, error) {
	exchange, err := s.exchangeService.GetByName(ctx, payload.Exchange)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()

	model := &stores.MarketHoliday{
		ExchangeID:  exchange.ID,
		Date:        payload.Date,
		Description: payload.Description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	marketHoliday, err := s.store.Create(ctx, model)
	if err != nil {
		return nil, err
	}

	return s.buildResp(marketHoliday, exchange), nil
}

func (s *marketHolidayService) Patch(ctx *gofr.Context, id int, payload *MarketHolidayUpdate) (*MarketHoliday, error) {
	marketHoliday, err := s.store.Retrieve(ctx, id)
	if err != nil {
		return nil, err
	}

	if payload.Exchange != "" {
		exchange, err := s.exchangeService.GetByName(ctx, payload.Exchange)
		if err != nil {
			return nil, err
		}

		marketHoliday.ExchangeID = exchange.ID
	}

	if payload.Date != (time.Time{}) {
		marketHoliday.Date = payload.Date
	}

	if payload.Description != "" {
		marketHoliday.Description = payload.Description
	}

	marketHoliday.UpdatedAt = time.Now()

	marketHoliday, err = s.store.Update(ctx, id, marketHoliday)
	if err != nil {
		return nil, err
	}

	exchange, err := s.exchangeService.GetByID(ctx, marketHoliday.ExchangeID)
	if err != nil {
		return nil, err
	}

	return s.buildResp(marketHoliday, exchange), nil
}

func (s *marketHolidayService) Delete(ctx *gofr.Context, id int) error {
	_, err := s.store.Retrieve(ctx, id)
	if err != nil {
		return err
	}

	err = s.store.Delete(ctx, id)
	if err != nil {
		return err
	}

	return nil
}

func (*marketHolidayService) buildResp(model *stores.MarketHoliday, exchange *Exchange) *MarketHoliday {
	resp := &MarketHoliday{
		ID:          model.ID,
		Exchange:    exchange.Name,
		Date:        model.Date,
		Description: model.Description,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}

	return resp
}
