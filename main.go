package main

import (
	"fmt"
	"time"
)

func main() {
	now := time.Now()
	fmt.Println(now.Format("Mon, 02 Jan 2006 15:04:05 MST")) // Sun, 19 Sep 2021 15:42:00 MSK
}
