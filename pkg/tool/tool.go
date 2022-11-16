package tool

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/Jeffail/gabs/v2"
	"github.com/charlieegan3/toolbelt/pkg/apis"
	"github.com/gorilla/mux"
)

//go:embed migrations
var migrations embed.FS

// VirginMediaOutageReporter creates RSS entries when new Virgin Media outages are reported
type VirginMediaOutageReporter struct {
	db     *sql.DB
	config *gabs.Container
}

func (v *VirginMediaOutageReporter) Name() string {
	return "virgin-media-outage-reporter"
}

func (v *VirginMediaOutageReporter) FeatureSet() apis.FeatureSet {
	return apis.FeatureSet{
		Config:   true,
		Database: true,
		Jobs:     true,
	}
}

func (v *VirginMediaOutageReporter) SetConfig(config map[string]any) error {
	v.config = gabs.Wrap(config)

	return nil
}
func (v *VirginMediaOutageReporter) Jobs() ([]apis.Job, error) {
	var j []apis.Job
	var path string
	var ok bool

	// load all config
	path = "jobs.check.schedule"
	schedule, ok := v.config.Path(path).Data().(string)
	if !ok {
		return j, fmt.Errorf("missing required config path: %s", path)
	}
	path = "jobs.check.endpoint"
	endpoint, ok := v.config.Path(path).Data().(string)
	if !ok {
		return j, fmt.Errorf("missing required config path: %s", path)
	}
	path = "jobs.check.headers"
	headers, ok := v.config.Path(path).Data().(map[string]interface{})
	if !ok {
		return j, fmt.Errorf("missing required config path: %s", path)
	}
	headersString := make(map[string]string)
	for k, v := range headers {
		headersString[k], ok = v.(string)
		if !ok {
			return j, fmt.Errorf("invalid header config at %s, expected string:string mapping", k)
		}
	}
	path = "jobs.check.webhook_rss_endpoint"
	webhookRSSEndpoint, ok := v.config.Path(path).Data().(string)
	if !ok {
		return j, fmt.Errorf("missing required config path: %s", path)
	}

	return []apis.Job{
		&Check{
			DB:                 v.db,
			ScheduleOverride:   schedule,
			Endpoint:           endpoint,
			Headers:            headersString,
			WebhookRSSEndpoint: webhookRSSEndpoint,
		},
	}, nil
}
func (v *VirginMediaOutageReporter) ExternalJobsFuncSet(f func(job apis.ExternalJob) error) {}

func (v *VirginMediaOutageReporter) DatabaseMigrations() (*embed.FS, string, error) {
	return &migrations, "migrations", nil
}
func (v *VirginMediaOutageReporter) DatabaseSet(db *sql.DB) {
	v.db = db
}
func (v *VirginMediaOutageReporter) HTTPPath() string                    { return "" }
func (v *VirginMediaOutageReporter) HTTPAttach(router *mux.Router) error { return nil }
