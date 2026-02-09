package bloomfilter

import (
	"errors"
	"hash/fnv"
	"math"
)

type Bloom struct {
	m    uint64 // total number of bits to use for presence detection
	k    uint64
	bits []uint64 // the actual bit values, stored as a slice of 64-bit words for space efficiency.
}

// New builds a Bloom filter sized for
// n = expected insertions
// p = desired false positive probability
func New(n uint64, p float64) (*Bloom, error) {
	if n == 0 {
		return nil, errors.New("error invalid n")
	}
	if p <= 0 || p >= 1 {
		return nil, errors.New("error invalid p")
	}

	// TODO: where do these formulas come from?
	m := uint64(math.Ceil(-(float64(n) * math.Log(p)) / (math.Ln2 * math.Ln2)))
	k := uint64(math.Ceil((float64(m) / float64(n)) * math.Ln2))

	if k == 0 {
		k = 1
	}

	return &Bloom{
		m:    m,
		k:    k,
		bits: make([]uint64, (m+63)/64),
	}, nil
}

func (b *Bloom) Add(s string) {
	h1, h2 := hash2(s)

	for i := range b.k {
		// TODO: why use this particular formula?
		bitIdx := (h1 + i*h2) % b.m
		b.setBit(bitIdx)
	}
}

// MightContain returns true if the given string **might** exists and false
// if it is sure to not exist.
func (b *Bloom) MightContain(s string) bool {
	h1, h2 := hash2(s)

	for i := range b.k {
		bitIdx := (h1 + i*h2) % b.m
		if !b.getBit(bitIdx) {
			return false
		}
	}

	return true
}

func (b *Bloom) setBit(bitIdx uint64) {
	wordIdx := bitIdx / 64
	// create a bit mask like `00000100...000`, so that it can be applied
	// to the target word to set the correct bit to 1.
	mask := uint64(1) << (bitIdx % 64)
	b.bits[wordIdx] |= mask
}

// getBit returns the state of a bit at the given index. Returns true if
// this bit is 1, otherwise false.
func (b *Bloom) getBit(bitIdx uint64) bool {
	wordIdx := bitIdx / 64
	mask := uint64(1) << (bitIdx % 64)
	return (b.bits[wordIdx] & mask) != 0
}

// hash2 generates two different hash values for the given string.
// why do we need two hash values?
func hash2(s string) (uint64, uint64) {
	h1 := fnv.New64a()
	h1.Write([]byte("salt-1:"))
	h1.Write([]byte(s))
	sum1 := h1.Sum64()

	h2 := fnv.New64a()
	h2.Write([]byte("salt-2:"))
	h2.Write([]byte(s))
	sum2 := h2.Sum64()

	return sum1, sum2
}
