package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"text/tabwriter"

	"golang.org/x/sys/unix"
)

func main() {
	fd, err := unix.Open("/proc", unix.O_RDONLY, 0)
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
	fmt.Fprintln(writer, "pid\tname\tstate\tppid\tuid\tthreads\tvmsize\tvmrss")

	pidRegex := regexp.MustCompile("^[0-9]+$")
	for _, entryName := range entryNames {
		if !pidRegex.MatchString(entryName) {
			continue
		}

		// A process is ephemeral, so it and the associated `/proc/{pid}` directory
		// might disappear right after we read that entry into memory.
		// Therefore, we will silently ignore any errors that occur while retrieving process info.
		stat, err := getProcessStat(entryName)
		if err != nil {
			continue
		}

		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			entryName, stat.name, stat.state, stat.ppid, stat.realUID, stat.threads, stat.vmSize, stat.vmRSS)
	}

	if err := writer.Flush(); err != nil {
		panic(err)
	}
}

type processStat struct {
	name    string
	state   string
	ppid    string
	realUID string
	threads string
	vmSize  string
	vmRSS   string
}

func getProcessStat(pid string) (processStat, error) {
	file, err := os.Open(fmt.Sprintf("/proc/%s/status", pid))
	if err != nil {
		return processStat{}, err
	}
	defer file.Close()
	return parseProcessStatus(file)
}

func parseProcessStatus(r io.Reader) (processStat, error) {
	var stat processStat

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		key, value, ok := strings.Cut(scanner.Text(), ":")
		if !ok {
			continue
		}
		value = strings.TrimSpace(value)
		switch key {
		case "Name":
			stat.name = value
		case "State":
			stat.state = value
		case "PPid":
			stat.ppid = value
		case "Uid":
			// [real uid] [effective uid] [saved set uid] [filesystem uid]
			if fields := strings.Fields(value); len(fields) > 0 {
				stat.realUID = fields[0]
			}
		case "Threads":
			stat.threads = value
		case "VmSize":
			stat.vmSize = value
		case "VmRSS":
			stat.vmRSS = value
		}
	}

	return stat, scanner.Err()
}
