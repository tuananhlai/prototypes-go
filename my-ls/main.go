package main

import (
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

func main() {
	var wd string
	var err error
	if len(os.Args) == 1 {
		wd, err = unix.Getwd()
	} else {
		wd = os.Args[1]
	}

	if err != nil {
		panic(err)
	}

	fd, err := unix.Open(wd, unix.O_RDONLY, 0o644)
	if err != nil {
		panic(err)
	}
	defer unix.Close(fd)

	buf := make([]byte, 4096)
	for {
		// Read the next directory entries from the cwd's file descriptor.
		// The method tries to read as many **complete** directory entries into
		// buf as possible before returning.
		//
		// Regular `Read` system call will return EISDIR
		n, err := unix.ReadDirent(fd, buf)
		if err != nil {
			panic(err)
		}
		if n == 0 {
			break
		}

		// Parse the raw bytes returned by `ReadDirent` into a slice of entry names.
		_, _, names := unix.ParseDirent(buf[:n], -1, nil)
		for _, name := range names {
			fmt.Println(name)
		}
	}
}
