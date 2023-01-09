package temp

import (
	"fmt"
	"math"
	"time"
)

const (
	MAX_RAW_READING = 4095
	MAX_TEMP        = 50
	MIN_TEMP        = -50
)

type (
	TemperatureReading struct {
		Temperature float64
		TimeStamp   time.Time
	}

	TemperatureMeasurement struct {
		Time    MeasurementTime `json:"time"`
		Min     float64         `json:"min"`
		Max     float64         `json:"max"`
		Average float64         `json:"avg"`
	}

	MeasurementTime struct {
		Start time.Time `json:"start"`
		End   time.Time `json:"end"`
	}
)

func CalcMeasurement(tempReadings <-chan TemperatureReading, tempMeasurements chan<- TemperatureMeasurement, postingInterval time.Duration) {

	initial := TemperatureMeasurement{
		MeasurementTime{time.Now().UTC(), time.Now().UTC()}, math.MaxFloat64, -math.MaxFloat64, 0,
	}

	var accumulated TemperatureMeasurement = initial
	var readingCount uint = 0

	for {

		reading := <-tempReadings
		fmt.Println("reading: ", reading)
		accumulated = accumulateReadings(accumulated, reading, readingCount)
		readingCount++

		if shouldPublish(accumulated, postingInterval) {

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

	average := (acc.Average*float64(readingCount) + reading.Temperature) / float64(readingCount+1)

	min := math.Min(acc.Min, reading.Temperature)
	max := math.Max(acc.Max, reading.Temperature)

	return TemperatureMeasurement{
		MeasurementTime{startTime, reading.TimeStamp},
		math.Round(min) / 100,
		math.Round(max) / 100,
		math.Round(average) / 100,
	}
}

func shouldPublish(acc TemperatureMeasurement, threshold time.Duration) bool {
	return acc.Time.End.Sub(acc.Time.Start) >= threshold
}

func ReadTemperatures(temps []uint, tempReadings chan<- TemperatureReading) {

	ticker := time.NewTicker(time.Millisecond * 100)

	getTemperature := func() float64 {
		<-ticker.C
		temp := temps[0]
		temps = temps[1:]
		return rawTempToFloat(temp)
	}

	for range temps {

		temp := getTemperature()
		timeStamp := time.Now().UTC()

		tempReadings <- TemperatureReading{temp, timeStamp}
	}
}

func rawTempToFloat(raw uint) float64 {
	return lerp(float64(raw)/MAX_RAW_READING, MIN_TEMP, MAX_TEMP)
}

func lerp(val float64, min float64, max float64) float64 {
	return val*(max-min) + min
}
