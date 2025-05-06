package dataProviders

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"os"
	"slices"
	"strconv"
	"sync"
	"time"

	"gofr.dev/pkg/gofr"
)

type client struct {
	apiKey                string
	clientID              string
	isinSecurityIDMapping map[string]int
	securityIDISINMapping map[int]string
	lastMarketFeedAPICall time.Time
	mu                    *sync.Mutex
}

func NewDhanHQClient(app *gofr.App) *client {
	apiKey := app.Config.Get("DHAN_API_KEY")
	if apiKey == "" {
		app.Logger().Fatalf("missing DHAN_API_KEY")
	}

	clientID := app.Config.Get("DHAN_CLIENT_ID")
	if clientID == "" {
		app.Logger().Fatalf("missing DHAN_CLIENT_ID")
	}

	file, err := os.Open("./dhan-scrip-master_NSE_EQ.csv")
	if err != nil {
		app.Logger().Fatalf("failed to load dhan-scrip-master_NSE_EQ.csv")
	}

	defer file.Close()

	records, err := csv.NewReader(file).ReadAll()
	if err != nil {
		app.Logger().Fatalf("failed to read dhan-scrip-master_NSE_EQ.csv")
	}

	isinSecurityIDMapping := make(map[string]int)
	securityIDISINMapping := make(map[int]string)
	headers := records[0]

	for _, row := range records[1:] {
		isin := row[slices.Index(headers, "ISIN")]
		securityIDStr := row[slices.Index(headers, "SECURITY_ID")]
		securityID, _ := strconv.Atoi(securityIDStr)

		isinSecurityIDMapping[isin] = securityID
		securityIDISINMapping[securityID] = isin
	}

	app.AddHTTPService("dhan-api", "https://api.dhan.co")

	return &client{
		apiKey:                apiKey,
		clientID:              clientID,
		isinSecurityIDMapping: isinSecurityIDMapping,
		securityIDISINMapping: securityIDISINMapping,
		lastMarketFeedAPICall: time.Date(2001, 1, 1, 1, 1, 1, 1, time.UTC),
		mu:                    &sync.Mutex{},
	}
}

func (c *client) GetLTP(ctx *gofr.Context, isin string) (*LTPData, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	payload := map[string][]int{
		"NSE_EQ": {c.isinSecurityIDMapping[isin]},
	}

	body, _ := json.Marshal(payload)
	headers := map[string]string{"Content-Type": "application/json", "access-token": c.apiKey, "client-id": c.clientID}

	if time.Now().UTC().Sub(c.lastMarketFeedAPICall) <= 2*time.Second {
		time.Sleep(2 * time.Second)
	}

	resp, err := ctx.GetHTTPService("dhan-api").PostWithHeaders(ctx, "v2/marketfeed/ltp", nil, body, headers)
	if err != nil {
		return nil, errors.New("failed POST /v2/marketfeed/ltp, err: " + err.Error())
	}

	c.lastMarketFeedAPICall = time.Now().UTC()

	defer resp.Body.Close()

	var res struct {
		Data struct {
			NseEQ map[string]struct {
				LTP float64 `json:"last_price"`
			} `json:"NSE_EQ"`
		} `json:"data"`
	}

	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return nil, errors.New("unexpected resp POST /v2/marketfeed/ltp, err: " + err.Error())
	}

	return &LTPData{
		ISIN: isin,
		LTP:  res.Data.NseEQ[strconv.Itoa(c.isinSecurityIDMapping[isin])].LTP,
	}, nil
}

func (c *client) GetLTPBulk(ctx *gofr.Context, isins []string) ([]*LTPData, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(isins) > 1000 {
		return nil, errors.New("max limit is 1000 for bulk ltp fetch")
	}

	payload := map[string][]int{
		"NSE_EQ": make([]int, len(isins)),
	}

	for i := range isins {
		payload["NSE_EQ"][i] = c.isinSecurityIDMapping[isins[i]]
	}

	body, _ := json.Marshal(payload)
	headers := map[string]string{"Content-Type": "application/json", "access-token": c.apiKey, "client-id": c.clientID}

	if time.Now().UTC().Sub(c.lastMarketFeedAPICall) <= 2*time.Second {
		time.Sleep(2 * time.Second)
	}

	resp, err := ctx.GetHTTPService("dhan-api").PostWithHeaders(ctx, "v2/marketfeed/ltp", nil, body, headers)
	if err != nil {
		return nil, errors.New("failed POST /v2/marketfeed/ltp, err: " + err.Error())
	}

	c.lastMarketFeedAPICall = time.Now().UTC()

	defer resp.Body.Close()

	var res struct {
		Data struct {
			NseEQ map[string]struct {
				LTP float64 `json:"last_price"`
			} `json:"NSE_EQ"`
		} `json:"data"`
	}

	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return nil, errors.New("unexpected resp POST /v2/marketfeed/ltp, err: " + err.Error())
	}

	var ltps = make([]*LTPData, len(res.Data.NseEQ))

	for securityIDStr, data := range res.Data.NseEQ {
		securityID, _ := strconv.Atoi(securityIDStr)

		ltps = append(ltps, &LTPData{
			ISIN: c.securityIDISINMapping[securityID],
			LTP:  data.LTP,
		})
	}

	return ltps, nil
}

func (c *client) GetMarketData(ctx *gofr.Context, isin string, date time.Time) (*MarketData, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	currentTime := time.Now().UTC()
	today := currentTime.Format(time.DateOnly)

	if date.After(currentTime) {
		return nil, errors.New("cannot fetch market data for future date")
	}

	if date.Format(time.DateOnly) != today {
		return c.getHistoricalData(ctx, isin, date)
	}

	payload := map[string][]int{
		"NSE_EQ": {c.isinSecurityIDMapping[isin]},
	}

	body, _ := json.Marshal(payload)
	headers := map[string]string{"Content-Type": "application/json", "access-token": c.apiKey, "client-id": c.clientID}

	if time.Now().UTC().Sub(c.lastMarketFeedAPICall) <= 2*time.Second {
		time.Sleep(2 * time.Second)
	}

	resp, err := ctx.GetHTTPService("dhan-api").PostWithHeaders(ctx, "v2/marketfeed/quote", nil, body, headers)
	if err != nil {
		return nil, errors.New("failed POST /v2/marketfeed/quote, err: " + err.Error())
	}

	c.lastMarketFeedAPICall = time.Now().UTC()

	defer resp.Body.Close()

	var res struct {
		Data struct {
			NseEQ map[string]struct {
				Volume int `json:"volume"`
				Ohlc   struct {
					Open  float64 `json:"open"`
					Close float64 `json:"close"`
					High  float64 `json:"high"`
					Low   float64 `json:"low"`
				} `json:"ohlc"`
			} `json:"NSE_EQ"`
		} `json:"data"`
	}

	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return nil, errors.New("unexpected resp POST /v2/marketfeed/quote, err: " + err.Error())
	}

	stats := res.Data.NseEQ[strconv.Itoa(c.isinSecurityIDMapping[isin])]

	return &MarketData{
		ISIN:   isin,
		Open:   stats.Ohlc.Open,
		Close:  stats.Ohlc.Close,
		High:   stats.Ohlc.High,
		Low:    stats.Ohlc.Low,
		Volume: stats.Volume,
	}, nil
}

func (c *client) GetMarketDataBulk(ctx *gofr.Context, isins []string, date time.Time) ([]*MarketData, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(isins) > 1000 {
		return nil, errors.New("max limit is 1000 for bulk ltp fetch")
	}

	currentTime := time.Now().UTC()
	today := currentTime.Format(time.DateOnly)

	if date.After(currentTime) {
		return nil, errors.New("cannot fetch market data for future date")
	}

	if date.Format(time.DateOnly) != today {
		var marketData = make([]*MarketData, len(isins))

		for _, isin := range isins {
			md, err := c.getHistoricalData(ctx, isin, date)
			if err != nil {
				return nil, err
			}

			marketData = append(marketData, md)
		}

		return marketData, nil
	}

	payload := map[string][]int{
		"NSE_EQ": make([]int, len(isins)),
	}

	for i := range isins {
		payload["NSE_EQ"][i] = c.isinSecurityIDMapping[isins[i]]
	}

	body, _ := json.Marshal(payload)
	headers := map[string]string{"Content-Type": "application/json", "access-token": c.apiKey, "client-id": c.clientID}

	if time.Now().UTC().Sub(c.lastMarketFeedAPICall) <= 2*time.Second {
		time.Sleep(2 * time.Second)
	}

	resp, err := ctx.GetHTTPService("dhan-api").PostWithHeaders(ctx, "v2/marketfeed/quote", nil, body, headers)
	if err != nil {
		return nil, errors.New("failed POST /v2/marketfeed/quote, err: " + err.Error())
	}

	c.lastMarketFeedAPICall = time.Now().UTC()

	defer resp.Body.Close()

	var res struct {
		Data struct {
			NseEQ map[string]struct {
				Volume int `json:"volume"`
				Ohlc   struct {
					Open  float64 `json:"open"`
					Close float64 `json:"close"`
					High  float64 `json:"high"`
					Low   float64 `json:"low"`
				} `json:"ohlc"`
			} `json:"NSE_EQ"`
		} `json:"data"`
	}

	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return nil, errors.New("unexpected resp POST /v2/marketfeed/quote, err: " + err.Error())
	}

	var marketData = make([]*MarketData, len(res.Data.NseEQ))

	for securityIDStr, data := range res.Data.NseEQ {
		securityID, _ := strconv.Atoi(securityIDStr)

		marketData = append(marketData, &MarketData{
			ISIN:   c.securityIDISINMapping[securityID],
			Open:   data.Ohlc.Open,
			Close:  data.Ohlc.Close,
			High:   data.Ohlc.High,
			Low:    data.Ohlc.Low,
			Volume: data.Volume,
		})
	}

	return marketData, nil
}

func (c *client) getHistoricalData(ctx *gofr.Context, isin string, date time.Time) (*MarketData, error) {
	payload := map[string]any{
		"securityId":      c.isinSecurityIDMapping[isin],
		"exchangeSegment": "NSE_EQ",
		"instrument":      "EQUITY",
		"expiryCode":      0,
		"oi":              false,
		"fromDate":        date.Format(time.DateOnly),
		"toDate":          date.AddDate(0, 0, 1).Format(time.DateOnly),
	}

	body, _ := json.Marshal(payload)
	headers := map[string]string{"Content-Type": "application/json", "access-token": c.apiKey}

	resp, err := ctx.GetHTTPService("dhan-api").PostWithHeaders(ctx, "v2/charts/historical", nil, body, headers)
	if err != nil {
		return nil, errors.New("failed POST /v2/charts/historical, err: " + err.Error())
	}

	defer resp.Body.Close()

	var res struct {
		Open   []float64 `json:"open"`
		High   []float64 `json:"high"`
		Low    []float64 `json:"low"`
		Close  []float64 `json:"close"`
		Volume []int     `json:"volume"`
	}

	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return nil, errors.New("unexpected resp POST /v2/charts/historical, err: " + err.Error())
	}

	return &MarketData{
		ISIN:   isin,
		Open:   res.Open[0],
		Close:  res.Close[0],
		High:   res.High[0],
		Low:    res.Low[0],
		Volume: res.Volume[0],
	}, nil
}
