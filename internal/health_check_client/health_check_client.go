package health_check_client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/myshkins/fetch_takehome/internal/logger"
)

type endpoint struct {
	Name    string            `yaml:"name"`
	Url     string            `yaml:"url"`
	Method  string            `yaml:"method,omitempty"`
	Body    string            `yaml:"body,omitempty"`
	Headers map[string]string `yaml:"headers,omitempty"`
}

type HealthCheckClient struct {
	httpClient   *http.Client
	endpoints    []endpoint
	timeInterval int
	stats        map[string]map[string]int // eg. {"endpoint.Name":{"up":5}}
}

func NewHealthCheckClient(fp string, t int) *HealthCheckClient {
	var hc HealthCheckClient
	hc.httpClient = &http.Client{
		Timeout: 500 * time.Millisecond,
	}

	data, err := os.ReadFile(fp)
	if err != nil {
		logger.Logger.Error(err.Error())
		os.Exit(1)
	}
	hc.endpoints = parseEndpointConfig(data)
	hc.timeInterval = t
	hc.stats = make(map[string]map[string]int)

	// initialize inner map for each endpoint
	for _, e := range hc.endpoints {
		hc.stats[e.Name] = make(map[string]int)
	}
	return &hc
}

func parseEndpointConfig(data []byte) []endpoint {
	endpoints := []endpoint{}

	err := yaml.Unmarshal(data, &endpoints)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	return endpoints
}

func (hc *HealthCheckClient) printStats() {
	for _, v := range hc.endpoints {
		total := hc.stats[v.Name]["up"] + hc.stats[v.Name]["down"]
		a := (float64(hc.stats[v.Name]["up"]) / float64(total)) * float64(100)
		a = math.Round(a)
		logger.Logger.Info("availablility", v.Name, a)
	}
}

func (hc *HealthCheckClient) PingEndpoints() {
	ticker := time.NewTicker(time.Duration(hc.timeInterval) * time.Second)
	defer ticker.Stop()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	msg := fmt.Sprintf("Starting endpoint healthcheck. Pinging every %d seconds. Press Ctrl+C to stop", hc.timeInterval)
	logger.Logger.Info(msg)

	err := hc.ping()
	if err != nil {
		logger.Logger.Error("error pinging endpoints", "error", err)
		// log.Print("error pinging endpoints: ", err)
	}

	for {
		select {
		case <-ticker.C:
			err := hc.ping()
			if err != nil {
				logger.Logger.Error("error pinging endpoints", "error", err)
				// log.Printf("error pinging endpoints: %v", err)
			}
		case <-sigChan:
			logger.Logger.Info("Received interrupt signal. Exiting..")
			return
		}
	}
}

func formRequestBody(endpoint endpoint) io.Reader {
  if endpoint.Body == "" {
    return nil
  }

  var j map[string]interface{}
  err := json.Unmarshal([]byte(endpoint.Body), &j)
  if err != nil {
    logger.Logger.Error("error parsing JSON body", endpoint.Name, err)
    return nil
  }

  b, err := json.Marshal(j)
  if err != nil {
    logger.Logger.Error("error marshalling JSON body", endpoint.Name, err)
  }
	return bytes.NewBuffer(b)
}

func formRequest(endpoint endpoint) *http.Request {
	method := "GET"
	body := formRequestBody(endpoint)
	if endpoint.Method != "" {
		method = endpoint.Method
	}
	req, err := http.NewRequest(method, endpoint.Url, body)
	if err != nil {
		log.Printf("error forming request: %v", err)
	}

	if endpoint.Headers != nil {
		for k, v := range endpoint.Headers {
			req.Header.Add(k, v)
		}
	}
	return req
}

func (hc *HealthCheckClient) ping() error {
	var wg sync.WaitGroup

	for _, ep := range hc.endpoints {
		wg.Add(1)
		go func(e endpoint) {
			defer wg.Done()
			req := formRequest(e)

			resp, err := hc.httpClient.Do(req)
			if err != nil {
				hc.stats[e.Name]["down"]++
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				hc.stats[e.Name]["up"]++
			} else {
				hc.stats[e.Name]["down"]++
			}
			_, err = io.Copy(io.Discard, resp.Body)
			if err != nil {
				logger.Logger.Error("error draining body", "error", err)
			}
		}(ep)
	}
	wg.Wait()
	hc.printStats()
	fmt.Println("...")
	return nil
}
