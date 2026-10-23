package services

import (
	"time"

	"gofr.dev/pkg/gofr"
	"golang.org/x/sync/errgroup"

	"github.com/stratifyr/security-service/internal/stores"
)

type SecurityService interface {
	List(ctx *gofr.Context) ([]*Security, error)
	Upsert(ctx *gofr.Context, payload *SecurityUpsert) (*Security, error)
	GetHistoricalData(ctx *gofr.Context, id int) ([]*MarketData, error)
}

type Security struct {
	ID              int
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
	*MarketData
}

type MarketData struct {
	SecurityStat    *stores.SecurityStat
	SecurityMetrics []*SecurityMetric
}

type SecurityUpsert struct {
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
	securityStatStore     stores.SecurityStatStore
	store                 stores.SecurityStore
}

func NewSecurityService(marketDayService MarketDayService, securityMetricService SecurityMetricService,
	securityStatStore stores.SecurityStatStore, store stores.SecurityStore) *securityService {
	return &securityService{
		marketDayService:      marketDayService,
		securityMetricService: securityMetricService,
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

	var securityIDs = make([]int, len(securities))

	for i := range securities {
		securityIDs[i] = securities[i].ID
	}

	prevMarketDay, err := s.getPrevMarketDay(ctx)
	if err != nil {
		return nil, err
	}

	var (
		securityStats   map[int]*stores.SecurityStat
		securityMetrics map[int][]*SecurityMetric
	)

	var g errgroup.Group

	g.Go(func() error {
		var err error

		securityStats, err = s.getStatsMap(ctx, securityIDs, prevMarketDay)

		return err
	})

	g.Go(func() error {
		var err error

		securityMetrics, err = s.getSecurityMetricsMap(ctx, securityIDs, prevMarketDay)

		return err
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	resp := make([]*Security, len(securities))

	for i := range securities {
		resp[i] = s.buildResp(securities[i], securityStats, securityMetrics)
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

	model := &stores.Security{
		ISIN:            payload.ISIN,
		Symbol:          payload.Symbol,
		Industry:        industry,
		Name:            payload.Name,
		Image:           payload.Image,
		LTP:             payload.LTP,
		Volume:          payload.Volume,
		FreeFloatShares: payload.FreeFloatShares,
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}

	security, err := s.store.Create(ctx, model)
	if err != nil {
		return nil, err
	}

	prevMarketDay, err := s.getPrevMarketDay(ctx)
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

	return s.buildResp(security, securityStats, securityMetrics), nil
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

	var historicalData = make([]*MarketData, len(securityStats))

	for i, ss := range securityStats {
		securityMetrics, err := s.securityMetricService.List(ctx, []int{id}, ss.Date)
		if err != nil {
			return nil, err
		}

		historicalData[i] = &MarketData{
			SecurityStat:    securityStats[i],
			SecurityMetrics: securityMetrics,
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

	security, err = s.store.Update(ctx, id, security)
	if err != nil {
		return nil, err
	}

	prevMarketDay, err := s.getPrevMarketDay(ctx)
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

	return s.buildResp(security, securityStats, securityMetrics), nil
}

func (s *securityService) getPrevMarketDay(ctx *gofr.Context) (time.Time, error) {
	dates, _, err := s.marketDayService.List(ctx, &MarketDayFilter{LastNDaysFromReference: &struct {
		N         int
		Reference time.Time
	}{N: 2, Reference: time.Now()}})
	if err != nil {
		return time.Time{}, err
	}

	marketDay := dates[0]
	if dates[0].Format(time.DateOnly) == time.Now().Format(time.DateOnly) {
		marketDay = dates[1]
	}

	return marketDay, nil
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
	securityMetrics, err := s.securityMetricService.List(ctx, securityIDs, date)
	if err != nil {
		return nil, err
	}

	securityMetricsMap := make(map[int][]*SecurityMetric, len(securityIDs))

	for _, securityID := range securityIDs {
		securityMetricsMap[securityID] = make([]*SecurityMetric, 0)
	}

	for _, sm := range securityMetrics {
		securityMetricsMap[sm.SecurityID] = append(securityMetricsMap[sm.SecurityID], sm)
	}

	return securityMetricsMap, nil
}

func (s *securityService) buildResp(model *stores.Security, securityStats map[int]*stores.SecurityStat,
	securityMetrics map[int][]*SecurityMetric) *Security {
	resp := &Security{
		ID:              model.ID,
		ISIN:            model.ISIN,
		Symbol:          model.Symbol,
		Industry:        model.Industry.String(),
		Name:            model.Name,
		Image:           model.Image,
		LTP:             model.LTP,
		Volume:          model.Volume,
		FreeFloatShares: model.FreeFloatShares,
		CreatedAt:       model.CreatedAt,
		UpdatedAt:       model.CreatedAt,
		MarketData:      new(MarketData),
	}

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
