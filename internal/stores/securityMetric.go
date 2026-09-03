package stores

import (
	"fmt"
	"time"

	"github.com/vmihailenco/msgpack/v5"
	"gofr.dev/pkg/gofr"
)

type SecurityMetricStore interface {
	Index(ctx *gofr.Context, securityIDs []int, date time.Time) ([]*SecurityMetric, error)
	Create(ctx *gofr.Context, securityMetrics []*SecurityMetric, date time.Time) error
}

type SecurityMetric struct {
	SecurityID int
	MetricID   int
	Date       time.Time
	Value      float64
	ZValue     float64
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

func (*securityMetricStore) Index(ctx *gofr.Context, securityIDs []int, date time.Time) ([]*SecurityMetric, error) {
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

		var metricVals []*securityMetricCacheValue
		if err := msgpack.Unmarshal([]byte(val.(string)), &metricVals); err != nil {
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

func (*securityMetricStore) Create(ctx *gofr.Context, securityMetrics []*SecurityMetric, date time.Time) error {
	grouped := make(map[int][]*securityMetricCacheValue)

	for _, metric := range securityMetrics {
		grouped[metric.SecurityID] = append(grouped[metric.SecurityID], &securityMetricCacheValue{
			MetricID: metric.MetricID,
			Value:    metric.Value,
			ZValue:   metric.ZValue,
		})
	}

	pipe := ctx.Redis.Pipeline()

	for securityID, metrics := range grouped {
		key := fmt.Sprintf(SecurityMetricsServerCacheKey, securityID, date.Format(time.DateOnly))

		bytes, err := msgpack.Marshal(metrics)
		if err != nil {
			return err
		}

		pipe.SetEx(ctx, key, bytes, 5*time.Hour) //nolint:mnd // cache TTL
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return err
	}

	return nil
}
