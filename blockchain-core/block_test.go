package core

import (
	"bytes"
	"math/big"
	"testing"
	"time"

	crypto "github.com/AzlanAmjad/DreamscapeCanvas-Blockchain/cryptography"
	types "github.com/AzlanAmjad/DreamscapeCanvas-Blockchain/data-types"
	"github.com/stretchr/testify/assert"
)

func getRandomBlock(t *testing.T, index uint32, prevBlockHash types.Hash) *Block {
	// generate random transaction
	tx := randomTransactionWithSignature(t)

	// create a new block
	b := NewBlock()
	b.Header.PrevBlockHash = prevBlockHash
	b.Header.Index = index
	b.Header.Timestamp = time.Now().UnixNano()
	b.Transactions = append(b.Transactions, tx)
	assert.NotNil(t, b)

	// data hash
	dataHash, err := CalculateDataHash(b.Transactions, b.TransactionEncoder)
	if err != nil {
		t.Fatalf("failed to calculate data hash: %s", err)
	}
	b.Header.DataHash = dataHash

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
	b.Signature = nil

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

func TestBlockEncodeDecode(t *testing.T) {
	// Create a new block header
	b := getRandomBlock(t, 0, types.Hash{})

	// Encode the block
	buf := bytes.Buffer{}
	enc := NewBlockEncoder()
	err := b.Encode(&buf, enc)
	if err != nil {
		t.Fatalf("failed to encode block: %s", err)
	}
	assert.NoError(t, err)

	// Decode the block
	dec := NewBlockDecoder()
	decodedBlock := NewBlock()
	err = decodedBlock.Decode(&buf, dec)
	if err != nil {
		t.Fatalf("failed to decode block: %s", err.Error())
	}
	assert.NoError(t, err)

	// verify the decoded block
	assert.Equal(t, b, decodedBlock)
	ok, _ := decodedBlock.VerifySignature()
	assert.True(t, ok)
}
