package main

import (
	"fmt"
	"net"
	"os"

	"golang.org/x/sys/unix"
)

func dialTCPNoNetDial(addr string) (*os.File, error) {
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	var port int
	_, err = fmt.Sscanf(portStr, "%d", &port)
	if err != nil {
		return nil, err
	}

	ips, err := net.LookupIP(host)
	if err != nil {
		return nil, err
	}
	if len(ips) == 0 {
		return nil, fmt.Errorf("no IPs for host %q", host)
	}

	fd, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM, unix.IPPROTO_TCP)
	if err != nil {
		return nil, err
	}

	// If connect fails, close the fd.
	ok := false
	defer func() {
		if !ok {
			_ = unix.Close(fd)
		}
	}()

	ip4 := ips[0].To4()
	if ip4 == nil {
		return nil, fmt.Errorf("only IPv4 shown here (got %v)", ips[0])
	}

	sa := &unix.SockaddrInet4{Port: port}
	copy(sa.Addr[:], ip4)

	if err := unix.Connect(fd, sa); err != nil {
		return nil, err
	}

	ok = true
	// Wrap fd as *os.File (you can unix.Read/Write directly too)
	return os.NewFile(uintptr(fd), "tcp:"+addr), nil
}

func main() {
	f, err := dialTCPNoNetDial("example.com:80")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	_, _ = f.WriteString("GET / HTTP/1.1\r\nHost: example.com\r\nConnection: close\r\n\r\n")

	buf := make([]byte, 4096)
	for {
		n, err := f.Read(buf)
		if n > 0 {
			fmt.Print(string(buf[:n]))
		}
		if err != nil {
			break
		}
	}
}
