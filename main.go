package main

import (
    "os"
    "temperature-sensor/temp"
)

func main() {

    tempFile, err := os.Open("temperature.txt")
    defer tempFile.Close()
  
    if err != nil {
        panic(err)
    }

    tempReadings := make(chan float64, 1)
    tempMeasurements := make(chan temp.TemperatureMeasurement, 1)

    go temp.CalcMeasurement(tempReadings, tempMeasurements)
    temp.ReadTemperatures(tempFile, tempReadings)
}

