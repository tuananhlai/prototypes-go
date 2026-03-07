package cheatsheetbenchmark_test

import (
	"testing"

	"golang.org/x/sys/unix"
)

const (
	testFilePath = "testdata/sample.txt"
)

// BenchmarkRawReadOneBytes benchmarks reading a file one byte at a time without buffer.
func BenchmarkRawReadOneBytes(b *testing.B) {
	fd, err := unix.Open(testFilePath, unix.O_RDONLY, 0644)
	if err != nil {
		b.Fatalf("cannot open test file: %v\n", err)
	}

	for b.Loop() {
		_, err := unix.Seek(fd, 0, unix.SEEK_SET)
		if err != nil {
			b.Fatalf("cannot seek to the beginning of the file: %v", err)
		}

		buf := make([]byte, 1)
		for {
			n, err := unix.Read(fd, buf)
			if err != nil {
				b.Fatalf("cannot read from test file: %v", err)
			}
			if n == 0 {
				break
			}
		}
	}

	_ = unix.Close(fd)
}

// BenchmarkRawReadMultipleBytes benchmarks reading a file multiple bytes at a time without a buffer.
func BenchmarkRawReadMultipleBytes(b *testing.B) {
	fd, err := unix.Open(testFilePath, unix.O_RDONLY, 0644)
	if err != nil {
		b.Fatalf("cannot open test file: %v\n", err)
	}

	for b.Loop() {
		_, err := unix.Seek(fd, 0, unix.SEEK_SET)
		if err != nil {
			b.Fatalf("cannot seek to the beginning of the file: %v", err)
		}

		buf := make([]byte, 1024)
		for {
			n, err := unix.Read(fd, buf)
			if err != nil {
				b.Fatalf("cannot read from test file: %v", err)
			}
			if n == 0 {
				break
			}
		}
	}

	_ = unix.Close(fd)
}
