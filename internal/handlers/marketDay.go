package handlers

import (
	"strconv"
	"strings"
	"time"

	"gofr.dev/pkg/gofr"
	"gofr.dev/pkg/gofr/http"
	"gofr.dev/pkg/gofr/http/response"

	"github.com/stratifyr/security-service/internal/services"
)

type marketDayHandler struct {
	svc services.MarketDayService
}

func NewMarketDayHandler(svc services.MarketDayService) *marketDayHandler {
	return &marketDayHandler{svc: svc}
}

func (h *marketDayHandler) Index(ctx *gofr.Context) (interface{}, error) {
	var (
		filter services.MarketDayFilter
		err    error
	)

	if ctx.Param("lastNDays") != "" {
		filter.LastNDays, err = strconv.Atoi(ctx.Param("lastNDays"))
		if err != nil {
			return nil, http.ErrorInvalidParam{Params: []string{"lastNDays"}}
		}
	}

	if ctx.Param("dateBetween") != "" {
		dates := strings.Split(ctx.Param("dateBetween"), ",")
		if len(dates) != 2 {
			return nil, http.ErrorInvalidParam{Params: []string{"dateBetween"}}
		}

		filter.DateBetween = &struct {
			StartDate time.Time
			EndDate   time.Time
		}{}

		filter.DateBetween.StartDate, err = time.Parse(time.DateOnly, dates[0])
		if err != nil {
			return nil, http.ErrorInvalidParam{Params: []string{"dateBetween"}}
		}

		filter.DateBetween.EndDate, err = time.Parse(time.DateOnly, dates[1])
		if err != nil {
			return nil, http.ErrorInvalidParam{Params: []string{"dateBetween"}}
		}
	}

	marketDays, count, err := h.svc.Index(ctx, &filter)
	if err != nil {
		return nil, err
	}

	var resp = make([]string, len(marketDays))

	for i := range marketDays {
		resp[i] = marketDays[i].Format(time.DateOnly)
	}

	return response.Raw{Data: map[string]any{
		"data": resp,
		"meta": map[string]any{
			"total": count,
		},
	}}, nil
}
