package core

import (
	"bytes"
	"testing"
	"time"

	types "github.com/AzlanAmjad/DreamscapeCanvas-Blockchain/data-types"
	"github.com/stretchr/testify/assert"
)

func TestHeaderEncodeDecode(t *testing.T) {
	// Create a new block header
	bh := &BlockHeader{
		Version:       1,
		PrevBlockHash: types.Hash{0x01, 0x02, 0x03},
		MerkleRoot:    types.Hash{0x04, 0x05, 0x06},
		Timestamp:     time.Now().UnixNano(),
		Index:         1,
		Nonce:         123456,
	}

	// Create a buffer to write the binary data to
	buf := new(bytes.Buffer)

	// Encode the block header to binary
	err := bh.BinaryEncode(buf)
	if err != nil {
		t.Fatalf("BinaryEncode failed: %v", err)
	}

	// Create a new block header to decode into
	bh2 := &BlockHeader{}

	// Decode the block header from binary
	err = bh2.BinaryDecode(buf)
	if err != nil {
		t.Fatalf("BinaryDecode failed: %v", err)
	}

	// Compare the original block header with the decoded block header
	assert.Equal(t, bh, bh2)
}

func TestBlockEncodeDecode(t *testing.T) {
	// Create a new block header
	bh := &BlockHeader{
		Version:       1,
		PrevBlockHash: types.Hash{0x01, 0x02, 0x03},
		MerkleRoot:    types.Hash{0x04, 0x05, 0x06},
		Timestamp:     time.Now().UnixNano(),
		Index:         1,
		Nonce:         123456,
	}

	// Create a new block
	b := &Block{
		Header:       *bh,
		Transactions: []*Transaction{},
	}

	// Create a buffer to write the binary data to
	buf := new(bytes.Buffer)

	// Encode the block to binary
	err := b.BinaryEncode(buf)
	if err != nil {
		t.Fatalf("BinaryEncode failed: %v", err)
	}

	// Create a new block to decode into
	b2 := &Block{}

	// Decode the block from binary
	err = b2.BinaryDecode(buf)
	if err != nil {
		t.Fatalf("BinaryDecode failed: %v", err)
	}

	// Compare the original block with the decoded block
	assert.Equal(t, b, b2)
}

func TestBlockHash(t *testing.T) {
	// Create a new block header
	bh := &BlockHeader{
		Version:       1,
		PrevBlockHash: types.Hash{0x01, 0x02, 0x03},
		MerkleRoot:    types.Hash{0x04, 0x05, 0x06},
		Timestamp:     time.Now().UnixNano(),
		Index:         1,
		Nonce:         123456,
	}

	// Create a new block
	b := &Block{
		Header:       *bh,
		Transactions: []*Transaction{},
	}

	// Calculate the hash of the block
	hash := b.Hash()

	// Ensure the hash is not zero
	assert.False(t, hash.IsZero())
}
