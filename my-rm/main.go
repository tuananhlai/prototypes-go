package main

import (
	"flag"
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

func main() {
	var recursive bool
	flag.BoolVar(&recursive, "r", false, "remove directories and their contents recursively")
	flag.Parse()

	var err error
	for _, path := range flag.Args() {
		err = unix.Unlink(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error removing %s: %v\n", path, err)
		}
	}
}
