package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gofr.dev/pkg/gofr"

	dataProviders "github.com/stratifyr/security-service/crons/marketdata-populator/data-providers"
)

func main() {
	app := gofr.NewCMD()

	app.AddHTTPService("security-service", app.Config.Get("SECURITY_SERVICE_HOST"))

	client, err := dataProviders.New(app)
	if err != nil {
		fmt.Println("error:", err)

		return
	}

	h := &marketDataHandler{client: client}

	app.SubCommand("run", h.LoadMarketData)

	app.Run()
}

type marketDataHandler struct {
	client dataProviders.DataProvider
}

func (h *marketDataHandler) LoadMarketData(ctx *gofr.Context) (any, error) {
	var (
		today = time.Now().UTC().Truncate(24 * time.Hour)
		date  = today
		err   error
	)

	if ctx.Param("date") != "" {
		date, err = time.Parse(time.DateOnly, ctx.Param("date"))
		if err != nil {
			return nil, errors.New("invalid value for arg date")
		}

		if date.After(today) {
			return nil, errors.New("cannot load data for future date - " + date.Format(time.DateOnly))
		}
	}

	if date.Weekday() == time.Saturday || date.Weekday() == time.Sunday {
		return nil, errors.New("cannot load data on market holiday - " + date.Format(time.DateOnly))
	}

	securityISINs, securityIDMap, err := h.getSecurityDetails(ctx)
	if err != nil {
		return nil, err
	}

	marketData, err := h.client.GetMarketDataBulk(ctx, securityISINs, date)
	if err != nil {
		return nil, errors.New("failed to get marketData, err: " + err.Error())
	}

	for i := range marketData {
		securityID := securityIDMap[marketData[i].ISIN]

		payload := map[string]any{
			"userId":     1,
			"securityId": securityID,
			"date":       date.Format(time.DateOnly),
			"open":       marketData[i].Open,
			"close":      marketData[i].Close,
			"high":       marketData[i].High,
			"low":        marketData[i].Low,
			"volume":     marketData[i].Volume,
		}

		body, _ := json.Marshal(payload)

		resp, err := ctx.GetHTTPService("security-service").PostWithHeaders(ctx, "security-stats", nil, body, nil)
		if err != nil {
			return nil, errors.New("failed POST /security-service/security-stats, err: " + err.Error())
		}

		if resp.StatusCode != 201 {
			return nil, errors.New("non 201 resp POST /security-service/security-stats")
		}
	}

	return "successfully loaded data for " + date.Format(time.DateOnly), err
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
			return nil, nil, errors.New("failed GET /security-service/securities, err: " + err.Error())
		}

		if resp.StatusCode != 200 {
			return nil, nil, errors.New("non 200 resp GET /security-service/securities")
		}

		var res struct {
			Data []*struct {
				ID   int    `json:"id"`
				ISIN string `json:"isin"`
			} `json:"data"`
		}

		err = json.NewDecoder(resp.Body).Decode(&res)
		if err != nil {
			return nil, nil, errors.New("unexpected resp GET /security-service/securities, " + err.Error())
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
