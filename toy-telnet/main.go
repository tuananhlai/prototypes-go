package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"time"
)

const (
	timeout = 5 * time.Second
)

func main() {
	if len(os.Args) < 3 {
		panic("not enough arguments")
	}
	host, port := os.Args[1], os.Args[2]

	conn, err := net.DialTimeout("tcp4", fmt.Sprintf("%s:%s", host, port), timeout)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Connection successful to %s, port %s\n", host, port)

	bufrw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

	go func() {
		rawRes := make([]byte, 4096)
		for {
			n, err := bufrw.Read(rawRes)
			if err != nil {
				if err != io.EOF {
					fmt.Fprintf(os.Stderr, "error reading response: %v\n", err)
				}
				break
			}
			fmt.Println(string(rawRes[:n]))
		}
	}()

	for {
		fmt.Print("")
		_, err := io.Copy(bufrw, os.Stdin)
		if err != nil {
			panic(err)
		}
	}
}
