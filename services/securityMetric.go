package services

import (
	"time"

	"gofr.dev/pkg/gofr"

	"github.com/stratifyr/security-service/stores"
)

type SecurityMetricService interface {
	Index(ctx *gofr.Context, f *SecurityMetricFilter, page, perPage int) ([]*SecurityMetric, int, error)
	Read(ctx *gofr.Context, id int) (*SecurityMetric, error)
	Create(ctx *gofr.Context, payload *SecurityMetricCreate) (*SecurityMetric, error)
	Patch(ctx *gofr.Context, id int, payload *SecurityMetricUpdate) (*SecurityMetric, error)
}

type SecurityMetricFilter struct {
	Date       time.Time
	SecurityID int
}

type SecurityMetric struct {
	ID         int
	SecurityID int
	MetricID   int
	Date       time.Time
	Value      float64
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type SecurityMetricCreate struct {
	UserID     int
	SecurityID int
	MetricID   int
	Date       time.Time
	Value      float64
}

type SecurityMetricUpdate struct {
	UserID int
	Value  float64
}

type securityMetricService struct {
	store stores.SecurityMetricStore
}

func NewSecurityMetricService(store stores.SecurityMetricStore) *securityMetricService {
	return &securityMetricService{store: store}
}

func (s *securityMetricService) Index(ctx *gofr.Context, f *SecurityMetricFilter, page, perPage int) ([]*SecurityMetric, int, error) {
	limit := perPage
	offset := limit * (page - 1)

	filter := &stores.SecurityMetricFilter{
		SecurityID: f.SecurityID,
		Date:       f.Date,
	}

	securityMetrics, err := s.store.Index(ctx, filter, limit, offset)
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

	var resp = make([]*SecurityMetric, len(securityMetrics))

	for i := range securityMetrics {
		resp[i] = s.buildResp(securityMetrics[i])
	}

	return resp, count, nil
}

func (s *securityMetricService) Read(ctx *gofr.Context, id int) (*SecurityMetric, error) {
	securityMetric, err := s.store.Retrieve(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.buildResp(securityMetric), nil
}

func (s *securityMetricService) Create(ctx *gofr.Context, payload *SecurityMetricCreate) (*SecurityMetric, error) {
	if payload.UserID != 1 {
		return nil, &ErrResp{Code: 403, Message: "Forbidden"}
	}

	model := &stores.SecurityMetric{
		SecurityID: payload.SecurityID,
		MetricID:   payload.MetricID,
		Date:       payload.Date,
		Value:      payload.Value,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}

	securityMetric, err := s.store.Create(ctx, model)
	if err != nil {
		return nil, err
	}

	return s.buildResp(securityMetric), nil
}

func (s *securityMetricService) Patch(ctx *gofr.Context, id int, payload *SecurityMetricUpdate) (*SecurityMetric, error) {
	if payload.UserID != 1 {
		return nil, &ErrResp{Code: 403, Message: "Forbidden"}
	}

	oldSecurityMetric, err := s.store.Retrieve(ctx, id)
	if err != nil {
		return nil, err
	}

	newSecurityMetric := *oldSecurityMetric

	if payload.Value != 0 {
		newSecurityMetric.Value = payload.Value
	}

	securityMetric, err := s.store.Update(ctx, id, &newSecurityMetric)
	if err != nil {
		return nil, err
	}

	return s.buildResp(securityMetric), nil
}

func (s *securityMetricService) buildResp(model *stores.SecurityMetric) *SecurityMetric {
	resp := &SecurityMetric{
		ID:         model.ID,
		SecurityID: model.SecurityID,
		MetricID:   model.MetricID,
		Date:       model.Date,
		Value:      model.Value,
		CreatedAt:  model.CreatedAt,
		UpdatedAt:  model.UpdatedAt,
	}

	return resp
}
