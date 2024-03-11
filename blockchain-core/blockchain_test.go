package core

import (
	"testing"

	types "github.com/AzlanAmjad/DreamscapeCanvas-Blockchain/data-types"
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
	bc, err := NewBlockchain(storage, getRandomBlock(t, 0, types.Hash{}))
	assert.NotNil(t, bc.Storage)
	assert.NotNil(t, bc.Validator)
	assert.Nil(t, err)
	assert.Equal(t, uint32(0), bc.GetHeight())
}

func TestHasBlock(t *testing.T) {
	// create the default LevelDB storage
	storage, err := NewLevelDBStorage("./leveldb/test")
	if err != nil {
		t.Errorf("NewLevelDBStorage failed %v", err)
	}
	defer storage.Shutdown()

	// create new blockchain with the LevelDB storage
	bc, err := NewBlockchain(storage, getRandomBlock(t, 0, types.Hash{}))
	assert.Nil(t, err)

	// create a new block
	block := getRandomBlock(t, 1, getPrevBlockHash(t, bc, 1))

	// add the block to the blockchain
	err = bc.AddBlock(block)
	assert.Nil(t, err)

	// check if the block exists
	assert.True(t, bc.HasBlock(block))

	// create a new block
	block2 := getRandomBlock(t, 2, getPrevBlockHash(t, bc, 2))
	assert.False(t, bc.HasBlock(block2))
}

func TestAddBlock(t *testing.T) {
	// create the default LevelDB storage
	storage, err := NewLevelDBStorage("./leveldb/test")
	if err != nil {
		t.Errorf("NewLevelDBStorage failed %v", err)
	}
	defer storage.Shutdown()

	// create new blockchain with the LevelDB storage
	bc, err := NewBlockchain(storage, getRandomBlock(t, 0, types.Hash{}))
	assert.Nil(t, err)

	for i := 1; i < 10; i++ {
		// create a new block
		block := getRandomBlock(t, uint32(i), getPrevBlockHash(t, bc, uint32(i)))

		// add the block to the blockchain
		err = bc.AddBlock(block)
		assert.Nil(t, err)
		assert.Equal(t, uint32(i), bc.GetHeight())
	}

	// add the same block again
	block := getRandomBlock(t, 9, getPrevBlockHash(t, bc, 9))
	err = bc.AddBlock(block)
	assert.NotNil(t, err)
}

func TestAddBlockToHigh(t *testing.T) {
	// create the default LevelDB storage
	storage, err := NewLevelDBStorage("./leveldb/test")
	if err != nil {
		t.Errorf("NewLevelDBStorage failed %v", err)
	}
	defer storage.Shutdown()

	// create new blockchain with the LevelDB storage
	bc, err := NewBlockchain(storage, getRandomBlock(t, 0, types.Hash{}))
	assert.Nil(t, err)

	// create a new block, with a high index
	block := getRandomBlock(t, 10, types.Hash{})
	valid_block := getRandomBlock(t, 1, getPrevBlockHash(t, bc, 1))

	// add the block to the blockchain
	assert.NotNil(t, bc.AddBlock(block))
	assert.Nil(t, bc.AddBlock(valid_block))
}

func TestGetHeaderByIndex(t *testing.T) {
	// create the default LevelDB storage
	storage, err := NewLevelDBStorage("./leveldb/test")
	if err != nil {
		t.Errorf("NewLevelDBStorage failed %v", err)
	}
	defer storage.Shutdown()

	// create new blockchain with the LevelDB storage
	bc, err := NewBlockchain(storage, getRandomBlock(t, 0, types.Hash{}))
	assert.Nil(t, err)

	for i := 1; i < 10; i++ {
		// create a new block
		block := getRandomBlock(t, uint32(i), getPrevBlockHash(t, bc, uint32(i)))

		// add the block to the blockchain
		err = bc.AddBlock(block)
		assert.Nil(t, err)

		// get the block header by index
		blockHeader, err := bc.GetHeaderByIndex(uint32(i))
		assert.Nil(t, err)
		assert.Equal(t, uint32(i), blockHeader.Index)
	}
}
