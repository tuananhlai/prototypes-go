package main

import (
	"fmt"
	"io"
	"net"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		panic("not enough arguments")
	}

	conn, err := net.Dial("tcp", os.Args[1])
	if err != nil {
		panic(err)
	}

	for {
		fmt.Print("> ")
		_, err := io.Copy(conn, os.Stdin)
		if err != nil {
			panic(err)
		}
	}
}
