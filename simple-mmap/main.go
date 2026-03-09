package main

import (
	"fmt"
	"log"

	"golang.org/x/sys/unix"
)

func main() {
	fd, err := unix.Open("target.txt", unix.O_RDWR, 0o644)
	if err != nil {
		log.Fatalf("error opening target.txt: %v", err)
	}
	defer unix.Close(fd)

	// `offset` needs to be byte-aligned.
	data, err := unix.Mmap(fd, 0, 128, unix.PROT_READ|unix.PROT_WRITE, unix.MAP_SHARED)
	if err != nil {
		log.Fatalf("error creating memory map: %v", err)
	}
	defer unix.Munmap(data)

	fmt.Println(string(data))

	// Modifying the memory map slice will update the file content as well.
	data[0] = 'W'
}
