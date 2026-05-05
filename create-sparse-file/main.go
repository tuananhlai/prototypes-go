package main

import "golang.org/x/sys/unix"

func main() {
	fd, err := unix.Open("./test1.txt", unix.O_RDWR|unix.O_CREAT|unix.O_TRUNC, 0o644)
	if err != nil {
		panic(err)
	}

	_, err = unix.Write(fd, []byte("hello,"))
	if err != nil {
		panic(err)
	}

	_, err = unix.Seek(fd, 15000, unix.SEEK_CUR)
	if err != nil {
		panic(err)
	}

	_, err = unix.Write(fd, []byte("world!"))
	if err != nil {
		panic(err)
	}
}
