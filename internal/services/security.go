package services

import (
	"time"

	"gofr.dev/pkg/gofr"

	"github.com/stratifyr/security-service/internal/stores"
)

type SecurityService interface {
	List(ctx *gofr.Context) ([]*Security, error)
	Upsert(ctx *gofr.Context, payload *SecurityUpsert) (*Security, error)
	GetHistoricalData(ctx *gofr.Context, id int) ([]*MarketData, error)
}

type Security struct {
	ID              int
	ExchangeID      int
	ISIN            string
	Symbol          string
	Industry        string
	Name            string
	Image           string
	LTP             float64
	Volume          int
	FreeFloatShares int
	PreviousClose   float64
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Exchange        *Exchange
	*MarketData
}

type MarketData struct {
	SecurityStat    *stores.SecurityStat
	SecurityMetrics []*SecurityMetric
}

type SecurityUpsert struct {
	Exchange        string
	ISIN            string
	Symbol          string
	Industry        string
	Name            string
	Image           string
	LTP             float64
	Volume          int
	FreeFloatShares int
}

type securityService struct {
	marketDayService      MarketDayService
	securityMetricService SecurityMetricService
	exchangeService       ExchangeService
	securityStatStore     stores.SecurityStatStore
	store                 stores.SecurityStore
}

func NewSecurityService(marketDayService MarketDayService, securityMetricService SecurityMetricService,
	exchangeService ExchangeService, securityStatStore stores.SecurityStatStore, store stores.SecurityStore) *securityService {
	return &securityService{
		marketDayService:      marketDayService,
		securityMetricService: securityMetricService,
		exchangeService:       exchangeService,
		securityStatStore:     securityStatStore,
		store:                 store,
	}
}

func (s *securityService) List(ctx *gofr.Context) ([]*Security, error) {
	securities, err := s.store.List(ctx, &stores.SecurityFilter{}, 0, 0)
	if err != nil {
		return nil, err
	}

	if len(securities) == 0 {
		return nil, nil
	}

	exchanges, err := s.getExchangesMap(ctx)
	if err != nil {
		return nil, err
	}

	securityStats, securityMetrics, err := s.getMarketDataMaps(ctx, securities)
	if err != nil {
		return nil, err
	}

	resp := make([]*Security, len(securities))

	for i := range securities {
		resp[i] = s.buildResp(securities[i], exchanges, securityStats, securityMetrics)
	}

	return resp, nil
}

func (s *securityService) Upsert(ctx *gofr.Context, payload *SecurityUpsert) (*Security, error) {
	securities, err := s.store.List(ctx, &stores.SecurityFilter{Symbol: payload.Symbol}, 1, 0)
	if err != nil {
		return nil, err
	}

	if len(securities) > 0 {
		return s.patch(ctx, securities[0].ID, payload)
	}

	industry, err := stores.IndustryFromString(payload.Industry)
	if err != nil {
		return nil, err
	}

	exchange, err := s.exchangeService.GetByName(ctx, payload.Exchange)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()

	model := &stores.Security{
		ExchangeID:      exchange.ID,
		ISIN:            payload.ISIN,
		Symbol:          payload.Symbol,
		Industry:        industry,
		Name:            payload.Name,
		Image:           payload.Image,
		LTP:             payload.LTP,
		Volume:          payload.Volume,
		FreeFloatShares: payload.FreeFloatShares,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	security, err := s.store.Create(ctx, model)
	if err != nil {
		return nil, err
	}

	exchanges, err := s.getExchangesMap(ctx)
	if err != nil {
		return nil, err
	}

	prevMarketDay, err := s.marketDayService.GetPreviousDay(ctx, security.ExchangeID)
	if err != nil {
		return nil, err
	}

	securityStats, err := s.getStatsMap(ctx, []int{security.ID}, prevMarketDay)
	if err != nil {
		return nil, err
	}

	securityMetrics, err := s.getSecurityMetricsMap(ctx, []int{security.ID}, prevMarketDay)
	if err != nil {
		return nil, err
	}

	return s.buildResp(security, exchanges, securityStats, securityMetrics), nil
}

func (s *securityService) GetHistoricalData(ctx *gofr.Context, id int) ([]*MarketData, error) {
	today := time.Now()
	threeYearsEarlier := today.AddDate(-3, 0, 0)
	startDate, endDate := threeYearsEarlier, today.AddDate(0, 0, -1)

	if _, err := s.store.Retrieve(ctx, id); err != nil {
		return nil, err
	}

	dateRange := &struct {
		Start time.Time
		End   time.Time
	}{Start: startDate, End: endDate}

	securityStats, err := s.securityStatStore.List(ctx, &stores.SecurityStatFilter{SecurityIDs: []int{id}, DateBetween: dateRange}, 0, 0)
	if err != nil {
		return nil, err
	}

	if len(securityStats) == 0 {
		return nil, nil
	}

	dates := make([]time.Time, len(securityStats))

	for i := range securityStats {
		dates[i] = securityStats[i].Date
	}

	securityMetricsByDate, err := s.securityMetricService.GetBySecurityID(ctx, id, dates)
	if err != nil {
		return nil, err
	}

	var historicalData = make([]*MarketData, len(securityStats))

	for i := range securityStats {
		historicalData[i] = &MarketData{
			SecurityStat:    securityStats[i],
			SecurityMetrics: securityMetricsByDate[securityStats[i].Date],
		}
	}

	return historicalData, nil
}

//nolint:gocyclo // patch logic requires multiple validations
func (s *securityService) patch(ctx *gofr.Context, id int, payload *SecurityUpsert) (*Security, error) {
	security, err := s.store.Retrieve(ctx, id)
	if err != nil {
		return nil, err
	}

	if payload.Exchange != "" {
		exchange, err := s.exchangeService.GetByName(ctx, payload.Exchange)
		if err != nil {
			return nil, err
		}

		security.ExchangeID = exchange.ID
	}

	if payload.ISIN != "" {
		security.ISIN = payload.ISIN
	}

	if payload.Industry != "" {
		security.Industry, err = stores.IndustryFromString(payload.Industry)
		if err != nil {
			return nil, err
		}
	}

	if payload.Name != "" {
		security.Name = payload.Name
	}

	if payload.Image != "" {
		security.Image = payload.Image
	}

	if payload.LTP != 0 {
		security.LTP = payload.LTP
	}

	if payload.Volume != 0 {
		security.Volume = payload.Volume
	}

	if payload.FreeFloatShares != 0 {
		security.FreeFloatShares = payload.FreeFloatShares
	}

	security.UpdatedAt = time.Now()

	security, err = s.store.Update(ctx, id, security)
	if err != nil {
		return nil, err
	}

	exchanges, err := s.getExchangesMap(ctx)
	if err != nil {
		return nil, err
	}

	prevMarketDay, err := s.marketDayService.GetPreviousDay(ctx, security.ExchangeID)
	if err != nil {
		return nil, err
	}

	securityStats, err := s.getStatsMap(ctx, []int{security.ID}, prevMarketDay)
	if err != nil {
		return nil, err
	}

	securityMetrics, err := s.getSecurityMetricsMap(ctx, []int{security.ID}, prevMarketDay)
	if err != nil {
		return nil, err
	}

	return s.buildResp(security, exchanges, securityStats, securityMetrics), nil
}

func (s *securityService) getExchangesMap(ctx *gofr.Context) (map[int]*Exchange, error) {
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

func (s *securityService) getMarketDataMaps(ctx *gofr.Context, securities []*stores.Security) (
	securityStats map[int]*stores.SecurityStat, securityMetrics map[int][]*SecurityMetric, err error) {
	securityIDsByExchange := make(map[int][]int)

	for _, security := range securities {
		securityIDsByExchange[security.ExchangeID] = append(securityIDsByExchange[security.ExchangeID], security.ID)
	}

	securityIDsByDate := make(map[time.Time][]int)

	for exchangeID, securityIDs := range securityIDsByExchange {
		prevMarketDay, err := s.marketDayService.GetPreviousDay(ctx, exchangeID)
		if err != nil {
			return nil, nil, err
		}

		securityIDsByDate[prevMarketDay] = append(securityIDsByDate[prevMarketDay], securityIDs...)
	}

	securityStats = make(map[int]*stores.SecurityStat)
	securityMetrics = make(map[int][]*SecurityMetric)

	for date, securityIDs := range securityIDsByDate {
		stats, err := s.getStatsMap(ctx, securityIDs, date)
		if err != nil {
			return nil, nil, err
		}

		for securityID, stat := range stats {
			securityStats[securityID] = stat
		}

		metrics, err := s.getSecurityMetricsMap(ctx, securityIDs, date)
		if err != nil {
			return nil, nil, err
		}

		for securityID, securityMetric := range metrics {
			securityMetrics[securityID] = securityMetric
		}
	}

	return securityStats, securityMetrics, nil
}

func (s *securityService) getStatsMap(ctx *gofr.Context, securityIDs []int, date time.Time) (map[int]*stores.SecurityStat, error) {
	securityStats, err := s.securityStatStore.List(ctx, &stores.SecurityStatFilter{SecurityIDs: securityIDs, Date: date}, 0, 0)
	if err != nil {
		return nil, err
	}

	var securityStatsMap = make(map[int]*stores.SecurityStat)

	for i := range securityStats {
		securityStatsMap[securityStats[i].SecurityID] = securityStats[i]
	}

	return securityStatsMap, nil
}

func (s *securityService) getSecurityMetricsMap(ctx *gofr.Context, securityIDs []int, date time.Time) (map[int][]*SecurityMetric, error) {
	securityMetrics, err := s.securityMetricService.List(ctx, date)
	if err != nil {
		return nil, err
	}

	securityMetricsMap := make(map[int][]*SecurityMetric, len(securityIDs))

	for _, securityID := range securityIDs {
		securityMetricsMap[securityID] = make([]*SecurityMetric, 0)
	}

	for _, sm := range securityMetrics {
		if _, ok := securityMetricsMap[sm.SecurityID]; !ok {
			continue
		}

		securityMetricsMap[sm.SecurityID] = append(securityMetricsMap[sm.SecurityID], sm)
	}

	return securityMetricsMap, nil
}

func (s *securityService) buildResp(model *stores.Security, exchanges map[int]*Exchange,
	securityStats map[int]*stores.SecurityStat, securityMetrics map[int][]*SecurityMetric) *Security {
	resp := &Security{
		ID:              model.ID,
		ExchangeID:      model.ExchangeID,
		ISIN:            model.ISIN,
		Symbol:          model.Symbol,
		Industry:        model.Industry.String(),
		Name:            model.Name,
		Image:           model.Image,
		LTP:             model.LTP,
		Volume:          model.Volume,
		FreeFloatShares: model.FreeFloatShares,
		CreatedAt:       model.CreatedAt,
		UpdatedAt:       model.UpdatedAt,
		MarketData:      new(MarketData),
	}

	resp.Exchange = exchanges[model.ExchangeID]

	s.bindSecurityStat(resp, securityStats)
	s.bindSecurityMetricsDetails(resp, securityMetrics)

	return resp
}

func (*securityService) bindSecurityStat(resp *Security, securityStats map[int]*stores.SecurityStat) {
	securityStat, ok := securityStats[resp.ID]
	if !ok {
		return
	}

	resp.SecurityStat = securityStat
	resp.PreviousClose = securityStat.Close
}

func (*securityService) bindSecurityMetricsDetails(resp *Security, securityMetricsMap map[int][]*SecurityMetric) {
	if resp.SecurityStat == nil {
		return
	}

	securityMetrics, ok := securityMetricsMap[resp.ID]
	if !ok {
		return
	}

	resp.SecurityMetrics = make([]*SecurityMetric, len(securityMetrics))

	for i := range securityMetrics {
		resp.SecurityMetrics[i] = &SecurityMetric{
			Metric: securityMetrics[i].Metric,
			Value:  securityMetrics[i].Value,
			ZValue: securityMetrics[i].ZValue,
		}
	}
}
