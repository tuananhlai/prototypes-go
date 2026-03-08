package main

import (
	"golang.org/x/sys/unix"
)

func main() {
	fd, err := unix.Open("test.txt", unix.O_CREAT|unix.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	buf1 := []byte("hello")
	buf2 := []byte("world")
	buf3 := []byte("linux")

	_, err = unix.Writev(fd, [][]byte{buf1, buf2, buf3})
	if err != nil {
		panic(err)
	}
}
