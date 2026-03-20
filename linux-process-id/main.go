package main

import (
	"fmt"

	"golang.org/x/sys/unix"
)

func main() {
	jobName := "demoSetpgid"

	switch jobName {
	case "demoSetsid":
		demoSetsid()
	case "demoSetpgid":
		demoSetpgid()
	default:
		fmt.Println("jobName not specified or invalid")
	}
}

func demoSetpgid() {
	printIDs("before setpgid")
	err := unix.Setpgid(0, 0)
	if err != nil {
		panic(err)
	}
	printIDs("after setpgid")
}

func demoSetsid() {
	printIDs("before setsid")
	_, err := unix.Setsid()
	if err != nil {
		panic(err)
	}
	printIDs("after setsid")
}

func printIDs(prefix string) {
	pid := unix.Getpid()
	ppid := unix.Getppid()
	pgid := unix.Getpgrp()

	// get current process's session ID
	sid, err := unix.Getsid(0)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s\t\tpid=%d ppid=%d pgid=%d sid=%d\n", prefix, pid, ppid, pgid, sid)
}
