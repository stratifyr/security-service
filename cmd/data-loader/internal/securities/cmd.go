package securities

import (
	_ "embed"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"strconv"
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

	idxISIN := slices.Index(headers, "ISIN Code")
	idxSymbol := slices.Index(headers, "Symbol")
	idxIndustry := slices.Index(headers, "Industry")
	idxName := slices.Index(headers, "Company Name")

	var errs []string

	for {
		row, readErr := reader.Read()
		if readErr == io.EOF {
			break
		}

		if readErr != nil {
			return nil, fmt.Errorf("failed to read csv file, err: %v", err)
		}

		fmt.Println(row[idxSymbol])

		if err = h.createOrUpdateSecurity(ctx, row[idxISIN], row[idxSymbol], row[idxIndustry], row[idxName], "0"); err != nil {
			errs = append(errs, fmt.Sprintf("[%s] %v", row[idxSymbol], err))
			continue
		}
	}

	if len(errs) > 0 {
		return "\nERRORS:", fmt.Errorf(strings.Join(errs, "\n"))
	}

	return "\nOK", nil
}

func (h *handler) createOrUpdateSecurity(ctx *gofr.Context, ISIN, symbol, industry, name, tier string) error {
	tierInt, _ := strconv.Atoi(tier)

	securityID, exists, err := h.checkIfSecurityAlreadyExists(ctx, symbol)
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

func (h *handler) checkIfSecurityAlreadyExists(ctx *gofr.Context, symbol string) (int, bool, error) {
	securityService := ctx.GetHTTPService("security-service")

	resp, err := securityService.Get(ctx, "securities", map[string]any{"symbol": symbol})
	if err != nil {
		return 0, false, fmt.Errorf("failed GET /security-service/securities, err: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)

		return 0, false, fmt.Errorf("non 200 resp GET /security-service/securities, resp: %s", body)
	}

	var res struct {
		Data []*struct {
			ID int `json:"id"`
		} `json:"data"`
	}

	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return 0, false, fmt.Errorf("unexpected resp GET /security-service/securities, unmarshallErr: %v", err)
	}

	if len(res.Data) > 0 {
		return res.Data[0].ID, true, nil
	}

	return 0, false, nil
}

func (h *handler) updateSecurity(ctx *gofr.Context, securityID int, ISIN, symbol, industry, name string, tier int) error {
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
		return fmt.Errorf("failed PATCH /security-service/securities/%d, err: %s", securityID, err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)

		return fmt.Errorf("non 200 resp PATCH /security-service/securities/%d, resp: %s", securityID, b)
	}

	return nil
}

func (h *handler) createSecurity(ctx *gofr.Context, ISIN, symbol, industry, name string, tier int) error {
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
		return fmt.Errorf("failed POST /security-service/securities, err: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		b, _ := io.ReadAll(resp.Body)

		return fmt.Errorf("non 201 resp POST /security-service/securities, resp: %s", b)
	}

	return nil
}
