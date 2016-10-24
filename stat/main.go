package main

import (
	"fmt"

	"github.com/montanaflynn/stats"
)

func main() {
	tpArray := []float64{7, 15, 36, 39, 40}

	a4, _ := stats.Percentile(tpArray, 20)
	b4, _ := stats.Percentile(tpArray, 40)
	c4, _ := stats.Percentile(tpArray, 60)
	d4, _ := stats.Percentile(tpArray, 80)
	fmt.Printf("%v %v %v %v", a4, b4, c4, d4) // {15 37.5 40}}
}
