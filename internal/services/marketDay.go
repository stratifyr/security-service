package services

import (
	"slices"
	"time"

	"gofr.dev/pkg/gofr"
	"gofr.dev/pkg/gofr/http"

	"github.com/stratifyr/security-service/internal/stores"
)

type MarketDayService interface {
	Index(ctx *gofr.Context, f *MarketDayFilter) ([]time.Time, int, error)
}

type MarketDayFilter struct {
	LastNDays              int
	LastNDaysFromReference *struct {
		N         int
		Reference time.Time
	}
	DateBetween *struct {
		StartDate time.Time
		EndDate   time.Time
	}
}

type marketDayService struct {
	marketHolidayStore stores.MarketHolidayStore
}

func NewMarketDayService(marketHolidayStore stores.MarketHolidayStore) *marketDayService {
	return &marketDayService{marketHolidayStore: marketHolidayStore}
}

func (s *marketDayService) Index(ctx *gofr.Context, f *MarketDayFilter) ([]time.Time, int, error) {
	var (
		startDate time.Time
		endDate   time.Time
		n         int
	)

	switch {
	case f.LastNDays > 0:
		endDate = time.Now().UTC()
		n = f.LastNDays
	case f.LastNDaysFromReference != nil:
		endDate = f.LastNDaysFromReference.Reference
		n = f.LastNDaysFromReference.N
	case f.DateBetween != nil:
		startDate = f.DateBetween.StartDate
		endDate = f.DateBetween.EndDate
		n = int(endDate.Sub(startDate).Hours()/24) + 1 //nolint:mnd // hours in a day
	default:
		return nil, 0, http.ErrorMissingParam{Params: []string{"lastNDays", "dateBetween"}}
	}

	if n > 366*5 {
		return nil, 0, ErrDateRangeTooLong
	}

	if startDate.IsZero() {
		lookBackDays := max(n*2, 14) //nolint:mnd // lookback multiplier
		startDate = endDate.Add(-time.Duration(lookBackDays) * 24 * time.Hour)
	}

	marketHolidays, err := s.marketHolidayStore.Index(ctx, &stores.MarketHolidayFilter{
		DateBetween: &struct {
			StartDate time.Time
			EndDate   time.Time
		}{StartDate: startDate, EndDate: endDate},
	}, 0, 0)
	if err != nil {
		return nil, 0, err
	}

	var marketDays []time.Time

	for date := endDate; len(marketDays) < n && date.Unix() >= startDate.Unix(); date = date.Add(-24 * time.Hour) {
		if date.Weekday() == time.Saturday || date.Weekday() == time.Sunday {
			continue
		}

		isHoliday := slices.ContainsFunc(marketHolidays, func(marketHoliday *stores.MarketHoliday) bool {
			return marketHoliday.Date.Format(time.DateOnly) == date.Format(time.DateOnly)
		})

		if isHoliday {
			continue
		}

		marketDays = append(marketDays, date)
	}

	return marketDays, len(marketDays), nil
}
