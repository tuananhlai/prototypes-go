package main

import (
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

func main() {
	demoSeekPastEnd()
	// demoWrite()
}

func demoSeekPastEnd() {
	fd, err := unix.Open("/tmp/demoSeekPastEnd.txt", unix.O_CREAT|unix.O_RDWR, 0644)
	if err != nil {
		panic(err)
	}

	_, err = unix.Write(fd, []byte("hello"))
	if err != nil {
		panic(err)
	}

	// Seek past the end of the file. If a write happens after this point,
	// bytes in the middle will be padding with zeros.
	offset, err := unix.Seek(fd, 10, unix.SEEK_END)
	if err != nil {
		panic(err)
	}

	println("offset:", offset)

	_, err = unix.Write(fd, []byte(",world!"))
	if err != nil {
		panic(err)
	}

	_, err = unix.Seek(fd, 0, unix.SEEK_SET)
	if err != nil {
		panic(err)
	}

	buf := make([]byte, 4096)
	n, err := unix.Read(fd, buf)
	if err != nil {
		panic(err)
	}

	// Print file content in escaped form so that null bytes are visible.
	fmt.Printf("%q\n", buf[:n])
}

func demoWrite() {
	// unix.O_WRONLY|unix.O_RDONLY isn't the same as unix.O_RDWR
	fd, err := unix.Open("/tmp/demoWrite.txt", unix.O_CREAT|unix.O_RDWR, 0644)
	if err != nil {
		panic(err)
	}

	_, err = unix.Write(fd, []byte("hello, world!"))
	if err != nil {
		panic(err)
	}

	// Not exactly necessary in this example, but fsync ensures that a write has been written
	// to disk instead of being stored inside OS buffer.
	err = unix.Fsync(fd)
	if err != nil {
		panic(err)
	}

	// Move the file descriptor offset to the start of the file.
	_, err = unix.Seek(fd, 0, unix.SEEK_SET)
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

	err = unix.Close(fd)
	if err != nil {
		panic(err)
	}
}

func panic(err error) {
	if errno, ok := err.(unix.Errno); ok {
		println("errno:", errno, unix.ErrnoName(errno), errno.Error())
	}
	os.Exit(1)
}
