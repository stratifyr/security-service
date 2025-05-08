package dataProviders

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
	lastAPICallTime       time.Time
	mu                    *sync.Mutex
}

func NewDhanHQClient(app *gofr.App) (*client, error) {
	apiKey := app.Config.Get("DHAN_API_KEY")
	if apiKey == "" {
		return nil, errors.New("missing DHAN_API_KEY")
	}

	clientID := app.Config.Get("DHAN_CLIENT_ID")
	if clientID == "" {
		return nil, errors.New("missing DHAN_CLIENT_ID")
	}

	file, err := os.Open("./data-providers/dhan-scrip-master_NSE_EQ.csv")
	if err != nil {
		return nil, errors.New("failed to load dhan-scrip-master_NSE_EQ.csv")
	}

	defer file.Close()

	records, err := csv.NewReader(file).ReadAll()
	if err != nil {
		return nil, errors.New("failed to read dhan-scrip-master_NSE_EQ.csv")
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
		lastAPICallTime:       time.Date(2001, 1, 1, 1, 1, 1, 1, time.UTC),
		mu:                    &sync.Mutex{},
	}, nil
}

func (c *client) LTP(ctx *gofr.Context, isin string) (*LTPData, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	payload := map[string][]int{
		"NSE_EQ": {c.isinSecurityIDMapping[isin]},
	}

	body, _ := json.Marshal(payload)
	headers := map[string]string{"Content-Type": "application/json", "access-token": c.apiKey, "client-id": c.clientID}

	if time.Now().UTC().Sub(c.lastAPICallTime) <= time.Second {
		time.Sleep(time.Second)
	}

	resp, err := ctx.GetHTTPService("dhan-api").PostWithHeaders(ctx, "v2/marketfeed/ltp", nil, body, headers)
	if err != nil {
		return nil, errors.New("failed POST /v2/marketfeed/ltp, err: " + err.Error())
	}

	defer resp.Body.Close()

	c.lastAPICallTime = time.Now().UTC()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)

		return nil, errors.New("non 200 resp POST /v2/marketfeed/ltp, resp: " + string(b))
	}

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

func (c *client) LTPBulk(ctx *gofr.Context, isins []string) ([]*LTPData, error) {
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

	c.mu.Lock()
	defer c.mu.Unlock()

	if time.Now().UTC().Sub(c.lastAPICallTime) <= time.Second {
		time.Sleep(time.Second)
	}

	resp, err := ctx.GetHTTPService("dhan-api").PostWithHeaders(ctx, "v2/marketfeed/ltp", nil, body, headers)
	if err != nil {
		return nil, errors.New("failed POST /v2/marketfeed/ltp, err: " + err.Error())
	}

	defer resp.Body.Close()

	c.lastAPICallTime = time.Now().UTC()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)

		return nil, errors.New("non 200 resp POST /v2/marketfeed/ltp, resp: " + string(b))
	}

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

	var ltpData = make([]*LTPData, len(isins))

	for i := range isins {
		securityID := c.isinSecurityIDMapping[isins[i]]

		data, ok := res.Data.NseEQ[strconv.Itoa(securityID)]
		if !ok {
			return nil, errors.New(fmt.Sprintf("missing data for %s, POST /v2/marketfeed/ltp", isins[i]))
		}

		ltpData[i] = &LTPData{
			ISIN: isins[i],
			LTP:  data.LTP,
		}
	}

	return ltpData, nil
}

func (c *client) OHLC(ctx *gofr.Context, isin string) (*OHLCData, error) {
	payload := map[string][]int{
		"NSE_EQ": {c.isinSecurityIDMapping[isin]},
	}

	body, _ := json.Marshal(payload)
	headers := map[string]string{"Content-Type": "application/json", "access-token": c.apiKey, "client-id": c.clientID}

	c.mu.Lock()
	defer c.mu.Unlock()

	if time.Now().UTC().Sub(c.lastAPICallTime) <= time.Second {
		time.Sleep(time.Second)
	}

	resp, err := ctx.GetHTTPService("dhan-api").PostWithHeaders(ctx, "v2/marketfeed/quote", nil, body, headers)
	if err != nil {
		return nil, errors.New("failed POST /v2/marketfeed/quote, err: " + err.Error())
	}

	defer resp.Body.Close()

	c.lastAPICallTime = time.Now().UTC()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)

		return nil, errors.New("non 200 resp POST /v2/marketfeed/quote, resp: " + string(b))
	}

	var res struct {
		Data struct {
			NseEQ map[string]struct {
				Volume int `json:"volume"`
				Ohlc   struct {
					Open  float64 `json:"open"`
					High  float64 `json:"high"`
					Low   float64 `json:"low"`
					Close float64 `json:"close"`
				} `json:"ohlc"`
			} `json:"NSE_EQ"`
		} `json:"data"`
	}

	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return nil, errors.New("unexpected resp POST /v2/marketfeed/quote, err: " + err.Error())
	}

	stats, ok := res.Data.NseEQ[strconv.Itoa(c.isinSecurityIDMapping[isin])]
	if !ok {
		return nil, errors.New("missing ohlc data /v2/marketfeed/quote, err: " + err.Error())
	}

	return &OHLCData{
		ISIN:   isin,
		Open:   stats.Ohlc.Open,
		High:   stats.Ohlc.High,
		Low:    stats.Ohlc.Low,
		Close:  stats.Ohlc.Close,
		Volume: stats.Volume,
	}, nil
}

func (c *client) OHLCBulk(ctx *gofr.Context, isins []string) ([]*OHLCData, error) {
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

	c.mu.Lock()
	defer c.mu.Unlock()

	if time.Now().UTC().Sub(c.lastAPICallTime) <= time.Second {
		time.Sleep(time.Second)
	}

	resp, err := ctx.GetHTTPService("dhan-api").PostWithHeaders(ctx, "v2/marketfeed/quote", nil, body, headers)
	if err != nil {
		return nil, errors.New("failed POST /v2/marketfeed/quote, err: " + err.Error())
	}

	defer resp.Body.Close()

	c.lastAPICallTime = time.Now().UTC()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)

		return nil, errors.New("non 200 resp POST /v2/marketfeed/quote, resp: " + string(b))
	}

	var res struct {
		Data struct {
			NseEQ map[string]struct {
				Volume float64 `json:"volume"`
				Ohlc   struct {
					Open  float64 `json:"open"`
					High  float64 `json:"high"`
					Low   float64 `json:"low"`
					Close float64 `json:"close"`
				} `json:"ohlc"`
			} `json:"NSE_EQ"`
		} `json:"data"`
	}

	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return nil, errors.New("unexpected resp POST /v2/marketfeed/quote, err: " + err.Error())
	}

	var ohlcData = make([]*OHLCData, len(isins))

	for i := range isins {
		securityID := c.isinSecurityIDMapping[isins[i]]

		data, ok := res.Data.NseEQ[strconv.Itoa(securityID)]
		if !ok {
			return nil, errors.New(fmt.Sprintf("missing data for %s, POST /v2/marketfeed/quote", isins[i]))
		}

		ohlcData[i] = &OHLCData{
			ISIN:   isins[i],
			Open:   data.Ohlc.Open,
			High:   data.Ohlc.High,
			Low:    data.Ohlc.Low,
			Close:  data.Ohlc.Close,
			Volume: int(data.Volume),
		}
	}

	return ohlcData, nil
}

func (c *client) HistoricalOHLC(ctx *gofr.Context, isin string, startDate, endDate time.Time) ([]*HistoricalOHLC, error) {
	payload := map[string]any{
		"securityId":      c.isinSecurityIDMapping[isin],
		"exchangeSegment": "NSE_EQ",
		"instrument":      "EQUITY",
		"expiryCode":      0,
		"oi":              false,
		"fromDate":        startDate.Format(time.DateOnly),
		"toDate":          endDate.AddDate(0, 0, 1).Format(time.DateOnly),
	}

	body, _ := json.Marshal(payload)
	headers := map[string]string{"Content-Type": "application/json", "access-token": c.apiKey}

	c.mu.Lock()
	defer c.mu.Unlock()

	if time.Now().UTC().Sub(c.lastAPICallTime) <= time.Second {
		time.Sleep(time.Second)
	}

	c.lastAPICallTime = time.Now().UTC()

	resp, err := ctx.GetHTTPService("dhan-api").PostWithHeaders(ctx, "v2/charts/historical", nil, body, headers)
	if err != nil {
		return nil, errors.New("failed POST /v2/charts/historical, err: " + err.Error())
	}

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)

		return nil, errors.New("non 200 resp POST /v2/charts/historical, resp: " + string(b))
	}

	defer resp.Body.Close()

	var res struct {
		Open      []float64 `json:"open"`
		High      []float64 `json:"high"`
		Low       []float64 `json:"low"`
		Close     []float64 `json:"close"`
		Volume    []float64 `json:"volume"`
		Timestamp []float64 `json:"timestamp"`
	}

	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return nil, errors.New("unexpected resp POST /v2/charts/historical, err: " + err.Error())
	}

	var historicalData = make([]*HistoricalOHLC, len(res.Timestamp))

	istLocation, _ := time.LoadLocation("Asia/Kolkata")

	for i := range res.Timestamp {
		historicalData[i] = &HistoricalOHLC{
			Date: time.Unix(int64(res.Timestamp[i]), 0).In(istLocation),
			OHLCData: &OHLCData{
				ISIN:   isin,
				Open:   res.Open[i],
				High:   res.Open[i],
				Low:    res.Open[i],
				Close:  res.Open[i],
				Volume: int(res.Volume[i]),
			},
		}
	}

	return historicalData, nil
}
