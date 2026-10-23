package stores

import (
	"errors"
	"fmt"
	"time"

	"github.com/vmihailenco/msgpack/v5"
	"gofr.dev/pkg/gofr"
)

const (
	SecurityMetricCacheTTL = 24 * time.Hour
)

var ErrUnexpectedCacheValueType = errors.New("unexpected security metrics cache value type")

type SecurityMetricStore interface {
	List(ctx *gofr.Context, securityIDs []int, date time.Time) ([]*SecurityMetric, error)
	GetBySecurityID(ctx *gofr.Context, securityID int, dates []time.Time) (map[time.Time][]*SecurityMetric, []time.Time, error)
	Create(ctx *gofr.Context, entries []*SecurityMetricCacheEntry) error
}

type SecurityMetric struct {
	SecurityID int
	MetricID   int
	Date       time.Time
	Value      float64
	ZValue     float64
}

// SecurityMetricCacheEntry is a single cache write for one security and date. Metrics may be empty,
// which stores an empty value so that days without computable metrics are not recomputed on every request.
type SecurityMetricCacheEntry struct {
	SecurityID int
	Date       time.Time
	Metrics    []*SecurityMetric
}

type securityMetricCacheValue struct {
	MetricID int
	Value    float64
	ZValue   float64
}

type securityMetricStore struct{}

func NewSecurityMetricStore() *securityMetricStore {
	return &securityMetricStore{}
}

func (*securityMetricStore) List(ctx *gofr.Context, securityIDs []int, date time.Time) ([]*SecurityMetric, error) {
	keys := make([]string, len(securityIDs))

	for i, securityID := range securityIDs {
		keys[i] = fmt.Sprintf(SecurityMetricsServerCacheKey, securityID, date.Format(time.DateOnly))
	}

	vals, err := ctx.Redis.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}

	var securityMetrics []*SecurityMetric

	for i, val := range vals {
		if val == nil {
			ctx.Logger.Warnf("cache miss: %s", keys[i])
			continue
		}

		metricVals, err := decodeSecurityMetricCacheValue(val)
		if err != nil {
			return nil, err
		}

		for _, mv := range metricVals {
			securityMetrics = append(securityMetrics, &SecurityMetric{
				SecurityID: securityIDs[i],
				MetricID:   mv.MetricID,
				Date:       date,
				Value:      mv.Value,
				ZValue:     mv.ZValue,
			})
		}
	}

	return securityMetrics, nil
}

func (*securityMetricStore) GetBySecurityID(ctx *gofr.Context, securityID int,
	dates []time.Time) (map[time.Time][]*SecurityMetric, []time.Time, error) {
	keys := make([]string, len(dates))

	for i, date := range dates {
		keys[i] = fmt.Sprintf(SecurityMetricsServerCacheKey, securityID, date.Format(time.DateOnly))
	}

	vals, err := ctx.Redis.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, nil, err
	}

	var (
		cached       = make(map[time.Time][]*SecurityMetric, len(dates))
		missingDates = make([]time.Time, 0)
	)

	for i, val := range vals {
		if val == nil {
			ctx.Logger.Warnf("cache miss: %s", keys[i])
			missingDates = append(missingDates, dates[i])

			continue
		}

		metricVals, err := decodeSecurityMetricCacheValue(val)
		if err != nil {
			return nil, nil, err
		}

		metrics := make([]*SecurityMetric, len(metricVals))

		for j := range metricVals {
			metrics[j] = &SecurityMetric{
				SecurityID: securityID,
				MetricID:   metricVals[j].MetricID,
				Date:       dates[i],
				Value:      metricVals[j].Value,
				ZValue:     metricVals[j].ZValue,
			}
		}

		cached[dates[i]] = metrics
	}

	return cached, missingDates, nil
}

func (*securityMetricStore) Create(ctx *gofr.Context, entries []*SecurityMetricCacheEntry) error {
	if len(entries) == 0 {
		return nil
	}

	pipe := ctx.Redis.Pipeline()

	for _, entry := range entries {
		metricVals := make([]*securityMetricCacheValue, len(entry.Metrics))

		for i := range entry.Metrics {
			metricVals[i] = &securityMetricCacheValue{
				MetricID: entry.Metrics[i].MetricID,
				Value:    entry.Metrics[i].Value,
				ZValue:   entry.Metrics[i].ZValue,
			}
		}

		bytes, err := msgpack.Marshal(metricVals)
		if err != nil {
			return err
		}

		key := fmt.Sprintf(SecurityMetricsServerCacheKey, entry.SecurityID, entry.Date.Format(time.DateOnly))

		pipe.SetEx(ctx, key, bytes, SecurityMetricCacheTTL)
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return err
	}

	return nil
}

func decodeSecurityMetricCacheValue(val any) ([]*securityMetricCacheValue, error) {
	var bytes []byte

	switch value := val.(type) {
	case string:
		bytes = []byte(value)
	case []byte:
		bytes = value
	default:
		return nil, fmt.Errorf("%w: %T", ErrUnexpectedCacheValueType, val)
	}

	var metricVals []*securityMetricCacheValue
	if err := msgpack.Unmarshal(bytes, &metricVals); err != nil {
		return nil, err
	}

	return metricVals, nil
}
