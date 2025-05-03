package services

import (
	"gofr.dev/pkg/gofr/http"
	"slices"
	"strconv"
	"time"

	"gofr.dev/pkg/gofr"

	"github.com/stratifyr/security-service/stores"
)

type UniverseService interface {
	Index(ctx *gofr.Context, f *UniverseFilter, page, perPage int) ([]*Universe, int, error)
	Read(ctx *gofr.Context, id, userID int) (*Universe, error)
	Create(ctx *gofr.Context, payload *UniverseCreate) (*Universe, error)
	Patch(ctx *gofr.Context, id int, payload *UniverseUpdate) (*Universe, error)
}

type UniverseFilter struct {
	UserID         int
	IncludeDefault bool
}

type Universe struct {
	ID                 int
	UserID             int
	Name               string
	CreatedAt          time.Time
	UpdatedAt          time.Time
	UniverseSecurities []*struct {
		ID         int
		UniverseID int
		SecurityID int
		Status     string
		Security   *struct {
			ID           int
			ISIN         string
			Symbol       string
			Industry     string
			Name         string
			Image        string
			LTP          float64
			SecurityStat *struct {
				ID         int
				SecurityID int
				Date       time.Time
				Open       float64
				Close      float64
				High       float64
				Low        float64
				Volume     int
			}
			SecurityMetrics []*struct {
				ID         int
				SecurityID int
				MetricID   int
				Date       time.Time
				Value      float64
				Metric     *struct {
					ID   int
					Name string
					Type string
				}
			}
		}
	}
}

type UniverseCreate struct {
	UserID      int
	Name        string
	SecurityIDs []int
}

type UniverseUpdate struct {
	UserID int
	Name   string
}

type universeService struct {
	securityService SecurityService
	store           stores.UniverseStore
}

func NewUniverseService(securityService SecurityService, store stores.UniverseStore) *universeService {
	return &universeService{
		securityService: securityService,
		store:           store,
	}
}

func (s *universeService) Index(ctx *gofr.Context, f *UniverseFilter, page, perPage int) ([]*Universe, int, error) {
	limit := perPage
	offset := limit * (page - 1)

	filter := &stores.UniverseFilter{
		UserIDs: []int{f.UserID},
	}

	if f.IncludeDefault {
		filter.UserIDs = append(filter.UserIDs, 1)
	}

	universes, err := s.store.Index(ctx, filter, limit, offset)
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

	var resp = make([]*Universe, len(universes))

	for i := range universes {
		resp[i], err = s.buildResp(ctx, universes[i])
		if err != nil {
			return nil, 0, err
		}
	}

	return resp, count, nil
}

func (s *universeService) Read(ctx *gofr.Context, id, userID int) (*Universe, error) {
	universe, err := s.store.Retrieve(ctx, id)
	if err != nil {
		return nil, err
	}

	if universe.UserID != 1 && universe.UserID != userID {
		return nil, &ErrResp{Code: 403, Message: "Forbidden"}
	}

	return s.buildResp(ctx, universe)
}

func (s *universeService) Create(ctx *gofr.Context, payload *UniverseCreate) (*Universe, error) {
	securities, _, err := s.securityService.Index(ctx, &SecurityFilter{IDs: payload.SecurityIDs}, 0, 0)
	if err != nil {
		return nil, err
	}

	for i := range securities {
		if !slices.Contains(payload.SecurityIDs, securities[i].ID) {
			return nil, http.ErrorEntityNotFound{Name: "security", Value: strconv.Itoa(securities[i].ID)}
		}
	}

	model := &stores.Universe{
		UserID:             payload.UserID,
		Name:               payload.Name,
		CreatedAt:          time.Now().UTC(),
		UpdatedAt:          time.Now().UTC(),
		UniverseSecurities: make([]*stores.UniverseSecurity, len(payload.SecurityIDs)),
	}

	for i, securityID := range payload.SecurityIDs {
		model.UniverseSecurities[i] = &stores.UniverseSecurity{
			SecurityID: securityID,
			Status:     "ENABLED",
			CreatedAt:  time.Now().UTC(),
			UpdatedAt:  time.Now().UTC(),
		}
	}

	universe, err := s.store.Create(ctx, model)
	if err != nil {
		return nil, err
	}

	return s.buildResp(ctx, universe)
}

func (s *universeService) Patch(ctx *gofr.Context, id int, payload *UniverseUpdate) (*Universe, error) {
	oldUniverse, err := s.store.Retrieve(ctx, id)
	if err != nil {
		return nil, err
	}

	if oldUniverse.UserID != payload.UserID {
		return nil, &ErrResp{Code: 403, Message: "Forbidden"}
	}

	newUniverse := *oldUniverse

	if payload.Name != "" {
		newUniverse.Name = payload.Name
	}

	universe, err := s.store.Update(ctx, id, &newUniverse)
	if err != nil {
		return nil, err
	}

	return s.buildResp(ctx, universe)
}

func (s *universeService) buildResp(ctx *gofr.Context, model *stores.Universe) (*Universe, error) {
	resp := &Universe{
		ID:                 model.ID,
		UserID:             model.UserID,
		Name:               model.Name,
		CreatedAt:          model.CreatedAt,
		UpdatedAt:          model.UpdatedAt,
		UniverseSecurities: nil,
	}

	if len(model.UniverseSecurities) == 0 {
		return resp, nil
	}

	resp.UniverseSecurities = make([]*struct {
		ID         int
		UniverseID int
		SecurityID int
		Status     string
		Security   *struct {
			ID           int
			ISIN         string
			Symbol       string
			Industry     string
			Name         string
			Image        string
			LTP          float64
			SecurityStat *struct {
				ID         int
				SecurityID int
				Date       time.Time
				Open       float64
				Close      float64
				High       float64
				Low        float64
				Volume     int
			}
			SecurityMetrics []*struct {
				ID         int
				SecurityID int
				MetricID   int
				Date       time.Time
				Value      float64
				Metric     *struct {
					ID   int
					Name string
					Type string
				}
			}
		}
	}, len(model.UniverseSecurities))

	var securityIDs []int

	for i := range model.UniverseSecurities {
		securityIDs = append(securityIDs, model.UniverseSecurities[i].SecurityID)
	}

	securities, _, err := s.securityService.Index(ctx, &SecurityFilter{IDs: securityIDs}, 0, 0)
	if err != nil {
		return nil, err
	}

	var securityMapping = make(map[int]*Security)

	for i := range securities {
		securityMapping[securities[i].ID] = securities[i]
	}

	for i := range model.UniverseSecurities {
		security := securityMapping[model.UniverseSecurities[i].SecurityID]

		resp.UniverseSecurities[i] = &struct {
			ID         int
			UniverseID int
			SecurityID int
			Status     string
			Security   *struct {
				ID           int
				ISIN         string
				Symbol       string
				Industry     string
				Name         string
				Image        string
				LTP          float64
				SecurityStat *struct {
					ID         int
					SecurityID int
					Date       time.Time
					Open       float64
					Close      float64
					High       float64
					Low        float64
					Volume     int
				}
				SecurityMetrics []*struct {
					ID         int
					SecurityID int
					MetricID   int
					Date       time.Time
					Value      float64
					Metric     *struct {
						ID   int
						Name string
						Type string
					}
				}
			}
		}{
			ID:         model.UniverseSecurities[i].ID,
			UniverseID: model.UniverseSecurities[i].UniverseID,
			SecurityID: model.UniverseSecurities[i].SecurityID,
			Status:     model.UniverseSecurities[i].Status,
			Security: &struct {
				ID           int
				ISIN         string
				Symbol       string
				Industry     string
				Name         string
				Image        string
				LTP          float64
				SecurityStat *struct {
					ID         int
					SecurityID int
					Date       time.Time
					Open       float64
					Close      float64
					High       float64
					Low        float64
					Volume     int
				}
				SecurityMetrics []*struct {
					ID         int
					SecurityID int
					MetricID   int
					Date       time.Time
					Value      float64
					Metric     *struct {
						ID   int
						Name string
						Type string
					}
				}
			}{
				ID:              security.ID,
				ISIN:            security.ISIN,
				Symbol:          security.Symbol,
				Industry:        security.Industry,
				Name:            security.Name,
				Image:           security.Image,
				LTP:             security.LTP,
				SecurityStat:    security.SecurityStat,
				SecurityMetrics: security.SecurityMetrics,
			},
		}
	}

	return resp, nil
}
