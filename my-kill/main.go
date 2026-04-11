package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"golang.org/x/sys/unix"
)

func main() {
	var sig int
	flag.IntVar(&sig, "s", int(unix.SIGINT), "")
	flag.Parse()

	pidStrs := flag.Args()

	for _, pidStr := range pidStrs {
		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error converting pid %s to int: %v", pidStr, err)
			continue
		}
		err = unix.Kill(pid, unix.Signal(sig))
		if err != nil {
			fmt.Fprintf(os.Stderr, "error sending signals to pid %d: %v", pid, err)
		}
	}
}
