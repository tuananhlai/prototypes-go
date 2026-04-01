package main

import (
	"os"

	"golang.org/x/sys/unix"
)

func main() {
	if len(os.Args) == 1 {
		panic("no file specified")
	}

	filePath := os.Args[1]

	fd, err := unix.Open(filePath, unix.O_RDONLY, 0)
	if err != nil {
		panic(err)
	}

	buf := make([]byte, 4096)
	for {
		n, err := unix.Read(fd, buf)
		if err != nil {
			panic(err)
		}

		if n == 0 {
			break
		}

		_, _ = unix.Write(unix.Stdout, buf[:n])
	}
}
