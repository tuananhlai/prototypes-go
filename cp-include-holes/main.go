package main

import (
	"flag"
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

func main() {
	flag.Parse()

	if flag.NArg() <= 1 {
		fatalf("invalid number of arguments: expect 2, got %d\n", flag.NArg())
	}

	src, dst := flag.Arg(0), flag.Arg(1)
	srcFd, err := unix.Open(src, unix.O_RDONLY, 0)
	if err != nil {
		fatalf("opening %s: %v\n", src, err)
	}
	defer unix.Close(srcFd)

	// retrieve source file stat to extract file size.
	var stat unix.Stat_t
	err = unix.Fstat(srcFd, &stat)
	if err != nil {
		fatalf("fstat: %v\n", err)
	}

	// Assume destination file path doesn't exist.
	dstFd, err := unix.Open(dst, unix.O_WRONLY|unix.O_CREAT|unix.O_TRUNC, 0o644)
	if err != nil {
		fatalf("opening %s: %v\n", dst, err)
	}
	defer unix.Close(dstFd)

	// match length of source file
	err = unix.Ftruncate(dstFd, stat.Size)
	if err != nil {
		fatalf("ftruncate: %v\n", err)
	}

	dataOffset := int64(0)
	holeOffset := int64(0)

	for {
		// retrieve the next offset for a data region immediately after `holeOffset`.
		dataOffset, err = unix.Seek(srcFd, holeOffset, unix.SEEK_DATA)
		if err != nil {
			// no more data left
			if err == unix.ENXIO {
				break
			}
			fatalf("seek data: %v\n", err)
		}

		// retrieve the next offset for a file hole immediately after `dataOffset`.
		holeOffset, err = unix.Seek(srcFd, dataOffset, unix.SEEK_HOLE)
		if err != nil {
			fatalf("seek hole: %v\n", err)
		}

		// copy the data region between the two offset calculated earlier.
		_, err = unix.Sendfile(dstFd, srcFd, &dataOffset, int(holeOffset)-int(dataOffset))
		if err != nil {
			fatalf("sendfile: %v\n", err)
		}
	}
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}
