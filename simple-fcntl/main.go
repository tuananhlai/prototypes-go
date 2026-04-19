package main

import (
	"fmt"
	"time"

	"golang.org/x/sys/unix"
)

func main() {
	// Create a unique temp file path.
	outputFilePath := fmt.Sprintf("/tmp/test%v.txt", time.Now().UnixMicro())

	// Open the file for read/write, creating it if needed.
	fd, err := unix.Open(outputFilePath, unix.O_RDWR|unix.O_CREAT, 0o644)
	if err != nil {
		panic(err)
	}
	defer unix.Close(fd)

	// Write the first line at the current file offset.
	_, err = unix.Write(fd, []byte("first line\n"))
	if err != nil {
		panic(err)
	}

	// Rewind the file offset back to the start.
	_, err = unix.Seek(fd, 0, unix.SEEK_SET)
	if err != nil {
		panic(err)
	}

	// Read the current open-file status flags.
	flags, err := unix.FcntlInt(uintptr(fd), unix.F_GETFL, 0)
	if err != nil {
		panic(err)
	}

	// Enable append mode on this open file description.
	_, err = unix.FcntlInt(uintptr(fd), unix.F_SETFL, flags|unix.O_APPEND)
	if err != nil {
		panic(err)
	}

	// This write now goes to the end of the file, not the current offset.
	_, err = unix.Write(fd, []byte("second line\n"))
	if err != nil {
		panic(err)
	}

	fmt.Printf("run `cat %s` to see result.\n", outputFilePath)
}
