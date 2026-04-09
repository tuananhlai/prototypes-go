package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"golang.org/x/sys/unix"
)

func main() {
	var recursive bool
	flag.BoolVar(&recursive, "r", false, "remove directories and their contents recursively")
	flag.Parse()

	var err error
	for _, path := range flag.Args() {
		err = rm(path, recursive)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error removing %s: %v\n", path, err)
		}
	}
}

// stat returns an object containing metadata of the file at the given path.
func stat(filePath string) (unix.Stat_t, error) {
	var statT unix.Stat_t
	err := unix.Stat(filePath, &statT)
	return statT, err
}

func rm(path string, recursive bool) error {
	if !recursive {
		return singleRemove(path)
	}

	return recursiveRemove(path)
}

// singleRemove removes the file / empty directory at the target path.
func singleRemove(path string) error {
	statT, err := stat(path)
	if err != nil {
		return err
	}

	if !isDir(statT) {
		err = unix.Unlink(path)
		return err
	}

	err = unix.Rmdir(path)
	if err != nil {
		return err
	}
	return nil
}

// recursiveRemove removes the file / directory at the target path. If the target is an directory, its content
// is deleted recursively before the directory is removed.
func recursiveRemove(path string) error {
	statT, err := stat(path)
	if err != nil {
		return err
	}

	var entries []string
	if isDir(statT) {
		entries, err = readDir(path)
		if err != nil {
			return err
		}

		for _, entry := range entries {
			err = recursiveRemove(entry)
			if err != nil {
				return err
			}
		}

		err = unix.Rmdir(path)
		if err != nil {
			return err
		}

		return nil
	}

	return unix.Unlink(path)
}

// readDir returns the absolute path of all entries in the given directory.
func readDir(path string) ([]string, error) {
	fd, err := unix.Open(path, unix.O_RDONLY, 0o644)
	if err != nil {
		panic(err)
	}
	defer unix.Close(fd)

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

	// Transform entry names into absolute paths.
	for i := 0; i < len(retval); i++ {
		retval[i] = filepath.Join(path, retval[i])
	}

	return retval, nil
}

func isDir(statT unix.Stat_t) bool {
	return statT.Mode&unix.S_IFMT == unix.S_IFDIR
}
