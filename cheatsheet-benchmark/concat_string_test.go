package cheatsheetbenchmark_test

import (
	"strings"
	"testing"
)

func BenchmarkStringsConcat(b *testing.B) {
	parts := []string{"hello", " ", "world", "!", " ", "bench"}

	b.ReportAllocs()

	for b.Loop() {
		s := ""
		for _, p := range parts {
			s += p
		}
	}
}

func BenchmarkStringBuilder(b *testing.B) {
	parts := []string{"hello", " ", "world", "!", " ", "bench"}

	b.ReportAllocs()

	for b.Loop() {
		var sb strings.Builder
		for _, p := range parts {
			sb.WriteString(p)
		}
	}
}

func BenchmarkStringBuilderPreallocate(b *testing.B) {
	parts := []string{"hello", " ", "world", "!", " ", "bench"}

	b.ReportAllocs()

	for b.Loop() {
		var sb strings.Builder
		sb.Grow(64)
		for _, p := range parts {
			sb.WriteString(p)
		}
	}
}
