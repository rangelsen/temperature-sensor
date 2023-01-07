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

func main() {

	temps := getTempsFromFile("temperature.txt")

	tempReadings := make(chan temp.TemperatureReading, 1)
	tempMeasurements := make(chan temp.TemperatureMeasurement, 1)

	go temp.CalcMeasurement(tempReadings, tempMeasurements, time.Second)
	go temp.ReadTemperatures(temps, tempReadings)
	publishMeasurements(tempMeasurements)
}

func publishMeasurements(tempMeasurements <-chan temp.TemperatureMeasurement) {

	for {
		measurement := <-tempMeasurements
		fmt.Println("measurement:", measurement)

		body, _ := json.Marshal(measurement)
		buf := bytes.NewBuffer(body)
		resp, _ := http.Post("http://localhost:5000/api/temperature", "application/json", buf)
		fmt.Println(resp.Status)
		resp.Body.Close()
	}
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
