package main

import (
    "os"
    "temperature-sensor/temp"
    "time"
)

func main() {

    tempFile, err := os.Open("temperature.txt")
    defer tempFile.Close()
  
    if err != nil {
        panic(err)
    }

    tempReadings := make(chan temp.TemperatureReading, 1)
    tempMeasurements := make(chan temp.TemperatureMeasurement, 1)

    go temp.CalcMeasurement(tempReadings, tempMeasurements, time.Second)
    temp.ReadTemperatures(tempFile, tempReadings)
}

