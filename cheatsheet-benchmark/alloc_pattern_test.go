package cheatsheetbenchmark_test

import "testing"

var (
	sinkBytes []byte
)

func BenchmarkAllocPatterns(b *testing.B) {
	b.Run("allocate_each_time", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			for i := range 10 {
				buf := make([]byte, 1024)
				buf[i] = byte(i)
				sinkBytes = buf
			}
		}
	})

	b.Run("reuse_buffer", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			buf := make([]byte, 1024)
			for i := range buf {
				clear(buf)
				buf[i] = byte(i)
				sinkBytes = buf
			}
		}
	})
}
