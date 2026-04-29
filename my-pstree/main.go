package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
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

	var dfs func(pid int, level int)
	dfs = func(pid int, level int) {
		stat := pidToStat[pid]
		fmt.Printf("%s%s (%d)\n", strings.Repeat("    ", level), stat.name, pid)

		for _, childPid := range children[pid] {
			dfs(childPid, level+1)
		}
	}

	// start printing from init process
	dfs(1, 0)
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
