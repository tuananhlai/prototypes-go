package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

func main() {
	sortedFilePaths := []string{"testdata/file1.txt", "testdata/file2.txt"}
	var scanners []*bufio.Scanner

	for _, path := range sortedFilePaths {
		file, err := os.Open(path)
		if err != nil {
			panic(err)
		}
		defer file.Close()

		scanners = append(scanners, bufio.NewScanner(file))
	}

	outputFile, err := os.CreateTemp(os.TempDir(), "*.txt")
	if err != nil {
		panic(err)
	}
	defer outputFile.Close()
	outputFileWriter := bufio.NewWriter(outputFile)

	current := make([]string, len(scanners))
	ok := make([]bool, len(scanners))
	for i, s := range scanners {
		ok[i] = s.Scan()
		if ok[i] {
			current[i] = s.Text()
		}
	}

	for {
		chosen := -1

		for i := range scanners {
			if !ok[i] {
				continue
			}
			if chosen == -1 || current[i] < current[chosen] {
				chosen = i
			}
		}

		if chosen == -1 {
			break
		}

		fmt.Fprintln(outputFileWriter, current[chosen])

		ok[chosen] = scanners[chosen].Scan()
		if ok[chosen] {
			current[chosen] = scanners[chosen].Text()
		}
	}

	if err := outputFileWriter.Flush(); err != nil {
		panic(err)
	}

	fmt.Println("Merged output written to:", outputFile.Name())

	_, err = outputFile.Seek(0, 0)
	if err != nil {
		panic(err)
	}

	outputFileContent, err := io.ReadAll(outputFile)
	if err != nil {
		panic(err)
	}

	fmt.Print(string(outputFileContent))
}
