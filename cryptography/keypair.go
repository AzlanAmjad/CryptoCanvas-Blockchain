package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"math/big"

	types "github.com/AzlanAmjad/DreamscapeCanvas-Blockchain/data-types"
)

type PrivateKey struct {
	key *ecdsa.PrivateKey
}

type PublicKey struct {
	Key *ecdsa.PublicKey
}

type Signature struct {
	R, S *big.Int
}

// GeneratePrivateKey generates a new private key.
func GeneratePrivateKey() PrivateKey {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(err)
	}
	return PrivateKey{key}
}

// PublicKey returns the public key for the private key.
func (k *PrivateKey) GetPublicKey() PublicKey {
	return PublicKey{k.key.Public().(*ecdsa.PublicKey)}
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
	// Hash the public key and take the last 20 bytes.
	publicKeyHash := sha256.Sum256(k.ToBytes())
	return types.Address(publicKeyHash[len(publicKeyHash)-20:])
}

// Verify verifies the signature of the data using the public key.
func (k *PublicKey) Verify(data []byte, signature *Signature) bool {
	dataHash := sha256.Sum256(data)
	return ecdsa.Verify(k.Key, dataHash[:], signature.R, signature.S)
}
