package main

import (
	"fmt"
	"time"
)

// Андрей родился 26 ноября 1993 года.
// Посчитайте количество дней до его 100-летия — относительно сегодняшнего дня.

func main() {
	birthday := time.Date(2093, time.November, 26, 0, 0, 0, 0, time.Local)
	duration := time.Until(birthday)
	days := int(duration.Hours()) / 24
	fmt.Println(days)
}
