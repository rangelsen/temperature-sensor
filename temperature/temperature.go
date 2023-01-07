package temperature

import (
    "bufio"
    "os"
    "fmt"
    "strconv"
    "time"
    "strings"
)

func ReadTemperatures() {

    readFile, err := os.Open("temperature.txt")
  
    if err != nil {
        fmt.Println(err)
    }

    tempScanner := bufio.NewScanner(readFile)
    tempScanner.Split(bufio.ScanLines)

    ticker := time.NewTicker(time.Millisecond * 100)
  
    for {
        fmt.Println(getTemperature(tempScanner, ticker))
    }

    readFile.Close()
}

func getTemperature(tempScanner *bufio.Scanner, ticker *time.Ticker) float64 {

    <-ticker.C
    tempScanner.Scan()
    tempStr := strings.TrimSpace(tempScanner.Text())

    if temp, err := strconv.ParseFloat(tempStr, 64); err == nil {
        return temp
    } else {
        panic(err)
    }
}

