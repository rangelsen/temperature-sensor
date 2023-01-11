package temp

import (
    "testing"
    "time"
)

type (
    accumAvgTest struct {
        avg float64
        val float64
        n uint
        res float64
    }
)

func TestReadingsProcessorRun(t *testing.T) {

    // Arrange
    readings := make(chan TemperatureReading, 2)
    measurements := make(chan TemperatureMeasurement, 1)
    quit := make(chan bool, 1)

    processor := ReadingsProcessor {
        Readings: readings,
        PublishingInterval: time.Minute*10,
        Quit: quit,
    }

    r1 := TemperatureReading{8.42, time.Now().UTC()}
    r2 := TemperatureReading{-3.28, r1.TimeStamp.Add(time.Minute*10)} 
    readings <- r1
    readings <- r2

    average := (r1.Temperature + r2.Temperature) / 2
    expected := TemperatureMeasurement{
        Time: MeasurementTime{
            Start: r1.TimeStamp,
            End: r2.TimeStamp,
        },
        Min: r2.Temperature,
        Max: r1.Temperature,
        Average: round2(average),
    }

    // Act
    go processor.Run(measurements)
    actual := <-measurements
    quit <- true

    // Assert
    if expected != actual {
        t.Errorf("expected: %+v, actual: %+v", expected, actual)
    }
}

func TestAccumulateAverage(t *testing.T) {

    // Arrange
    testData := []accumAvgTest{
        accumAvgTest{0, 1, 0, 1},
        accumAvgTest{-2, 2, 1, 0},
        accumAvgTest{8.4, 3.2, 4, 7.36},
    }

    // Act
    for _, d := range testData {

        expected := d.res
        actual := accumulateAverage(d.avg, d.val, d.n)

        // Assert
        if expected != actual {
            t.Errorf("expected: %+v, actual: %+v", expected, actual)
        }
    }
}
