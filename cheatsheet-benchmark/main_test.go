package cheatsheetbenchmark_test

import (
	"strings"
	"testing"
)

var (
	sinkString string
)

func BenchmarkStringsConcat(b *testing.B) {
	parts := []string{"hello", " ", "world", "!", " ", "bench"}

	b.ReportAllocs()

	for b.Loop() {
		s := ""
		for _, p := range parts {
			s += p
		}
		sinkString = s
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
		sinkString = sb.String()
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
		sinkString = sb.String()
	}
}
