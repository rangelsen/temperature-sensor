package main

import (
	"bufio"
	"fmt"
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

	for {
		measurement := <-tempMeasurements
		fmt.Println("measurement:", measurement)
	}
}

func getTempsFromFile(filePath string) []float64 {

	tempFile, err := os.Open(filePath)
	defer tempFile.Close()

	if err != nil {
		panic(err)
	}

	tempScanner := bufio.NewScanner(tempFile)
	tempScanner.Split(bufio.ScanLines)

	return fileContentToFloat64Slice(tempScanner)
}

func fileContentToFloat64Slice(tempScanner *bufio.Scanner) []float64 {

	var temps []float64

	for tempScanner.Scan() {

		tempStr := strings.TrimSpace(tempScanner.Text())

		if temp, err := strconv.ParseInt(tempStr, 10, 32); err == nil {
			temps = append(temps, float64(temp))
		} else {
			panic(err)
		}
	}

	return temps
}
