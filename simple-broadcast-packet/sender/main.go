package main

import (
	"fmt"
	"net"
)

func main() {
	addr, err := net.ResolveUDPAddr("udp4", "255.255.255.255:9999")
	if err != nil {
		panic(err)
	}

	conn, err := net.DialUDP("udp4", nil, addr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	msg := []byte("hello broadcast")
	_, err = conn.Write(msg)
	if err != nil {
		panic(err)
	}

	fmt.Println("broadcast sent")
}
