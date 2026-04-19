package main

import (
	"fmt"

	"golang.org/x/sys/unix"
)

func main() {
	fd, err := unix.Open("go.mod", unix.O_RDONLY, 0)
	if err != nil {
		panic(err)
	}
	defer unix.Close(fd)

	newfd, err := unix.Dup(fd)
	if err != nil {
		panic(err)
	}

	buf := make([]byte, 64)
	_, err = unix.Read(fd, buf)
	if err != nil {
		panic(err)
	}

	offset, err := unix.Seek(fd, 0, unix.SEEK_CUR)
	if err != nil {
		panic(err)
	}

	newfdoffset, err := unix.Seek(newfd, 0, unix.SEEK_CUR)
	if err != nil {
		panic(err)
	}

	oflags, err := unix.FcntlInt(uintptr(fd), unix.F_GETFL, 0)
	if err != nil {
		panic(err)
	}

	newfdoflags, err := unix.FcntlInt(uintptr(fd), unix.F_GETFL, 0)
	if err != nil {
		panic(err)
	}

	fmt.Printf("fd %v: offset=%d,oflags=0x%x\n", fd, offset, oflags)
	fmt.Printf("newfd %v: offset=%d,oflags=0x%x\n", newfd, newfdoffset, newfdoflags)
}
