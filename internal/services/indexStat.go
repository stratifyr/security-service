package services

import (
	"time"

	"gofr.dev/pkg/gofr"

	"github.com/stratifyr/security-service/internal/stores"
)

type IndexStatService interface {
	List(ctx *gofr.Context, f *IndexStatFilter, page, perPage int) ([]*IndexStat, int, error)
	Upsert(ctx *gofr.Context, payload *IndexStatUpsert) (*IndexStat, error)
}

type IndexStatFilter struct {
	IndexID int
	Date    time.Time
}

type IndexStat struct {
	ID        int
	IndexID   int
	Date      time.Time
	Open      float64
	Close     float64
	High      float64
	Low       float64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type IndexStatUpsert struct {
	IndexID int
	Date    time.Time
	Open    float64
	Close   float64
	High    float64
	Low     float64
}

type indexStatService struct {
	marketDayService MarketDayService
	store            stores.IndexStatStore
}

func NewIndexStatService(marketDayService MarketDayService, store stores.IndexStatStore) *indexStatService {
	return &indexStatService{
		marketDayService: marketDayService,
		store:            store,
	}
}

func (s *indexStatService) List(ctx *gofr.Context, f *IndexStatFilter, page, perPage int) ([]*IndexStat, int, error) {
	limit := perPage
	offset := limit * (page - 1)

	var filter stores.IndexStatFilter

	if f.IndexID != 0 {
		filter.IndexIDs = []int{f.IndexID}
	}

	if !f.Date.IsZero() {
		filter.Date = f.Date
	}

	indexStats, err := s.store.List(ctx, &filter, limit, offset)
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

	var resp = make([]*IndexStat, len(indexStats))

	for i := range indexStats {
		resp[i] = s.buildResp(indexStats[i])
	}

	return resp, count, nil
}

func (s *indexStatService) Upsert(ctx *gofr.Context, payload *IndexStatUpsert) (*IndexStat, error) {
	indexStats, err := s.store.List(ctx, &stores.IndexStatFilter{IndexIDs: []int{payload.IndexID}, Date: payload.Date}, 1, 0)
	if err != nil {
		return nil, err
	}

	if len(indexStats) > 0 {
		return s.patch(ctx, indexStats[0].ID, payload)
	}

	marketDays, count, err := s.marketDayService.Index(ctx,
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

	model := &stores.IndexStat{
		IndexID:   payload.IndexID,
		Date:      payload.Date,
		Open:      payload.Open,
		Close:     payload.Close,
		High:      payload.High,
		Low:       payload.Low,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	indexStat, err := s.store.Create(ctx, model)
	if err != nil {
		return nil, err
	}

	return s.buildResp(indexStat), nil
}

func (s *indexStatService) patch(ctx *gofr.Context, id int, payload *IndexStatUpsert) (*IndexStat, error) {
	indexStat, err := s.store.Retrieve(ctx, id)
	if err != nil {
		return nil, err
	}

	if payload.Open != 0 {
		indexStat.Open = payload.Open
	}

	if payload.Close != 0 {
		indexStat.Close = payload.Close
	}

	if payload.High != 0 {
		indexStat.High = payload.High
	}

	if payload.Low != 0 {
		indexStat.Low = payload.Low
	}

	indexStat, err = s.store.Update(ctx, id, indexStat)
	if err != nil {
		return nil, err
	}

	return s.buildResp(indexStat), nil
}

func (*indexStatService) buildResp(model *stores.IndexStat) *IndexStat {
	resp := &IndexStat{
		ID:        model.ID,
		IndexID:   model.IndexID,
		Date:      model.Date,
		Open:      model.Open,
		Close:     model.Close,
		High:      model.High,
		Low:       model.Low,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}

	return resp
}
