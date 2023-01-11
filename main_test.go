package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"temperature-sensor/temp"
	"testing"
	"time"
)

func Test_processMeasurement_ShouldPostHttp(t *testing.T) {

	// Arrange
	measurement := temp.TemperatureMeasurement{
		Time: temp.MeasurementTime{
			Start: time.Now().UTC(),
			End:   time.Now().UTC().Add(time.Minute * 2),
		},
		Min:     -2.82,
		Max:     23,
		Average: 11,
	}

	var testServerHit = false

	// create http test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		testServerHit = true
		receivedMeasurement := getMeasurementFromHttpRequest(r)

		if receivedMeasurement != measurement {
			t.Errorf("received measurement: %+v, expected: %+v",
				receivedMeasurement, measurement)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	outbound := OutboundMeasurements{
		Measurements:   make(chan temp.TemperatureMeasurement, 1),
		missing:        make([]temp.TemperatureMeasurement, 0),
		Quit:           make(chan bool, 1),
		temperatureURL: server.URL,
		missingURL:     "",
	}

	// Act
	outbound.processMeasurement(measurement)

	// Assert
	if !testServerHit {
		t.Errorf("No request was made")
	}
}

func getMeasurementFromHttpRequest(r *http.Request) temp.TemperatureMeasurement {

	bodyBytes, _ := io.ReadAll(r.Body)
	var receivedMeasurement temp.TemperatureMeasurement
	json.Unmarshal(bodyBytes, &receivedMeasurement)
	r.Body.Close()

	return receivedMeasurement
}
