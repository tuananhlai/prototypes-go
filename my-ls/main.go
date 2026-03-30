package main

import (
	"flag"
	"fmt"
	"slices"

	"golang.org/x/sys/unix"
)

func main() {
	displayHiddenFiles := flag.Bool("a", false, "all")

	flag.Parse()

	wd := flag.Arg(0)
	var err error
	if wd == "" {
		wd, err = unix.Getwd()
		if err != nil {
			panic(err)
		}
	}

	fd, err := unix.Open(wd, unix.O_RDONLY, 0o644)
	if err != nil {
		panic(err)
	}
	defer unix.Close(fd)

	entryNames, err := readDir(fd)
	if err != nil {
		panic(err)
	}

	slices.Sort(entryNames)

	for _, name := range entryNames {
		if !*displayHiddenFiles && isHiddenFile(name) {
			continue
		}
		fmt.Println(name)
	}
}

func isHiddenFile(fileName string) bool {
	return len(fileName) > 0 && fileName[0] == '.'
}

func readDir(fd int) ([]string, error) {
	var retval []string

	buf := make([]byte, 4096)
	for {
		// Read the next directory entries from the cwd's file descriptor.
		// The method tries to read as many **complete** directory entries into
		// buf as possible before returning.
		//
		// Regular `Read` system call will return EISDIR
		n, err := unix.ReadDirent(fd, buf)
		if err != nil {
			return nil, err
		}
		if n == 0 {
			break
		}

		// Parse the raw bytes returned by `ReadDirent` into a slice of entry names.
		_, _, names := unix.ParseDirent(buf[:n], -1, nil)
		retval = slices.Concat(retval, names)
	}

	return retval, nil
}
