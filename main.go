package main

import (
	"fmt"
	"time"
)

func main() {
	now := time.Now()
	truncTime := now.Truncate(time.Hour * 24)
	fmt.Println(truncTime)
}
