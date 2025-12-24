package services

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"gofr.dev/pkg/gofr"

	"github.com/stratifyr/security-service/internal/stores"
)

type SecurityMetricService interface {
	Get(ctx *gofr.Context, securityID int, date time.Time) ([]*SecurityMetric, error)
}

type SecurityMetric struct {
	Metric *Metric
	Value  float64
}

type securityMetricService struct {
	mu                sync.Mutex
	marketDayService  MarketDayService
	metricStore       stores.MetricStore
	securityStatStore stores.SecurityStatStore
}

func NewSecurityMetricService(marketDayService MarketDayService, metricStore stores.MetricStore, securityStatStore stores.SecurityStatStore) *securityMetricService {
	return &securityMetricService{
		mu:                sync.Mutex{},
		marketDayService:  marketDayService,
		metricStore:       metricStore,
		securityStatStore: securityStatStore,
	}
}

func (s *securityMetricService) Get(ctx *gofr.Context, securityID int, date time.Time) ([]*SecurityMetric, error) {
	metrics, err := s.metricStore.Index(ctx, &stores.MetricFilter{}, 0, 0)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	metricValues, err := s.getMetricValues(ctx, securityID, date, metrics)
	if err != nil {
		return nil, err
	}
	s.mu.Unlock()

	var resp = make([]*SecurityMetric, 0)

	for i := range metrics {
		value, ok := metricValues[strconv.Itoa(metrics[i].ID)]
		if !ok {
			continue
		}

		valueFloat, _ := strconv.ParseFloat(value, 64)

		resp = append(resp, &SecurityMetric{
			Metric: &Metric{
				ID:        metrics[i].ID,
				Name:      metrics[i].Name,
				Type:      metrics[i].Type.String(),
				Period:    metrics[i].Period,
				Indicator: metrics[i].Indicator.String(),
				Tier:      metrics[i].Tier,
				CreatedAt: metrics[i].CreatedAt,
				UpdatedAt: metrics[i].UpdatedAt,
			},
			Value: valueFloat,
		})
	}

	return resp, nil
}

func (s *securityMetricService) getMetricValues(ctx *gofr.Context, securityID int, date time.Time, metrics []*stores.Metric) (map[string]string, error) {
	isCacheable := time.Since(date) <= 72*time.Hour

	if isCacheable {
		values, err := s.getMetricValuesFromCache(ctx, securityID, date)
		if err == nil {
			return values, nil
		}

		ctx.Logger.Warnf("failed to get metric values from cache: %v", map[string]any{
			"error":      err,
			"securityId": securityID,
			"date":       date,
		})
	}

	maxPeriod := 0
	for i := range metrics {
		if metrics[i].Period > maxPeriod {
			maxPeriod = metrics[i].Period
		}
	}

	marketDays, _, err := s.marketDayService.Index(ctx, &MarketDayFilter{LastNDaysFromReference: &struct {
		N         int
		Reference time.Time
	}{N: maxPeriod, Reference: date}})
	if err != nil {
		return nil, err
	}

	securityStats, err := s.securityStatStore.Index(ctx, &stores.SecurityStatFilter{SecurityIDs: []int{securityID}, Dates: marketDays}, 0, 0)
	if err != nil {
		return nil, err
	}

	sort.Slice(securityStats, func(i, j int) bool {
		return securityStats[i].Date.After(securityStats[j].Date)
	})

	var values = make(map[string]string)

	for i := range metrics {
		n := metrics[i].Period

		if len(securityStats) < n {
			ctx.Logger.Warnf("cannot compute securityId:%d_%s_%d, not enough data", securityID, metrics[i].Type.String(), metrics[i].Period)
			continue
		}

		value := s.computeMetricValue(metrics[i], securityStats[:n])
		values[strconv.Itoa(metrics[i].ID)] = fmt.Sprintf("%0.2f", value)
	}

	if isCacheable {
		if err = s.setMetricValuesToCache(ctx, values, securityID, date); err != nil {
			ctx.Logger.Warnf("failed to set metric values in cache: %v", map[string]any{
				"error":      err,
				"securityId": securityID,
				"date":       date,
			})
		}
	}

	return values, nil
}

func (s *securityMetricService) computeMetricValue(metric *stores.Metric, securityStats []*stores.SecurityStat) float64 {
	switch metric.Type {
	case stores.SMA:
		return s.computeSMA(securityStats)
	case stores.EMA:
		k := 2.0 / float64(len(securityStats)+1)
		smaSeed := s.computeSMA(securityStats)
		return s.computeEMA(k, smaSeed, securityStats)
	case stores.RSI:
		return s.computeRSI(securityStats)
	case stores.ROC:
		return s.computeROC(securityStats)
	case stores.ATR:
		return s.computeATR(securityStats)
	case stores.VMA:
		return s.computeVMA(securityStats)
	default:
		return 0
	}
}

func (s *securityMetricService) getMetricValuesFromCache(ctx *gofr.Context, securityID int, date time.Time) (map[string]string, error) {
	key := fmt.Sprintf("security_metrics:security_id:%d:date:%s", securityID, date.Format(time.DateOnly))

	res, err := ctx.Redis.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	if len(res) == 0 {
		return nil, redis.Nil
	}

	return res, nil
}

func (s *securityMetricService) setMetricValuesToCache(ctx *gofr.Context, values map[string]string, securityID int, date time.Time) error {
	key := fmt.Sprintf("security_metrics:security_id:%d:date:%s", securityID, date.Format(time.DateOnly))

	if err := ctx.Redis.HSet(ctx, key, values).Err(); err != nil {
		return err
	}

	if err := ctx.Redis.Expire(ctx, key, time.Hour).Err(); err != nil {
		return err
	}

	return nil
}

func (s *securityMetricService) buildResp(metric *stores.Metric, value float64) *SecurityMetric {
	resp := &SecurityMetric{
		Metric: &Metric{
			ID:        metric.ID,
			Name:      metric.Name,
			Type:      metric.Type.String(),
			Period:    metric.Period,
			Indicator: metric.Indicator.String(),
			Tier:      metric.Tier,
			CreatedAt: metric.CreatedAt,
			UpdatedAt: metric.UpdatedAt,
		},
		Value: value,
	}

	return resp
}

func (s *securityMetricService) computeSMA(lastNStats []*stores.SecurityStat) float64 {
	var (
		sumPrice float64
		n        = len(lastNStats)
	)

	for _, stat := range lastNStats {
		sumPrice += stat.Close
	}

	return sumPrice / float64(n)
}

func (s *securityMetricService) computeEMA(k, seeder float64, lastNStats []*stores.SecurityStat) float64 {
	n := len(lastNStats)
	if n == 0 {
		return 0
	}

	ema := seeder

	for i := n - 1; i >= 0; i-- {
		ema = lastNStats[i].Close*k + ema*(1-k)
	}

	return ema
}

func (s *securityMetricService) computeRSI(lastNStats []*stores.SecurityStat) float64 {
	var (
		totalProfit float64
		totalLoss   float64
		n           = len(lastNStats)
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

func (s *securityMetricService) computeROC(lastNStats []*stores.SecurityStat) float64 {
	n := len(lastNStats)
	currentPrice := lastNStats[0].Close
	nDaysPriorPrice := lastNStats[n-1].Close

	return ((currentPrice - nDaysPriorPrice) / nDaysPriorPrice) * 100
}

func (s *securityMetricService) computeATR(lastNStats []*stores.SecurityStat) float64 {
	var (
		totalTR float64
		n       = len(lastNStats)
	)

	for i := 1; i < n; i++ {
		high := lastNStats[i].High
		low := lastNStats[i].Low
		prevClose := lastNStats[i-1].Close

		tr := math.Max(high-low, math.Max(math.Abs(high-prevClose), math.Abs(low-prevClose)))
		totalTR += tr
	}

	return totalTR / float64(n)
}

func (s *securityMetricService) computeVMA(lastNStats []*stores.SecurityStat) float64 {
	var (
		sumVolume float64
		n         = len(lastNStats)
	)

	for _, stat := range lastNStats {
		sumVolume += float64(stat.Volume)
	}

	return sumVolume / float64(n)
}
