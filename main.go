package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"temperature-sensor/temp"
	"time"
)

type (
	OutboundMeasurements struct {
		Measurements               chan temp.TemperatureMeasurement
		missing                    []temp.TemperatureMeasurement
		Quit                       chan bool
		temperatureURL, missingURL string
	}
)

func main() {

	tempScanner, file := getTempScanner("temperature.txt")
	defer file.Close()

	baseURL := "http://localhost:5000"

	tempSensor := temp.NewSensor(tempScanner, time.NewTicker(time.Millisecond*100))
	processor := temp.NewReadingsProcessor(time.Minute * 2)
	outbound := NewOutboundMeasurements(baseURL+"/api/temperature", baseURL+"/api/temperature/missing")

	go processor.Run(outbound.Measurements)
	go tempSensor.Start(processor.Readings)
	go outbound.dispatchMesurements()
	stopWhenDone(tempSensor, outbound, processor)
}

func NewOutboundMeasurements(tempURL, missingURL string) OutboundMeasurements {

	return OutboundMeasurements{
		Measurements:   make(chan temp.TemperatureMeasurement, 1),
		missing:        make([]temp.TemperatureMeasurement, 0),
		Quit:           make(chan bool, 1),
		temperatureURL: tempURL,
		missingURL:     missingURL,
	}
}

func (outbound *OutboundMeasurements) dispatchMesurements() {

	for {
		select {
		case <-outbound.Quit:
			return
		default:
			measurement := <-outbound.Measurements
			outbound.processMeasurement(measurement)
		}
	}
}

func (outbound *OutboundMeasurements) processMeasurement(
	measurement temp.TemperatureMeasurement) {

	fmt.Println("measurement:", measurement)

	json, _ := json.Marshal(measurement)
	postSuccess := postJson(json, outbound.temperatureURL)

	if len(outbound.missing) > 0 {
		outbound.PublishMissing()
	}

	if !postSuccess {
		fmt.Println("Failed post to /api/temperature. Adding to missing")
		outbound.AddMissing(measurement)
	}
}

func postJson(json []byte, url string) bool {

	buf := bytes.NewBuffer(json)
	resp, _ := http.Post(url, "application/json", buf)
	defer resp.Body.Close()

	return resp.StatusCode != http.StatusInternalServerError
}

func getTempScanner(filePath string) (*bufio.Scanner, *os.File) {

	tempFile, err := os.Open(filePath)

	if err != nil {
		panic(err)
	}

	tempScanner := bufio.NewScanner(tempFile)
	tempScanner.Split(bufio.ScanLines)
	return tempScanner, tempFile
}

func (outbound *OutboundMeasurements) AddMissing(measurement temp.TemperatureMeasurement) {

	outbound.missing = append(outbound.missing, measurement)

	if len(outbound.missing) > 10 {
		outbound.missing = outbound.missing[1:11]
	}
}

func (outbound *OutboundMeasurements) PublishMissing() {

	json, _ := json.Marshal(outbound.missing)
	fmt.Println("Publishing missing measurements:", string(json))

	if postJson(json, outbound.missingURL) {
		fmt.Println("Missing successfully published")
		outbound.missing = outbound.missing[:0]
	} else {
		fmt.Println("Missing failed")
	}
}

func stopWhenDone(sensor temp.Sensor, outbound OutboundMeasurements, processor temp.ReadingsProcessor) {

	<-sensor.Quit
	outbound.Quit <- true
	processor.Quit <- true
}
