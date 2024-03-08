package core

import (
	"math/big"
	"testing"
	"time"

	crypto "github.com/AzlanAmjad/DreamscapeCanvas-Blockchain/cryptography"
	types "github.com/AzlanAmjad/DreamscapeCanvas-Blockchain/data-types"
	"github.com/stretchr/testify/assert"
)

func getRandomBlock() *Block {
	bh := &BlockHeader{
		Version:       1,
		PrevBlockHash: types.Hash{0x01, 0x02, 0x03},
		MerkleRoot:    types.Hash{0x04, 0x05, 0x06},
		Timestamp:     time.Now().UnixNano(),
		Index:         1,
	}

	return &Block{
		Header:       bh,
		Transactions: []*Transaction{},
	}
}

func TestBlockHash(t *testing.T) {
	// Create a new block header
	b := getRandomBlock()

	// Calculate the hash of the block
	hasher := &BlockHasher{}
	hash := b.GetHash(hasher)

	// Ensure the hash is not zero
	assert.False(t, hash.IsZero())
}

func TestBlockSignAndVerifyPass(t *testing.T) {
	// Create a new block header
	b := getRandomBlock()

	// Generate a new private key
	privateKey := crypto.GeneratePrivateKey()

	// Sign the block
	err := b.Sign(&privateKey)
	assert.NoError(t, err)

	// Verify the signature
	ok, err := b.VerifySignature()
	assert.NoError(t, err)
	assert.True(t, ok)
}

func TestBlockSignAndVerifyFailChangedHeader(t *testing.T) {
	// Create a new block header
	b := getRandomBlock()

	// Generate a new private key
	privateKey := crypto.GeneratePrivateKey()

	// Sign the block
	err := b.Sign(&privateKey)
	assert.NoError(t, err)

	// Modify the block header
	b.Header.Version = 2

	// Verify the signature
	ok, err := b.VerifySignature()
	assert.NoError(t, err)
	assert.False(t, ok)
}

func TestBlockSignAndVerifyFailNoSignature(t *testing.T) {
	// Create a new block header
	b := getRandomBlock()

	// Verify the signature
	ok, err := b.VerifySignature()
	assert.False(t, ok)
	assert.Error(t, err)
}

func TestBlockSignAndVerifyFailInvalidSignature(t *testing.T) {
	// Create a new block header
	b := getRandomBlock()

	// Generate a new private key
	privateKey := crypto.GeneratePrivateKey()

	// Sign the block
	err := b.Sign(&privateKey)
	assert.NoError(t, err)

	// Modify the signature
	b.Signature = &crypto.Signature{
		R: new(big.Int),
		S: new(big.Int),
	}

	// Verify the signature
	ok, err := b.VerifySignature()
	assert.False(t, ok)
	assert.NoError(t, err)
}

func TestBlockSignAndVerifyFailInvalidPublicKey(t *testing.T) {
	// Create a new block header
	b := getRandomBlock()

	// Generate a new private key
	privateKey := crypto.GeneratePrivateKey()

	// Sign the block
	err := b.Sign(&privateKey)
	assert.NoError(t, err)

	// Modify the public key
	otherPrivKey := crypto.GeneratePrivateKey()
	b.PublicKey = otherPrivKey.GetPublicKey()

	// Verify the signature
	ok, err := b.VerifySignature()
	assert.False(t, ok)
	assert.NoError(t, err)
}
