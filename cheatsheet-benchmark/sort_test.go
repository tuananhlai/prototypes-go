package cheatsheetbenchmark_test

import (
	"cmp"
	"slices"
	"testing"
)

func BenchmarkSort(b *testing.B) {
	var compares int64

	for b.Loop() {
		s := []int{5, 4, 3, 2, 1}
		slices.SortFunc(s, func(a, b int) int {
			compares++
			return cmp.Compare(a, b)
		})
	}
	// This metric is per-operation, so divide by b.N and
	// report it as a "/op" unit.
	b.ReportMetric(float64(compares)/float64(b.N), "compares/op")
	// This metric is per-time, so divide by b.Elapsed and
	// report it as a "/ns" unit.
	b.ReportMetric(float64(compares)/float64(b.Elapsed().Nanoseconds()), "compares/ns")

}
