package cheatsheetbenchmark_test

import (
	"testing"

	"golang.org/x/sys/unix"
)

const (
	testFilePath = "testdata/sample.dat"
)

// BenchmarkRawReadOneBytes benchmarks reading a file one byte at a time without buffer.
func BenchmarkRawReadOneBytes(b *testing.B) {
	fd, err := unix.Open(testFilePath, unix.O_RDONLY, 0644)
	if err != nil {
		b.Fatalf("cannot open test file: %v\n", err)
	}

	buf := make([]byte, 1)
	for b.Loop() {
		_ = readFile(fd, buf)
	}

	_ = unix.Close(fd)
}

// BenchmarkRawReadDiskAlignedBlockSize benchmarks reading a file multiple bytes at a time without a buffer.
// The number of bytes read at once is a multiple of the disk block size.
func BenchmarkRawReadDiskAlignedBlockSize(b *testing.B) {
	fd, err := unix.Open(testFilePath, unix.O_RDONLY, 0644)
	if err != nil {
		b.Fatalf("cannot open test file: %v\n", err)
	}

	buf := make([]byte, 4096) // 4096 is a common disk block size
	for b.Loop() {
		_ = readFile(fd, buf)
	}

	_ = unix.Close(fd)
}

// BenchmarkRawReadArbitraryBlockSize benchmarks reading a file multiple bytes at a time without a buffer.
// The number of bytes read at once is NOT a multiple of the disk block size.
func BenchmarkRawReadArbitraryBlockSize(b *testing.B) {
	fd, err := unix.Open(testFilePath, unix.O_RDONLY, 0644)
	if err != nil {
		b.Fatalf("cannot open test file: %v\n", err)
	}

	buf := make([]byte, 4241)
	for b.Loop() {
		_ = readFile(fd, buf)
	}

	_ = unix.Close(fd)
}

func readFile(fd int, buf []byte) error {
	_, err := unix.Seek(fd, 0, unix.SEEK_SET)
	if err != nil {
		return err
	}

	for {
		n, err := unix.Read(fd, buf)
		if err != nil {
			return err
		}
		if n == 0 {
			break
		}
	}

	return nil
}
