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
	// Assume destination file path doesn't exist.
	dstFd, err := unix.Open(dst, unix.O_WRONLY|unix.O_CREAT, 0o644)
	if err != nil {
		fatalf("opening %s: %v\n", dst, err)
	}

	var stat unix.Stat_t
	err = unix.Fstat(srcFd, &stat)
	if err != nil {
		fatalf("fstat: %v\n", err)
	}

	_, err = unix.Sendfile(dstFd, srcFd, nil, int(stat.Size))
	if err != nil {
		fatalf("sendfile: %v\n", err)
	}
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}
