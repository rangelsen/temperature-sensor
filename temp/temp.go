package temp

import (
    "fmt"
    "time"
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

    initial := TemperatureMeasurement {
        MeasurementTime { time.Now(), time.Now() }, math.MaxFloat64, -math.MaxFloat64, 0,
    }

    var accumulated TemperatureMeasurement = initial

    var readingCount uint = 0

    for {

        reading := <-tempReadings
        fmt.Println("reading: ", reading)
        accumulated = accumulateReadings(accumulated, reading, readingCount)
        readingCount++

        if shouldPost(accumulated, postingInterval) {

            tempMeasurements <- accumulated
            readingCount = 0
            accumulated = initial
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

func ReadTemperatures(temps []float64, tempReadings chan<- TemperatureReading) {

    ticker := time.NewTicker(time.Millisecond * 100)
  
    getTemperature := func() float64 {
        <-ticker.C
        temp := temps[0]
        temps = temps[1:]
        return temp
    }

    for range temps {

        temp := getTemperature()
        timeStamp := time.Now()

        tempReadings <- TemperatureReading { temp, timeStamp }
    }
}

