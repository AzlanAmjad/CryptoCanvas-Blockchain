package core

import (
	"math/big"
	"testing"
	"time"

	crypto "github.com/AzlanAmjad/DreamscapeCanvas-Blockchain/cryptography"
	types "github.com/AzlanAmjad/DreamscapeCanvas-Blockchain/data-types"
	"github.com/stretchr/testify/assert"
)

func getRandomBlock(t *testing.T, index uint32, prevBlockHash types.Hash) *Block {
	// create a new block
	bh := &BlockHeader{
		Version:       1,
		PrevBlockHash: prevBlockHash,
		MerkleRoot:    types.Hash{0x04, 0x05, 0x06},
		Timestamp:     time.Now().UnixNano(),
		Index:         index,
	}
	b := &Block{
		Header:       bh,
		Transactions: []*Transaction{},
	}
	assert.NotNil(t, b)
	return b
}

func getRandomBlockWithSignature(t *testing.T, index uint32, prevBlockHash types.Hash) *Block {
	// create a new block
	b := getRandomBlock(t, index, prevBlockHash)

	// sign the block
	privateKey := crypto.GeneratePrivateKey()
	b.Sign(&privateKey)

	return b
}

func getPrevBlockHash(t *testing.T, bc *Blockchain, index uint32) types.Hash {
	prevIndex := index - 1
	if prevIndex > bc.GetHeight() {
		t.Fatalf("invalid index: %d, blockchain height %d", index, bc.GetHeight())
	}

	prevBlockHeader, err := bc.GetHeaderByIndex(prevIndex)
	assert.Nil(t, err)
	assert.NotNil(t, prevBlockHeader)
	return bc.BlockHeaderHasher.Hash(prevBlockHeader)
}

func TestBlockHash(t *testing.T) {
	// Create a new block header
	b := getRandomBlock(t, 0, types.Hash{})

	// Calculate the hash of the block
	hasher := &BlockHeaderHasher{}
	hash := b.GetHash(hasher)

	// Ensure the hash is not zero
	assert.False(t, hash.IsZero())
}

func TestBlockSignAndVerifyPass(t *testing.T) {
	// Create a new block header
	b := getRandomBlock(t, 0, types.Hash{})

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
	b := getRandomBlock(t, 0, types.Hash{})

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
	b := getRandomBlock(t, 0, types.Hash{})

	// Verify the signature
	ok, err := b.VerifySignature()
	assert.False(t, ok)
	assert.Error(t, err)
}

func TestBlockSignAndVerifyFailInvalidSignature(t *testing.T) {
	// Create a new block header
	b := getRandomBlock(t, 0, types.Hash{})

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
	b := getRandomBlock(t, 0, types.Hash{})

	// Generate a new private key
	privateKey := crypto.GeneratePrivateKey()

	// Sign the block
	err := b.Sign(&privateKey)
	assert.NoError(t, err)

	// Modify the public key
	otherPrivKey := crypto.GeneratePrivateKey()
	b.Validator = otherPrivKey.GetPublicKey()

	// Verify the signature
	ok, err := b.VerifySignature()
	assert.False(t, ok)
	assert.NoError(t, err)
}
