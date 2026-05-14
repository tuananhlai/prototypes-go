package main

import (
	"fmt"
	"math/rand"
	"testing"
)

var sizes = []int{10, 100, 1000}

func makeIntSlice(n int) []int {
	s := make([]int, n)
	for i := range s {
		s[i] = rand.Int()
	}
	return s
}

func BenchmarkSerializeIntArray(b *testing.B) {
	for _, size := range sizes {
		arr := makeIntSlice(size)
		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			for b.Loop() {
				serializeIntArrayStrconv(arr)
			}
		})
	}
}

func BenchmarkDeserializeIntArray(b *testing.B) {
	for _, size := range sizes {
		data := serializeIntArrayStrconv(makeIntSlice(size))
		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			for b.Loop() {
				deserializeIntArrayStrconv(data)
			}
		})
	}
}
