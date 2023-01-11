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
		Measurements chan temp.TemperatureMeasurement
		missing      []temp.TemperatureMeasurement
        Quit chan bool
	}
)

func main() {

	tempScanner, file := getTempScanner("temperature.txt")
	defer file.Close()

	readings := make(chan temp.TemperatureReading, 1)
    quit := make(chan bool, 1)

	tempSensor := temp.Sensor{
		TempSource: tempScanner,
		Ticker:     time.NewTicker(time.Millisecond * 10),
        Quit: quit,
	}

	processor := temp.ReadingsProcessor{
		Readings:           readings,
		PublishingInterval: time.Second * 2,
        Quit: make(chan bool, 1),
	}

	outbound := OutboundMeasurements{
		make(chan temp.TemperatureMeasurement, 1),
		make([]temp.TemperatureMeasurement, 0),
        make(chan bool, 1),
	}

	go processor.Run(outbound.Measurements)
	go tempSensor.Start(readings)
	go publishMeasurements(&outbound)
    stop(quit, &outbound, &processor)
}

func publishMeasurements(outbound *OutboundMeasurements) {

	for {
        select {
        case <-outbound.Quit:
            return
        default:
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
}

func postJson(json []byte, url string) bool {

	buf := bytes.NewBuffer(json)
	resp, _ := http.Post(url, "application/json", buf)
	defer resp.Body.Close()

	return resp.StatusCode != 500
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

func stop(quit <-chan bool, outbound *OutboundMeasurements, processor *temp.ReadingsProcessor) {

    <-quit
    outbound.Quit <- true
    processor.Quit <- true
}
