package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
	"text/tabwriter"

	"golang.org/x/sys/unix"
)

func main() {
	fd, err := unix.Open("/proc", unix.O_RDONLY|unix.O_DIRECTORY, 0)
	if err != nil {
		panic(err)
	}
	defer unix.Close(fd)

	buf := make([]byte, 4096)
	var entryNames []string

	for {
		n, err := unix.ReadDirent(fd, buf)
		if err != nil {
			panic(err)
		}
		if n == 0 {
			break
		}

		_, _, entryNames = unix.ParseDirent(buf[:n], -1, entryNames)
	}

	writer := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	fmt.Fprint(writer, "pid\tuid\n")
	pidRegex := regexp.MustCompile("^[0-9]+$")
	for _, entryName := range entryNames {
		if !pidRegex.MatchString(entryName) {
			continue
		}

		// A process is ephemeral, so it and the associated `/proc/{pid}` directory
		// might disappear right after we read that entry into memory.
		// Therefore, we will silently ignore any error that occur while retrieving a process's uid.
		uid, err := getRealUid(entryName)
		if err != nil {
			continue
		}
		fmt.Fprintf(writer, "%s\t%s\n", entryName, uid)
	}

	err = writer.Flush()
	if err != nil {
		panic(err)
	}
}

// getRealUid returns the real uid of the user who started the process
// with the given ID.
func getRealUid(pid string) (string, error) {
	name := fmt.Sprintf("/proc/%s/status", pid)
	fd, err := unix.Open(name, unix.O_RDONLY, 0)
	if err != nil {
		return "", err
	}
	defer unix.Close(fd)

	scanner := bufio.NewScanner(os.NewFile(uintptr(fd), name))
	for {
		ok := scanner.Scan()
		if !ok {
			return "", scanner.Err()
		}

		line := scanner.Text()
		parts := strings.Split(line, ":")

		key := parts[0]
		if key != "Uid" {
			continue
		}

		// [real uid] [effective uid] [saved set uid] [filesystem uid]
		uidParts := strings.Fields(parts[1])
		return uidParts[0], nil
	}
}
