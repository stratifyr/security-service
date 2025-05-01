package services

import (
	"slices"
	"time"

	"gofr.dev/pkg/gofr"

	"github.com/stratifyr/security-service/stores"
)

type UniverseSecurityService interface {
	Index(ctx *gofr.Context, f *UniverseSecurityFilter, page, perPage int) ([]*UniverseSecurity, int, error)
	Read(ctx *gofr.Context, id, userID int) (*UniverseSecurity, error)
	Create(ctx *gofr.Context, payload *UniverseSecurityCreate) (*UniverseSecurity, error)
	Patch(ctx *gofr.Context, id int, payload *UniverseSecurityUpdate) (*UniverseSecurity, error)
	Delete(ctx *gofr.Context, id, userID int) error
}

type UniverseSecurityFilter struct {
	UserID int
	Status string
}

type UniverseSecurity struct {
	ID         int
	UniverseID int
	SecurityID int
	Status     string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type UniverseSecurityCreate struct {
	UserID     int
	UniverseID int
	SecurityID int
}

type UniverseSecurityUpdate struct {
	UserID int
	Status string
}

type universeSecurityService struct {
	securityService SecurityService
	universeService UniverseService
	store           stores.UniverseSecurityStore
}

func NewUniverseSecurityService(securitySvc SecurityService, universeSvc UniverseService, store stores.UniverseSecurityStore) *universeSecurityService {
	return &universeSecurityService{
		securityService: securitySvc,
		universeService: universeSvc,
		store:           store,
	}
}

func (s *universeSecurityService) Index(ctx *gofr.Context, f *UniverseSecurityFilter, page, perPage int) ([]*UniverseSecurity, int, error) {
	limit := perPage
	offset := limit * (page - 1)

	filter := &stores.UniverseSecurityFilter{
		Status: f.Status,
	}

	if f.UserID != 0 {
		universes, count, err := s.universeService.Index(ctx, &UniverseFilter{UserID: f.UserID}, 0, 0)
		if err != nil {
			return nil, 0, err
		}

		if count == 0 {
			return nil, 0, nil
		}

		for i := range universes {
			filter.UniverseIDs = append(filter.UniverseIDs, universes[i].ID)
		}
	}

	universeSecurities, err := s.store.Index(ctx, filter, limit, offset)
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

	var resp = make([]*UniverseSecurity, len(universeSecurities))

	for i := range universeSecurities {
		resp[i] = s.buildResp(universeSecurities[i])
	}

	return resp, count, nil
}

func (s *universeSecurityService) Read(ctx *gofr.Context, id, userID int) (*UniverseSecurity, error) {
	universeSecurity, err := s.store.Retrieve(ctx, id)
	if err != nil {
		return nil, err
	}

	universes, _, err := s.universeService.Index(ctx, &UniverseFilter{UserID: userID}, 0, 0)
	if err != nil {
		return nil, err
	}

	isUsersUniverse := slices.ContainsFunc(universes, func(universe *Universe) bool {
		return universe.ID == universeSecurity.UniverseID
	})

	if !isUsersUniverse {
		return nil, &ErrResp{Code: 403, Message: "Forbidden"}
	}

	return s.buildResp(universeSecurity), nil
}

func (s *universeSecurityService) Create(ctx *gofr.Context, payload *UniverseSecurityCreate) (*UniverseSecurity, error) {
	if _, err := s.securityService.Read(ctx, payload.SecurityID); err != nil {
		return nil, err
	}

	universes, _, err := s.universeService.Index(ctx, &UniverseFilter{UserID: payload.UserID}, 0, 0)
	if err != nil {
		return nil, err
	}

	isUsersUniverse := slices.ContainsFunc(universes, func(universe *Universe) bool {
		return universe.ID == payload.UniverseID
	})

	if !isUsersUniverse {
		return nil, &ErrResp{Code: 403, Message: "Forbidden"}
	}

	model := &stores.UniverseSecurity{
		UniverseID: payload.UniverseID,
		SecurityID: payload.SecurityID,
		Status:     "ENABLED",
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}

	security, err := s.store.Create(ctx, model)
	if err != nil {
		return nil, err
	}

	return s.buildResp(security), nil
}

func (s *universeSecurityService) Patch(ctx *gofr.Context, id int, payload *UniverseSecurityUpdate) (*UniverseSecurity, error) {
	oldUniverseSecurity, err := s.store.Retrieve(ctx, id)
	if err != nil {
		return nil, err
	}

	universes, _, err := s.universeService.Index(ctx, &UniverseFilter{UserID: payload.UserID}, 0, 0)
	if err != nil {
		return nil, err
	}

	isUsersUniverse := slices.ContainsFunc(universes, func(universe *Universe) bool {
		return universe.ID == oldUniverseSecurity.UniverseID
	})

	if !isUsersUniverse {
		return nil, &ErrResp{Code: 403, Message: "Forbidden"}
	}

	newUniverseSecurity := *oldUniverseSecurity

	if payload.Status != "" {
		newUniverseSecurity.Status = payload.Status
	}

	universeSecurity, err := s.store.Update(ctx, id, &newUniverseSecurity)
	if err != nil {
		return nil, err
	}

	return s.buildResp(universeSecurity), nil
}

func (s *universeSecurityService) Delete(ctx *gofr.Context, id, userID int) error {
	universeSecurity, err := s.store.Retrieve(ctx, id)
	if err != nil {
		return err
	}

	universes, _, err := s.universeService.Index(ctx, &UniverseFilter{UserID: userID}, 0, 0)
	if err != nil {
		return err
	}

	isUsersUniverse := slices.ContainsFunc(universes, func(universe *Universe) bool {
		return universe.ID == universeSecurity.UniverseID
	})

	if !isUsersUniverse {
		return &ErrResp{Code: 403, Message: "Forbidden"}
	}

	err = s.store.Delete(ctx, id)
	if err != nil {
		return err
	}

	return nil
}

func (s *universeSecurityService) buildResp(model *stores.UniverseSecurity) *UniverseSecurity {
	security := &UniverseSecurity{
		ID:         model.ID,
		UniverseID: model.UniverseID,
		SecurityID: model.SecurityID,
		Status:     model.Status,
		CreatedAt:  model.CreatedAt,
		UpdatedAt:  model.UpdatedAt,
	}

	return security
}
