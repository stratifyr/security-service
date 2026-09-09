package services

import (
	"gofr.dev/pkg/gofr"

	"github.com/stratifyr/security-service/internal/stores"
)

type MarketDataJobTypeService interface {
	List(ctx *gofr.Context) []stores.MarketDataJobType
}

type marketDataJobTypeService struct {
	store stores.MarketDataJobTypeStore
}

func NewMarketDataJobTypeService(store stores.MarketDataJobTypeStore) *marketDataJobTypeService {
	return &marketDataJobTypeService{store: store}
}

func (s *marketDataJobTypeService) List(ctx *gofr.Context) []stores.MarketDataJobType {
	return s.store.List(ctx)
}
