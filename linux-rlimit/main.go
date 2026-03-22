package main

import (
	"fmt"

	"golang.org/x/sys/unix"
)

func main() {
	fmt.Println("Default RLIMIT values")
	printRLimits()

	fmt.Println()
	// demoIncreaseHardLimit()
	demoIncreaseSoftLimit()

	fmt.Println("Updated RLIMIT values")
	printRLimits()
}

func demoIncreaseSoftLimit() {
	err := unix.Setrlimit(unix.RLIMIT_NOFILE, &unix.Rlimit{
		Cur: 10000,
		Max: 60000,
	})
	if err != nil {
		panic(err)
	}

	err = unix.Setrlimit(unix.RLIMIT_NOFILE, &unix.Rlimit{
		Cur: 60000,
		Max: 60000,
	})
	if err != nil {
		panic(err)
	}
}

// Only work when this program is run as (real) root. Will throw error otherwise due to permission issues.
func demoIncreaseHardLimit() {
	err := unix.Setrlimit(unix.RLIMIT_NOFILE, &unix.Rlimit{
		Cur: 60000,
		Max: 60000,
	})
	if err != nil {
		panic(err)
	}

	err = unix.Setrlimit(unix.RLIMIT_NOFILE, &unix.Rlimit{
		Cur: 60000,
		Max: 61000,
	})
	if err != nil {
		panic(err)
	}
}

func printRLimits() {
	limits := []struct {
		name     string
		resource int
	}{
		{"RLIMIT_CPU", unix.RLIMIT_CPU},
		{"RLIMIT_FSIZE", unix.RLIMIT_FSIZE},
		{"RLIMIT_DATA", unix.RLIMIT_DATA},
		{"RLIMIT_STACK", unix.RLIMIT_STACK},
		{"RLIMIT_CORE", unix.RLIMIT_CORE},
		{"RLIMIT_RSS", unix.RLIMIT_RSS},
		{"RLIMIT_NPROC", unix.RLIMIT_NPROC},
		{"RLIMIT_NOFILE", unix.RLIMIT_NOFILE},
		{"RLIMIT_MEMLOCK", unix.RLIMIT_MEMLOCK},
		{"RLIMIT_AS", unix.RLIMIT_AS},
		{"RLIMIT_LOCKS", unix.RLIMIT_LOCKS},
		{"RLIMIT_SIGPENDING", unix.RLIMIT_SIGPENDING},
		{"RLIMIT_MSGQUEUE", unix.RLIMIT_MSGQUEUE},
		{"RLIMIT_NICE", unix.RLIMIT_NICE},
		{"RLIMIT_RTPRIO", unix.RLIMIT_RTPRIO},
		{"RLIMIT_RTTIME", unix.RLIMIT_RTTIME},
	}

	var rlim unix.Rlimit
	for _, limit := range limits {
		if err := unix.Getrlimit(limit.resource, &rlim); err != nil {
			panic(err)
		}
		fmt.Printf("%s soft=%s hard=%s\n", limit.name, formatLimit(rlim.Cur), formatLimit(rlim.Max))
	}
}

func formatLimit(v uint64) string {
	if v == unix.RLIM_INFINITY {
		return "infinity"
	}
	return fmt.Sprintf("%d", v)
}
