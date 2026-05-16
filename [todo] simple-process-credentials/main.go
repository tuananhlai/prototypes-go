package main

import (
	"fmt"

	"golang.org/x/sys/unix"
)

func main() {
	printUID()
	// fmt.Println("setuid(1000)")
	// err := unix.Setuid(1000)
	fmt.Println("setreuid(-1, 1000)")
	err := unix.Setreuid(-1, 1000)
	if err != nil {
		panic(err)
	}
	printUID()
}

func printUID() {
	ruid, euid, suid := unix.Getresuid()
	fmt.Printf("real = %d, effective = %d, saved = %d\n", ruid, euid, suid)
}
