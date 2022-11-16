package outages

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestOutages(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bytes, err := os.ReadFile("fixtures/outage.json")
		require.NoError(t, err)
		w.WriteHeader(http.StatusOK)
		w.Write(bytes)
	}))

	opts := OutageOpts{
		Endpoint: testServer.URL,
		Headers:  map[string]string{},
	}

	outages, err := FetchOutages(opts)
	require.NoError(t, err)

	require.Equal(t, []Outage{
		{
			ID:                      "0000000",
			Description:             "You might find that you don't have any Virgin Fibre, Virgin TV, Virgin Phone, TiVo® or Interactive TV services at the moment. We are sorry about this and are working hard to restore your services as soon as possible.",
			Status:                  "Our technician is in your area and is working to fix things.",
			Type:                    "CHANGE",
			TicketNumber:            "C111111111",
			EstimatedResolutionTime: time.Date(2022, 11, 14, 15, 0, 0, 0, time.UTC),
		},
	}, outages)

}

func TestNoOutages(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bytes, err := os.ReadFile("fixtures/working.json")
		require.NoError(t, err)
		w.WriteHeader(http.StatusOK)
		w.Write(bytes)
	}))

	opts := OutageOpts{
		Endpoint: testServer.URL,
		Headers:  map[string]string{},
	}

	outages, err := FetchOutages(opts)
	require.NoError(t, err)

	require.Equal(t, []Outage{}, outages)

}
