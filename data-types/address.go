package types

import "encoding/hex"

// Address is a type that represents the address of an account.
type Address [20]byte

func (a Address) String() string {
	return hex.EncodeToString(a[:])
}
