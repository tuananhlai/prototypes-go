package main

import (
	"fmt"

	"golang.org/x/sys/unix"
)

func main() {
	fmt.Println("== demoReadRegularFile ==")
	demoReadRegularFile()

	fmt.Println("== demoTooManyOpenFiles ==")
	demoTooManyOpenFiles()

	fmt.Println("== demoWriteStdin ==")
	demoWriteStdin()

	fmt.Println("== demoExit ==")
	demoPanic("this is a sample error\n")
}

func demoWriteStdin() {
	n, err := unix.Write(unix.Stdin, []byte("Look ma, I'm writing to STDIN\n"))
	// err will be non-nil if stdin is read-only (like when stdin is piped from another command)
	fmt.Println(n, err)
}

func demoPanic(msg string) {
	unix.Write(unix.Stderr, []byte(msg))
	unix.Exit(1)

	fmt.Println("this will not be run")
}

func demoReadRegularFile() {
	fd, err := unix.Open("hello.txt", unix.O_RDONLY, 0)
	if err != nil {
		panic(err)
	}

	buf := make([]byte, 256)

	for {
		n, err := unix.Read(fd, buf)
		if err != nil {
			panic(err)
		}
		if n == 0 {
			break
		}

		_, err = unix.Write(unix.Stdout, buf[:n])
		if err != nil {
			panic(err)
		}
	}

	err = unix.Close(fd)
	if err != nil {
		panic(err)
	}
}

func demoTooManyOpenFiles() {
	if err := unix.Setrlimit(unix.RLIMIT_NOFILE, &unix.Rlimit{
		Cur: 10,
		Max: 10,
	}); err != nil {
		panic(err)
	}

	var fileDescriptors []int

	// Current process has 3 file descriptors by default (stdin, stdout, stderr)
	// So we can open 7 more file descriptors before we reach soft limit.
	for range 7 {
		fd, err := unix.Open("/dev/null", unix.O_RDONLY, 0)
		if err != nil {
			panic(err)
		}
		fileDescriptors = append(fileDescriptors, fd)
	}

	// Since we've already opened 10 file descriptors at this point, we shouldn't
	// be able to open more.
	_, err := unix.Open("/dev/null", unix.O_RDONLY, 0)
	if err == nil {
		panic("Should not be able to open file descriptor")
	}
	fmt.Println("Got error:", err)

	for _, fd := range fileDescriptors {
		_ = unix.Close(fd)
	}
}
