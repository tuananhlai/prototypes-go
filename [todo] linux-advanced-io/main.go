package main

import (
	"fmt"

	"golang.org/x/sys/unix"
)

func main() {
	demoReadv()
}

func demoReadv() {
	fd, err := unix.Open("readv.txt", unix.O_RDONLY, 0644)
	if err != nil {
		panic(err)
	}

	buf1 := make([]byte, 128)
	buf2 := make([]byte, 64)
	buf3 := make([]byte, 256)

	_, err = unix.Readv(fd, [][]byte{buf1, buf2, buf3})
	if err != nil {
		panic(err)
	}

	fmt.Println("[buf1]", string(buf1), "[buf2]", string(buf2), "[buf3]", string(buf3))
}
