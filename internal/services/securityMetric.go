package services

import (
	"fmt"
	"math"
	"time"

	"gofr.dev/pkg/gofr"

	"github.com/stratifyr/security-service/internal/stores"
)

type SecurityMetricService interface {
	Index(ctx *gofr.Context, f *SecurityMetricFilter, page, perPage int) ([]*SecurityMetric, int, error)
	Read(ctx *gofr.Context, id int) (*SecurityMetric, error)
	Create(ctx *gofr.Context, payload *SecurityMetricCreate) (*SecurityMetric, error)
	Patch(ctx *gofr.Context, id int, payload *SecurityMetricUpdate) (*SecurityMetric, error)
}

type SecurityMetricFilter struct {
	Date       time.Time
	SecurityID int
}

type SecurityMetric struct {
	ID         int
	SecurityID int
	MetricID   int
	Date       time.Time
	Value      float64
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type SecurityMetricCreate struct {
	UserID         int
	SecurityID     int
	MetricID       int
	Date           time.Time
	Value          float64
	CalculateValue bool
}

type SecurityMetricUpdate struct {
	UserID int
	Value  float64
}

type securityMetricService struct {
	metricStore       stores.MetricStore
	securityStatStore stores.SecurityStatStore
	store             stores.SecurityMetricStore
}

func NewSecurityMetricService(metricStore stores.MetricStore, securityStatStore stores.SecurityStatStore, store stores.SecurityMetricStore) *securityMetricService {
	return &securityMetricService{
		metricStore:       metricStore,
		securityStatStore: securityStatStore,
		store:             store,
	}
}

func (s *securityMetricService) Index(ctx *gofr.Context, f *SecurityMetricFilter, page, perPage int) ([]*SecurityMetric, int, error) {
	limit := perPage
	offset := limit * (page - 1)

	filter := &stores.SecurityMetricFilter{
		SecurityID: f.SecurityID,
		Date:       f.Date,
	}

	securityMetrics, err := s.store.Index(ctx, filter, limit, offset)
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

	var resp = make([]*SecurityMetric, len(securityMetrics))

	for i := range securityMetrics {
		resp[i] = s.buildResp(securityMetrics[i])
	}

	return resp, count, nil
}

func (s *securityMetricService) Read(ctx *gofr.Context, id int) (*SecurityMetric, error) {
	securityMetric, err := s.store.Retrieve(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.buildResp(securityMetric), nil
}

func (s *securityMetricService) Create(ctx *gofr.Context, payload *SecurityMetricCreate) (*SecurityMetric, error) {
	if payload.UserID != 1 {
		return nil, &ErrResp{Code: 403, Message: "Forbidden"}
	}

	model := &stores.SecurityMetric{
		SecurityID: payload.SecurityID,
		MetricID:   payload.MetricID,
		Date:       payload.Date,
		Value:      payload.Value,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}

	if payload.CalculateValue {
		var err error

		model.Value, err = s.computeMetricValue(ctx, payload.SecurityID, payload.MetricID, payload.Date)
		if err != nil {
			return nil, err
		}
	}

	securityMetric, err := s.store.Create(ctx, model)
	if err != nil {
		return nil, err
	}

	return s.buildResp(securityMetric), nil
}

func (s *securityMetricService) Patch(ctx *gofr.Context, id int, payload *SecurityMetricUpdate) (*SecurityMetric, error) {
	if payload.UserID != 1 {
		return nil, &ErrResp{Code: 403, Message: "Forbidden"}
	}

	oldSecurityMetric, err := s.store.Retrieve(ctx, id)
	if err != nil {
		return nil, err
	}

	newSecurityMetric := *oldSecurityMetric

	if payload.Value != 0 {
		newSecurityMetric.Value = payload.Value
	}

	securityMetric, err := s.store.Update(ctx, id, &newSecurityMetric)
	if err != nil {
		return nil, err
	}

	return s.buildResp(securityMetric), nil
}

func (s *securityMetricService) buildResp(model *stores.SecurityMetric) *SecurityMetric {
	resp := &SecurityMetric{
		ID:         model.ID,
		SecurityID: model.SecurityID,
		MetricID:   model.MetricID,
		Date:       model.Date,
		Value:      model.Value,
		CreatedAt:  model.CreatedAt,
		UpdatedAt:  model.UpdatedAt,
	}

	return resp
}

func (s *securityMetricService) computeMetricValue(ctx *gofr.Context, securityID, metricID int, date time.Time) (float64, error) {
	metric, err := s.metricStore.Retrieve(ctx, metricID)
	if err != nil {
		return 0, err
	}

	securityStats, err := s.securityStatStore.Index(ctx, &stores.SecurityStatFilter{SecurityID: securityID, CutoffDate: date}, metric.Period, 0)
	if err != nil {
		return 0, err
	}

	if len(securityStats) != metric.Period {
		return 0, &ErrResp{Code: 400, Message: fmt.Sprintf("Cannot compute %s_%d, not enough data", metric.Type.String(), metric.Period)}
	}

	switch metric.Type {
	case stores.SMA:
		return s.computeSMA(metric.Period, securityStats), nil
	case stores.EMA:
		return s.computeEMA(metric.Period, securityStats), nil
	case stores.RSI:
		return s.computeRSI(metric.Period, securityStats), nil
	case stores.ROC:
		return s.computeROC(metric.Period, securityStats), nil
	case stores.ATR:
		return s.computeATR(metric.Period, securityStats), nil
	case stores.VMA:
		return s.computeVMA(metric.Period, securityStats), nil
	default:
		return 0, nil
	}
}

func (s *securityMetricService) computeSMA(n int, lastNStats []*stores.SecurityStat) float64 {
	var sumPrice float64

	for _, stat := range lastNStats {
		sumPrice += stat.Close
	}

	return sumPrice / float64(n)
}

func (s *securityMetricService) computeEMA(n int, lastNStats []*stores.SecurityStat) float64 {
	var sumPrice float64

	for i := len(lastNStats) - 1; i >= len(lastNStats)-n; i-- {
		sumPrice += lastNStats[i].Close
	}

	ema := sumPrice / float64(n)

	k := 2.0 / float64(n+1)

	for i := len(lastNStats) - n - 1; i >= 0; i-- {
		ema = lastNStats[i].Close*k + ema*(1-k)
	}

	return ema
}

func (s *securityMetricService) computeRSI(n int, lastNStats []*stores.SecurityStat) float64 {
	var (
		totalProfit float64
		totalLoss   float64
	)

	for i := 1; i < n; i++ {
		deltaP := lastNStats[i].Close - lastNStats[i-1].Close

		if deltaP > 0 {
			totalProfit += deltaP

			continue
		}

		totalLoss += -deltaP
	}

	if totalLoss == 0 {
		return 100
	}

	rs := totalProfit / totalLoss

	return 100 - (100 / (1 + rs))
}

func (s *securityMetricService) computeROC(n int, lastNStats []*stores.SecurityStat) float64 {
	currentPrice := lastNStats[0].Close
	nDaysPriorPrice := lastNStats[n-1].Close

	return (currentPrice - nDaysPriorPrice) / nDaysPriorPrice
}

func (s *securityMetricService) computeATR(n int, lastNStats []*stores.SecurityStat) float64 {
	var totalTR float64

	for i := 1; i < n; i++ {
		high := lastNStats[i].High
		low := lastNStats[i].Low
		prevClose := lastNStats[i-1].Close

		tr := math.Max(high-low, math.Max(math.Abs(high-prevClose), math.Abs(low-prevClose)))
		totalTR += tr
	}

	return totalTR / float64(n)
}

func (s *securityMetricService) computeVMA(n int, lastNStats []*stores.SecurityStat) float64 {
	var sumVolume float64

	for _, stat := range lastNStats {
		sumVolume += float64(stat.Volume)
	}

	return sumVolume / float64(n)
}
