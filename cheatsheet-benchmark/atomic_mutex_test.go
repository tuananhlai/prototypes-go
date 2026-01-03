package cheatsheetbenchmark_test

import (
	"sync"
	"sync/atomic"
	"testing"
)

func BenchmarkAtomicVsMutex(b *testing.B) {
	b.Run("atomic", func(b *testing.B) {
		var x int64

		b.RunParallel(func(p *testing.PB) {
			for p.Next() {
				atomic.AddInt64(&x, 1)
			}
		})
	})

	b.Run("mutex", func(b *testing.B) {
		var (
			mu sync.Mutex
			x  int
		)

		b.RunParallel(func(p *testing.PB) {
			for p.Next() {
				mu.Lock()
				x++
				mu.Unlock()
			}
		})
	})
}
