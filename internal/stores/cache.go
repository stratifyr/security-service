package stores

import (
	"strings"

	"gofr.dev/pkg/gofr"
)

const (
	MetricsClientCacheKey                       = "security-service:client-cache:metrics"
	SecuritiesClientCacheKey                    = "security-service:client-cache:securities:date:%s"
	SecuritiesClientCachePattern                = "security-service:client-cache:securities:date:*"
	SecurityMetricsServerCacheKey               = "security-service:server-cache:security-metrics:security-id:%d:date:%s"
	SecurityMetricsServerCachePatternBySecurity = "security-service:server-cache:security-metrics:security-id:%d:date:*"
)

type cacheInvalidator interface {
	cacheInvalidations() []string
}

func invalidateCache(ctx *gofr.Context, invalidator cacheInvalidator) {
	var (
		keys     []string
		patterns []string
	)

	for _, invalidation := range invalidator.cacheInvalidations() {
		if strings.ContainsAny(invalidation, "*?[") {
			patterns = append(patterns, invalidation)
			continue
		}

		keys = append(keys, invalidation)
	}

	if len(patterns) > 0 {
		matchingKeys, err := getMatchingKeys(ctx, patterns...)
		if err != nil {
			ctx.Logger.Warnf("failed to get matching cache keys: %v", err)
			return
		}

		keys = append(keys, matchingKeys...)
	}

	if len(keys) == 0 {
		return
	}

	if err := ctx.Redis.Unlink(ctx, keys...).Err(); err != nil {
		ctx.Logger.Warnf("failed to invalidate cache keys: %v", err)

		return
	}

	ctx.Logger.Infof("invalidated %d cache keys (patterns: %v)", len(keys), patterns)
}

func getMatchingKeys(ctx *gofr.Context, patterns ...string) ([]string, error) {
	var keys []string

	for _, pattern := range patterns {
		iter := ctx.Redis.Scan(ctx, 0, pattern, 500).Iterator() //nolint:mnd // Redis scan count

		for iter.Next(ctx) {
			keys = append(keys, iter.Val())
		}

		if err := iter.Err(); err != nil {
			return nil, err
		}
	}

	return keys, nil
}
