package health_check_client

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

var configYaml string = `
- headers:
    user-agent: fetch-synthetic-monitor
  method: GET
  name: fetch index page
  url: https://fetch.com/
- headers:
    user-agent: fetch-synthetic-monitor
  method: GET
  name: fetch careers page
  url: https://fetch.com/careers
- body: '{"foo":"bar"}'
  headers:
    content-type: application/json
    user-agent: fetch-synthetic-monitor
  method: POST
  name: fetch some fake post endpoint
  url: https://fetch.com/some/post/endpoint
- name: fetch rewards index page
  url: https://www.fetchrewards.com/
  `

var expected_endpoints = []endpoint{
	{
		Name:   "fetch index page",
		Url:    "https://fetch.com/",
		Method: "GET",
		Headers: map[string]string{
			"user-agent": "fetch-synthetic-monitor",
		},
	},
	{
		Name:   "fetch careers page",
		Url:    "https://fetch.com/careers",
		Method: "GET",
		Headers: map[string]string{
			"user-agent": "fetch-synthetic-monitor",
		},
	},
	{
		Name:   "fetch some fake post endpoint",
		Url:    "https://fetch.com/some/post/endpoint",
		Method: "POST",
		Body:   `{"foo":"bar"}`,
		Headers: map[string]string{
			"content-type": "application/json",
			"user-agent":   "fetch-synthetic-monitor",
		},
	},
	{
		Name: "fetch rewards index page",
		Url:  "https://www.fetchrewards.com/",
	},
}

func TestParseEndpointConfig(t *testing.T) {
	endpoints := parseEndpointConfig([]byte(configYaml))

	if !reflect.DeepEqual(expected_endpoints, endpoints) {
		t.Errorf("\nendpoints = \n%+v, \nwant \n%+v", endpoints, expected_endpoints)
	}
}

func TestPing(t *testing.T) {
	// create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	endpoints := []endpoint{
		{Name: "MockEndpoint", Url: server.URL},
	}

	hc := &HealthCheckClient{
		httpClient:   server.Client(),
		endpoints:    endpoints,
		timeInterval: 1,
		stats:        make(map[string]map[string]int),
	}

	for _, e := range hc.endpoints {
		hc.stats[e.Name] = make(map[string]int)
	}

	err := hc.ping()
	if err != nil {
		t.Fatalf("ping returned an error: %v", err)
	}

	if hc.stats["MockEndpoint"]["up"] != 1 {
		t.Errorf("expected 'up' count to be 1, got %d", hc.stats["MockEndpoint"]["up"])
	}
}
