package main

import (
	"fmt"
	"log"
	"net"

	"golang.org/x/sys/unix"
)

func main() {
	host := "example.com"
	port := 80

	// Creates a new socket and returns its file descriptor.
	// - `domain` (`unix.AF_INET`): address family. `AF_INET` means IPv4. Use `AF_INET6` for IPv6, `AF_UNIX`
	// for Unix domain sockets, etc.
	// - `typ` (`unix.SOCK_STREAM`): socket type. `SOCK_STREAM` gives you a reliable, ordered byte stream (TCP).
	// `SOCK_DGRAM` would be UDP datagrams.
	// - `proto` (`0`): protocol number. `0` means “default for this domain+type”. For `AF_INET` + `SOCK_STREAM`,
	// the kernel picks TCP (`IPPROTO_TCP`). You could also pass `unix.IPPROTO_TCP` explicitly.
	fd, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM, 0)
	if err != nil {
		log.Fatal(err)
	}

	ipAddr, err := net.ResolveIPAddr("ip", host)
	if err != nil {
		log.Fatal(err)
	}

	err = unix.Connect(fd, &unix.SockaddrInet4{
		Addr: [4]byte(ipAddr.IP),
		Port: port,
	})

	body := []byte("GET / HTTP/1.1\r\nHost: example.com\r\nConnection: close\r\n\r\n")
	_, err = unix.Write(fd, body)
	if err != nil {
		log.Fatal(err)
	}

	var resp []byte

	buf := make([]byte, 256)
	for {
		n, err := unix.Read(fd, buf)
		if err != nil {
			log.Fatal(err)
		}
		if n == 0 {
			break
		}
		resp = append(resp, buf[:n]...)
	}

	fmt.Println(string(resp))
}
