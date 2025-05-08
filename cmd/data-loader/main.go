package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"slices"
	"strings"
	"time"

	"gofr.dev/pkg/gofr"

	dataProviders "github.com/stratifyr/security-service/cmd/data-loader/data-providers"
)

func main() {
	app := gofr.NewCMD()

	app.AddHTTPService("security-service", app.Config.Get("SECURITY_SERVICE_HOST"))

	client, err := dataProviders.New(app)
	if err != nil {
		log.Fatalf("failed to get data provider, err: %s", err)
	}

	istLocation, _ := time.LoadLocation("Asia/Kolkata")

	h := &marketDataHandler{
		client:         client,
		marketHolidays: strings.Split(app.Config.Get("MARKET_HOLIDAYS"), ","),
		tz:             istLocation,
	}

	app.SubCommand("load ltp", h.LoadLTPData)
	app.SubCommand("load ohlc", h.LoadOHLCData)
	app.SubCommand("load historical-ohlc", h.LoadHistoricalOHLCData)

	app.Run()
}

type marketDataHandler struct {
	client         dataProviders.DataProvider
	marketHolidays []string
	tz             *time.Location
}

func (h *marketDataHandler) LoadLTPData(ctx *gofr.Context) (any, error) {
	currentTime := time.Now().In(h.tz)

	if currentTime.Weekday() == time.Saturday || currentTime.Weekday() == time.Sunday ||
		slices.Contains(h.marketHolidays, currentTime.Format(time.DateOnly)) {
		return nil, errors.New("cannot load data on market holiday - " + currentTime.Format(time.DateOnly))
	}

	securityISINs, securityIDMap, err := h.getSecurityDetails(ctx)
	if err != nil {
		return nil, err
	}

	ltpData, err := h.client.LTPBulk(ctx, securityISINs)
	if err != nil {
		return nil, errors.New("failed to get ltpData, err: " + err.Error())
	}

	for i := range ltpData {
		securityID := securityIDMap[ltpData[i].ISIN]

		payload := map[string]any{
			"userId": 1,
			"ltp":    ltpData[i].LTP,
		}

		body, _ := json.Marshal(payload)

		resp, err := ctx.GetHTTPService("security-service").Patch(ctx, fmt.Sprintf("securities/%d", securityID), nil, body)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("failed PATCH /security-service/securities/%d, err: %s", securityID, err))
		}

		if resp.StatusCode != 200 {
			return nil, errors.New(fmt.Sprintf("non 200 resp POST /security-service/securities/%d, err: %s", securityID, err))
		}

		fmt.Println(fmt.Sprintf("[%d] successfully loaded ltp data for %s", i+1, securityISINs[i]))
	}

	return "------------\nsuccessfully loaded ltp data: " + currentTime.Format(time.DateTime), nil
}

func (h *marketDataHandler) LoadOHLCData(ctx *gofr.Context) (any, error) {
	today := time.Now().In(h.tz)

	if today.Weekday() == time.Saturday || today.Weekday() == time.Sunday || slices.Contains(h.marketHolidays, today.Format(time.DateOnly)) {
		return nil, errors.New("cannot load data on market holiday - " + today.Format(time.DateOnly))
	}

	securityISINs, securityIDMap, err := h.getSecurityDetails(ctx)
	if err != nil {
		return nil, err
	}

	ohlcData, err := h.client.OHLCBulk(ctx, securityISINs)
	if err != nil {
		return nil, errors.New("failed to get ohlcData, err: " + err.Error())
	}

	for i := range ohlcData {
		securityID := securityIDMap[ohlcData[i].ISIN]

		securityStatID, statExists, err := h.checkIfStatAlreadyExists(ctx, securityID, today)
		if err != nil {
			return nil, err
		}

		if statExists {
			if err = h.updateSecurityStat(ctx, securityStatID, today, ohlcData[i]); err != nil {
				return nil, err
			}

			fmt.Println(fmt.Sprintf("[%d] successfully updated ohlc data for %s", i+1, securityISINs[i]))

			continue
		}

		if err = h.createSecurityStat(ctx, securityID, today, ohlcData[i]); err != nil {
			return nil, err
		}

		fmt.Println(fmt.Sprintf("[%d] successfully loaded ohlc data for %s", i+1, securityISINs[i]))
	}

	return "------------\nsuccessfully loaded ohlc data for " + today.Format(time.DateOnly), nil
}

func (h *marketDataHandler) LoadHistoricalOHLCData(ctx *gofr.Context) (any, error) {
	startDate, err := time.Parse(time.DateOnly, ctx.Param("start-date"))
	if err != nil {
		return nil, errors.New("invalid start-date")
	}

	endDate, err := time.Parse(time.DateOnly, ctx.Param("end-date"))
	if err != nil {
		return nil, errors.New("invalid end-date")
	}

	if endDate.Before(startDate) {
		return nil, errors.New("invalid date range")
	}

	if endDate.Sub(startDate) > 365*(24*time.Hour) {
		return nil, errors.New("date range is too long, please pass interval within a year")
	}

	securityISINs, securityIDMap, err := h.getSecurityDetails(ctx)
	if err != nil {
		return nil, err
	}

	for i := range securityISINs {
		historicalData, err := h.client.HistoricalOHLC(ctx, securityISINs[i], startDate, endDate)
		if err != nil {
			fmt.Println(fmt.Sprintf("skipping: %s, failed to get historical data, err: %s", securityISINs[i], err))

			continue
		}

		securityID := securityIDMap[securityISINs[i]]

		for _, data := range historicalData {
			securityStatID, statExists, err := h.checkIfStatAlreadyExists(ctx, securityID, data.Date)
			if err != nil {
				return nil, err
			}

			if statExists {
				if err = h.updateSecurityStat(ctx, securityStatID, data.Date, data.OHLCData); err != nil {
					return nil, err
				}

				continue
			}

			if err = h.createSecurityStat(ctx, securityID, data.Date, data.OHLCData); err != nil {
				return nil, err
			}
		}

		fmt.Println(fmt.Sprintf("[%d] successfully loaded ohlc data for %s", i+1, securityISINs[i]))
	}

	return fmt.Println(fmt.Sprintf("------------\nsuccessfully loaded ohlc data for interval %s-%s", startDate.Format(time.DateOnly), endDate.Format(time.DateOnly)))
}

func (h *marketDataHandler) getSecurityDetails(ctx *gofr.Context) ([]string, map[string]int, error) {
	var (
		securityIDMap = make(map[string]int)
		securityISINs = make([]string, 0)
	)

	securityService := ctx.GetHTTPService("security-service")

	for page := 1; ; page++ {
		resp, err := securityService.Get(ctx, "securities", map[string]any{"userId": 1, "page": page, "perPage": 100})
		if err != nil {
			return nil, nil, errors.New(fmt.Sprintf("failed GET /security-service/securities?userId=1&page=%d&perPage=100, err: %s", page, err))
		}

		if resp.StatusCode != 200 {
			return nil, nil, errors.New(fmt.Sprintf("non 200 resp GET /security-service/securities?userId=1&page=%d&perPage=100", page))
		}

		var res struct {
			Data []*struct {
				ID   int    `json:"id"`
				ISIN string `json:"isin"`
			} `json:"data"`
		}

		err = json.NewDecoder(resp.Body).Decode(&res)
		if err != nil {
			return nil, nil, errors.New(fmt.Sprintf("unexpected resp GET /security-service/securities?userId=1&page=%d&perPage=100,"+
				" unmarshalErr: %s", page, err))
		}

		resp.Body.Close()

		if len(res.Data) == 0 {
			break
		}

		for i := range res.Data {
			securityISINs = append(securityISINs, res.Data[i].ISIN)
			securityIDMap[res.Data[i].ISIN] = res.Data[i].ID
		}
	}

	return securityISINs, securityIDMap, nil
}

func (h *marketDataHandler) checkIfStatAlreadyExists(ctx *gofr.Context, securityID int, date time.Time) (int, bool, error) {
	securityService := ctx.GetHTTPService("security-service")

	resp, err := securityService.Get(ctx, "security-stats", map[string]any{"securityId": securityID, "date": date.Format(time.DateOnly)})
	if err != nil {
		return 0, false, errors.New(fmt.Sprintf("failed GET /security-service/security-stats?securityId=%d&date=%s,"+
			" err: %s", securityID, date.Format(time.DateOnly), err))
	}

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)

		return 0, false, errors.New(fmt.Sprintf("non 200 resp GET /security-service/security-stats?securityId=%d&date=%s,"+
			" resp: %s", securityID, date.Format(time.DateOnly), string(body)))
	}

	var res struct {
		Data []*struct {
			ID int `json:"id"`
		} `json:"data"`
	}

	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return 0, false, errors.New(fmt.Sprintf("unexpected resp GET /security-service/security-stats?securityId=%d&date=%s,"+
			" unmarshallErr: %s", securityID, date.Format(time.DateOnly), err))
	}

	if len(res.Data) > 0 {
		return res.Data[0].ID, true, nil
	}

	return 0, false, nil
}

func (h *marketDataHandler) updateSecurityStat(ctx *gofr.Context, securityStatID int, date time.Time, ohlcData *dataProviders.OHLCData) error {
	payload := map[string]any{
		"userId": 1,
		"date":   date.Format(time.DateOnly),
		"open":   ohlcData.Open,
		"close":  ohlcData.Close,
		"high":   ohlcData.High,
		"low":    ohlcData.Low,
		"volume": ohlcData.Volume,
	}

	body, _ := json.Marshal(payload)

	resp, err := ctx.GetHTTPService("security-service").Patch(ctx, fmt.Sprintf("security-stats/%d", securityStatID), nil, body)
	if err != nil {
		return errors.New(fmt.Sprintf("failed PATCH /security-service/security-stats/%d, err: %s", securityStatID, err))
	}

	if resp.StatusCode != 200 {
		return errors.New(fmt.Sprintf("non 200 resp POST /security-service/security-stats/%d, err: %s", securityStatID, err))
	}

	return nil
}

func (h *marketDataHandler) createSecurityStat(ctx *gofr.Context, securityID int, date time.Time, ohlcData *dataProviders.OHLCData) error {
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
		return errors.New("failed POST /security-service/security-stats, err: " + err.Error())
	}

	if resp.StatusCode != 201 {
		return errors.New("non 201 resp POST /security-service/security-stats")
	}

	return nil
}
