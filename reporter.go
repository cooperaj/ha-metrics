package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type report struct {
	sensor string
	data   interface{}
}

type reporter struct {
	Prefix     string
	HAEndpoint string
	HAToken    string
	Stop       chan bool

	reports chan report
	client  *http.Client
}

// NewReporter create a reporter instance that will send out metrics to an endpoint
func NewReporter(endpoint string, token string, prefix string) *reporter {
	reporter := reporter{
		Prefix:     prefix,
		HAEndpoint: endpoint,
		HAToken:    token,
		Stop:       make(chan bool),
		reports:    make(chan report, 20),
	}

	return &reporter
}

// Report sends sensor data to the configured HA instance
func (r *reporter) Report(sensor string, data interface{}) {
	report := report{
		sensor: sensor,
		data:   data,
	}

	r.reports <- report
}

func (r *reporter) Run(wg *sync.WaitGroup) {
	r.client = &http.Client{
		Timeout: time.Duration(5 * time.Second),
	}

	wg.Add(1)

	go func() {
		for {
			select {
			case report := <-r.reports:
				json, err := json.Marshal(report.data)
				if err != nil {
					log.Fatalf("Unable to marshal JSON for sensor %s", report.sensor)
				}

				go r.pushReport(report.sensor, json)
			case <-r.Stop:
				wg.Done()
				return
			}
		}
	}()

	log.Println("Started metric reporter")
}

func (r *reporter) pushReport(sensorName string, json []byte) {
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("%s/api/states/sensor.%s%s", r.HAEndpoint, r.Prefix, sensorName),
		bytes.NewReader(json),
	)

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", r.HAToken))

	log.Debugf("Sending request to %s", req.URL)
	resp, err := r.client.Do(req)
	if err != nil {
		log.Fatalf("Error whilst sending metric to %s: %v", r.HAEndpoint, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		log.Errorf("Received %d error code when sending metric %s", resp.StatusCode, sensorName)
	}
}
