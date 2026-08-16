package jobprocessors

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"gofr.dev/pkg/gofr"

	dataProviders "github.com/stratifyr/security-service/daemon/stats-loader/internal/data-providers"
)

type MarketDataJob struct {
	ID   int    `json:"id"`
	Type string `json:"type"`
}

func GetNextPendingJob(ctx *gofr.Context) (*MarketDataJob, error) {
	securityService := ctx.GetHTTPService("security-service")

	resp, err := securityService.Get(ctx, "market-data-jobs", map[string]any{"userId": 1, "status": "CREATED", "page": 1, "perPage": 1})
	if err != nil {
		return nil, fmt.Errorf("failed GET /security-service/market-data-jobs, err: %v", err)
	}

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		return nil, fmt.Errorf("non 200 resp GET /security-service/market-data-jobs, resp: %s", body)
	}

	var res struct {
		Data []*MarketDataJob `json:"data"`
	}

	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		resp.Body.Close()

		return nil, fmt.Errorf("unexpected resp GET /security-service/market-data-jobs, unmarshalErr: %v" + err.Error())
	}

	resp.Body.Close()

	if len(res.Data) == 0 {
		return nil, nil
	}

	return res.Data[0], nil
}

func UpdateJobStatus(ctx *gofr.Context, jobID int, status string, logs *Logs) error {
	payload := map[string]any{
		"userId": 1,
		"status": status,
		"logs":   logs,
	}

	body, _ := json.Marshal(payload)

	resp, err := ctx.GetHTTPService("security-service").Patch(ctx, fmt.Sprintf("market-data-jobs/%d", jobID), nil, body)
	if err != nil {
		return fmt.Errorf("failed PATCH /security-service/market-data-jobs/%d, err: %s", jobID, err)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("non 200 resp PATCH /security-service/market-data-jobs/%d, err: %s", jobID, err)
	}

	return nil
}

func getSecurityDetails(ctx *gofr.Context, symbol string) ([]string, map[string]int, error) {
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

func getMarketDays(ctx *gofr.Context, startDate, endDate time.Time) ([]time.Time, error) {
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

func updateLTP(ctx *gofr.Context, securityID int, ltp float64) error {
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

func createSecurityStat(ctx *gofr.Context, securityID int, date time.Time, ohlcData *dataProviders.OHLCData) error {
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

func initializeJobLogs(jobName string) *Logs {
	return &Logs{
		Job: strings.ToLower(jobName),
		Meta: map[string]string{
			"start_time": time.Now().Format(time.DateTime),
		},
	}
}

func recordJobCompletionLogs(logs *Logs, err error) {
	if err != nil {
		logs.Errors = append(logs.Errors, err.Error())
	}

	logs.Meta["end_time"] = time.Now().Format(time.DateTime)
}
