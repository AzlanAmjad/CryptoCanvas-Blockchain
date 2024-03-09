package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBlockchain(t *testing.T) {
	// create the default LevelDB storage
	storage, err := NewLevelDBStorage("./leveldb/test")
	if err != nil {
		t.Error("NewLevelDBStorage failed")
	}
	defer storage.Shutdown()

	// create new blockchain with the LevelDB storage
	bc, err := NewBlockchain(storage, getRandomBlock(0))
	assert.NotNil(t, bc.Storage)
	assert.NotNil(t, bc.Validator)
	assert.Nil(t, err)
	assert.Equal(t, uint32(0), bc.GetHeight())
}

func TestAddBlock(t *testing.T) {
	// create the default LevelDB storage
	storage, err := NewLevelDBStorage("./leveldb/test")
	if err != nil {
		t.Errorf("NewLevelDBStorage failed %v", err)
	}
	defer storage.Shutdown()

	// create new blockchain with the LevelDB storage
	bc, err := NewBlockchain(storage, getRandomBlock(0))
	assert.Nil(t, err)

	for i := 1; i < 10; i++ {
		// create a new block
		block := getRandomBlockWithSignature(uint32(i))

		// add the block to the blockchain
		err = bc.AddBlock(block)
		assert.Nil(t, err)
		assert.Equal(t, uint32(i), bc.GetHeight())
	}

	// add the same block again
	block := getRandomBlockWithSignature(9)
	err = bc.AddBlock(block)
	assert.NotNil(t, err)
}
