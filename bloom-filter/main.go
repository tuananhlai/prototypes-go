package main

import "hash/fnv"

func main() {

}

type BloomFilter struct {
	m    uint
	k    uint
	bits []uint64
}

func NewBloomFilter(m, k uint) *BloomFilter {
	if m == 0 {
		m = 1
	}
	if k == 0 {
		k = 1
	}
	words := (m + 63) / 64
	return &BloomFilter{
		m:    m,
		k:    k,
		bits: make([]uint64, words),
	}
}

func (b *BloomFilter) Add(v string) {
	h1, h2 := hash64(v)
	for i := uint(0); i < b.k; i++ {
		idx := (h1 + uint64(i)*h2) % uint64(b.m)
		b.set(uint(idx))
	}
}

func (b *BloomFilter) MightContain(v string) bool {
	h1, h2 := hash64(v)
	for i := uint(0); i < b.k; i++ {
		idx := (h1 + uint64(i)*h2) % uint64(b.m)
		if !b.get(uint(idx)) {
			return false
		}
	}
	return true
}

func (b *BloomFilter) set(i uint) {
	word := i / 64
	bit := i % 64
	b.bits[word] |= 1 << bit
}

func (b *BloomFilter) get(i uint) bool {
	word := i / 64
	bit := i % 64
	return (b.bits[word] & (1 << bit)) != 0
}

func hash64(s string) (uint64, uint64) {
	h1 := fnv.New64a()
	_, _ = h1.Write([]byte(s))
	sum1 := h1.Sum64()

	h2 := fnv.New64()
	_, _ = h2.Write([]byte(s))
	sum2 := h2.Sum64()
	if sum2 == 0 {
		sum2 = 1
	}
	return sum1, sum2
}
