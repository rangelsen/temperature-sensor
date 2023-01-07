package temp

import (
    "bufio"
    "os"
    "fmt"
    "strconv"
    "time"
    "strings"
    "math"
)

type (
    TemperatureReading struct {
        Temperature float64
        TimeStamp time.Time
    }

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

func CalcMeasurement(tempReadings <-chan TemperatureReading, tempMeasurements chan<- TemperatureMeasurement, postingInterval time.Duration) {

    var accumulated = TemperatureMeasurement {
        MeasurementTime { time.Now(), time.Now() }, math.MaxFloat64, -math.MaxFloat64, 0,
    }

    var readingCount uint = 0

    for {

        reading := <-tempReadings
        fmt.Println("reading: ", reading)
        accumulated = accumulateReadings(accumulated, reading, readingCount)
        readingCount++

        if shouldPost(accumulated, postingInterval) {

            fmt.Println("Time passed. Posting: ", accumulated)
            tempMeasurements <- accumulated
            readingCount = 0
        }
    }
}

func accumulateReadings(acc TemperatureMeasurement, reading TemperatureReading, readingCount uint) TemperatureMeasurement {

    startTime := acc.Time.Start

    if readingCount == 0 {
        startTime = reading.TimeStamp
    }

    average := (acc.Average * float64(readingCount) + reading.Temperature) / float64(readingCount + 1)

    return TemperatureMeasurement {
        MeasurementTime { startTime, reading.TimeStamp },
        math.Min(acc.Min, reading.Temperature),
        math.Max(acc.Max, reading.Temperature),
        average,
    }
}

func shouldPost(acc TemperatureMeasurement, threshold time.Duration) bool {

    return acc.Time.End.Sub(acc.Time.Start) >= threshold
}

func ReadTemperatures(tempFile *os.File, tempReadings chan<- TemperatureReading) {

    tempScanner := bufio.NewScanner(tempFile)
    tempScanner.Split(bufio.ScanLines)

    ticker := time.NewTicker(time.Millisecond * 100)
  
    for {
        temp := getTemperature(tempScanner, ticker)
        timeStamp := time.Now()

        tempReadings <- TemperatureReading { temp, timeStamp }
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

