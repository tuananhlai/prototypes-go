package cheatsheetbenchmark_test

import (
	"crypto/rand"
	"crypto/sha256"
	"io"
	"testing"
)

func BenchmarkSHA256(b *testing.B) {
	data := make([]byte, 10*1024)
	_, err := io.ReadFull(rand.Reader, data)
	if err != nil {
		b.Fatal(err)
	}

	// Records the number of bytes processed per iteration
	b.SetBytes(int64(len(data)))

	// Traditionally, you need to reset the timer before entering the benchmark
	// loop for accurate results. However, since Go 1.24.0, `b.Loop()` was introduced which
	// resets the timer automatically before the benchmark loop is run.
	b.ResetTimer()
	for range b.N {
		sha256.Sum256(data)
	}
}
