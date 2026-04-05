package main

import (
	"golang.org/x/sys/unix"
)

func main() {
	sockfd, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM, 0)
	if err != nil {
		panic(err)
	}
	defer unix.Close(sockfd)

	err = unix.Bind(sockfd, &unix.SockaddrInet4{
		Port: 8080,
		Addr: [4]byte{0, 0, 0, 0},
	})
	if err != nil {
		panic(err)
	}

	err = unix.Listen(sockfd, 0)
	if err != nil {
		panic(err)
	}

	nfd, _, err := unix.Accept(sockfd)
	if err != nil {
		panic(err)
	}

	buf := make([]byte, 4096)

	// TODO: handle multiple connections in parallel.
	for {
		n, _, err := unix.Recvfrom(nfd, buf, 0)
		if err != nil {
			panic(err)
		}

		_, _ = unix.Write(unix.Stdout, buf[:n])
	}
}
