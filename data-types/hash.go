package types

import "encoding/hex"

// Hash is a type that represents the hash of a block or transaction.
type Hash [32]byte

func BytesToHash(b []byte) Hash {
	if len(b) != 32 {
		panic("invalid hash length")
	}
	var h Hash
	copy(h[:], b)
	return h
}

func (h Hash) String() string {
	return hex.EncodeToString(h[:])
}

func (h Hash) IsZero() bool {
	for _, b := range h {
		if b != 0 {
			return false
		}
	}
	return true
}
