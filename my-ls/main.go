package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"text/tabwriter"
	"time"

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
		if errors.Is(err, unix.ENOTDIR) {
			fmt.Println(wd)
			return
		}

		panic(err)
	}

	slices.Sort(entryNames)

	writer := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)

	for _, name := range entryNames {
		if !*displayHiddenFiles && isHiddenFile(name) {
			continue
		}

		fullPath := filepath.Join(wd, name)

		statT, err := stat(fullPath)
		if err != nil {
			panic(err)
		}

		lastAccessTime := timespecToUnix(statT.Atim)

		formattedName := name
		if isDir(statT) {
			formattedName = colorize(name, colorBlue)
		}

		username, err := lookupUsername(statT.Uid)
		if err != nil {
			panic(err)
		}

		groupName, err := lookupGroup(statT.Gid)
		if err != nil {
			panic(err)
		}

		fmt.Fprintf(writer, "%s\t%s\t%s\t%d\t%s\t%s\t%s\n",
			formatMode(statT.Mode), username, groupName, statT.Size,
			lastAccessTime.Format("Jan 2"),
			lastAccessTime.Format("15:04"),
			formattedName,
		)
	}

	err = writer.Flush()
	if err != nil {
		panic(err)
	}
}

func timespecToUnix(ts unix.Timespec) time.Time {
	return time.Unix(ts.Sec, ts.Nsec)
}

func isHiddenFile(fileName string) bool {
	return len(fileName) > 0 && fileName[0] == '.'
}

// stat returns an object containing metadata of the file at the given path.
func stat(filePath string) (unix.Stat_t, error) {
	var statT unix.Stat_t
	err := unix.Stat(filePath, &statT)
	return statT, err
}

func isDir(statT unix.Stat_t) bool {
	return statT.Mode&unix.S_IFMT == unix.S_IFDIR
}

// readDir returns a slice of entry names in the directory specified by the given file descriptor.
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

const (
	reset     = "\033[0m"
	colorBlue = "\033[34m"
)

func colorize(text, color string) string {
	return fmt.Sprintf("%s%s%s", color, text, reset)
}

// formatMode returns a string representation of the file mode.
func formatMode(mode uint32) string {
	retval := []byte("---------")
	if mode&unix.S_IRUSR != 0 {
		retval[0] = 'r'
	}
	if mode&unix.S_IWUSR != 0 {
		retval[1] = 'w'
	}
	if mode&unix.S_IXUSR != 0 {
		retval[2] = 'x'
	}
	if mode&unix.S_IRGRP != 0 {
		retval[3] = 'r'
	}
	if mode&unix.S_IWGRP != 0 {
		retval[4] = 'w'
	}
	if mode&unix.S_IXGRP != 0 {
		retval[5] = 'x'
	}
	if mode&unix.S_IROTH != 0 {
		retval[6] = 'r'
	}
	if mode&unix.S_IWOTH != 0 {
		retval[7] = 'w'
	}
	if mode&unix.S_IXOTH != 0 {
		retval[8] = 'x'
	}

	return string(retval)
}

// lookupUsername converts uid to a username.
func lookupUsername(uid uint32) (string, error) {
	f, err := os.Open("/etc/passwd")
	if err != nil {
		return "", err
	}
	defer f.Close()

	uidS := fmt.Sprintf("%d", uid)
	scanner := bufio.NewScanner(f)
	for {
		ok := scanner.Scan()
		if !ok {
			return "", scanner.Err()
		}

		line := scanner.Text()
		if line == "" || line[0] == '#' {
			continue
		}

		parts := strings.Split(line, ":")
		if len(parts) >= 3 && parts[2] == uidS {
			return parts[0], nil
		}
	}
}

// lookupGroup converts gid to a group name.
func lookupGroup(gid uint32) (string, error) {
	f, err := os.Open("/etc/group")
	if err != nil {
		return "", err
	}
	defer f.Close()

	gidS := fmt.Sprintf("%d", gid)
	scanner := bufio.NewScanner(f)
	for {
		ok := scanner.Scan()
		if !ok {
			return "", scanner.Err()
		}

		line := scanner.Text()
		if line == "" || line[0] == '#' {
			continue
		}

		parts := strings.Split(line, ":")
		if len(parts) >= 3 && parts[2] == gidS {
			return parts[0], nil
		}
	}
}
