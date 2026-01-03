package objectid

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"io"
	"sync/atomic"
	"time"
)

var (
	processUnique [5]byte
	counter       atomic.Uint32
)

func init() {
	if _, err := io.ReadFull(rand.Reader, processUnique[:]); err != nil {
		panic(err)
	}

	var c [4]byte
	if _, err := io.ReadFull(rand.Reader, c[1:]); err != nil {
		panic(err)
	}
	counter.Store(binary.BigEndian.Uint32(c[:]))
}

// https://www.mongodb.com/docs/manual/reference/method/ObjectId/#description:~:text=Returns,restarts,-%2E
func New() [12]byte {
	var id [12]byte

	// Traditionally, unix time is stored in 4 bytes. However, Go uses `int64` to store unix time
	// for a much larger range, so we need to convert it back to a 4-byte unsigned integer.
	unixTime := uint32(time.Now().Unix())
	binary.BigEndian.PutUint32(id[:4], unixTime)

	copy(id[4:9], processUnique[:])

	c := counter.Add(1) & 0xFFFFFF
	// We cannot use binary.BigEndian.PutUint32 here, since our
	// counter is only 3 bytes, not 4.
	id[9] = byte(c >> 16)
	id[10] = byte(c >> 8)
	id[11] = byte(c)

	return id
}

func NewString() string {
	id := New()
	return hex.EncodeToString(id[:])
}
