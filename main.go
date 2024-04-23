package main

import (
	"fmt"
	"log"
	"time"
)

func main() {
	currentTimeStr := "2021-09-19T15:59:41+03:00"
	layout := "2006-01-02T15:04:05-07:00"
	currentTime, err := time.Parse(layout, currentTimeStr)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(currentTime)
}
