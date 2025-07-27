package services

import (
	"time"

	"gofr.dev/pkg/gofr"

	"github.com/stratifyr/security-service/internal/stores"
)

type MetricService interface {
	Index(ctx *gofr.Context, f *MetricFilter, page, perPage int) ([]*Metric, int, error)
	Read(ctx *gofr.Context, id int) (*Metric, error)
	Create(ctx *gofr.Context, payload *MetricCreate) (*Metric, error)
	Patch(ctx *gofr.Context, id int, payload *MetricUpdate) (*Metric, error)
}

type MetricFilter struct {
	Type   string
	Period int
}

type Metric struct {
	ID        int
	Name      string
	Type      string
	Period    int
	Indicator string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type MetricCreate struct {
	UserID int
	Name   string
	Type   string
	Period int
}

type MetricUpdate struct {
	UserID int
	Name   string
}

var MetricTypeIndicator = map[stores.MetricType]stores.MetricIndicator{
	stores.SMA: stores.Trend,
	stores.EMA: stores.Trend,
	stores.RSI: stores.Momentum,
	stores.ROC: stores.Momentum,
	stores.ATR: stores.Volatility,
	stores.VMA: stores.Volume,
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
		Type:   nil,
		Period: f.Period,
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
		Period:    payload.Period,
		Indicator: MetricTypeIndicator[metricType],
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

	metric, err := s.store.Retrieve(ctx, id)
	if err != nil {
		return nil, err
	}

	if payload.Name != "" {
		metric.Name = payload.Name
	}

	metric, err = s.store.Update(ctx, id, metric)
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
		Period:    model.Period,
		Indicator: model.Indicator.String(),
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}

	return resp
}
