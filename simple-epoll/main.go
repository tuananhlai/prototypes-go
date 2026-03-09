package main

import (
	"fmt"
	"log"

	"golang.org/x/sys/unix"
)

func main() {
	epfd, err := unix.EpollCreate1(0)
	if err != nil {
		log.Fatal(err)
	}
	defer unix.Close(epfd)

	event := unix.EpollEvent{
		Events: unix.EPOLLIN | unix.EPOLLET,
		Fd:     int32(unix.Stdin),
	}
	if err := unix.EpollCtl(epfd, unix.EPOLL_CTL_ADD, 0, &event); err != nil {
		log.Fatal(err)
	}

	events := make([]unix.EpollEvent, 10)

	for {
		n, err := unix.EpollWait(epfd, events, -1)
		if err != nil {
			log.Println("EpollWait:", err)
			continue
		}

		for i := range n {
			fmt.Printf("Event on fd %d: %v\n", events[i].Fd, events[i].Events)
			buf := make([]byte, 1024)
			nr, err := unix.Read(int(events[i].Fd), buf)
			if err != nil {
				log.Println("Read:", err)
			} else if nr > 0 {
				fmt.Printf("Read %d bytes: %s", nr, buf[:nr])
			}
		}
	}
}
