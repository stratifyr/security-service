package services

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/vmihailenco/msgpack/v5"
	"gofr.dev/pkg/gofr"

	"github.com/stratifyr/security-service/internal/stores"
)

type SecurityMetricService interface {
	Get(ctx *gofr.Context, securityIDs []int, date time.Time) (map[int][]*SecurityMetric, error)
}

type SecurityMetric struct {
	Metric *Metric
	Value  float64
	ZValue float64
}

type securityMetricService struct {
	marketDayService  MarketDayService
	metricService     MetricService
	securityStatStore stores.SecurityStatStore
}

func NewSecurityMetricService(marketDayService MarketDayService, metricService MetricService, securityStatStore stores.SecurityStatStore) *securityMetricService {
	return &securityMetricService{
		marketDayService:  marketDayService,
		metricService:     metricService,
		securityStatStore: securityStatStore,
	}
}

func (s *securityMetricService) Get(ctx *gofr.Context, securityIDs []int, date time.Time) (map[int][]*SecurityMetric, error) {
	metrics, err := s.metricService.Index(ctx)
	if err != nil {
		return nil, err
	}

	cachedSecurityMetrics, err := s.getSecurityMetricsFromCache(ctx, securityIDs, date, metrics)
	if err != nil {
		ctx.Logger.Warnf("failed to get security metrics from cache: %v", map[string]any{
			"error":       err,
			"securityIds": securityIDs,
			"date":        date,
		})
	}

	if len(cachedSecurityMetrics) == len(securityIDs) {
		return cachedSecurityMetrics, nil
	}

	var cacheMisses []int

	for _, securityID := range securityIDs {
		if _, ok := cachedSecurityMetrics[securityID]; !ok {
			cacheMisses = append(cacheMisses, securityID)
		}
	}

	securityMetrics, err := s.computeSecurityMetrics(ctx, cacheMisses, date, metrics)
	if err != nil {
		return nil, err
	}

	if err = s.setSecurityMetricsToCache(ctx, securityMetrics, date); err != nil {
		ctx.Logger.Warnf("failed to set security metrics in cache: %v", map[string]any{
			"error":       err,
			"securityIds": cacheMisses,
			"date":        date,
		})
	}

	mergeMaps(securityMetrics, cachedSecurityMetrics)

	return securityMetrics, nil
}

type metricValues struct {
	ID     int
	Value  float64
	ZValue float64
}

func (s *securityMetricService) getSecurityMetricsFromCache(ctx *gofr.Context, securityIDs []int, date time.Time, metrics []*Metric) (map[int][]*SecurityMetric, error) {
	metricsMap := make(map[int]*Metric)
	for _, m := range metrics {
		metricsMap[m.ID] = m
	}

	var keys = make([]string, len(securityIDs))

	for i, securityID := range securityIDs {
		keys[i] = fmt.Sprintf("security_metrics:security_id:%d:date:%s", securityID, date.Format(time.DateOnly))
	}

	vals, err := ctx.Redis.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}

	securityMetrics := make(map[int][]*SecurityMetric)

	for i, val := range vals {
		if val == nil {
			ctx.Logger.Warnf("cache miss: %s", keys[i])
			continue
		}

		var metricVals []metricValues
		if err = msgpack.Unmarshal([]byte(val.(string)), &metricVals); err != nil {
			return nil, err
		}

		sm := make([]*SecurityMetric, 0)

		for j := range metricVals {
			mv := &metricVals[j]

			metric, ok := metricsMap[mv.ID]
			if !ok {
				continue
			}

			sm = append(sm, &SecurityMetric{
				Metric: metric,
				Value:  mv.Value,
				ZValue: mv.ZValue,
			})
		}

		securityMetrics[securityIDs[i]] = sm
	}

	return securityMetrics, nil
}

func (s *securityMetricService) computeSecurityMetrics(ctx *gofr.Context, securityIDs []int, date time.Time, metrics []*Metric) (map[int][]*SecurityMetric, error) {
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

	startDate, endDate := marketDays[len(marketDays)-1], marketDays[0]
	if startDate.After(endDate) {
		startDate, endDate = endDate, startDate
	}

	securityStats, err := s.securityStatStore.Index(ctx, &stores.SecurityStatFilter{SecurityIDs: securityIDs,
		DateBetween: &struct {
			Start time.Time
			End   time.Time
		}{Start: startDate, End: endDate}}, 0, 0)
	if err != nil {
		return nil, err
	}

	securityStatsMap := make(map[int][]*stores.SecurityStat, len(securityIDs))

	for _, stat := range securityStats {
		securityStatsMap[stat.SecurityID] = append(securityStatsMap[stat.SecurityID], stat)
	}

	var securityMetrics = make(map[int][]*SecurityMetric)

	for _, securityID := range securityIDs {
		stats := securityStatsMap[securityID]

		sort.Slice(stats, func(i, j int) bool {
			return stats[i].Date.After(stats[j].Date)
		})

		var sm = make([]*SecurityMetric, 0)

		for i := range metrics {
			n := metrics[i].Period

			if len(stats) < n {
				ctx.Logger.Warnf("cannot compute securityId:%d_%s_%d, not enough data", securityID, metrics[i].Type.String(), metrics[i].Period)
				continue
			}

			value, normalizedValue := s.computeMetricValue(metrics[i], stats[:n])

			sm = append(sm, &SecurityMetric{
				Metric: &Metric{
					ID:        metrics[i].ID,
					Name:      metrics[i].Name,
					Type:      metrics[i].Type,
					Period:    metrics[i].Period,
					Indicator: metrics[i].Indicator,
					Tier:      metrics[i].Tier,
					CreatedAt: metrics[i].CreatedAt,
					UpdatedAt: metrics[i].UpdatedAt,
				},
				Value:  value,
				ZValue: normalizedValue,
			})
		}

		securityMetrics[securityID] = sm
	}

	return securityMetrics, nil
}

func (s *securityMetricService) setSecurityMetricsToCache(ctx *gofr.Context, securityMetrics map[int][]*SecurityMetric, date time.Time) error {
	pipe := ctx.Redis.Pipeline()

	for securityID, metrics := range securityMetrics {
		key := fmt.Sprintf("security_metrics:security_id:%d:date:%s", securityID, date.Format(time.DateOnly))

		metricValue := make([]*metricValues, len(metrics))

		for i, metric := range metrics {
			metricValue[i] = &metricValues{
				ID:     metric.Metric.ID,
				Value:  metric.Value,
				ZValue: metric.ZValue,
			}
		}

		bytes, err := msgpack.Marshal(metricValue)
		if err != nil {
			return err
		}

		pipe.SetEx(ctx, key, bytes, 5*time.Hour)
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return err
	}

	return nil
}

func (s *securityMetricService) computeMetricValue(metric *Metric, securityStats []*stores.SecurityStat) (float64, float64) {
	dayStat := securityStats[0]

	switch metric.Type {
	case stores.SMA:
		value := s.computeSMA(securityStats)
		return value, (dayStat.Close - value) / value
	case stores.EMA:
		k := 2.0 / float64(len(securityStats)+1)
		smaSeed := s.computeSMA(securityStats)
		value := s.computeEMA(k, smaSeed, securityStats)
		return value, (dayStat.Close - value) / value
	case stores.RSI:
		value := s.computeRSI(securityStats)
		return value, value / 100
	case stores.ROC:
		value := s.computeROC(securityStats)
		return value, value / 100
	case stores.ATR:
		value := s.computeATR(securityStats)
		return value, -value / dayStat.Close
	case stores.VMA:
		value := s.computeVMA(securityStats)
		return value, (float64(dayStat.Volume) - value) / value
	default:
		return 0, 0
	}
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
		deltaP := lastNStats[i-1].Close - lastNStats[i].Close

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

func mergeMaps[K comparable, V any](dst, src map[K]V) {
	for k, v := range src {
		dst[k] = v
	}
}
