package services

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"gofr.dev/pkg/gofr"

	"github.com/stratifyr/security-service/internal/stores"
)

type SecurityService interface {
	Index(ctx *gofr.Context, f *SecurityFilter, page, perPage int) ([]*Security, int, error)
	Read(ctx *gofr.Context, id, userID int) (*Security, error)
	Create(ctx *gofr.Context, payload *SecurityCreate) (*Security, error)
	Patch(ctx *gofr.Context, id int, payload *SecurityUpdate) (*Security, error)
}

type SecurityFilter struct {
	UserID int
	IDs    []int
	ISIN   string
	Symbol string
	Date   time.Time
}

type Security struct {
	ID              int
	ISIN            string
	Symbol          string
	Industry        string
	Name            string
	Image           string
	LTP             float64
	PreviousClose   float64
	Tier            int
	CreatedAt       time.Time
	UpdatedAt       time.Time
	SecurityStat    *SecurityStat
	SecurityMetrics []*SecurityMetric
}

type SecurityCreate struct {
	UserID   int
	ISIN     string
	Symbol   string
	Industry string
	Name     string
	Image    string
	LTP      float64
	Tier     int
}

type SecurityUpdate struct {
	UserID   int
	Symbol   string
	Industry string
	Name     string
	Image    string
	LTP      float64
	Tier     *int
}

type securityService struct {
	marketDayService      MarketDayService
	securityMetricService SecurityMetricService
	metricsStore          stores.MetricStore
	securityStatStore     stores.SecurityStatStore
	store                 stores.SecurityStore
}

func NewSecurityService(marketDayService MarketDayService, securityMetricService SecurityMetricService, metricStore stores.MetricStore,
	securityStatStore stores.SecurityStatStore, store stores.SecurityStore) *securityService {
	return &securityService{
		marketDayService:      marketDayService,
		securityMetricService: securityMetricService,
		metricsStore:          metricStore,
		securityStatStore:     securityStatStore,
		store:                 store,
	}
}

func (s *securityService) Index(ctx *gofr.Context, f *SecurityFilter, page, perPage int) ([]*Security, int, error) {
	limit := perPage
	offset := limit * (page - 1)

	if f.Date.IsZero() {
		f.Date = time.Now()
	}

	filter := &stores.SecurityFilter{
		IDs:     f.IDs,
		Symbol:  f.Symbol,
		ISIN:    f.ISIN,
		MaxTier: nil,
	}

	if f.UserID != 0 {
		userTier, err := s.getUserTier(ctx, f.UserID)
		if err != nil {
			return nil, 0, err
		}

		filter.MaxTier = &userTier
	}

	securities, err := s.store.Index(ctx, filter, limit, offset)
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

	var securityIDs = make([]int, len(securities))

	for i := range securities {
		securityIDs[i] = securities[i].ID
	}

	prevMarketDay, err := s.getPrevMarketDay(ctx, f.Date)
	if err != nil {
		return nil, 0, err
	}

	securityStats, err := s.getStatsMap(ctx, securityIDs, prevMarketDay)
	if err != nil {
		return nil, 0, err
	}

	securityMetrics, err := s.securityMetricService.Get(ctx, f.UserID, securityIDs, prevMarketDay)
	if err != nil {
		return nil, 0, err
	}

	resp := make([]*Security, len(securities))

	for i := range securities {
		resp[i] = s.buildResp(securities[i], securityStats, securityMetrics)
	}

	return resp, count, nil
}

func (s *securityService) Read(ctx *gofr.Context, id, userID int) (*Security, error) {
	security, err := s.store.Retrieve(ctx, id)
	if err != nil {
		return nil, err
	}

	prevMarketDay, err := s.getPrevMarketDay(ctx, time.Now())
	if err != nil {
		return nil, err
	}

	securityStats, err := s.getStatsMap(ctx, []int{security.ID}, prevMarketDay)
	if err != nil {
		return nil, err
	}

	securityMetrics, err := s.securityMetricService.Get(ctx, userID, []int{security.ID}, prevMarketDay)
	if err != nil {
		return nil, err
	}

	return s.buildResp(security, securityStats, securityMetrics), nil
}

func (s *securityService) Create(ctx *gofr.Context, payload *SecurityCreate) (*Security, error) {
	if payload.UserID != 1 {
		return nil, &ErrResp{Code: 403, Message: "Forbidden"}
	}

	industry, err := stores.IndustryFromString(payload.Industry)
	if err != nil {
		return nil, err
	}

	model := &stores.Security{
		ISIN:      payload.ISIN,
		Symbol:    payload.Symbol,
		Industry:  industry,
		Name:      payload.Name,
		Image:     payload.Image,
		LTP:       payload.LTP,
		Tier:      payload.Tier,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	security, err := s.store.Create(ctx, model)
	if err != nil {
		return nil, err
	}

	prevMarketDay, err := s.getPrevMarketDay(ctx, time.Now())
	if err != nil {
		return nil, err
	}

	securityStats, err := s.getStatsMap(ctx, []int{security.ID}, prevMarketDay)
	if err != nil {
		return nil, err
	}

	securityMetrics, err := s.securityMetricService.Get(ctx, payload.UserID, []int{security.ID}, prevMarketDay)
	if err != nil {
		return nil, err
	}

	return s.buildResp(security, securityStats, securityMetrics), nil
}

func (s *securityService) Patch(ctx *gofr.Context, id int, payload *SecurityUpdate) (*Security, error) {
	if payload.UserID != 1 {
		return nil, &ErrResp{Code: 403, Message: "Forbidden"}
	}

	security, err := s.store.Retrieve(ctx, id)
	if err != nil {
		return nil, err
	}

	if payload.Symbol != "" {
		security.Symbol = payload.Symbol
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

	if payload.Tier != nil {
		security.Tier = *payload.Tier
	}

	security, err = s.store.Update(ctx, id, security)
	if err != nil {
		return nil, err
	}

	prevMarketDay, err := s.getPrevMarketDay(ctx, time.Now())
	if err != nil {
		return nil, err
	}

	securityStats, err := s.getStatsMap(ctx, []int{security.ID}, prevMarketDay)
	if err != nil {
		return nil, err
	}

	securityMetrics, err := s.securityMetricService.Get(ctx, payload.UserID, []int{security.ID}, prevMarketDay)
	if err != nil {
		return nil, err
	}

	return s.buildResp(security, securityStats, securityMetrics), nil
}

func (s *securityService) getUserTier(ctx *gofr.Context, userID int) (int, error) {
	httpService := ctx.GetHTTPService("account-service")

	resp, err := httpService.Get(ctx, fmt.Sprintf("users/%d", userID), nil)
	if err != nil {
		ctx.Logger.Errorf("failed GET /account-service/users/{id}, %v", map[string]interface{}{
			"err":    err.Error(),
			"userId": userID,
		})

		return 0, &ErrResp{Code: 503, Message: "something went wrong!"}
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)

		ctx.Logger.Errorf("non 200 resp GET /account-service/users/{id}, %v", map[string]interface{}{
			"code": resp.StatusCode,
			"resp": string(body),
		})

		return 0, &ErrResp{Code: 503, Message: "something went wrong!"}
	}

	var res struct {
		Data *struct {
			Plan *struct {
				Tier int `json:"tier"`
			} `json:"plan"`
		} `json:"data"`
	}

	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		ctx.Logger.Error("unexpected response from GET /account-service/users/{id}, %v", map[string]interface{}{
			"resp":         fmt.Sprintf("%s", resp.Body),
			"unmarshalErr": err,
		})

		return 0, &ErrResp{Code: 503, Message: "something went wrong!"}
	}

	return res.Data.Plan.Tier, nil
}

func (s *securityService) getPrevMarketDay(ctx *gofr.Context, referenceDate time.Time) (time.Time, error) {
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

func (s *securityService) getStatsMap(ctx *gofr.Context, securityIDs []int, date time.Time) (map[int]*stores.SecurityStat, error) {
	securityStats, err := s.securityStatStore.Index(ctx, &stores.SecurityStatFilter{SecurityIDs: securityIDs, Dates: []time.Time{date}}, 0, 0)
	if err != nil {
		return nil, err
	}

	var securityStatsMap = make(map[int]*stores.SecurityStat)

	for i := range securityStats {
		securityStatsMap[securityStats[i].SecurityID] = securityStats[i]
	}

	return securityStatsMap, nil
}

func (s *securityService) buildResp(model *stores.Security, securityStats map[int]*stores.SecurityStat, securityMetrics map[int][]*SecurityMetric) *Security {
	resp := &Security{
		ID:              model.ID,
		ISIN:            model.ISIN,
		Symbol:          model.Symbol,
		Industry:        model.Industry.String(),
		Name:            model.Name,
		Image:           model.Image,
		LTP:             model.LTP,
		Tier:            model.Tier,
		CreatedAt:       model.CreatedAt,
		UpdatedAt:       model.CreatedAt,
		SecurityStat:    nil,
		SecurityMetrics: nil,
	}

	s.bindSecurityStat(resp, securityStats)
	s.bindSecurityMetricsDetails(resp, securityMetrics)

	return resp
}

func (s *securityService) bindSecurityStat(resp *Security, securityStats map[int]*stores.SecurityStat) {
	securityStat, ok := securityStats[resp.ID]
	if !ok {
		return
	}

	resp.SecurityStat = &SecurityStat{
		ID:         securityStat.ID,
		SecurityID: securityStat.SecurityID,
		Date:       securityStat.Date,
		Open:       securityStat.Open,
		Close:      securityStat.Close,
		High:       securityStat.High,
		Low:        securityStat.Low,
		Volume:     securityStat.Volume,
	}

	resp.PreviousClose = securityStat.Close

	return
}

func (s *securityService) bindSecurityMetricsDetails(resp *Security, securityMetricsMap map[int][]*SecurityMetric) {
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

	return
}
