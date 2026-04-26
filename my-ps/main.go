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
	fmt.Fprint(writer, "pid\tuid\n")
	pidRegex := regexp.MustCompile("^[0-9]+$")
	for _, entryName := range entryNames {
		if !pidRegex.MatchString(entryName) {
			continue
		}

		// A process is ephemeral, so it and the associated `/proc/{pid}` directory
		// might disappear right after we read that entry into memory.
		// Therefore, we will silently ignore any error that occur while retrieving a process's uid.
		procStat, err := getProcessStat(entryName)
		if err != nil {
			continue
		}
		fmt.Fprintf(writer, "%s\t%s\n", entryName, procStat.realUID)
	}

	err = writer.Flush()
	if err != nil {
		panic(err)
	}
}

type processStat struct {
	realUID string
	ppid    string
}

// getProcessStat returns information related to a particular process.
func getProcessStat(pid string) (processStat, error) {
	var retval processStat

	keyToSetFn := map[string]func(value string) error{
		"Uid": func(value string) error {
			// [real uid] [effective uid] [saved set uid] [filesystem uid]
			uidParts := strings.Fields(value)
			retval.realUID = uidParts[0]
			return nil
		},
		"PPid": func(value string) error {
			retval.ppid = strings.TrimSpace(value)
			return nil
		},
	}

	file, err := os.Open(fmt.Sprintf("/proc/%s/status", pid))
	if err != nil {
		return processStat{}, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for {
		ok := scanner.Scan()
		if !ok {
			err = scanner.Err()
			if err == nil {
				break
			}

			return processStat{}, err
		}

		line := scanner.Text()
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			return processStat{}, fmt.Errorf("error malformed line: %s", line)
		}

		setFn, ok := keyToSetFn[key]
		if !ok {
			continue
		}

		err = setFn(value)
		if err != nil {
			return processStat{}, err
		}
	}

	return retval, nil
}
