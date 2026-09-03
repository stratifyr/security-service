package services

import (
	"encoding/json"
	"time"

	"gofr.dev/pkg/gofr"

	"github.com/stratifyr/security-service/internal/stores"
)

type MarketDataJobService interface {
	Index(ctx *gofr.Context, f *MarketDataJobFilter, page, perPage int) ([]*MarketDataJob, int, error)
	Read(ctx *gofr.Context, id int) (*MarketDataJob, error)
	Create(ctx *gofr.Context, payload *MarketDataJobCreate) (*MarketDataJob, error)
	Patch(ctx *gofr.Context, id int, payload *MarketDataJobUpdate) (*MarketDataJob, error)
}

type MarketDataJobFilter struct {
	UserID int
	Status string
}

type MarketDataJob struct {
	ID        int
	Type      stores.MarketDataJobType
	Status    string
	Logs      *json.RawMessage
	CreatedAt time.Time
	UpdatedAt time.Time
}

type MarketDataJobCreate struct {
	UserID int
	Type   string
}

type MarketDataJobUpdate struct {
	UserID int
	Status string
	Logs   *json.RawMessage
}

type marketDataJobService struct {
	store stores.MarketDataJobStore
}

func NewMarketDataJobService(store stores.MarketDataJobStore) *marketDataJobService {
	return &marketDataJobService{store: store}
}

func (s *marketDataJobService) Index(ctx *gofr.Context, f *MarketDataJobFilter, page, perPage int) ([]*MarketDataJob, int, error) {
	limit := perPage
	offset := limit * (page - 1)

	if f.UserID != 1 {
		return nil, 0, ErrForbidden
	}

	filter := &stores.MarketDataJobFilter{
		Status: f.Status,
	}

	marketDataJobs, err := s.store.Index(ctx, filter, limit, offset)
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

	var resp = make([]*MarketDataJob, len(marketDataJobs))

	for i := range marketDataJobs {
		resp[i] = s.buildResp(marketDataJobs[i])
	}

	return resp, count, nil
}

func (s *marketDataJobService) Read(ctx *gofr.Context, id int) (*MarketDataJob, error) {
	marketDataJob, err := s.store.Retrieve(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.buildResp(marketDataJob), nil
}

func (s *marketDataJobService) Create(ctx *gofr.Context, payload *MarketDataJobCreate) (*MarketDataJob, error) {
	if payload.UserID != 1 {
		return nil, ErrForbidden
	}

	jobType, err := stores.MarketDataJobTypeFromString(payload.Type)
	if err != nil {
		return nil, err
	}

	model := &stores.MarketDataJob{
		Type:      jobType,
		Status:    "CREATED",
		Logs:      nil,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	marketDataJob, err := s.store.Create(ctx, model)
	if err != nil {
		return nil, err
	}

	return s.buildResp(marketDataJob), nil
}

func (s *marketDataJobService) Patch(ctx *gofr.Context, id int, payload *MarketDataJobUpdate) (*MarketDataJob, error) {
	if payload.UserID != 1 {
		return nil, ErrForbidden
	}

	marketDataJob, err := s.store.Retrieve(ctx, id)
	if err != nil {
		return nil, err
	}

	if payload.Status != "" && payload.Status != marketDataJob.Status {
		if marketDataJob.Status != "CREATED" || (payload.Status != "COMPLETED" && payload.Status != "FAILED") {
			return nil, ErrInvalidStatusChange
		}

		marketDataJob.Status = payload.Status
	}

	if payload.Logs != nil {
		marketDataJob.Logs = payload.Logs
	}

	marketDataJob, err = s.store.Update(ctx, id, marketDataJob)
	if err != nil {
		return nil, err
	}

	return s.buildResp(marketDataJob), nil
}

func (*marketDataJobService) buildResp(model *stores.MarketDataJob) *MarketDataJob {
	resp := &MarketDataJob{
		ID:        model.ID,
		Type:      model.Type,
		Status:    model.Status,
		Logs:      model.Logs,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}

	return resp
}
