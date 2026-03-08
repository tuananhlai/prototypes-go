package cheatsheetbenchmark_test

import (
	"testing"

	"golang.org/x/sys/unix"
)

const (
	readTargetFilePath = "testdata/sample.dat"
)

// BenchmarkRawReadOneBytes benchmarks reading a file one byte at a time.
func BenchmarkRawReadOneBytes(b *testing.B) {
	fd, err := unix.Open(readTargetFilePath, unix.O_RDONLY, 0644)
	if err != nil {
		b.Fatalf("cannot open test file: %v\n", err)
	}

	buf := make([]byte, 1)
	for b.Loop() {
		_ = process(fd, buf)
	}

	_ = unix.Close(fd)
}

// BenchmarkRawReadDiskAlignedBlockSize benchmarks reading a file multiple bytes at a time.
// The number of bytes read at once is a multiple of the disk block size.
func BenchmarkRawReadDiskAlignedBlockSize(b *testing.B) {
	fd, err := unix.Open(readTargetFilePath, unix.O_RDONLY, 0644)
	if err != nil {
		b.Fatalf("cannot open test file: %v\n", err)
	}

	buf := make([]byte, 4096) // 4096 is a common disk block size
	for b.Loop() {
		_ = process(fd, buf)
	}

	_ = unix.Close(fd)
}

// BenchmarkRawReadArbitraryBlockSize benchmarks reading a file multiple bytes at a time.
// The number of bytes read at once is NOT a multiple of the disk block size.
func BenchmarkRawReadArbitraryBlockSize(b *testing.B) {
	fd, err := unix.Open(readTargetFilePath, unix.O_RDONLY, 0644)
	if err != nil {
		b.Fatalf("cannot open test file: %v\n", err)
	}

	buf := make([]byte, 4241)
	for b.Loop() {
		_ = process(fd, buf)
	}

	_ = unix.Close(fd)
}

// BenchmarkWriteConcatBytes writes multiple byte arrays in a single `write` system call
// by concatenating them together.
func BenchmarkWriteConcatBytes(b *testing.B) {
	filePath := "testdata/write_concat_bytes.dat"
	fd, err := unix.Open(filePath, unix.O_CREAT|unix.O_WRONLY, 0644)
	if err != nil {
		b.Fatalf("cannot open target file: %v\n", err)
	}
	defer unix.Close(fd)

	slices := make([][]byte, 4)
	for i := range slices {
		slices[i] = make([]byte, 256)
	}

	for b.Loop() {
		var totalLen int
		for _, s := range slices {
			totalLen += len(s)
		}

		concatBuf := make([]byte, 0, totalLen)
		for _, s := range slices {
			concatBuf = append(concatBuf, s...)
		}

		_, err := unix.Write(fd, concatBuf)
		if err != nil {
			b.Fatalf("write failed: %v", err)
		}

		b.StopTimer()
		err = unix.Truncate(filePath, 0)
		if err != nil {
			b.Fatalf("truncate failed: %v", err)
		}
		b.StartTimer()
	}
}

// BenchmarkWriteWritev writes multiple byte arrays in a single `writev` system call, avoiding
// the need to concatenate input byte arrays.
func BenchmarkWriteWritev(b *testing.B) {
	filePath := "testdata/write_writev.dat"

	fd, err := unix.Open(filePath, unix.O_CREAT|unix.O_WRONLY, 0644)
	if err != nil {
		b.Fatalf("cannot open target file: %v\n", err)
	}
	defer unix.Close(fd)

	slices := make([][]byte, 4)
	for i := range slices {
		slices[i] = make([]byte, 256)
	}

	for b.Loop() {
		_, err := unix.Writev(fd, slices)
		if err != nil {
			b.Fatalf("writev failed: %v", err)
		}

		b.StopTimer()
		err = unix.Truncate(filePath, 0)
		if err != nil {
			b.Fatalf("truncate failed: %v", err)
		}
		b.StartTimer()
	}
}

// process reads the given file descriptor, then do something with it, while making use of
// the given `buf` byte slice to store temporary read data.
func process(fd int, buf []byte) error {
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

		// process the data inside `buf`
	}

	return nil
}
