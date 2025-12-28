package stats

import (
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"strings"
	"time"

	"gofr.dev/pkg/gofr"

	"github.com/stratifyr/security-service/cmd/data-loader/internal/data-providers"
)

func NewCMDHandler(dataProvider dataProviders.Provider) *handler {
	return &handler{dataProvider: dataProvider}
}

type handler struct {
	dataProvider dataProviders.Provider
}

func (h *handler) LoadLTP(ctx *gofr.Context) (any, error) {
	symbolFilter := ctx.Param("symbol")

	securities, securityIDMap, err := h.getSecurityDetails(ctx, symbolFilter)
	if err != nil {
		return nil, err
	}

	if symbolFilter != "" && !slices.Contains(securities, symbolFilter) {
		return nil, fmt.Errorf("security not found with symbol %v", symbolFilter)
	}

	ltpMap, err := h.dataProvider.LTP(ctx, securities)
	if err != nil {
		return nil, fmt.Errorf("failed to get ltp data, err: %v", err)
	}

	var errs []string

	for i := range securities {
		fmt.Println(securities[i])

		ltp, ok := ltpMap[securities[i]]
		if !ok {
			errs = append(errs, fmt.Sprintf("[%s] ltp data not found", securities[i]))
			continue
		}

		if err = h.updateLTP(ctx, securityIDMap[securities[i]], ltp); err != nil {
			errs = append(errs, fmt.Sprint(securities[i], err))
			continue
		}
	}

	if len(errs) > 0 {
		return "\nERRORS:", fmt.Errorf(strings.Join(errs, "\n"))
	}

	return "\nOK", nil
}

func (h *handler) LoadSecurityStats(ctx *gofr.Context) (any, error) {
	today := time.Now()

	marketDays, err := h.getMarketDays(ctx, today, today)
	if err != nil {
		return nil, err
	}

	if len(marketDays) != 1 || marketDays[0].Format(time.DateOnly) != today.Format(time.DateOnly) {
		return nil, fmt.Errorf("cannot load stats on market holiday %v", today.Format(time.DateOnly))
	}

	symbolFilter := ctx.Param("symbol")

	securities, securityIDMap, err := h.getSecurityDetails(ctx, symbolFilter)
	if err != nil {
		return nil, err
	}

	if symbolFilter != "" && !slices.Contains(securities, symbolFilter) {
		return nil, fmt.Errorf("security not found with symbol %s", symbolFilter)
	}

	ohlcData, err := h.dataProvider.OHLC(ctx, securities)
	if err != nil {
		return nil, fmt.Errorf("failed to get ohlc data, err: %v", err)
	}

	var errs []string

	for i := range securities {
		fmt.Println(securities[i])

		ohlc, ok := ohlcData[securities[i]]
		if !ok {
			errs = append(errs, fmt.Sprintf("[%s] ohlc data not found", securities[i]))
			continue
		}

		if err = h.createSecurityStat(ctx, securityIDMap[securities[i]], today, ohlc); err != nil {
			errs = append(errs, fmt.Sprintf("[%s] %v", securities[i], err))
			continue
		}
	}

	if len(errs) > 0 {
		return "\nERRORS:", fmt.Errorf(strings.Join(errs, "\n"))
	}

	return "\nOK", nil
}

func (h *handler) BackFillSecurityStats(ctx *gofr.Context) (any, error) {
	today := time.Now()
	threeYearsEarlier := today.AddDate(-2, 0, 0)
	startDate, endDate := threeYearsEarlier, today.AddDate(0, 0, -1)

	symbolFilter := ctx.Param("symbol")

	securities, securityIDMap, err := h.getSecurityDetails(ctx, ctx.Param("symbol"))
	if err != nil {
		return nil, err
	}

	if symbolFilter != "" && !slices.Contains(securities, symbolFilter) {
		return nil, fmt.Errorf("security not found with symbol %s", symbolFilter)
	}

	marketDays, err := h.getMarketDays(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}

	var errs []string

	for i := range securities {
		fmt.Println(securities[i])

		historicalData, err := h.dataProvider.HistoricalOHLC(ctx, securities[i], startDate, endDate)
		if err != nil {
			errs = append(errs, fmt.Sprintf("[%s] %v", securities[i], err))
			continue
		}

		for _, date := range marketDays {
			fmt.Sprintln(securities[i], date.Format(time.DateOnly))

			idx := slices.IndexFunc(historicalData, func(ohlc *dataProviders.HistoricalOHLC) bool {
				return ohlc.Date.Format(time.DateOnly) == date.Format(time.DateOnly)
			})

			if idx == -1 {
				errs = append(errs, fmt.Sprintf("[%s %s] historical data not found", securities[i], date.Format(time.DateOnly)))
				continue
			}

			if err = h.createSecurityStat(ctx, securityIDMap[securities[i]], date, historicalData[idx].OHLCData); err != nil {
				errs = append(errs, fmt.Sprintf("[%s %s] %v", securities[i], date.Format(time.DateOnly), err))
				continue
			}
		}
	}

	if len(errs) > 0 {
		return "\nERRORS:", fmt.Errorf(strings.Join(errs, "\n"))
	}

	return "\nOK", nil
}

func (h *handler) getSecurityDetails(ctx *gofr.Context, symbol string) ([]string, map[string]int, error) {
	var (
		securityIDMap   = make(map[string]int)
		securitySymbols = make([]string, 0)
	)

	securityService := ctx.GetHTTPService("security-service")

	for page := 1; ; page++ {
		resp, err := securityService.Get(ctx, "securities", map[string]any{"userId": 1, "symbol": symbol, "page": page, "perPage": 100})
		if err != nil {
			return nil, nil, fmt.Errorf("failed GET /security-service/securities, err: %v", err)
		}

		if resp.StatusCode != 200 {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()

			return nil, nil, fmt.Errorf("non 200 resp GET /security-service/securities, resp: %s", body)
		}

		var res struct {
			Data []*struct {
				ID     int    `json:"id"`
				Symbol string `json:"symbol"`
			} `json:"data"`
		}

		err = json.NewDecoder(resp.Body).Decode(&res)
		if err != nil {
			resp.Body.Close()

			return nil, nil, fmt.Errorf("unexpected resp GET /security-service/securities, unmarshalErr: %v" + err.Error())
		}

		resp.Body.Close()

		if len(res.Data) == 0 {
			break
		}

		for i := range res.Data {
			securitySymbols = append(securitySymbols, res.Data[i].Symbol)
			securityIDMap[res.Data[i].Symbol] = res.Data[i].ID
		}
	}

	return securitySymbols, securityIDMap, nil
}

func (h *handler) getMarketDays(ctx *gofr.Context, startDate, endDate time.Time) ([]time.Time, error) {
	securityService := ctx.GetHTTPService("security-service")

	resp, err := securityService.Get(ctx, "market-days", map[string]any{"dateBetween": fmt.Sprintf("%s,%s", startDate.Format(time.DateOnly), endDate.Format(time.DateOnly))})
	if err != nil {
		return nil, fmt.Errorf("failed GET /security-service/market-days, err: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)

		return nil, fmt.Errorf("non 200 resp GET /security-service/market-days, resp: %s", body)
	}

	var res struct {
		Data []string `json:"data"`
	}

	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return nil, fmt.Errorf("unexpected resp GET /security-service/market-days, unmarshalErr: %v", err)
	}

	var marketDays = make([]time.Time, len(res.Data))

	for i := range marketDays {
		marketDays[i], _ = time.Parse(time.DateOnly, res.Data[i])
	}

	return marketDays, nil
}

func (h *handler) updateLTP(ctx *gofr.Context, securityID int, ltp float64) error {
	payload := map[string]any{
		"userId": 1,
		"ltp":    ltp,
	}

	body, _ := json.Marshal(payload)

	resp, err := ctx.GetHTTPService("security-service").Patch(ctx, fmt.Sprintf("securities/%d", securityID), nil, body)
	if err != nil {
		return fmt.Errorf("failed PATCH /security-service/securities/%d, err: %s", securityID, err)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("non 200 resp POST /security-service/securities/%d, err: %s", securityID, err)
	}

	return nil
}

func (h *handler) createSecurityStat(ctx *gofr.Context, securityID int, date time.Time, ohlcData *dataProviders.OHLCData) error {
	payload := map[string]any{
		"userId":     1,
		"securityId": securityID,
		"date":       date.Format(time.DateOnly),
		"open":       ohlcData.Open,
		"close":      ohlcData.Close,
		"high":       ohlcData.High,
		"low":        ohlcData.Low,
		"volume":     ohlcData.Volume,
	}

	body, _ := json.Marshal(payload)

	resp, err := ctx.GetHTTPService("security-service").Post(ctx, "security-stats", nil, body)
	if err != nil {
		return fmt.Errorf("failed POST /security-service/security-stats, err: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		b, _ := io.ReadAll(resp.Body)

		return fmt.Errorf("non 201 resp POST /security-service/security-stats, resp: %s", b)
	}

	return nil
}
