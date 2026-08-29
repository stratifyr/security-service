package services

import (
	"fmt"

	"gofr.dev/pkg/gofr"

	"github.com/stratifyr/security-service/internal/stores"
)

type MetricService interface {
	Index(ctx *gofr.Context) []*Metric
}

type Metric struct {
	ID        int
	Name      string
	Type      stores.MetricType
	Period    int
	Indicator stores.MetricIndicator
}

type metricService struct {
	store stores.MetricStore
}

func NewMetricService(store stores.MetricStore) *metricService {
	return &metricService{store: store}
}

func (s *metricService) Index(ctx *gofr.Context) []*Metric {
	metrics := s.store.Index(ctx)

	var resp = make([]*Metric, len(metrics))

	for i := range metrics {
		resp[i] = s.buildResp(metrics[i])
	}

	return resp
}

func (s *metricService) buildResp(model *stores.Metric) *Metric {
	resp := &Metric{
		ID:        model.ID,
		Name:      fmt.Sprintf("%s_%d", model.Type, model.Period),
		Type:      model.Type,
		Period:    model.Period,
		Indicator: model.Indicator,
	}

	return resp
}
