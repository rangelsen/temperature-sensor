package temp

import (
    "bufio"
    "os"
    "fmt"
    "strconv"
    "time"
    "strings"
)

type (
    TemperatureMeasurement struct {
        Time MeasurementTime    `json:"time"`
        Min float64             `json:"min"`
        Max float64             `json:"max"`
        Average float64         `json:"average"`

    }

    MeasurementTime struct {
        Start time.Time `json:"start"`
        End time.Time   `json:"end"`
    }
)

func CalcMeasurement(tempReadings <-chan float64, tempMeasurements chan<- TemperatureMeasurement) {
    for {
        reading := <-tempReadings
        fmt.Println("reading: ", reading)
    }
}

func ReadTemperatures(tempFile *os.File, tempReadings chan<- float64) {

    tempScanner := bufio.NewScanner(tempFile)
    tempScanner.Split(bufio.ScanLines)

    ticker := time.NewTicker(time.Millisecond * 100)
  
    for {
        tempReadings <- getTemperature(tempScanner, ticker)
    }
}

func getTemperature(tempScanner *bufio.Scanner, ticker *time.Ticker) float64 {

    // block until ticker signals ready
    <-ticker.C
    tempScanner.Scan()
    tempStr := strings.TrimSpace(tempScanner.Text())

    if temp, err := strconv.ParseInt(tempStr, 10, 32); err == nil {
        return float64(temp)
    } else {
        panic(err)
    }
}

