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
	List(ctx *gofr.Context, date time.Time) ([]*SecurityMetric, error)
	Create(ctx *gofr.Context, date time.Time, securityMetrics []*SecurityMetric) error
}

type SecurityMetric struct {
	SecurityID int
	MetricID   int
	Date       time.Time
	Value      float64
	ZValue     float64
}

type securityMetricStore struct{}

func NewSecurityMetricStore() *securityMetricStore {
	return &securityMetricStore{}
}

func (*securityMetricStore) List(ctx *gofr.Context, date time.Time) ([]*SecurityMetric, error) {
	key := fmt.Sprintf(SecurityMetricsServerCacheKey, date.Format(time.DateOnly))

	val, err := ctx.Redis.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			ctx.Logger.Warnf("cache miss: %s", key)
		}

		return nil, datasource.ErrorDB{Err: err}
	}

	return decodeSecurityMetricCacheValue(val)
}

func (*securityMetricStore) Create(ctx *gofr.Context, date time.Time, securityMetrics []*SecurityMetric) error {
	if len(securityMetrics) == 0 {
		return nil
	}

	bytes, err := msgpack.Marshal(securityMetrics)
	if err != nil {
		return datasource.ErrorDB{Err: err}
	}

	key := fmt.Sprintf(SecurityMetricsServerCacheKey, date.Format(time.DateOnly))

	if err = ctx.Redis.SetEx(ctx, key, bytes, SecurityMetricCacheTTL).Err(); err != nil {
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
