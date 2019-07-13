package lib

import (
	"encoding/binary"
	"math/rand"
)

func uint64tob(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return b
}

func randInt64Range(min, max int64) int64 {
	if min == max {
		return min
	}
	return rand.Int63n((max+1)-min) + min
}

