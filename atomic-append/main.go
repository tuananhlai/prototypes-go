package main

import (
	"os"
	"strconv"

	"golang.org/x/sys/unix"
)

func main() {
	outputFilePath := os.Args[1]
	numBytes, err := strconv.Atoi(os.Args[2])
	if err != nil {
		panic(err)
	}

	var nonAtomic bool
	if len(os.Args) == 4 {
		nonAtomic = os.Args[3] == "x"
	}

	if nonAtomic {
		fd, err := unix.Open(outputFilePath, unix.O_WRONLY|unix.O_CREAT, unix.S_IRUSR|unix.S_IWUSR)
		if err != nil {
			panic(err)
		}
		err = doNonAtomicWrite(fd, numBytes)
		if err != nil {
			panic(err)
		}
		return
	}

	fd, err := unix.Open(outputFilePath, unix.O_WRONLY|unix.O_CREAT|unix.O_APPEND, unix.S_IRUSR|unix.S_IWUSR)
	if err != nil {
		panic(err)
	}
	err = doAtomicWrite(fd, numBytes)
	if err != nil {
		panic(err)
	}
}

func doAtomicWrite(fd int, numBytes int) error {
	var err error

	for range numBytes {
		_, err = unix.Write(fd, []byte{0})
		if err != nil {
			return err
		}
	}

	return nil
}

func doNonAtomicWrite(fd int, numBytes int) error {
	var err error

	for range numBytes {
		_, err = unix.Write(fd, []byte{0})
		if err != nil {
			return err
		}

		_, err = unix.Seek(fd, 0, unix.SEEK_END)
		if err != nil {
			return err
		}
	}

	return nil
}
