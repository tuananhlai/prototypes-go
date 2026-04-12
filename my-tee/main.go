package main

import (
	"flag"
	"fmt"

	"golang.org/x/sys/unix"
)

func main() {
	var appendOutput bool
	flag.BoolVar(&appendOutput, "a", false, "")
	flag.Parse()

	openFlags := unix.O_WRONLY | unix.O_CREAT
	if appendOutput {
		openFlags |= unix.O_APPEND
	} else {
		openFlags |= unix.O_TRUNC
	}

	outputFilePaths := flag.Args()
	outputFds := make([]int, 0, len(outputFilePaths)+1)

	for _, path := range outputFilePaths {
		fd, err := unix.Open(path, openFlags,
			unix.S_IRUSR|unix.S_IWUSR|unix.S_IRGRP|unix.S_IROTH)
		if err != nil {
			panic(fmt.Sprintf("error opening path %s: %v", path, err))
		}
		defer unix.Close(fd)

		outputFds = append(outputFds, fd)
	}
	outputFilePaths = append(outputFilePaths, "stdout")
	outputFds = append(outputFds, unix.Stdout)

	buf := make([]byte, 4096)

	for {
		numByteRead, err := unix.Read(unix.Stdin, buf)
		if err != nil {
			panic(fmt.Sprintf("error reading from stdin: %v", err))
		}

		for i, fd := range outputFds {
			numByteWritten, err := unix.Write(fd, buf[:numByteRead])
			if err != nil {
				panic(fmt.Sprintf("error writing to %s: %v", outputFilePaths[i], err))
			}
			if numByteWritten != numByteRead {
				panic(fmt.Sprintf("error writing all bytes to %s", outputFilePaths[i]))
			}
		}
	}
}
