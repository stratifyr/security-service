package services

import (
	"time"

	"gofr.dev/pkg/gofr"

	"github.com/stratifyr/security-service/internal/stores"
)

type MetricService interface {
	Index(ctx *gofr.Context) ([]*Metric, error)
	Read(ctx *gofr.Context, id int) (*Metric, error)
	Create(ctx *gofr.Context, payload *MetricCreate) (*Metric, error)
	Patch(ctx *gofr.Context, id int, payload *MetricUpdate) (*Metric, error)
}

type Metric struct {
	ID        int
	Name      string
	Type      stores.MetricType
	Period    int
	Indicator stores.MetricIndicator
	Tier      int
	CreatedAt time.Time
	UpdatedAt time.Time
}

type MetricCreate struct {
	UserID int
	Name   string
	Type   string
	Period int
	Tier   int
}

type MetricUpdate struct {
	UserID int
	Name   string
	Tier   *int
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

func (s *metricService) Index(ctx *gofr.Context) ([]*Metric, error) {
	metrics, err := s.store.Index(ctx)
	if err != nil {
		return nil, err
	}

	var resp = make([]*Metric, len(metrics))

	for i := range metrics {
		resp[i] = s.buildResp(metrics[i])
	}

	return resp, nil
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
		Type:      model.Type,
		Period:    model.Period,
		Indicator: model.Indicator,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}

	return resp
}
