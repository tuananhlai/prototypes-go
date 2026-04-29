package main

import (
	"fmt"
	"os"
	"strconv"

	"golang.org/x/sys/unix"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: %s <path>\n", os.Args[0])
		os.Exit(1)
	}
	targetPath := os.Args[1]

	possiblePids, err := os.ReadDir("/proc")
	if err != nil {
		panic(err)
	}

	var retval []int
	buf := make([]byte, 4096)
	for _, possiblePid := range possiblePids {
		pid, err := strconv.Atoi(possiblePid.Name())
		if err != nil {
			// entry is not a procfs directory.
			continue
		}

		fds, err := os.ReadDir(fmt.Sprintf("/proc/%d/fd", pid))
		if err != nil {
			// process might have disappeared.
			continue
		}

		for _, fd := range fds {
			n, err := unix.Readlink(fmt.Sprintf("/proc/%d/fd/%s", pid, fd.Name()), buf)
			if err != nil {
				// process might have disappeared.
				continue
			}

			if string(buf[:n]) == targetPath {
				retval = append(retval, pid)
			}
		}
	}

	fmt.Println(retval)
}
