package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
	f, err := os.CreateTemp("", "demo-*.txt")
	if err != nil {
		fmt.Fprintln(os.Stderr, "create temp file:", err)
		os.Exit(1)
	}
	defer f.Close()

	path := f.Name()
	fmt.Println("temp file:", path)
	fmt.Println("pid:", os.Getpid())

	// Stay alive indefinitely until killed.
	for {
		time.Sleep(time.Hour)
	}
}
