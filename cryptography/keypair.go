package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"math/big"

	types "github.com/AzlanAmjad/DreamscapeCanvas-Blockchain/data-types"
)

type PrivateKey struct {
	key *ecdsa.PrivateKey
}

type PublicKey struct {
	Key     *ecdsa.PublicKey
	Address types.Address
}

type Signature struct {
	R, S *big.Int
}

// GeneratePrivateKey generates a new private key.
func GeneratePrivateKey() PrivateKey {
	key, _ := GeneratePrivateKeyFromBytes(rand.Reader)
	return *key
}

// GeneratePrivateKeyFromBytes generates a new private key from a byte array.
func GeneratePrivateKeyFromBytes(r io.Reader) (*PrivateKey, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), r)
	if err != nil {
		return nil, err
	}
	return &PrivateKey{key}, nil
}

// PublicKey returns the public key for the private key.
func (k *PrivateKey) GetPublicKey() PublicKey {
	return PublicKey{
		Key: k.key.Public().(*ecdsa.PublicKey),
	}
}

// Sign signs the data with the private key.
func (k *PrivateKey) Sign(data []byte) (*Signature, error) {
	dataHash := sha256.Sum256(data)
	r, s, err := ecdsa.Sign(rand.Reader, k.key, dataHash[:])
	if err != nil {
		return nil, err
	}
	return &Signature{
		R: r,
		S: s,
	}, nil
}

// ToBytes returns the public key as a byte array.
func (k *PublicKey) ToBytes() []byte {
	if k.Key == nil {
		return []byte{}
	}
	publicKeyBytes := elliptic.MarshalCompressed(k.Key.Curve, k.Key.X, k.Key.Y)
	return publicKeyBytes
}

// String returns the public key as a string.
func (k *PublicKey) String() string {
	k_bytes := k.ToBytes()

	// return string of bytes, not shown as ascii
	return hex.EncodeToString(k_bytes)
}

// GetAddress returns the address of the public key.
func (k *PublicKey) GetAddress() types.Address {
	if k.Address != [20]byte{} { // address is not 0 byte array
		return k.Address
	}

	// Hash the public key and take the last 20 bytes.
	publicKeyHash := sha256.Sum256(k.ToBytes())
	k.Address = types.Address(publicKeyHash[len(publicKeyHash)-20:])
	return k.Address
}

// Verify verifies the signature of the data using the public key.
func (k *PublicKey) Verify(data []byte, signature *Signature) bool {
	dataHash := sha256.Sum256(data)
	return ecdsa.Verify(k.Key, dataHash[:], signature.R, signature.S)
}
