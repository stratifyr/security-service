package services

import (
	"math"
	"sort"
	"time"

	"gofr.dev/pkg/gofr"

	"github.com/stratifyr/security-service/internal/stores"
)

type SecurityMetricService interface {
	Index(ctx *gofr.Context, securityIDs []int, date time.Time) ([]*SecurityMetric, error)
}

type SecurityMetric struct {
	SecurityID int
	Date       time.Time
	Metric     *Metric
	Value      float64
	ZValue     float64
}

type securityMetricService struct {
	marketDayService    MarketDayService
	metricService       MetricService
	securityMetricStore stores.SecurityMetricStore
	securityStatStore   stores.SecurityStatStore
}

func NewSecurityMetricService(marketDayService MarketDayService, metricService MetricService,
	securityStatStore stores.SecurityStatStore, securityMetricStore stores.SecurityMetricStore) *securityMetricService {
	return &securityMetricService{
		marketDayService:    marketDayService,
		metricService:       metricService,
		securityMetricStore: securityMetricStore,
		securityStatStore:   securityStatStore,
	}
}

func (s *securityMetricService) Index(ctx *gofr.Context, securityIDs []int, date time.Time) ([]*SecurityMetric, error) {
	securityMetrics, err := s.securityMetricStore.Index(ctx, securityIDs, date)
	if err != nil {
		ctx.Logger.Warnf("failed to get security metrics from store: %v", map[string]any{
			"error":       err,
			"securityIds": securityIDs,
			"date":        date,
		})
	}

	missingSecurityIDs := findMissingSecurityIDs(securityIDs, securityMetrics)

	if len(missingSecurityIDs) == 0 {
		return s.buildResponse(ctx, securityMetrics), nil
	}

	computedSecurityMetrics, err := s.computeSecurityMetrics(ctx, missingSecurityIDs, date)
	if err != nil {
		return nil, err
	}

	if err = s.securityMetricStore.Create(ctx, computedSecurityMetrics, date); err != nil {
		ctx.Logger.Warnf("failed to save security metrics in store: %v", map[string]any{
			"error":       err,
			"securityIds": missingSecurityIDs,
			"date":        date,
		})
	}

	securityMetrics = append(securityMetrics, computedSecurityMetrics...)

	return s.buildResponse(ctx, securityMetrics), nil
}

func (s *securityMetricService) computeSecurityMetrics(ctx *gofr.Context,
	securityIDs []int, date time.Time) ([]*stores.SecurityMetric, error) {
	metrics := s.metricService.Index(ctx)

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

	var securityMetrics = make([]*stores.SecurityMetric, 0)

	for _, securityID := range securityIDs {
		stats := securityStatsMap[securityID]

		sort.Slice(stats, func(i, j int) bool {
			return stats[i].Date.After(stats[j].Date)
		})

		for i := range metrics {
			n := metrics[i].Period

			if len(stats) < n {
				ctx.Logger.Warnf("cannot compute securityId:%d_%s_%d, not enough data", securityID, metrics[i].Type.String(), metrics[i].Period)
				continue
			}

			value, normalizedValue := s.computeMetricValue(metrics[i], stats[:n])

			securityMetrics = append(securityMetrics, &stores.SecurityMetric{
				SecurityID: securityID,
				MetricID:   metrics[i].ID,
				Date:       date,
				Value:      value,
				ZValue:     normalizedValue,
			})
		}
	}

	return securityMetrics, nil
}

func (s *securityMetricService) buildResponse(ctx *gofr.Context, models []*stores.SecurityMetric) []*SecurityMetric {
	metrics := s.metricService.Index(ctx)

	var metricsMap = make(map[int]*Metric)

	for _, metric := range metrics {
		metricsMap[metric.ID] = metric
	}

	var securityMetrics []*SecurityMetric

	for i := range models {
		metric, ok := metricsMap[models[i].MetricID]
		if !ok {
			continue
		}

		securityMetrics = append(securityMetrics, &SecurityMetric{
			SecurityID: models[i].SecurityID,
			Date:       models[i].Date,
			Metric:     metric,
			Value:      models[i].Value,
			ZValue:     models[i].ZValue,
		})
	}

	return securityMetrics
}

func (s *securityMetricService) computeMetricValue(metric *Metric, securityStats []*stores.SecurityStat) (val, zVal float64) {
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
	case stores.ROC:
		value := s.computeROC(securityStats)
		return value, value / 100 //nolint:mnd // percentage divisor
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

func (*securityMetricService) computeSMA(lastNStats []*stores.SecurityStat) float64 {
	var (
		sumPrice float64
		n        = len(lastNStats)
	)

	for _, stat := range lastNStats {
		sumPrice += stat.Close
	}

	return sumPrice / float64(n)
}

func (*securityMetricService) computeEMA(k, seeder float64, lastNStats []*stores.SecurityStat) float64 {
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

func (*securityMetricService) computeROC(lastNStats []*stores.SecurityStat) float64 {
	n := len(lastNStats)
	currentPrice := lastNStats[0].Close
	nDaysPriorPrice := lastNStats[n-1].Close

	return ((currentPrice - nDaysPriorPrice) / nDaysPriorPrice) * 100 //nolint:mnd // percentage multiplier
}

func (*securityMetricService) computeATR(lastNStats []*stores.SecurityStat) float64 {
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

func (*securityMetricService) computeVMA(lastNStats []*stores.SecurityStat) float64 {
	var (
		sumVolume float64
		n         = len(lastNStats)
	)

	for _, stat := range lastNStats {
		sumVolume += float64(stat.Volume)
	}

	return sumVolume / float64(n)
}

func findMissingSecurityIDs(requestedIDs []int, storedMetrics []*stores.SecurityMetric) []int {
	found := make(map[int]struct{}, len(storedMetrics))

	for _, metric := range storedMetrics {
		found[metric.SecurityID] = struct{}{}
	}

	missing := make([]int, 0)

	for _, securityID := range requestedIDs {
		if _, ok := found[securityID]; !ok {
			missing = append(missing, securityID)
		}
	}

	return missing
}
