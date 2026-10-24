package services

import (
	"math"
	"sort"
	"time"

	"gofr.dev/pkg/gofr"

	"github.com/stratifyr/security-service/internal/stores"
)

type SecurityMetricService interface {
	List(ctx *gofr.Context, securityIDs []int, date time.Time) ([]*SecurityMetric, error)
	GetBySecurityID(ctx *gofr.Context, securityID int, dates []time.Time) (map[time.Time][]*SecurityMetric, error)
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

func (s *securityMetricService) List(ctx *gofr.Context, securityIDs []int, date time.Time) ([]*SecurityMetric, error) {
	securityMetrics, err := s.securityMetricStore.List(ctx, securityIDs, date)
	if err != nil {
		ctx.Logger.Warnf("failed to get security metrics from store: %v", err)
	}

	missingSecurityIDs := findMissingSecurityIDs(securityIDs, securityMetrics)
	metricsMap := s.getMetricsMap(ctx)

	if len(missingSecurityIDs) == 0 {
		return s.buildSecurityMetrics(metricsMap, securityMetrics), nil
	}

	computedSecurityMetrics, err := s.computeSecurityMetrics(ctx, missingSecurityIDs, date)
	if err != nil {
		return nil, err
	}

	if err = s.securityMetricStore.Create(ctx, buildCacheEntries(missingSecurityIDs, date, computedSecurityMetrics)); err != nil {
		ctx.Logger.Warnf("failed to save security metrics in store: %v", err)
	}

	securityMetrics = append(securityMetrics, computedSecurityMetrics...)

	return s.buildSecurityMetrics(metricsMap, securityMetrics), nil
}

func (s *securityMetricService) GetBySecurityID(ctx *gofr.Context,
	securityID int, dates []time.Time) (map[time.Time][]*SecurityMetric, error) {
	if len(dates) == 0 {
		return make(map[time.Time][]*SecurityMetric), nil
	}

	cachedMetrics, missingDates, err := s.securityMetricStore.GetBySecurityID(ctx, securityID, dates)
	if err != nil {
		ctx.Logger.Warnf("failed to get security metrics from store: %v", err)

		cachedMetrics, missingDates = make(map[time.Time][]*stores.SecurityMetric), dates
	}

	computedMetrics := make(map[time.Time][]*stores.SecurityMetric, len(missingDates))

	if len(missingDates) > 0 {
		computedMetrics, err = s.computeSecurityMetricsForDates(ctx, securityID, missingDates)
		if err != nil {
			return nil, err
		}

		entries := make([]*stores.SecurityMetricCacheEntry, 0, len(missingDates))

		for _, date := range missingDates {
			entries = append(entries, &stores.SecurityMetricCacheEntry{
				SecurityID: securityID,
				Date:       date,
				Metrics:    computedMetrics[date],
			})
		}

		if err = s.securityMetricStore.Create(ctx, entries); err != nil {
			ctx.Logger.Warnf("failed to save security metrics in store: %v", err)
		}
	}

	resp := make(map[time.Time][]*SecurityMetric, len(dates))
	metricsMap := s.getMetricsMap(ctx)

	for _, date := range dates {
		if metrics, ok := cachedMetrics[date]; ok {
			resp[date] = s.buildSecurityMetrics(metricsMap, metrics)
			continue
		}

		resp[date] = s.buildSecurityMetrics(metricsMap, computedMetrics[date])
	}

	return resp, nil
}

func (s *securityMetricService) computeSecurityMetrics(ctx *gofr.Context,
	securityIDs []int, date time.Time) ([]*stores.SecurityMetric, error) {
	metrics := s.metricService.List(ctx)
	maxPeriod := getMaxPeriod(metrics)

	marketDays, _, err := s.marketDayService.List(ctx, &MarketDayFilter{LastNDaysFromReference: &struct {
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

	securityStats, err := s.securityStatStore.List(ctx, &stores.SecurityStatFilter{SecurityIDs: securityIDs,
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

func (s *securityMetricService) computeSecurityMetricsForDates(ctx *gofr.Context,
	securityID int, dates []time.Time) (map[time.Time][]*stores.SecurityMetric, error) {
	metrics := s.metricService.List(ctx)
	maxPeriod := getMaxPeriod(metrics)
	startDate, endDate := minMaxDates(dates)

	marketDays, _, err := s.marketDayService.List(ctx, &MarketDayFilter{LastNDaysFromReference: &struct {
		N         int
		Reference time.Time
	}{N: maxPeriod, Reference: startDate}})
	if err != nil {
		return nil, err
	}

	statsStartDate, statsEndDate := marketDays[len(marketDays)-1], endDate

	securityStats, err := s.securityStatStore.List(ctx, &stores.SecurityStatFilter{SecurityIDs: []int{securityID},
		DateBetween: &struct {
			Start time.Time
			End   time.Time
		}{Start: statsStartDate, End: statsEndDate}}, 0, 0)
	if err != nil {
		return nil, err
	}

	sort.Slice(securityStats, func(i, j int) bool {
		return securityStats[i].Date.After(securityStats[j].Date)
	})

	positionByDate := make(map[time.Time]int, len(securityStats))

	for i := range securityStats {
		positionByDate[securityStats[i].Date] = i
	}

	computed := make(map[time.Time][]*stores.SecurityMetric, len(dates))

	for _, date := range dates {
		var window []*stores.SecurityStat

		if position, ok := positionByDate[date]; ok {
			window = securityStats[position:]

			if len(window) > maxPeriod {
				window = window[:maxPeriod]
			}
		}

		computed[date] = s.computeMetricsForWindow(securityID, date, metrics, window)
	}

	return computed, nil
}

func (s *securityMetricService) computeMetricsForWindow(securityID int, date time.Time,
	metrics []*Metric, window []*stores.SecurityStat) []*stores.SecurityMetric {
	securityMetrics := make([]*stores.SecurityMetric, 0, len(metrics))

	for i := range metrics {
		n := metrics[i].Period

		if len(window) < n {
			continue
		}

		value, normalizedValue := s.computeMetricValue(metrics[i], window[:n])

		securityMetrics = append(securityMetrics, &stores.SecurityMetric{
			SecurityID: securityID,
			MetricID:   metrics[i].ID,
			Date:       date,
			Value:      value,
			ZValue:     normalizedValue,
		})
	}

	return securityMetrics
}

func (s *securityMetricService) getMetricsMap(ctx *gofr.Context) map[int]*Metric {
	metrics := s.metricService.List(ctx)

	var metricsMap = make(map[int]*Metric, len(metrics))

	for _, metric := range metrics {
		metricsMap[metric.ID] = metric
	}

	return metricsMap
}

func (*securityMetricService) buildSecurityMetrics(metricsMap map[int]*Metric, models []*stores.SecurityMetric) []*SecurityMetric {
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

func buildCacheEntries(securityIDs []int, date time.Time, securityMetrics []*stores.SecurityMetric) []*stores.SecurityMetricCacheEntry {
	grouped := make(map[int][]*stores.SecurityMetric, len(securityIDs))

	for _, metric := range securityMetrics {
		grouped[metric.SecurityID] = append(grouped[metric.SecurityID], metric)
	}

	entries := make([]*stores.SecurityMetricCacheEntry, 0, len(securityIDs))

	for _, securityID := range securityIDs {
		if len(grouped[securityID]) == 0 {
			continue
		}

		entries = append(entries, &stores.SecurityMetricCacheEntry{
			SecurityID: securityID,
			Date:       date,
			Metrics:    grouped[securityID],
		})
	}

	return entries
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

func getMaxPeriod(metrics []*Metric) int {
	maxPeriod := 0

	for i := range metrics {
		if metrics[i].Period > maxPeriod {
			maxPeriod = metrics[i].Period
		}
	}

	return maxPeriod
}

func minMaxDates(dates []time.Time) (startDate, endDate time.Time) {
	startDate, endDate = dates[0], dates[0]

	for _, date := range dates {
		if date.Before(startDate) {
			startDate = date
		}

		if date.After(endDate) {
			endDate = date
		}
	}

	return startDate, endDate
}
