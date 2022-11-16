package tool

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/doug-martin/goqu/v9"
	"net/http"
	"time"

	"github.com/charlieegan3/tool-virgin-media-outage-reporter/pkg/outages"
)

// Check is a job which scrapes the VM website for any ongoing incidents
type Check struct {
	DB *sql.DB

	ScheduleOverride string

	WebhookRSSEndpoint string

	Endpoint string
	Headers  map[string]string
}

func (c *Check) Name() string {
	return "check"
}

func (c *Check) Run(ctx context.Context) error {
	doneCh := make(chan bool)
	errCh := make(chan error)

	goquDB := goqu.New("postgres", c.DB)

	go func() {
		currentOutages, err := outages.FetchOutages(outages.OutageOpts{
			Endpoint: c.Endpoint,
			Headers:  c.Headers,
		})
		if err != nil {
			errCh <- fmt.Errorf("failed to fetch outages: %w", err)
			return
		}

		if len(currentOutages) == 0 {
			doneCh <- true
			return
		}

		var outageIDs []string
		for _, outage := range currentOutages {
			outageIDs = append(outageIDs, outage.ID)
		}

		var rows []struct {
			ID string `db:"vm_outage_id"`
		}
		sel := goquDB.From("virgin_media_outage_reporter.outages").
			Select("vm_outage_id").
			Where(goqu.C("vm_outage_id").In(outageIDs))
		err = sel.Executor().ScanStructs(&rows)
		if err != nil {
			errCh <- fmt.Errorf("failed to get existing outages: %w", err)
			return
		}

		var newOutages []outages.Outage
		for i, outage := range currentOutages {
			found := false
			for _, row := range rows {
				if outage.ID == row.ID {
					found = true
					break
				}
			}
			if !found {
				newOutages = append(newOutages, currentOutages[i])
			}
		}

		if len(newOutages) == 0 {
			doneCh <- true
			return
		}

		if len(newOutages) > 0 {
			var rows []struct {
				ID   string `db:"vm_outage_id"`
				Data string `db:"data"`
			}
			for _, outage := range newOutages {
				outageData, err := json.MarshalIndent(outage, "", "  ")
				if err != nil {
					errCh <- fmt.Errorf("failed to marshal outage: %w", err)
					return
				}

				rssEntryData := []map[string]string{
					{
						"title": fmt.Sprintf("Virgin Media outage: %s", outage.ID),
						"body":  string(outageData),
						"url":   "",
					},
				}

				b, err := json.Marshal(rssEntryData)
				if err != nil {
					errCh <- fmt.Errorf("failed to marshal rss data: %w", err)
					return
				}

				client := &http.Client{}
				req, err := http.NewRequest("POST", c.WebhookRSSEndpoint, bytes.NewBuffer(b))
				if err != nil {
					errCh <- fmt.Errorf("failed to create request: %w", err)
					return
				}

				req.Header.Add("Content-Type", "application/json; charset=utf-8")

				resp, err := client.Do(req)
				if err != nil {
					errCh <- fmt.Errorf("failed to send request: %w", err)
					return
				}
				if resp.StatusCode != http.StatusOK {
					errCh <- fmt.Errorf("request was not 200OK: %d", resp.StatusCode)
					return
				}

				// add row to be saved
				rows = append(rows, struct {
					ID   string `db:"vm_outage_id"`
					Data string `db:"data"`
				}{
					ID:   outage.ID,
					Data: string(outageData),
				})
			}

			ins := goquDB.Insert("virgin_media_outage_reporter.outages").Rows(rows).Executor()
			_, err = ins.Exec()
			if err != nil {
				errCh <- fmt.Errorf("failed to insert new outages: %w", err)
				return
			}
		}

		doneCh <- true
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case e := <-errCh:
		return fmt.Errorf("job failed with error: %s", e)
	case <-doneCh:
		return nil
	}
}

func (c *Check) Timeout() time.Duration {
	return 30 * time.Second
}

func (c *Check) Schedule() string {
	if c.ScheduleOverride != "" {
		return c.ScheduleOverride
	}
	return "0 0 6 * * *"
}
