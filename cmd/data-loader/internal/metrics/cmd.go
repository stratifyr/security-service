package metrics

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

	idxName := slices.Index(headers, "Name")
	idxType := slices.Index(headers, "Type")
	idxPeriod := slices.Index(headers, "Period")
	idxTier := slices.Index(headers, "Tier")

	var errs []string

	for {
		row, readErr := reader.Read()
		if readErr == io.EOF {
			break
		}

		if readErr != nil {
			return nil, fmt.Errorf("failed to read csv file, err: %v", err)
		}

		fmt.Println(row[idxName])

		if err = h.createOrUpdateMetric(ctx, row[idxName], row[idxType], row[idxPeriod], row[idxTier]); err != nil {
			errs = append(errs, fmt.Sprintf("[%s] %v", row[idxName], err))
			continue
		}
	}

	if len(errs) > 0 {
		return "\nERRORS:", fmt.Errorf(strings.Join(errs, "\n"))
	}

	return "\nOK", nil
}

func (h *handler) getMetricIDs(ctx *gofr.Context) ([]int, map[int]string, error) {
	var (
		metricIDs    []int
		metricsNames = make(map[int]string)
	)

	securityService := ctx.GetHTTPService("security-service")

	for page := 1; ; page++ {
		resp, err := securityService.Get(ctx, "metrics", map[string]any{"userId": 1, "page": page, "perPage": 100})
		if err != nil {
			resp.Body.Close()

			return nil, nil, fmt.Errorf("failed GET /security-service/metrics, err: %v", err)
		}

		if resp.StatusCode != 200 {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()

			return nil, nil, fmt.Errorf("non 200 resp GET /security-service/metrics, resp: %s", body)
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

			return nil, nil, fmt.Errorf("unexpected resp GET /security-service/metrics, unmarshalErr: %v", err)
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

func (h *handler) createOrUpdateMetric(ctx *gofr.Context, name, typ, period, tier string) error {
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

func (h *handler) checkIfMetricAlreadyExists(ctx *gofr.Context, typ string, period int) (int, bool, error) {
	securityService := ctx.GetHTTPService("security-service")

	resp, err := securityService.Get(ctx, "metrics", map[string]any{"type": typ, "period": period})
	if err != nil {
		return 0, false, fmt.Errorf("failed GET /security-service/metrics, err: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)

		return 0, false, fmt.Errorf("non 200 resp GET /security-service/metrics, resp: %s", body)
	}

	var res struct {
		Data []*struct {
			ID int `json:"id"`
		} `json:"data"`
	}

	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return 0, false, fmt.Errorf("unexpected resp GET /security-service/metrics, unmarshallErr: %v", err)
	}

	if len(res.Data) > 0 {
		return res.Data[0].ID, true, nil
	}

	return 0, false, nil
}

func (h *handler) updateMetric(ctx *gofr.Context, metricID int, name string, tier int) error {
	payload := map[string]any{
		"userId": 1,
		"name":   name,
		"tier":   tier,
	}

	body, _ := json.Marshal(payload)

	resp, err := ctx.GetHTTPService("security-service").Patch(ctx, fmt.Sprintf("metrics/%d", metricID), nil, body)
	if err != nil {
		return fmt.Errorf("failed PATCH /security-service/metrics/%d, err: %s", metricID, err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)

		return fmt.Errorf("non 200 resp PATCH /security-service/metrics/%d, resp: %s", metricID, b)
	}

	return nil
}

func (h *handler) createMetric(ctx *gofr.Context, name, typ string, period, tier int) error {
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
		return fmt.Errorf("failed POST /security-service/metrics, err: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		b, _ := io.ReadAll(resp.Body)

		return fmt.Errorf("non 201 resp POST /security-service/metrics, resp: %s", b)
	}

	return nil
}
