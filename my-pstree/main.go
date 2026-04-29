package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	var err error
	rootPid := 1
	if len(os.Args) > 1 {
		rootPid, err = strconv.Atoi(os.Args[1])
		if err != nil {
			panic(err)
		}
	}

	entries, err := os.ReadDir("/proc")
	if err != nil {
		panic(err)
	}

	pidToStat := map[int]processStat{}
	for _, entry := range entries {
		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			// entry is not a process directory.
			continue
		}

		stat, err := getProcessStat(pid)
		if err != nil {
			fmt.Fprintf(os.Stderr, "get stat for pid %d: %v\n", pid, err)
			continue
		}
		pidToStat[pid] = stat
	}

	children := map[int][]int{}
	for pid, stat := range pidToStat {
		children[stat.ppid] = append(children[stat.ppid], pid)
	}

	// recursively print out the process id tree. `connector` is the string
	// to print before the process name. `prefix` is the string to print
	// before printing this process's descendants.
	var printTree func(pid int, connector, prefix string)
	printTree = func(pid int, connector, prefix string) {
		stat, ok := pidToStat[pid]
		if !ok {
			return
		}

		fmt.Printf("%s%s (%d)\n", connector, stat.name, pid)

		for i, child := range children[pid] {
			isLastChild := i == len(children[pid])-1

			if isLastChild {
				printTree(child, prefix+"└── ", prefix+"    ")
				continue
			}

			printTree(child, prefix+"├── ", prefix+"│   ")
		}
	}

	// start printing from init process
	printTree(rootPid, "", "")
}

type processStat struct {
	ppid int
	name string
}

func getProcessStat(pid int) (processStat, error) {
	f, err := os.Open(fmt.Sprintf("/proc/%d/status", pid))
	if err != nil {
		return processStat{}, err
	}
	defer f.Close()

	var retval processStat
	scanner := bufio.NewScanner(f)
	for {
		ok := scanner.Scan()
		if !ok {
			if err := scanner.Err(); err != nil {
				return processStat{}, err
			}

			break
		}

		line := scanner.Text()
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			return processStat{}, fmt.Errorf("read /proc/%d/status: malformed line", pid)
		}

		value = strings.TrimSpace(value)
		switch key {
		case "PPid":
			retval.ppid, err = strconv.Atoi(value)
			if err != nil {
				return processStat{}, fmt.Errorf("parse PPid: non-numeric value %s", value)
			}
		case "Name":
			retval.name = value
		}
	}

	return retval, nil
}
