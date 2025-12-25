package dataProviders

import (
	_ "embed"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"
	"time"

	"gofr.dev/pkg/gofr"
)

//go:embed dhan-scrip-master.csv
var dhanMasterScrip string

type client struct {
	apiKey              string
	clientID            string
	symbolDhanIDMapping map[string]int
	dhanIDSymbolMapping map[int]string
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

	records, err := csv.NewReader(strings.NewReader(dhanMasterScrip)).ReadAll()
	if err != nil {
		return nil, errors.New("failed to read dhan master scrip")
	}

	symbolDhanIDMapping := make(map[string]int)
	dhanIDSymbolMapping := make(map[int]string)
	headers := records[0]

	for _, row := range records[1:] {
		symbol := row[slices.Index(headers, "UNDERLYING_SYMBOL")]
		dhanIDStr := row[slices.Index(headers, "SECURITY_ID")]
		dhanID, _ := strconv.Atoi(dhanIDStr)

		symbolDhanIDMapping[symbol] = dhanID
		dhanIDSymbolMapping[dhanID] = symbol
	}

	app.AddHTTPService("dhan-api", "https://api.dhan.co")

	return &client{
		apiKey:              apiKey,
		clientID:            clientID,
		symbolDhanIDMapping: symbolDhanIDMapping,
		dhanIDSymbolMapping: dhanIDSymbolMapping,
	}, nil
}

func (c *client) LTP(ctx *gofr.Context, symbol string) (*LTPData, error) {
	payload := map[string][]int{
		"NSE_EQ": {c.symbolDhanIDMapping[symbol]},
	}

	body, _ := json.Marshal(payload)
	headers := map[string]string{"Content-Type": "application/json", "access-token": c.apiKey, "client-id": c.clientID}

	resp, err := ctx.GetHTTPService("dhan-api").PostWithHeaders(ctx, "v2/marketfeed/ltp", nil, body, headers)
	if err != nil {
		return nil, errors.New("failed POST /v2/marketfeed/ltp, err: " + err.Error())
	}

	defer resp.Body.Close()

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
		Symbol: symbol,
		LTP:    res.Data.NseEQ[strconv.Itoa(c.symbolDhanIDMapping[symbol])].LTP,
	}, nil
}

func (c *client) LTPBulk(ctx *gofr.Context, symbols []string) ([]*LTPData, error) {
	if len(symbols) > 1000 {
		return nil, errors.New("max limit is 1000 for bulk ltp fetch")
	}

	payload := map[string][]int{
		"NSE_EQ": make([]int, len(symbols)),
	}

	for i := range symbols {
		payload["NSE_EQ"][i] = c.symbolDhanIDMapping[symbols[i]]
	}

	body, _ := json.Marshal(payload)
	headers := map[string]string{"Content-Type": "application/json", "access-token": c.apiKey, "client-id": c.clientID}

	resp, err := ctx.GetHTTPService("dhan-api").PostWithHeaders(ctx, "v2/marketfeed/ltp", nil, body, headers)
	if err != nil {
		return nil, errors.New("failed POST /v2/marketfeed/ltp, err: " + err.Error())
	}

	defer resp.Body.Close()

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

	var ltpData []*LTPData

	for i := range symbols {
		securityID := c.symbolDhanIDMapping[symbols[i]]

		data, ok := res.Data.NseEQ[strconv.Itoa(securityID)]
		if !ok {
			ctx.Warnf(fmt.Sprintf("missing data for %s, POST /v2/marketfeed/ltp", symbols[i]))
			continue
		}

		ltpData = append(ltpData, &LTPData{
			Symbol: symbols[i],
			LTP:    data.LTP,
		})
	}

	return ltpData, nil
}

func (c *client) OHLC(ctx *gofr.Context, symbol string) (*OHLCData, error) {
	payload := map[string][]int{
		"NSE_EQ": {c.symbolDhanIDMapping[symbol]},
	}

	body, _ := json.Marshal(payload)
	headers := map[string]string{"Content-Type": "application/json", "access-token": c.apiKey, "client-id": c.clientID}

	resp, err := ctx.GetHTTPService("dhan-api").PostWithHeaders(ctx, "v2/marketfeed/quote", nil, body, headers)
	if err != nil {
		return nil, errors.New("failed POST /v2/marketfeed/quote, err: " + err.Error())
	}

	defer resp.Body.Close()

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

	stats, ok := res.Data.NseEQ[strconv.Itoa(c.symbolDhanIDMapping[symbol])]
	if !ok {
		return nil, errors.New("missing ohlc data /v2/marketfeed/quote, err: " + err.Error())
	}

	return &OHLCData{
		Symbol: symbol,
		Open:   stats.Ohlc.Open,
		High:   stats.Ohlc.High,
		Low:    stats.Ohlc.Low,
		Close:  stats.Ohlc.Close,
		Volume: stats.Volume,
	}, nil
}

func (c *client) OHLCBulk(ctx *gofr.Context, symbols []string) ([]*OHLCData, error) {
	if len(symbols) > 1000 {
		return nil, errors.New("max limit is 1000 for bulk ltp fetch")
	}

	payload := map[string][]int{
		"NSE_EQ": make([]int, len(symbols)),
	}

	for i := range symbols {
		payload["NSE_EQ"][i] = c.symbolDhanIDMapping[symbols[i]]
	}

	body, _ := json.Marshal(payload)
	headers := map[string]string{"Content-Type": "application/json", "access-token": c.apiKey, "client-id": c.clientID}

	resp, err := ctx.GetHTTPService("dhan-api").PostWithHeaders(ctx, "v2/marketfeed/quote", nil, body, headers)
	if err != nil {
		return nil, errors.New("failed POST /v2/marketfeed/quote, err: " + err.Error())
	}

	defer resp.Body.Close()

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

	var ohlcData []*OHLCData

	for i := range symbols {
		securityID := c.symbolDhanIDMapping[symbols[i]]

		data, ok := res.Data.NseEQ[strconv.Itoa(securityID)]
		if !ok {
			ctx.Warnf(fmt.Sprintf("missing data for %s, POST /v2/marketfeed/quote", symbols[i]))
			continue
		}

		ohlcData = append(ohlcData, &OHLCData{
			Symbol: symbols[i],
			Open:   data.Ohlc.Open,
			High:   data.Ohlc.High,
			Low:    data.Ohlc.Low,
			Close:  data.Ohlc.Close,
			Volume: int(data.Volume),
		})
	}

	return ohlcData, nil
}

func (c *client) HistoricalOHLC(ctx *gofr.Context, symbol string, startDate, endDate time.Time) ([]*HistoricalOHLC, error) {
	payload := map[string]any{
		"securityId":      c.symbolDhanIDMapping[symbol],
		"exchangeSegment": "NSE_EQ",
		"instrument":      "EQUITY",
		"expiryCode":      0,
		"oi":              false,
		"fromDate":        startDate.Format(time.DateOnly),
		"toDate":          endDate.AddDate(0, 0, 1).Format(time.DateOnly),
	}

	body, _ := json.Marshal(payload)
	headers := map[string]string{"Content-Type": "application/json", "access-token": c.apiKey}

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
				Symbol: symbol,
				Open:   res.Open[i],
				High:   res.High[i],
				Low:    res.Low[i],
				Close:  res.Close[i],
				Volume: int(res.Volume[i]),
			},
		}
	}

	return historicalData, nil
}
