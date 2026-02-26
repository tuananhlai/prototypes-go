package main

import (
	"fmt"
	"os"
	"syscall"
)

func main() {
	_, err := os.Open("nonexistent.txt")
	if err != nil {
		pathErr := err.(*os.PathError)
		if errno, ok := pathErr.Err.(syscall.Errno); ok {
			fmt.Println("Raw errno:", errno)
			fmt.Println("Numeric value:", int(errno))
			fmt.Println("errno == syscall.ENOENT:", errno == syscall.ENOENT)
		}
	}
}
