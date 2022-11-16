package outages

import (
	"encoding/json"
	"fmt"
	"github.com/Jeffail/gabs/v2"
	"io"
	"net/http"
	"time"
)

type OutageOpts struct {
	Endpoint string
	Headers  map[string]string
}

type Outage struct {
	ID                      string
	Description             string
	Status                  string
	Type                    string
	TicketNumber            string
	EstimatedResolutionTime time.Time
}

func (o *Outage) UnmarshalJSON(b []byte) error {
	var raw map[string]interface{}
	err := json.Unmarshal(b, &raw)
	if err != nil {
		return fmt.Errorf("failed to unmarshal outage: %w", err)
	}

	o.ID = raw["outageId"].(string)
	o.Description = raw["description"].(string)
	o.Status = raw["outageStatus"].(string)
	o.Type = raw["outageType"].(string)
	o.TicketNumber = raw["ticketNumber"].(string)

	resoutionTime, err := time.Parse("2006-01-02T15:04:05", raw["estimatedResolutionDate"].(string))
	if err != nil {
		return fmt.Errorf("failed to parse resolution time: %w", err)
	}

	o.EstimatedResolutionTime = resoutionTime.UTC()

	return nil
}

var path = []string{"care2Session", "serviceStatusResponse", "currentOutagesByProductType", "BROADBAND", "outages"}

func FetchOutages(opts OutageOpts) ([]Outage, error) {
	client := &http.Client{}
	//req, err := http.NewRequest("GET", fmt.Sprintf("%s/rou-compax/v2/session-data", opts.Endpoint), nil)
	req, err := http.NewRequest("GET", "http://localhost:8000/outage.json", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.1 Safari/605.1.15")
	for k, v := range opts.Headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch outage data: %w", err)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	container, err := gabs.ParseJSON(respBody)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	if !container.Exists(path...) {
		return []Outage{}, nil
	}

	var outages []Outage
	for _, o := range container.Search(path...).Children() {
		var outage Outage
		err := json.Unmarshal(o.Bytes(), &outage)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal outage: %w", err)
		}

		outages = append(outages, outage)
	}

	return outages, nil
}
