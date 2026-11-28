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
	ID            int
	Exchange      string
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
	Exchange    string
	Name        string
	Value       float64
	SecurityIDs []int
}

type indexService struct {
	marketDayService MarketDayService
	securityService  SecurityService
	exchangeService  ExchangeService
	indexStatStore   stores.IndexStatStore
	store            stores.IndexStore
}

func NewIndexService(marketDayService MarketDayService, securityService SecurityService,
	exchangeService ExchangeService, indexStatStore stores.IndexStatStore,
	store stores.IndexStore) *indexService {
	return &indexService{
		marketDayService: marketDayService,
		securityService:  securityService,
		exchangeService:  exchangeService,
		indexStatStore:   indexStatStore,
		store:            store,
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

	indexStats, err := s.getStatsMap(ctx, indices)
	if err != nil {
		return nil, err
	}

	exchanges, err := s.getExchangesMap(ctx)
	if err != nil {
		return nil, err
	}

	securities, err := s.securityService.List(ctx)
	if err != nil {
		return nil, err
	}

	resp := make([]*Index, len(indices))

	for i := range indices {
		resp[i] = s.buildResp(indices[i], indexStats, exchanges, securities)
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

	exchange, err := s.exchangeService.GetByName(ctx, payload.Exchange)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	model := &stores.Index{
		ExchangeID:   exchange.ID,
		Name:         payload.Name,
		Value:        payload.Value,
		CreatedAt:    now,
		UpdatedAt:    now,
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

	prevMarketDay, err := s.marketDayService.GetPreviousDay(ctx, index.ExchangeID)
	if err != nil {
		return nil, err
	}

	indexStats, err := s.getStatsForDate(ctx, []int{index.ID}, prevMarketDay)
	if err != nil {
		return nil, err
	}

	exchanges, err := s.getExchangesMap(ctx)
	if err != nil {
		return nil, err
	}

	securities, err := s.securityService.List(ctx)
	if err != nil {
		return nil, err
	}

	return s.buildResp(index, indexStats, exchanges, securities), nil
}

//nolint:gocyclo // patch logic requires multiple validations
func (s *indexService) patch(ctx *gofr.Context, id int, payload *IndexUpsert) (*Index, error) {
	index, err := s.store.Retrieve(ctx, id)
	if err != nil {
		return nil, err
	}

	if payload.Name != "" {
		index.Name = payload.Name
	}

	if payload.Value != 0 {
		index.Value = payload.Value
	}

	if payload.Exchange != "" {
		exchange, err := s.exchangeService.GetByName(ctx, payload.Exchange)
		if err != nil {
			return nil, err
		}

		index.ExchangeID = exchange.ID
	}

	if len(payload.SecurityIDs) > 0 {
		index.Constituents = make([]*stores.IndexConstituent, len(payload.SecurityIDs))

		for i := range payload.SecurityIDs {
			index.Constituents[i] = &stores.IndexConstituent{
				IndexID:    id,
				SecurityID: payload.SecurityIDs[i],
			}
		}
	}

	index.UpdatedAt = time.Now()

	index, err = s.store.Update(ctx, index.ID, index)
	if err != nil {
		return nil, err
	}

	prevMarketDay, err := s.marketDayService.GetPreviousDay(ctx, index.ExchangeID)
	if err != nil {
		return nil, err
	}

	indexStats, err := s.getStatsForDate(ctx, []int{id}, prevMarketDay)
	if err != nil {
		return nil, err
	}

	exchanges, err := s.getExchangesMap(ctx)
	if err != nil {
		return nil, err
	}

	securities, err := s.securityService.List(ctx)
	if err != nil {
		return nil, err
	}

	return s.buildResp(index, indexStats, exchanges, securities), nil
}

func (s *indexService) getStatsMap(ctx *gofr.Context, indices []*stores.Index) (map[int]*stores.IndexStat, error) {
	indexIDsByDate := make(map[time.Time][]int)
	prevDayByExchange := make(map[int]time.Time)

	for _, index := range indices {
		prevMarketDay, ok := prevDayByExchange[index.ExchangeID]
		if !ok {
			var err error

			prevMarketDay, err = s.marketDayService.GetPreviousDay(ctx, index.ExchangeID)
			if err != nil {
				return nil, err
			}

			prevDayByExchange[index.ExchangeID] = prevMarketDay
		}

		indexIDsByDate[prevMarketDay] = append(indexIDsByDate[prevMarketDay], index.ID)
	}

	indexStats := make(map[int]*stores.IndexStat)

	for date, indexIDs := range indexIDsByDate {
		stats, err := s.getStatsForDate(ctx, indexIDs, date)
		if err != nil {
			return nil, err
		}

		for indexID, stat := range stats {
			indexStats[indexID] = stat
		}
	}

	return indexStats, nil
}

func (s *indexService) getStatsForDate(ctx *gofr.Context, indexIDs []int, date time.Time) (map[int]*stores.IndexStat, error) {
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

func (s *indexService) getExchangesMap(ctx *gofr.Context) (map[int]*Exchange, error) {
	exchanges, err := s.exchangeService.List(ctx)
	if err != nil {
		return nil, err
	}

	var exchangeByID = make(map[int]*Exchange)

	for i := range exchanges {
		exchangeByID[exchanges[i].ID] = exchanges[i]
	}

	return exchangeByID, nil
}

func (s *indexService) buildResp(model *stores.Index, indexStats map[int]*stores.IndexStat,
	exchanges map[int]*Exchange, securities []*Security) *Index {
	resp := &Index{
		ID:            model.ID,
		Exchange:      exchanges[model.ExchangeID].Name,
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
