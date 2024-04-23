package main

import (
	"fmt"
	"log"
	"time"
)

func main() {
	currentTimeStr := "2021-09-19T15:59:41+03:00"
	layout := time.RFC3339
	currentTime, err := time.Parse(layout, currentTimeStr)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(currentTime)
}
