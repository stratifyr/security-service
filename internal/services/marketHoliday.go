package services

import (
	"time"

	"gofr.dev/pkg/gofr"

	"github.com/stratifyr/security-service/internal/stores"
)

type MarketHolidayService interface {
	Index(ctx *gofr.Context, f *MarketHolidayFilter, page, perPage int) ([]*MarketHoliday, int, error)
	Read(ctx *gofr.Context, id int) (*MarketHoliday, error)
	Create(ctx *gofr.Context, payload *MarketHolidayCreate) (*MarketHoliday, error)
	Patch(ctx *gofr.Context, id int, payload *MarketHolidayUpdate) (*MarketHoliday, error)
	Delete(ctx *gofr.Context, id, userID int) error
}

type MarketHolidayFilter struct {
	Date        time.Time
	DateBetween *struct {
		StartDate time.Time
		EndDate   time.Time
	}
}

type MarketHoliday struct {
	ID          int
	Date        time.Time
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type MarketHolidayCreate struct {
	UserID      int
	Date        time.Time
	Description string
}

type MarketHolidayUpdate struct {
	UserID      int
	Date        time.Time
	Description string
}

type marketHolidayService struct {
	store stores.MarketHolidayStore
}

func NewMarketHolidayService(store stores.MarketHolidayStore) *marketHolidayService {
	return &marketHolidayService{store: store}
}

func (s *marketHolidayService) Index(ctx *gofr.Context, f *MarketHolidayFilter, page, perPage int) ([]*MarketHoliday, int, error) {
	limit := perPage
	offset := limit * (page - 1)

	filter := &stores.MarketHolidayFilter{
		Date:        f.Date,
		DateBetween: f.DateBetween,
	}

	marketHolidays, err := s.store.Index(ctx, filter, limit, offset)
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
		resp[i] = s.buildResp(marketHolidays[i])
	}

	return resp, count, nil
}

func (s *marketHolidayService) Read(ctx *gofr.Context, id int) (*MarketHoliday, error) {
	marketHoliday, err := s.store.Retrieve(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.buildResp(marketHoliday), nil
}

func (s *marketHolidayService) Create(ctx *gofr.Context, payload *MarketHolidayCreate) (*MarketHoliday, error) {
	if payload.UserID != 1 {
		return nil, &ErrResp{Code: 403, Message: "Forbidden"}
	}

	model := &stores.MarketHoliday{
		Date:        payload.Date,
		Description: payload.Description,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	marketHoliday, err := s.store.Create(ctx, model)
	if err != nil {
		return nil, err
	}

	return s.buildResp(marketHoliday), nil
}

func (s *marketHolidayService) Patch(ctx *gofr.Context, id int, payload *MarketHolidayUpdate) (*MarketHoliday, error) {
	if payload.UserID != 1 {
		return nil, &ErrResp{Code: 403, Message: "Forbidden"}
	}

	marketHoliday, err := s.store.Retrieve(ctx, id)
	if err != nil {
		return nil, err
	}

	if payload.Date != (time.Time{}) {
		marketHoliday.Date = payload.Date
	}

	if payload.Description != "" {
		marketHoliday.Description = payload.Description
	}

	marketHoliday, err = s.store.Update(ctx, id, marketHoliday)
	if err != nil {
		return nil, err
	}

	return s.buildResp(marketHoliday), nil
}

func (s *marketHolidayService) Delete(ctx *gofr.Context, id, userID int) error {
	if userID != 1 {
		return &ErrResp{Code: 403, Message: "Forbidden"}
	}

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

func (s *marketHolidayService) buildResp(model *stores.MarketHoliday) *MarketHoliday {
	resp := &MarketHoliday{
		ID:          model.ID,
		Date:        model.Date,
		Description: model.Description,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}

	return resp
}
