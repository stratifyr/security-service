package services

import (
	"gofr.dev/pkg/gofr"

	"github.com/stratifyr/security-service/internal/stores"
)

type IndustryService interface {
	Index(ctx *gofr.Context) []stores.Industry
}

type industryService struct {
	store stores.IndustryStore
}

func NewIndustryService(store stores.IndustryStore) *industryService {
	return &industryService{store: store}
}

func (s *industryService) Index(ctx *gofr.Context) []stores.Industry {
	return s.store.Index(ctx)
}
