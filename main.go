package main

import (
	"fmt"
	"time"
)

func main() {
	start := time.Now()
	ticker := time.NewTicker(2 * time.Second)
	for i := 0; i < 10; i++ {
		t := <-ticker.C
		res := t.Sub(start).Seconds()
		fmt.Println(int(res))
	}
}
