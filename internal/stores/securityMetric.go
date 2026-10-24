package stores

import (
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/vmihailenco/msgpack/v5"
	"gofr.dev/pkg/gofr"
	"gofr.dev/pkg/gofr/datasource"
)

const (
	SecurityMetricCacheTTL = 24 * time.Hour
)

var ErrUnexpectedCacheValueType = errors.New("unexpected security metrics cache value type")

type SecurityMetricStore interface {
	List(ctx *gofr.Context, securityIDs []int) ([]*SecurityMetric, error)
	Create(ctx *gofr.Context, securityMetrics []*SecurityMetric) error
}

type SecurityMetric struct {
	SecurityID int
	MetricID   int
	Date       time.Time
	Value      float64
	ZValue     float64
}

type SecurityMetricCacheEntry struct {
	SecurityID int
	Date       time.Time
	Metrics    []*SecurityMetric
}

type securityMetricStore struct{}

func NewSecurityMetricStore() *securityMetricStore {
	return &securityMetricStore{}
}

func (*securityMetricStore) List(ctx *gofr.Context, securityIDs []int) ([]*SecurityMetric, error) {
	val, err := ctx.Redis.Get(ctx, SecurityMetricsServerCacheKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			ctx.Logger.Warnf("cache miss: %s", SecurityMetricsServerCacheKey)
		}

		return nil, datasource.ErrorDB{Err: err}
	}

	securityMetrics, err := decodeSecurityMetricCacheValue(val)
	if err != nil {
		return nil, datasource.ErrorDB{Err: err}
	}

	var requiredSecurities = make(map[int]struct{})

	for i := range securityIDs {
		requiredSecurities[securityIDs[i]] = struct{}{}
	}

	var filteredMetrics []*SecurityMetric

	for _, securityMetric := range securityMetrics {
		if _, ok := requiredSecurities[securityMetric.SecurityID]; ok {
			filteredMetrics = append(filteredMetrics, securityMetric)
		}
	}

	return filteredMetrics, nil
}

func (*securityMetricStore) Create(ctx *gofr.Context, securityMetrics []*SecurityMetric) error {
	if len(securityMetrics) == 0 {
		return nil
	}

	bytes, err := msgpack.Marshal(securityMetrics)
	if err != nil {
		return datasource.ErrorDB{Err: err}
	}

	if err = ctx.Redis.SetEx(ctx, SecurityMetricsServerCacheKey, bytes, SecurityMetricCacheTTL).Err(); err != nil {
		return datasource.ErrorDB{Err: err}
	}

	return nil
}

func decodeSecurityMetricCacheValue(val any) ([]*SecurityMetric, error) {
	var bytes []byte

	switch value := val.(type) {
	case string:
		bytes = []byte(value)
	case []byte:
		bytes = value
	default:
		return nil, fmt.Errorf("%w: %T", ErrUnexpectedCacheValueType, val)
	}

	var metricVals []*SecurityMetric
	if err := msgpack.Unmarshal(bytes, &metricVals); err != nil {
		return nil, err
	}

	return metricVals, nil
}
