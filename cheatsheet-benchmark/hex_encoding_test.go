package cheatsheetbenchmark_test

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func BenchmarkHexEncoding(b *testing.B) {
	data := make([]byte, 256)
	for i := range data {
		data[i] = byte(i)
	}

	cases := []struct {
		name string
		fn   func([]byte) string
	}{
		{
			name: "hex.EncodeToString",
			fn: func(b []byte) string {
				return hex.EncodeToString(b)
			},
		},
		{
			name: "manual (bytes.Buffer)",
			fn: func(p []byte) string {
				var buf bytes.Buffer
				buf.Grow(len(p) * 2)
				dst := make([]byte, len(p)*2)
				hex.Encode(dst, p)
				buf.Write(dst)
				return buf.String()
			},
		},
	}

	b.ReportAllocs()
	b.SetBytes(int64(len(data)))

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			for b.Loop() {
				tc.fn(data)
			}
		})
	}
}
