package main

import (
	"fmt"

	"github.com/tuananhlai/prototypes/detect-race-condition/counter"
)

func main() {
	count := counter.CountToFour()
	fmt.Printf("count=%d\n", count)
}
