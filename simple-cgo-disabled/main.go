package main

import (
	"fmt"
	"net"
)

func main() {
	ips, err := net.LookupHost("example.com")
	if err != nil {
		panic(err)
	}

	fmt.Println(ips)
}
