package services

import (
	"time"

	"gofr.dev/pkg/gofr"

	"github.com/stratifyr/security-service/internal/stores"
)

type SecurityService interface {
	Index(ctx *gofr.Context, f *SecurityFilter, page, perPage int) ([]*Security, int, error)
	Read(ctx *gofr.Context, id int) (*Security, error)
	Create(ctx *gofr.Context, payload *SecurityCreate) (*Security, error)
	Patch(ctx *gofr.Context, id int, payload *SecurityUpdate) (*Security, error)
}

type SecurityFilter struct {
	IDs    []int
	ISIN   string
	Symbol string
}

type Security struct {
	ID           int
	ISIN         string
	Symbol       string
	Industry     string
	Name         string
	Image        string
	LTP          float64
	CreatedAt    time.Time
	UpdatedAt    time.Time
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
		ID              int
		SecurityID      int
		MetricID        int
		Date            time.Time
		Value           float64
		NormalizedValue float64
		Metric          *struct {
			ID        int
			Name      string
			Type      string
			Period    int
			Indicator string
		}
	}
}

type SecurityCreate struct {
	UserID   int
	ISIN     string
	Symbol   string
	Industry string
	Name     string
	Image    string
	LTP      float64
}

type SecurityUpdate struct {
	UserID   int
	Symbol   string
	Industry string
	Name     string
	Image    string
	LTP      float64
}

type securityService struct {
	securityStatService   SecurityStatService
	metricsService        MetricService
	securityMetricService SecurityMetricService
	store                 stores.SecurityStore
}

func NewSecurityService(securityStatService SecurityStatService, metricsService MetricService,
	securityMetricService SecurityMetricService, store stores.SecurityStore) *securityService {
	return &securityService{
		securityStatService:   securityStatService,
		metricsService:        metricsService,
		securityMetricService: securityMetricService,
		store:                 store,
	}
}

func (s *securityService) Index(ctx *gofr.Context, f *SecurityFilter, page, perPage int) ([]*Security, int, error) {
	limit := perPage
	offset := limit * (page - 1)

	filter := &stores.SecurityFilter{
		IDs:    f.IDs,
		Symbol: f.Symbol,
		ISIN:   f.ISIN,
	}

	securities, err := s.store.Index(ctx, filter, limit, offset)
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

	metrics, _, err := s.metricsService.Index(ctx, &MetricFilter{}, 0, 0)
	if err != nil {
		return nil, 0, err
	}

	var metricsMap = make(map[int]*Metric)

	for i := range metrics {
		metricsMap[metrics[i].ID] = metrics[i]
	}

	var (
		sem   = make(chan struct{}, 5)
		resp  = make([]*Security, len(securities))
		errCh = make(chan error, len(securities))
	)

	for i := range securities {
		sem <- struct{}{}
		i := i
		resp[i] = &Security{}

		go func() {
			defer func() { <-sem }()

			resp[i], err = s.buildResp(ctx, securities[i], metricsMap)
			errCh <- err
		}()
	}

	for j := 0; j < len(securities); j++ {
		if err := <-errCh; err != nil {
			return nil, 0, err
		}
	}

	return resp, count, nil
}

func (s *securityService) Read(ctx *gofr.Context, id int) (*Security, error) {
	security, err := s.store.Retrieve(ctx, id)
	if err != nil {
		return nil, err
	}

	metrics, _, err := s.metricsService.Index(ctx, &MetricFilter{}, 0, 0)
	if err != nil {
		return nil, err
	}

	var metricsMap = make(map[int]*Metric)

	for i := range metrics {
		metricsMap[metrics[i].ID] = metrics[i]
	}

	return s.buildResp(ctx, security, metricsMap)
}

func (s *securityService) Create(ctx *gofr.Context, payload *SecurityCreate) (*Security, error) {
	if payload.UserID != 1 {
		return nil, &ErrResp{Code: 403, Message: "Forbidden"}
	}

	industry, err := stores.IndustryFromString(payload.Industry)
	if err != nil {
		return nil, err
	}

	model := &stores.Security{
		ISIN:      payload.ISIN,
		Symbol:    payload.Symbol,
		Industry:  industry,
		Name:      payload.Name,
		Image:     payload.Image,
		LTP:       payload.LTP,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	security, err := s.store.Create(ctx, model)
	if err != nil {
		return nil, err
	}

	metrics, _, err := s.metricsService.Index(ctx, &MetricFilter{}, 0, 0)
	if err != nil {
		return nil, err
	}

	var metricsMap = make(map[int]*Metric)

	for i := range metrics {
		metricsMap[metrics[i].ID] = metrics[i]
	}

	return s.buildResp(ctx, security, metricsMap)
}

func (s *securityService) Patch(ctx *gofr.Context, id int, payload *SecurityUpdate) (*Security, error) {
	if payload.UserID != 1 {
		return nil, &ErrResp{Code: 403, Message: "Forbidden"}
	}

	security, err := s.store.Retrieve(ctx, id)
	if err != nil {
		return nil, err
	}

	if payload.Symbol != "" {
		security.Symbol = payload.Symbol
	}

	if payload.Industry != "" {
		security.Industry, err = stores.IndustryFromString(payload.Industry)
		if err != nil {
			return nil, err
		}
	}

	if payload.Name != "" {
		security.Name = payload.Name
	}

	if payload.Image != "" {
		security.Image = payload.Image
	}

	if payload.LTP != 0 {
		security.LTP = payload.LTP
	}

	security, err = s.store.Update(ctx, id, security)
	if err != nil {
		return nil, err
	}

	metrics, _, err := s.metricsService.Index(ctx, &MetricFilter{}, 0, 0)
	if err != nil {
		return nil, err
	}

	var metricsMap = make(map[int]*Metric)

	for i := range metrics {
		metricsMap[metrics[i].ID] = metrics[i]
	}

	return s.buildResp(ctx, security, metricsMap)
}

func (s *securityService) buildResp(ctx *gofr.Context, model *stores.Security, metricsMap map[int]*Metric) (*Security, error) {
	resp := &Security{
		ID:              model.ID,
		ISIN:            model.ISIN,
		Symbol:          model.Symbol,
		Industry:        model.Industry.String(),
		Name:            model.Name,
		Image:           model.Image,
		LTP:             model.LTP,
		CreatedAt:       model.CreatedAt,
		UpdatedAt:       model.CreatedAt,
		SecurityStat:    nil,
		SecurityMetrics: nil,
	}

	if err := s.bindSecurityStat(ctx, resp); err != nil {
		return nil, err
	}

	if err := s.bindSecurityMetricsDetails(ctx, resp, metricsMap); err != nil {
		return nil, err
	}

	s.computeAndSetNormalizedValues(resp)

	return resp, nil
}

func (s *securityService) bindSecurityStat(ctx *gofr.Context, resp *Security) error {
	securityStat, _, err := s.securityStatService.Index(ctx, &SecurityStatFilter{SecurityID: resp.ID}, 1, 0)
	if err != nil {
		return err
	}

	if len(securityStat) < 1 {
		return nil
	}

	resp.SecurityStat = &struct {
		ID         int
		SecurityID int
		Date       time.Time
		Open       float64
		Close      float64
		High       float64
		Low        float64
		Volume     int
	}{
		ID:         securityStat[0].ID,
		SecurityID: securityStat[0].SecurityID,
		Date:       securityStat[0].Date,
		Open:       securityStat[0].Open,
		Close:      securityStat[0].Close,
		High:       securityStat[0].High,
		Low:        securityStat[0].Low,
		Volume:     securityStat[0].Volume,
	}

	return nil
}

func (s *securityService) bindSecurityMetricsDetails(ctx *gofr.Context, resp *Security, metricsMap map[int]*Metric) error {
	if resp.SecurityStat == nil {
		return nil
	}

	date := resp.SecurityStat.Date

	securityMetrics, _, err := s.securityMetricService.Index(ctx, &SecurityMetricFilter{SecurityID: resp.ID, Date: date}, 0, 0)
	if err != nil {
		return err
	}

	resp.SecurityMetrics = make([]*struct {
		ID              int
		SecurityID      int
		MetricID        int
		Date            time.Time
		Value           float64
		NormalizedValue float64
		Metric          *struct {
			ID        int
			Name      string
			Type      string
			Period    int
			Indicator string
		}
	}, len(securityMetrics))

	for i := range securityMetrics {
		resp.SecurityMetrics[i] = &struct {
			ID              int
			SecurityID      int
			MetricID        int
			Date            time.Time
			Value           float64
			NormalizedValue float64
			Metric          *struct {
				ID        int
				Name      string
				Type      string
				Period    int
				Indicator string
			}
		}{
			ID:              securityMetrics[i].ID,
			SecurityID:      securityMetrics[i].SecurityID,
			MetricID:        securityMetrics[i].MetricID,
			Date:            securityMetrics[i].Date,
			Value:           securityMetrics[i].Value,
			NormalizedValue: 0,
			Metric: &struct {
				ID        int
				Name      string
				Type      string
				Period    int
				Indicator string
			}{
				ID:        securityMetrics[i].MetricID,
				Name:      metricsMap[securityMetrics[i].MetricID].Name,
				Type:      metricsMap[securityMetrics[i].MetricID].Type,
				Period:    metricsMap[securityMetrics[i].MetricID].Period,
				Indicator: metricsMap[securityMetrics[i].MetricID].Indicator,
			},
		}
	}

	return nil
}

func (s *securityService) computeAndSetNormalizedValues(resp *Security) {
	for _, metric := range resp.SecurityMetrics {
		metricType, _ := stores.MetricTypeFromString(metric.Metric.Type)

		switch metricType {
		case stores.SMA, stores.EMA:
			metric.NormalizedValue = (resp.SecurityStat.Close - metric.Value) / resp.SecurityStat.Close

		case stores.RSI:
			metric.NormalizedValue = metric.Value / 100

			if metric.Value > 70 {
				metric.NormalizedValue = 0.7 - metric.NormalizedValue
			}

			if metric.Value < 30 {
				metric.NormalizedValue = metric.NormalizedValue - 0.3
			}

		case stores.ROC:
			metric.NormalizedValue = metric.Value / 100

		case stores.ATR:
			metric.NormalizedValue = metric.Value / resp.SecurityStat.Close

		case stores.VMA:
			metric.NormalizedValue = (float64(resp.SecurityStat.Volume) - metric.Value) / float64(resp.SecurityStat.Volume)

		default:
			metric.NormalizedValue = metric.Value
		}
	}
}
