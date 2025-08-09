package main

import (
	_ "embed"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"slices"
	"strconv"
	"strings"
	"time"

	"gofr.dev/pkg/gofr"

	dataProviders "github.com/stratifyr/security-service/cmd/data-loader/data-providers"
)

//go:embed data/securities-master.csv
var securitesMaster string

//go:embed data/metrics-master.csv
var metricsMaster string

//go:embed data/market-holidays.csv
var marketHolidaysMaster string

func main() {
	app := gofr.NewCMD()

	app.AddHTTPService("security-service", app.Config.Get("SECURITY_SERVICE_HOST"))

	client, err := dataProviders.New(app)
	if err != nil {
		log.Fatalf("failed to get data provider, err: %s", err)
	}

	istLocation, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		log.Fatalf("failed to load timezone: %v", err)
	}

	h := &marketDataHandler{
		client: client,
		tz:     istLocation,
	}

	app.SubCommand("load securities", h.LoadSecurities)
	app.SubCommand("load metrics", h.LoadMetrics)
	app.SubCommand("load market-holidays", h.LoadMarketHolidays)
	app.SubCommand("load ltp", h.LoadLTP)
	app.SubCommand("load security-stats", h.LoadSecurityStats)
	app.SubCommand("load security-metrics", h.LoadSecurityMetrics)

	app.Run()
}

type marketDataHandler struct {
	client dataProviders.DataProvider
	tz     *time.Location
}

func (h *marketDataHandler) LoadSecurities(ctx *gofr.Context) (any, error) {
	reader := csv.NewReader(strings.NewReader(securitesMaster))

	headers, err := reader.Read()
	if err != nil {
		return nil, errors.New("failed to read securitiesMasterFile headers")
	}

	idxISIN := slices.Index(headers, "ISIN Code")
	idxSymbol := slices.Index(headers, "Symbol")
	idxIndustry := slices.Index(headers, "Industry")
	idxName := slices.Index(headers, "Company Name")
	idxTier := slices.Index(headers, "Tier")

	for {
		row, readErr := reader.Read()
		if readErr == io.EOF {
			break
		}

		if readErr != nil {
			return nil, errors.New("failed to read securitiesMasterFile row")
		}

		if err = h.createOrUpdateSecurity(ctx, row[idxISIN], row[idxSymbol], row[idxIndustry], row[idxName], row[idxTier]); err != nil {
			fmt.Println(fmt.Sprintf("-[%s] fail, %s", row[idxISIN], err))
			continue
		}

		fmt.Println(fmt.Sprintf("-[%s] success", row[idxISIN]))
	}

	return "\nsuccessfully loaded securities", nil
}

func (h *marketDataHandler) LoadMetrics(ctx *gofr.Context) (any, error) {
	reader := csv.NewReader(strings.NewReader(metricsMaster))

	headers, err := reader.Read()
	if err != nil {
		return nil, errors.New("failed to read metricsMasterFile headers")
	}

	idxName := slices.Index(headers, "Name")
	idxType := slices.Index(headers, "Type")
	idxPeriod := slices.Index(headers, "Period")
	idxTier := slices.Index(headers, "Tier")

	for {
		row, readErr := reader.Read()
		if readErr == io.EOF {
			break
		}

		if readErr != nil {
			return nil, errors.New("failed to read metricsMasterFile row")
		}

		if err = h.createOrUpdateMetric(ctx, row[idxName], row[idxType], row[idxPeriod], row[idxTier]); err != nil {
			fmt.Println(fmt.Sprintf("-[%s] fail, %s", row[idxName], err))
			continue
		}

		fmt.Println(fmt.Sprintf("-[%s] success", row[idxName]))
	}

	return "\nsuccessfully loaded metrics", nil
}

func (h *marketDataHandler) LoadMarketHolidays(ctx *gofr.Context) (any, error) {
	reader := csv.NewReader(strings.NewReader(marketHolidaysMaster))

	headers, err := reader.Read()
	if err != nil {
		return nil, errors.New("failed to read marketHolidaysMasterFile headers")
	}

	idxDate := slices.Index(headers, "Date")
	idxDescription := slices.Index(headers, "Description")

	for {
		row, readErr := reader.Read()
		if readErr == io.EOF {
			break
		}

		if readErr != nil {
			return nil, errors.New("failed to read marketHolidaysMasterFile row")
		}

		if err = h.createOrUpdateMarketHolidays(ctx, row[idxDate], row[idxDescription]); err != nil {
			fmt.Println(fmt.Sprintf("-[%s] fail, %s", row[idxDate], err))
			continue
		}

		fmt.Println(fmt.Sprintf("-[%s] success", row[idxDate]))
	}

	return "\nsuccessfully loaded market-holidays", nil
}

func (h *marketDataHandler) LoadLTP(ctx *gofr.Context) (any, error) {
	currentTime := time.Now().In(h.tz)
	isinFilter := ctx.Param("isin")

	securityISINs, securityIDMap, err := h.getSecurityDetails(ctx, isinFilter)
	if err != nil {
		return nil, err
	}

	if isinFilter != "" && !slices.Contains(securityISINs, isinFilter) {
		return nil, errors.New("security not found with isin - " + isinFilter)
	}

	ltpData, err := h.client.LTPBulk(ctx, securityISINs)
	if err != nil {
		return nil, errors.New("failed to get ltpData, err: " + err.Error())
	}

	for i := range securityISINs {
		securityID := securityIDMap[securityISINs[i]]

		idx := slices.IndexFunc(ltpData, func(data *dataProviders.LTPData) bool {
			return data.ISIN == securityISINs[i]
		})

		if idx == -1 {
			fmt.Println(fmt.Sprintf("-[%s] fail, ltp data not found", securityISINs[i]))
			continue
		}

		data := ltpData[idx]

		if err = h.updateLTP(ctx, securityID, data.LTP); err != nil {
			fmt.Println(fmt.Sprintf("-[%s] fail, %s", securityISINs[i], err))
			continue
		}

		fmt.Println(fmt.Sprintf("-[%s] success", securityISINs[i]))
	}

	return "\nsuccessfully loaded ltp data @ " + currentTime.Format(time.DateTime), nil
}

func (h *marketDataHandler) LoadSecurityStats(ctx *gofr.Context) (any, error) {
	if ctx.Param("start-date") == "" && ctx.Param("end-date") == "" {
		return h.LoadTodaysSecurityStats(ctx)
	}

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

	if endDate.Sub(startDate) > 366*(24*time.Hour) {
		return nil, errors.New("date range is too long, please pass interval within a year")
	}

	isinFilter := ctx.Param("isin")

	securityISINs, securityIDMap, err := h.getSecurityDetails(ctx, ctx.Param("isin"))
	if err != nil {
		return nil, err
	}

	if isinFilter != "" && !slices.Contains(securityISINs, isinFilter) {
		return nil, errors.New("security not found with isin - " + isinFilter)
	}

	marketDays, err := h.getMarketDays(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}

	for i := range securityISINs {
		securityID := securityIDMap[securityISINs[i]]

		historicalData, err := h.client.HistoricalOHLC(ctx, securityISINs[i], startDate, endDate)
		if err != nil {
			fmt.Println(fmt.Sprintf("-[%s] fail, %s", securityISINs[i], err))
			continue
		}

		for _, date := range marketDays {
			idx := slices.IndexFunc(historicalData, func(ohlc *dataProviders.HistoricalOHLC) bool {
				return ohlc.Date.Format(time.DateOnly) == date.Format(time.DateOnly)
			})

			if idx == -1 {
				fmt.Println(fmt.Sprintf("--[%s][%s] fail, historical data not found", securityISINs[i], date.Format(time.DateOnly)))
				continue
			}

			if err = h.createOrUpdateSecurityStat(ctx, securityID, date, historicalData[idx].OHLCData); err != nil {
				fmt.Println(fmt.Sprintf("--[%s][%s] fail, %s", securityISINs[i], date.Format(time.DateOnly), err))
				continue
			}

			fmt.Println(fmt.Sprintf("--[%s][%s] success", securityISINs[i], date.Format(time.DateOnly)))
		}

		fmt.Println(fmt.Sprintf("-[%s] success", securityISINs[i]))
	}

	return fmt.Println(fmt.Sprintf("\nsuccessfully loaded ohlc data for interval %s to %s", startDate.Format(time.DateOnly), endDate.Format(time.DateOnly)))
}

func (h *marketDataHandler) LoadSecurityMetrics(ctx *gofr.Context) (any, error) {
	if ctx.Param("start-date") == "" || ctx.Param("end-date") == "" {
		return h.LoadTodaysSecurityMetrics(ctx)
	}

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

	if endDate.Sub(startDate) > 366*(24*time.Hour) {
		return nil, errors.New("date range is too long, please pass interval within a year")
	}

	isinFilter := ctx.Param("isin")

	securityISINs, securityIDMap, err := h.getSecurityDetails(ctx, ctx.Param("isin"))
	if err != nil {
		return nil, err
	}

	if isinFilter != "" && !slices.Contains(securityISINs, isinFilter) {
		return nil, errors.New("security not found with isin - " + isinFilter)
	}

	metricIDs, metricNames, err := h.getMetricIDs(ctx)
	if err != nil {
		return nil, err
	}

	marketDays, err := h.getMarketDays(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}

	for i := range securityISINs {
		securityID := securityIDMap[securityISINs[i]]

		for j := range metricIDs {
			metricID := metricIDs[j]

			for _, date := range marketDays {
				if err = h.createOrUpdateSecurityMetric(ctx, securityID, metricID, date); err != nil {
					fmt.Println(fmt.Sprintf("---[%s][%s][%s] fail, %s", securityISINs[i], metricNames[metricID], date.Format(time.DateOnly), err))
					continue
				}

				fmt.Println(fmt.Sprintf("---[%s][%s][%s] success", securityISINs[i], metricNames[metricID], date.Format(time.DateOnly)))
			}

			fmt.Println(fmt.Sprintf("--[%s][%s] success", securityISINs[i], metricNames[metricID]))
		}

		fmt.Println(fmt.Sprintf("-[%s] success", securityISINs[i]))
	}

	return fmt.Println(fmt.Sprintf("\nsuccessfully loaded security metrics data for interval %s to %s", startDate.Format(time.DateOnly), endDate.Format(time.DateOnly)))
}

func (h *marketDataHandler) LoadTodaysSecurityStats(ctx *gofr.Context) (any, error) {
	today := time.Now().In(h.tz)

	marketDays, err := h.getMarketDays(ctx, today, today)
	if err != nil {
		return nil, err
	}

	if len(marketDays) != 1 || marketDays[0].Format(time.DateOnly) != today.Format(time.DateOnly) {
		return nil, errors.New("cannot load data on market holiday - " + today.Format(time.DateOnly))
	}

	isinFilter := ctx.Param("isin")

	securityISINs, securityIDMap, err := h.getSecurityDetails(ctx, isinFilter)
	if err != nil {
		return nil, err
	}

	if isinFilter != "" && !slices.Contains(securityISINs, isinFilter) {
		return nil, errors.New("security not found with isin - " + isinFilter)
	}

	ohlcData, err := h.client.OHLCBulk(ctx, securityISINs)
	if err != nil {
		return nil, errors.New("failed to get ohlcData, err: " + err.Error())
	}

	for i := range securityISINs {
		securityID := securityIDMap[securityISINs[i]]

		idx := slices.IndexFunc(ohlcData, func(data *dataProviders.OHLCData) bool {
			return data.ISIN == securityISINs[i]
		})

		if idx == -1 {
			fmt.Println(fmt.Sprintf("-[%s] fail, ltp data not found", securityISINs[i]))
			continue
		}

		if err = h.createOrUpdateSecurityStat(ctx, securityID, today, ohlcData[idx]); err != nil {
			fmt.Println(fmt.Sprintf("-[%s] fail, %s", securityISINs[i], err))
			continue
		}

		fmt.Println(fmt.Sprintf("-[%s] success", securityISINs[i]))
	}

	return "\nsuccessfully loaded ohlc data @ " + today.Format(time.DateOnly), nil
}

func (h *marketDataHandler) LoadTodaysSecurityMetrics(ctx *gofr.Context) (any, error) {
	today := time.Now().Truncate(-24 * time.Hour)

	marketDays, err := h.getMarketDays(ctx, today, today)
	if err != nil {
		return nil, err
	}

	if len(marketDays) != 1 || marketDays[0].Format(time.DateOnly) != today.Format(time.DateOnly) {
		return nil, errors.New("cannot load data on market holiday - " + today.Format(time.DateOnly))
	}

	isinFilter := ctx.Param("isin")

	securityISINs, securityIDMap, err := h.getSecurityDetails(ctx, isinFilter)
	if err != nil {
		return nil, err
	}

	if isinFilter != "" && !slices.Contains(securityISINs, isinFilter) {
		return nil, errors.New("security not found with isin - " + isinFilter)
	}

	metricIDs, metricsNames, err := h.getMetricIDs(ctx)
	if err != nil {
		return nil, err
	}

	for i := range securityISINs {
		securityID := securityIDMap[securityISINs[i]]

		for j := range metricIDs {
			metricID := metricIDs[j]

			if err = h.createSecurityMetric(ctx, securityID, metricID, today); err != nil {
				fmt.Println(fmt.Sprintf("--[%s][%s] fail, %s", securityISINs[i], metricsNames[metricID], err))
				continue
			}

			fmt.Println(fmt.Sprintf("--[%s][%s] success", securityISINs[i], metricsNames[metricID]))
		}

		fmt.Println(fmt.Sprintf("-[%s] success", securityISINs[i]))
	}

	return "\nsuccessfully loaded security metrics data @ " + today.Format(time.DateTime), nil
}

func (h *marketDataHandler) createOrUpdateSecurity(ctx *gofr.Context, ISIN, symbol, industry, name, tier string) error {
	tierInt, _ := strconv.Atoi(tier)

	securityID, exists, err := h.checkIfSecurityAlreadyExists(ctx, ISIN)
	if err != nil {
		return err
	}

	if exists {
		if err = h.updateSecurity(ctx, securityID, ISIN, symbol, industry, name, tierInt); err != nil {
			return err
		}

		return nil
	}

	if err = h.createSecurity(ctx, ISIN, symbol, industry, name, tierInt); err != nil {
		return err
	}

	return nil
}

func (h *marketDataHandler) checkIfSecurityAlreadyExists(ctx *gofr.Context, ISIN string) (int, bool, error) {
	securityService := ctx.GetHTTPService("security-service")

	resp, err := securityService.Get(ctx, "securities", map[string]any{"isin": ISIN})
	if err != nil {
		return 0, false, errors.New("failed GET /security-service/securities, err: " + err.Error())
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)

		return 0, false, errors.New("non 200 resp GET /security-service/securities, resp: " + string(body))
	}

	var res struct {
		Data []*struct {
			ID int `json:"id"`
		} `json:"data"`
	}

	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return 0, false, errors.New("unexpected resp GET /security-service/securities, unmarshallErr: " + err.Error())
	}

	if len(res.Data) > 0 {
		return res.Data[0].ID, true, nil
	}

	return 0, false, nil
}

func (h *marketDataHandler) updateSecurity(ctx *gofr.Context, securityID int, ISIN, symbol, industry, name string, tier int) error {
	payload := map[string]any{
		"userId":   1,
		"isin":     ISIN,
		"symbol":   symbol,
		"industry": industry,
		"name":     name,
		"tier":     tier,
	}

	body, _ := json.Marshal(payload)

	resp, err := ctx.GetHTTPService("security-service").Patch(ctx, fmt.Sprintf("securities/%d", securityID), nil, body)
	if err != nil {
		return errors.New(fmt.Sprintf("failed PATCH /security-service/securities/%d, err: %s", securityID, err))
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)

		return errors.New(fmt.Sprintf("non 200 resp PATCH /security-service/securities/%d, resp: %s", securityID, string(b)))
	}

	return nil
}

func (h *marketDataHandler) createSecurity(ctx *gofr.Context, ISIN, symbol, industry, name string, tier int) error {
	payload := map[string]any{
		"userId":   1,
		"isin":     ISIN,
		"symbol":   symbol,
		"industry": industry,
		"name":     name,
		"tier":     tier,
	}

	body, _ := json.Marshal(payload)

	resp, err := ctx.GetHTTPService("security-service").Post(ctx, "securities", nil, body)
	if err != nil {
		return errors.New("failed POST /security-service/securities, err: " + err.Error())
	}

	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		b, _ := io.ReadAll(resp.Body)

		return errors.New("non 201 resp POST /security-service/securities, resp: " + string(b))
	}

	return nil
}

func (h *marketDataHandler) createOrUpdateMetric(ctx *gofr.Context, name, typ, period, tier string) error {
	interval, _ := strconv.Atoi(period)
	tierInt, _ := strconv.Atoi(tier)

	metricID, exists, err := h.checkIfMetricAlreadyExists(ctx, typ, interval)
	if err != nil {
		return err
	}

	if exists {
		if err = h.updateMetric(ctx, metricID, name, tierInt); err != nil {
			return err
		}

		return nil
	}

	if err = h.createMetric(ctx, name, typ, interval, tierInt); err != nil {
		return err
	}

	return nil
}

func (h *marketDataHandler) checkIfMetricAlreadyExists(ctx *gofr.Context, typ string, period int) (int, bool, error) {
	securityService := ctx.GetHTTPService("security-service")

	resp, err := securityService.Get(ctx, "metrics", map[string]any{"type": typ, "period": period})
	if err != nil {
		return 0, false, errors.New("failed GET /security-service/metrics, err: " + err.Error())
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)

		return 0, false, errors.New("non 200 resp GET /security-service/metrics, resp: " + string(body))
	}

	var res struct {
		Data []*struct {
			ID int `json:"id"`
		} `json:"data"`
	}

	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return 0, false, errors.New("unexpected resp GET /security-service/metrics, unmarshallErr: " + err.Error())
	}

	if len(res.Data) > 0 {
		return res.Data[0].ID, true, nil
	}

	return 0, false, nil
}

func (h *marketDataHandler) updateMetric(ctx *gofr.Context, metricID int, name string, tier int) error {
	payload := map[string]any{
		"userId": 1,
		"name":   name,
		"tier":   tier,
	}

	body, _ := json.Marshal(payload)

	resp, err := ctx.GetHTTPService("security-service").Patch(ctx, fmt.Sprintf("metrics/%d", metricID), nil, body)
	if err != nil {
		return errors.New(fmt.Sprintf("failed PATCH /security-service/metrics/%d, err: %s", metricID, err))
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)

		return errors.New(fmt.Sprintf("non 200 resp PATCH /security-service/metrics/%d, resp: %s", metricID, string(b)))
	}

	return nil
}

func (h *marketDataHandler) createMetric(ctx *gofr.Context, name, typ string, period, tier int) error {
	payload := map[string]any{
		"userId": 1,
		"name":   name,
		"type":   typ,
		"period": period,
		"tier":   tier,
	}

	body, _ := json.Marshal(payload)

	resp, err := ctx.GetHTTPService("security-service").Post(ctx, "metrics", nil, body)
	if err != nil {
		return errors.New("failed POST /security-service/metrics, err: " + err.Error())
	}

	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		b, _ := io.ReadAll(resp.Body)

		return errors.New("non 201 resp POST /security-service/metrics, resp: " + string(b))
	}

	return nil
}

func (h *marketDataHandler) createOrUpdateMarketHolidays(ctx *gofr.Context, date, description string) error {
	marketHolidayID, exists, err := h.checkIfMarketHolidayAlreadyExists(ctx, date)
	if err != nil {
		return err
	}

	if exists {
		if err = h.updateMarketHoliday(ctx, marketHolidayID, description); err != nil {
			return err
		}

		return nil
	}

	if err = h.createMarketHoliday(ctx, date, description); err != nil {
		return err
	}

	return nil
}

func (h *marketDataHandler) checkIfMarketHolidayAlreadyExists(ctx *gofr.Context, date string) (int, bool, error) {
	securityService := ctx.GetHTTPService("security-service")

	resp, err := securityService.Get(ctx, "market-holidays", map[string]any{"date": date})
	if err != nil {
		return 0, false, errors.New("failed GET /security-service/market-holidays, err: " + err.Error())
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)

		return 0, false, errors.New("non 200 resp GET /security-service/market-holidays, resp: " + string(body))
	}

	var res struct {
		Data []*struct {
			ID int `json:"id"`
		} `json:"data"`
	}

	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return 0, false, errors.New("unexpected resp GET /security-service/market-holidays, unmarshallErr: " + err.Error())
	}

	if len(res.Data) > 0 {
		return res.Data[0].ID, true, nil
	}

	return 0, false, nil
}

func (h *marketDataHandler) updateMarketHoliday(ctx *gofr.Context, marketHolidayID int, description string) error {
	payload := map[string]any{
		"userId":      1,
		"description": description,
	}

	body, _ := json.Marshal(payload)

	resp, err := ctx.GetHTTPService("security-service").Patch(ctx, fmt.Sprintf("market-holidays/%d", marketHolidayID), nil, body)
	if err != nil {
		return errors.New(fmt.Sprintf("failed PATCH /security-service/market-holidays/%d, err: %s", marketHolidayID, err))
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)

		return errors.New(fmt.Sprintf("non 200 resp PATCH /security-service/market-holidays/%d, resp: %s", marketHolidayID, string(b)))
	}

	return nil
}

func (h *marketDataHandler) createMarketHoliday(ctx *gofr.Context, date, description string) error {
	payload := map[string]any{
		"userId":      1,
		"date":        date,
		"description": description,
	}

	body, _ := json.Marshal(payload)

	resp, err := ctx.GetHTTPService("security-service").Post(ctx, "market-holidays", nil, body)
	if err != nil {
		return errors.New("failed POST /security-service/market-holidays, err: " + err.Error())
	}

	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		b, _ := io.ReadAll(resp.Body)

		return errors.New("non 201 resp POST /security-service/market-holidays, resp: " + string(b))
	}

	return nil
}

func (h *marketDataHandler) getMarketDays(ctx *gofr.Context, startDate, endDate time.Time) ([]time.Time, error) {
	securityService := ctx.GetHTTPService("security-service")

	resp, err := securityService.Get(ctx, "market-days", map[string]any{"dateBetween": fmt.Sprintf("%s,%s", startDate.Format(time.DateOnly), endDate.Format(time.DateOnly))})
	if err != nil {
		return nil, errors.New("failed GET /security-service/market-days, err: " + err.Error())
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)

		return nil, errors.New("non 200 resp GET /security-service/market-days, resp: " + string(body))
	}

	var res struct {
		Data []string `json:"data"`
	}

	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return nil, errors.New("unexpected resp GET /security-service/market-days, unmarshalErr: " + err.Error())
	}

	var marketDays = make([]time.Time, len(res.Data))

	for i := range marketDays {
		marketDays[i], _ = time.Parse(time.DateOnly, res.Data[i])
	}

	return marketDays, nil
}

func (h *marketDataHandler) getSecurityDetails(ctx *gofr.Context, ISIN string) ([]string, map[string]int, error) {
	var (
		securityIDMap = make(map[string]int)
		securityISINs = make([]string, 0)
	)

	securityService := ctx.GetHTTPService("security-service")

	for page := 1; ; page++ {
		resp, err := securityService.Get(ctx, "securities", map[string]any{"userId": 1, "isin": ISIN, "page": page, "perPage": 100})
		if err != nil {
			return nil, nil, errors.New("failed GET /security-service/securities, err: " + err.Error())
		}

		if resp.StatusCode != 200 {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()

			return nil, nil, errors.New("non 200 resp GET /security-service/securities, resp: " + string(body))
		}

		var res struct {
			Data []*struct {
				ID   int    `json:"id"`
				ISIN string `json:"isin"`
			} `json:"data"`
		}

		err = json.NewDecoder(resp.Body).Decode(&res)
		if err != nil {
			resp.Body.Close()

			return nil, nil, errors.New("unexpected resp GET /security-service/securities, unmarshalErr: " + err.Error())
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

func (h *marketDataHandler) getMetricIDs(ctx *gofr.Context) ([]int, map[int]string, error) {
	var (
		metricIDs    []int
		metricsNames = make(map[int]string)
	)

	securityService := ctx.GetHTTPService("security-service")

	for page := 1; ; page++ {
		resp, err := securityService.Get(ctx, "metrics", map[string]any{"userId": 1, "page": page, "perPage": 100})
		if err != nil {
			resp.Body.Close()

			return nil, nil, errors.New("failed GET /security-service/metrics, err: " + err.Error())
		}

		if resp.StatusCode != 200 {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()

			return nil, nil, errors.New("non 200 resp GET /security-service/metrics, resp: " + string(body))
		}

		var res struct {
			Data []*struct {
				ID   int    `json:"id"`
				Name string `json:"name"`
			} `json:"data"`
		}

		err = json.NewDecoder(resp.Body).Decode(&res)
		if err != nil {
			resp.Body.Close()

			return nil, nil, errors.New("unexpected resp GET /security-service/metrics, unmarshalErr: " + err.Error())
		}

		resp.Body.Close()

		if len(res.Data) == 0 {
			break
		}

		for i := range res.Data {
			metricIDs = append(metricIDs, res.Data[i].ID)
			metricsNames[res.Data[i].ID] = res.Data[i].Name
		}
	}

	return metricIDs, metricsNames, nil
}

func (h *marketDataHandler) updateLTP(ctx *gofr.Context, securityID int, ltp float64) error {
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

func (h *marketDataHandler) createOrUpdateSecurityStat(ctx *gofr.Context, securityID int, date time.Time, ohlcData *dataProviders.OHLCData) error {
	securityStatID, statExists, err := h.checkIfStatAlreadyExists(ctx, securityID, date)
	if err != nil {
		return err
	}

	if statExists {
		if err = h.updateSecurityStat(ctx, securityStatID, date, ohlcData); err != nil {
			return err
		}

		return nil
	}

	if err = h.createSecurityStat(ctx, securityID, date, ohlcData); err != nil {
		return err
	}

	return nil
}

func (h *marketDataHandler) checkIfStatAlreadyExists(ctx *gofr.Context, securityID int, date time.Time) (int, bool, error) {
	securityService := ctx.GetHTTPService("security-service")

	resp, err := securityService.Get(ctx, "security-stats", map[string]any{"securityId": securityID, "date": date.Format(time.DateOnly)})
	if err != nil {
		return 0, false, errors.New("failed GET /security-service/security-stats, err: " + err.Error())
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)

		return 0, false, errors.New("non 200 resp GET /security-service/security-stats, resp: " + string(body))
	}

	var res struct {
		Data []*struct {
			ID int `json:"id"`
		} `json:"data"`
	}

	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return 0, false, errors.New("unexpected resp GET /security-service/security-stats, unmarshallErr: " + err.Error())
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

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)

		return errors.New(fmt.Sprintf("non 200 resp POST /security-service/security-stats/%d, resp: %s", securityStatID, string(b)))
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

	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		b, _ := io.ReadAll(resp.Body)

		return errors.New("non 201 resp POST /security-service/security-stats, resp: " + string(b))
	}

	return nil
}

func (h *marketDataHandler) createOrUpdateSecurityMetric(ctx *gofr.Context, securityID, metricID int, date time.Time) error {
	securityMetricID, metricExists, err := h.checkIfSecurityMetricAlreadyExists(ctx, securityID, metricID, date)
	if err != nil {
		return err
	}

	if metricExists {
		if err = h.updateSecurityMetric(ctx, securityMetricID); err != nil {
			return err
		}

		return nil
	}

	if err = h.createSecurityMetric(ctx, securityID, metricID, date); err != nil {
		return err
	}

	return nil
}

func (h *marketDataHandler) checkIfSecurityMetricAlreadyExists(ctx *gofr.Context, securityID, metricID int, date time.Time) (int, bool, error) {
	securityService := ctx.GetHTTPService("security-service")

	resp, err := securityService.Get(ctx, "security-metrics", map[string]any{"securityId": securityID, "metricId": metricID, "date": date.Format(time.DateOnly)})
	if err != nil {
		return 0, false, errors.New("failed GET /security-service/security-metrics, err: " + err.Error())
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)

		return 0, false, errors.New("non 200 resp GET /security-service/security-metrics, resp: " + string(body))
	}

	var res struct {
		Data []*struct {
			ID int `json:"id"`
		} `json:"data"`
	}

	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return 0, false, errors.New("unexpected resp GET /security-service/security-metrics, unmarshallErr: " + err.Error())
	}

	if len(res.Data) > 0 {
		return res.Data[0].ID, true, nil
	}

	return 0, false, nil
}

func (h *marketDataHandler) updateSecurityMetric(ctx *gofr.Context, securityMetricID int) error {
	payload := map[string]any{
		"userId":         1,
		"recomputeValue": true,
	}

	body, _ := json.Marshal(payload)

	resp, err := ctx.GetHTTPService("security-service").Patch(ctx, fmt.Sprintf("security-metrics/%d", securityMetricID), nil, body)
	if err != nil {
		return errors.New(fmt.Sprintf("failed PATCH /security-service/security-metrics/%d, err: %s", securityMetricID, err))
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)

		return errors.New(fmt.Sprintf("non 200 resp POST /security-service/security-metrics/%d, resp: %s", securityMetricID, string(b)))
	}

	return nil
}

func (h *marketDataHandler) createSecurityMetric(ctx *gofr.Context, securityID, metricID int, date time.Time) error {
	payload := map[string]any{
		"userId":     1,
		"securityId": securityID,
		"metricId":   metricID,
		"date":       date.Format(time.DateOnly),
	}

	body, _ := json.Marshal(payload)

	resp, err := ctx.GetHTTPService("security-service").Post(ctx, "security-metrics", nil, body)
	if err != nil {
		return errors.New("failed POST /security-service/security-metrics, err: " + err.Error())
	}

	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		b, _ := io.ReadAll(resp.Body)

		return errors.New("non 201 resp POST /security-service/security-metrics, resp: " + string(b))
	}

	return nil
}
