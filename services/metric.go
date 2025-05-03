package services

import (
	"time"

	"gofr.dev/pkg/gofr"

	"github.com/stratifyr/security-service/stores"
)

type MetricService interface {
	Index(ctx *gofr.Context, f *MetricFilter, page, perPage int) ([]*Metric, int, error)
	Read(ctx *gofr.Context, id int) (*Metric, error)
	Create(ctx *gofr.Context, payload *MetricCreate) (*Metric, error)
	Patch(ctx *gofr.Context, id int, payload *MetricUpdate) (*Metric, error)
}

type MetricFilter struct {
	Type string
}

type Metric struct {
	ID        int
	Name      string
	Type      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type MetricCreate struct {
	UserID int
	Name   string
	Type   string
}

type MetricUpdate struct {
	UserID int
	Name   string
	Type   string
}

type metricService struct {
	store stores.MetricStore
}

func NewMetricService(store stores.MetricStore) *metricService {
	return &metricService{store: store}
}

func (s *metricService) Index(ctx *gofr.Context, f *MetricFilter, page, perPage int) ([]*Metric, int, error) {
	limit := perPage
	offset := limit * (page - 1)

	filter := &stores.MetricFilter{
		Type: nil,
	}

	if f.Type != "" {
		metricType, err := stores.MetricTypeFromString(f.Type)
		if err != nil {
			return nil, 0, err
		}

		filter.Type = &metricType
	}

	metrics, err := s.store.Index(ctx, filter, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	count, err := s.store.Count(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	if count == 0 {
		return nil, 0, nil
	}

	var resp = make([]*Metric, len(metrics))

	for i := range metrics {
		resp[i] = s.buildResp(metrics[i])
	}

	return resp, count, nil
}

func (s *metricService) Read(ctx *gofr.Context, id int) (*Metric, error) {
	metric, err := s.store.Retrieve(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.buildResp(metric), nil
}

func (s *metricService) Create(ctx *gofr.Context, payload *MetricCreate) (*Metric, error) {
	if payload.UserID != 1 {
		return nil, &ErrResp{Code: 403, Message: "Forbidden"}
	}

	metricType, err := stores.MetricTypeFromString(payload.Type)
	if err != nil {
		return nil, err
	}

	model := &stores.Metric{
		Name:      payload.Name,
		Type:      metricType,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	metric, err := s.store.Create(ctx, model)
	if err != nil {
		return nil, err
	}

	return s.buildResp(metric), nil
}

func (s *metricService) Patch(ctx *gofr.Context, id int, payload *MetricUpdate) (*Metric, error) {
	if payload.UserID != 1 {
		return nil, &ErrResp{Code: 403, Message: "Forbidden"}
	}

	oldMetric, err := s.store.Retrieve(ctx, id)
	if err != nil {
		return nil, err
	}

	newMetric := *oldMetric

	if payload.Name != "" {
		newMetric.Name = payload.Name
	}

	if payload.Type != "" {
		newMetric.Type, err = stores.MetricTypeFromString(payload.Type)
		if err != nil {
			return nil, err
		}
	}

	metric, err := s.store.Update(ctx, id, &newMetric)
	if err != nil {
		return nil, err
	}

	return s.buildResp(metric), nil
}

func (s *metricService) buildResp(model *stores.Metric) *Metric {
	resp := &Metric{
		ID:        model.ID,
		Name:      model.Name,
		Type:      model.Type.String(),
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}

	return resp
}
