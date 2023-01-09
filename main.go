package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"temperature-sensor/temp"
	"time"
)

type (
	OutboundMeasurements struct {
		Measurements chan temp.TemperatureMeasurement
		missing      []temp.TemperatureMeasurement
	}

	Serializable interface {
		Json() []byte
	}
)

func main() {

	temps := getTempsFromFile("temperature.txt")

	tempReadings := make(chan temp.TemperatureReading, 1)
	outbound := OutboundMeasurements{
		make(chan temp.TemperatureMeasurement, 1),
		make([]temp.TemperatureMeasurement, 0),
	}

	go temp.CalcMeasurement(tempReadings, outbound.Measurements, time.Second)
	go temp.ReadTemperatures(temps, tempReadings)
	publishMeasurements(&outbound)
}

func publishMeasurements(outbound *OutboundMeasurements) {

	for {
		measurement := <-outbound.Measurements
		fmt.Println("measurement:", measurement)

		json, _ := json.Marshal(measurement)
		postSuccess := postJson(json, "http://localhost:5000/api/temperature")
		outbound.PublishMissing()

		if !postSuccess {
			fmt.Println("Failed post to /api/temperature. Adding to missing")
			outbound.AddMissing(measurement)
		}
	}
}

func postJson(json []byte, url string) bool {

	buf := bytes.NewBuffer(json)
	resp, _ := http.Post(url, "application/json", buf)
	defer resp.Body.Close()

	return resp.StatusCode != 500
}

func getTempsFromFile(filePath string) []uint {

	tempFile, err := os.Open(filePath)
	defer tempFile.Close()

	if err != nil {
		panic(err)
	}

	tempScanner := bufio.NewScanner(tempFile)
	tempScanner.Split(bufio.ScanLines)
	return fileContentToUintSlice(tempScanner)
}

func fileContentToUintSlice(tempScanner *bufio.Scanner) []uint {

	var temps []uint

	for tempScanner.Scan() {

		tempStr := strings.TrimSpace(tempScanner.Text())

		if temp, err := strconv.ParseUint(tempStr, 10, 16); err == nil {
			temps = append(temps, uint(temp))
		} else {
			panic(err)
		}
	}

	return temps
}

func (outbound *OutboundMeasurements) AddMissing(measurement temp.TemperatureMeasurement) {

	outbound.missing = append(outbound.missing, measurement)

	if len(outbound.missing) > 10 {
		outbound.missing = outbound.missing[1:11]
	}
}

func (outbound *OutboundMeasurements) PublishMissing() {

	if len(outbound.missing) == 0 {
		return
	}

	json, _ := json.Marshal(outbound.missing)
	fmt.Println("Publishing missing measurements:", string(json))

	if postJson(json, "http://localhost:5000/api/temperature/missing") {
		fmt.Println("Missing successfully published")
		outbound.missing = outbound.missing[:0]
	} else {
		fmt.Println("Missing failed")
	}
}
