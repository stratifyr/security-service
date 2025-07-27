package services

import (
	"time"

	"gofr.dev/pkg/gofr"

	"github.com/stratifyr/security-service/internal/stores"
)

type SecurityStatService interface {
	Index(ctx *gofr.Context, f *SecurityStatFilter, page, perPage int) ([]*SecurityStat, int, error)
	Read(ctx *gofr.Context, id int) (*SecurityStat, error)
	Create(ctx *gofr.Context, payload *SecurityStatCreate) (*SecurityStat, error)
	Patch(ctx *gofr.Context, id int, payload *SecurityStatUpdate) (*SecurityStat, error)
}

type SecurityStatFilter struct {
	Date       time.Time
	SecurityID int
}

type SecurityStat struct {
	ID         int
	SecurityID int
	Date       time.Time
	Open       float64
	Close      float64
	High       float64
	Low        float64
	Volume     int
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type SecurityStatCreate struct {
	UserID     int
	SecurityID int
	Date       time.Time
	Open       float64
	Close      float64
	High       float64
	Low        float64
	Volume     int
}

type SecurityStatUpdate struct {
	UserID int
	Open   float64
	Close  float64
	High   float64
	Low    float64
	Volume int
}

type securityStatService struct {
	marketDayService MarketDayService
	store            stores.SecurityStatStore
}

func NewSecurityStatService(marketDayService MarketDayService, store stores.SecurityStatStore) *securityStatService {
	return &securityStatService{
		marketDayService: marketDayService,
		store:            store,
	}
}

func (s *securityStatService) Index(ctx *gofr.Context, f *SecurityStatFilter, page, perPage int) ([]*SecurityStat, int, error) {
	limit := perPage
	offset := limit * (page - 1)

	var filter stores.SecurityStatFilter

	if f.SecurityID != 0 {
		filter.SecurityIDs = []int{f.SecurityID}
	}

	if f.Date != (time.Time{}) {
		filter.Dates = []time.Time{f.Date}
	}

	securityStats, err := s.store.Index(ctx, &filter, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	count, err := s.store.Count(ctx, &filter)
	if err != nil {
		return nil, 0, err
	}

	if count == 0 {
		return nil, 0, nil
	}

	var resp = make([]*SecurityStat, len(securityStats))

	for i := range securityStats {
		resp[i] = s.buildResp(securityStats[i])
	}

	return resp, count, nil
}

func (s *securityStatService) Read(ctx *gofr.Context, id int) (*SecurityStat, error) {
	securityStat, err := s.store.Retrieve(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.buildResp(securityStat), nil
}

func (s *securityStatService) Create(ctx *gofr.Context, payload *SecurityStatCreate) (*SecurityStat, error) {
	if payload.UserID != 1 {
		return nil, &ErrResp{Code: 403, Message: "Forbidden"}
	}

	marketDays, count, err := s.marketDayService.Index(ctx,
		&MarketDayFilter{DateBetween: &struct {
			StartDate time.Time
			EndDate   time.Time
		}{StartDate: payload.Date, EndDate: payload.Date}})
	if count != 1 || marketDays[0].Format(time.DateOnly) != payload.Date.Format(time.DateOnly) {
		return nil, &ErrResp{Code: 400, Message: "cannot add stat for market holiday - " + payload.Date.Format(time.DateOnly)}
	}

	model := &stores.SecurityStat{
		SecurityID: payload.SecurityID,
		Date:       payload.Date,
		Open:       payload.Open,
		Close:      payload.Close,
		High:       payload.High,
		Low:        payload.Low,
		Volume:     payload.Volume,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}

	securityStat, err := s.store.Create(ctx, model)
	if err != nil {
		return nil, err
	}

	return s.buildResp(securityStat), nil
}

func (s *securityStatService) Patch(ctx *gofr.Context, id int, payload *SecurityStatUpdate) (*SecurityStat, error) {
	if payload.UserID != 1 {
		return nil, &ErrResp{Code: 403, Message: "Forbidden"}
	}

	securityStat, err := s.store.Retrieve(ctx, id)
	if err != nil {
		return nil, err
	}

	if payload.Open != 0 {
		securityStat.Open = payload.Open
	}

	if payload.Close != 0 {
		securityStat.Close = payload.Close
	}

	if payload.High != 0 {
		securityStat.High = payload.High
	}

	if payload.Low != 0 {
		securityStat.Low = payload.Low
	}

	if payload.Volume != 0 {
		securityStat.Volume = payload.Volume
	}

	securityStat, err = s.store.Update(ctx, id, securityStat)
	if err != nil {
		return nil, err
	}

	return s.buildResp(securityStat), nil
}

func (s *securityStatService) buildResp(model *stores.SecurityStat) *SecurityStat {
	resp := &SecurityStat{
		ID:         model.ID,
		SecurityID: model.SecurityID,
		Date:       model.Date,
		Open:       model.Open,
		Close:      model.Close,
		High:       model.High,
		Low:        model.Low,
		Volume:     model.Volume,
		CreatedAt:  model.CreatedAt,
		UpdatedAt:  model.UpdatedAt,
	}

	return resp
}
