package services

import (
	"time"

	"gofr.dev/pkg/gofr"

	"github.com/stratifyr/security-service/internal/stores"
)

type IndexService interface {
	List(ctx *gofr.Context, filter *IndexFilter) ([]*Index, error)
	Upsert(ctx *gofr.Context, payload *IndexUpsert) (*Index, error)
}

type IndexFilter struct {
	Date time.Time
}

type Index struct {
	ID            int
	Name          string
	Value         float64
	PreviousClose float64
	CreatedAt     time.Time
	UpdatedAt     time.Time
	IndexStat     *IndexStat
	Constituents  []*IndexConstituent
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
	Value       float64
	SecurityIDs []int
}

type indexService struct {
	marketDayService MarketDayService
	securityService  SecurityService
	indexStatStore   stores.IndexStatStore
	store            stores.IndexStore
}

func NewIndexService(marketDayService MarketDayService, securityService SecurityService,
	indexStatStore stores.IndexStatStore, store stores.IndexStore) *indexService {
	return &indexService{
		marketDayService: marketDayService,
		securityService:  securityService,
		indexStatStore:   indexStatStore,
		store:            store,
	}
}

func (s *indexService) List(ctx *gofr.Context, filter *IndexFilter) ([]*Index, error) {
	if filter.Date.IsZero() {
		filter.Date = time.Now()
	}

	indices, err := s.store.List(ctx, &stores.IndexFilter{}, 0, 0)
	if err != nil {
		return nil, err
	}

	if len(indices) == 0 {
		return nil, nil
	}

	prevMarketDay, err := s.getPrevMarketDay(ctx, filter.Date)
	if err != nil {
		return nil, err
	}

	var indexIDs = make([]int, len(indices))
	for i, index := range indices {
		indexIDs[i] = index.ID
	}

	indexStats, err := s.getStatsMap(ctx, indexIDs, prevMarketDay)
	if err != nil {
		return nil, err
	}

	securities, err := s.securityService.Index(ctx, &SecurityFilter{})
	if err != nil {
		return nil, err
	}

	resp := make([]*Index, len(indices))

	for i := range indices {
		resp[i] = s.buildResp(indices[i], indexStats, securities)
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
		Value:        payload.Value,
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

	prevMarketDay, err := s.getPrevMarketDay(ctx, time.Now())
	if err != nil {
		return nil, err
	}

	indexStats, err := s.getStatsMap(ctx, []int{index.ID}, prevMarketDay)
	if err != nil {
		return nil, err
	}

	securities, err := s.securityService.Index(ctx, &SecurityFilter{})
	if err != nil {
		return nil, err
	}

	return s.buildResp(index, indexStats, securities), nil
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

	if payload.Value != 0 {
		index.Value = payload.Value
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

	prevMarketDay, err := s.getPrevMarketDay(ctx, time.Now())
	if err != nil {
		return nil, err
	}

	indexStats, err := s.getStatsMap(ctx, []int{id}, prevMarketDay)
	if err != nil {
		return nil, err
	}

	securities, err := s.securityService.Index(ctx, &SecurityFilter{})
	if err != nil {
		return nil, err
	}

	return s.buildResp(index, indexStats, securities), nil
}

func (s *indexService) getPrevMarketDay(ctx *gofr.Context, referenceDate time.Time) (time.Time, error) {
	dates, _, err := s.marketDayService.Index(ctx, &MarketDayFilter{LastNDaysFromReference: &struct {
		N         int
		Reference time.Time
	}{N: 2, Reference: referenceDate}})
	if err != nil {
		return time.Time{}, err
	}

	marketDay := dates[0]
	if dates[0].Format(time.DateOnly) == time.Now().Format(time.DateOnly) {
		marketDay = dates[1]
	}

	return marketDay, nil
}

func (s *indexService) getStatsMap(ctx *gofr.Context, indexIDs []int, date time.Time) (map[int]*stores.IndexStat, error) {
	indexStats, err := s.indexStatStore.List(ctx, &stores.IndexStatFilter{IndexIDs: indexIDs, Date: date}, 0, 0)
	if err != nil {
		return nil, err
	}

	var indexStatsMap = make(map[int]*stores.IndexStat)

	for i := range indexStats {
		indexStatsMap[indexStats[i].IndexID] = indexStats[i]
	}

	return indexStatsMap, nil
}

func (s *indexService) buildResp(model *stores.Index, indexStats map[int]*stores.IndexStat, securities []*Security) *Index {
	resp := &Index{
		ID:            model.ID,
		Name:          model.Name,
		Value:         model.Value,
		PreviousClose: 0,
		CreatedAt:     model.CreatedAt,
		UpdatedAt:     model.UpdatedAt,
		Constituents:  make([]*IndexConstituent, len(model.Constituents)),
	}

	var securitiesByID = make(map[int]*Security)

	for _, security := range securities {
		securitiesByID[security.ID] = security
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

	s.bindIndexStat(resp, indexStats)

	return resp
}

func (*indexService) bindIndexStat(resp *Index, indexStats map[int]*stores.IndexStat) {
	indexStat, ok := indexStats[resp.ID]
	if !ok {
		return
	}

	resp.IndexStat = &IndexStat{
		ID:      indexStat.ID,
		IndexID: indexStat.IndexID,
		Date:    indexStat.Date,
		Open:    indexStat.Open,
		Close:   indexStat.Close,
		High:    indexStat.High,
		Low:     indexStat.Low,
	}

	resp.PreviousClose = indexStat.Close
}
