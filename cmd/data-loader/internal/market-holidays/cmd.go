package marketholidays

import (
	_ "embed"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"strings"

	"gofr.dev/pkg/gofr"
)

//go:embed data.csv
var data string

func NewCMDHandler() *handler {
	return &handler{}
}

type handler struct{}

func (h *handler) Load(ctx *gofr.Context) (any, error) {
	reader := csv.NewReader(strings.NewReader(data))

	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read csv file, err: %v", err)
	}

	idxDate := slices.Index(headers, "Date")
	idxDescription := slices.Index(headers, "Description")

	var errs []string

	for {
		row, readErr := reader.Read()
		if readErr == io.EOF {
			break
		}

		if readErr != nil {
			return nil, fmt.Errorf("failed to read csv file, err: %v", err)
		}

		fmt.Println(row[idxDate])

		if err = h.createOrUpdateMarketHolidays(ctx, row[idxDate], row[idxDescription]); err != nil {
			errs = append(errs, fmt.Sprintf("[%s] %v", row[idxDate], err))
			continue
		}
	}

	if len(errs) > 0 {
		return "\nERRORS:", fmt.Errorf(strings.Join(errs, "\n"))
	}

	return "\nOK", nil
}

func (h *handler) createOrUpdateMarketHolidays(ctx *gofr.Context, date, description string) error {
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

func (h *handler) checkIfMarketHolidayAlreadyExists(ctx *gofr.Context, date string) (int, bool, error) {
	securityService := ctx.GetHTTPService("security-service")

	resp, err := securityService.Get(ctx, "market-holidays", map[string]any{"date": date})
	if err != nil {
		return 0, false, fmt.Errorf("failed GET /security-service/market-holidays, err: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)

		return 0, false, fmt.Errorf("non 200 resp GET /security-service/market-holidays, resp: %s", body)
	}

	var res struct {
		Data []*struct {
			ID int `json:"id"`
		} `json:"data"`
	}

	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return 0, false, fmt.Errorf("unexpected resp GET /security-service/market-holidays, unmarshallErr: %v", err)
	}

	if len(res.Data) > 0 {
		return res.Data[0].ID, true, nil
	}

	return 0, false, nil
}

func (h *handler) updateMarketHoliday(ctx *gofr.Context, marketHolidayID int, description string) error {
	payload := map[string]any{
		"userId":      1,
		"description": description,
	}

	body, _ := json.Marshal(payload)

	resp, err := ctx.GetHTTPService("security-service").Patch(ctx, fmt.Sprintf("market-holidays/%d", marketHolidayID), nil, body)
	if err != nil {
		return fmt.Errorf("failed PATCH /security-service/market-holidays/%d, err: %s", marketHolidayID, err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)

		return fmt.Errorf("non 200 resp PATCH /security-service/market-holidays/%d, resp: %s", marketHolidayID, b)
	}

	return nil
}

func (h *handler) createMarketHoliday(ctx *gofr.Context, date, description string) error {
	payload := map[string]any{
		"userId":      1,
		"date":        date,
		"description": description,
	}

	body, _ := json.Marshal(payload)

	resp, err := ctx.GetHTTPService("security-service").Post(ctx, "market-holidays", nil, body)
	if err != nil {
		return fmt.Errorf("failed POST /security-service/market-holidays, err: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		b, _ := io.ReadAll(resp.Body)

		return fmt.Errorf("non 201 resp POST /security-service/market-holidays, resp: %s", b)
	}

	return nil
}
