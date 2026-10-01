package services

import (
	"time"

	"gofr.dev/pkg/gofr"

	"github.com/stratifyr/security-service/internal/stores"
)

type SecurityStatService interface {
	Upsert(ctx *gofr.Context, payload *SecurityStatUpsert) (*SecurityStat, error)
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

type SecurityStatUpsert struct {
	SecurityID int
	Date       time.Time
	Open       float64
	Close      float64
	High       float64
	Low        float64
	Volume     int
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

func (s *securityStatService) Upsert(ctx *gofr.Context, payload *SecurityStatUpsert) (*SecurityStat, error) {
	securityStats, err := s.store.List(ctx, &stores.SecurityStatFilter{SecurityIDs: []int{payload.SecurityID}, Date: payload.Date}, 1, 0)
	if err != nil {
		return nil, err
	}

	if len(securityStats) > 0 {
		return s.patch(ctx, securityStats[0].ID, payload)
	}

	marketDays, count, err := s.marketDayService.List(ctx,
		&MarketDayFilter{DateBetween: &struct {
			StartDate time.Time
			EndDate   time.Time
		}{StartDate: payload.Date, EndDate: payload.Date}})
	if err != nil {
		return nil, err
	}

	if count != 1 || marketDays[0].Format(time.DateOnly) != payload.Date.Format(time.DateOnly) {
		return nil, ErrMarketHolidayStat
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

func (s *securityStatService) patch(ctx *gofr.Context, id int, payload *SecurityStatUpsert) (*SecurityStat, error) {
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

func (*securityStatService) buildResp(model *stores.SecurityStat) *SecurityStat {
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
