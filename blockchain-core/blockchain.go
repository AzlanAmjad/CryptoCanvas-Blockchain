package core

import (
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
)

type Blockchain struct {
	Lock         sync.RWMutex
	BlockHeaders []*BlockHeader
	Storage      Storage
	Validator    Validator
	// block encoder and decoder
	Encoder Encoder[*Block]
	Decoder Decoder[*Block]
	// block header hasher
	BlockHeaderHasher Hasher[*BlockHeader]
}

// NewBlockchain creates a new empty blockchain. With the default block validator.
func NewBlockchain(storage Storage, genesis *Block) (*Blockchain, error) {
	bc := &Blockchain{
		BlockHeaders: make([]*BlockHeader, 0),
		Storage:      storage,
	}
	// add default block encoder and decoder
	bc.Encoder = NewBlockEncoder()
	bc.Decoder = NewBlockDecoder()
	// add default block hasher
	bc.BlockHeaderHasher = NewBlockHeaderHasher()
	// set the default block validator
	bc.SetValidator(NewBlockValidator(bc))
	// add the genesis block to the blockchain
	err := bc.addBlockWithoutValidation(genesis)
	return bc, err
}

// addBlockWithoutValidation adds a block to the blockchain without validation.
func (bc *Blockchain) addBlockWithoutValidation(block *Block) error {
	logrus.WithFields(logrus.Fields{
		"block_index": block.Header.Index,
		"block_hash":  block.GetHash(bc.BlockHeaderHasher),
	}).Log(logrus.InfoLevel, "adding block to the blockchain")

	bc.Lock.Lock()
	// add the block to the storage
	err := bc.Storage.Put(block, bc.Encoder)
	if err != nil {
		panic(err)
	}
	// add the block to the blockchain headers.
	bc.BlockHeaders = append(bc.BlockHeaders, block.Header)
	bc.Lock.Unlock()

	return err
}

// SetValidator sets the validator of the blockchain.
func (bc *Blockchain) SetValidator(validator Validator) {
	bc.Validator = validator
}

// [0, 1, 2, 3] -> len = 4
// [0, 1, 2, 3] -> height = 3
// GetHeight returns the largest index of the blockchain.
func (bc *Blockchain) GetHeight() uint32 {
	bc.Lock.RLock()
	defer bc.Lock.RUnlock()

	return uint32(len(bc.BlockHeaders) - 1)
}

// AddBlock adds a block to the blockchain.
func (bc *Blockchain) AddBlock(block *Block) error {
	if bc.Validator == nil {
		return fmt.Errorf("no validator to validate the block")
	}

	// validate the block before adding it to the blockchain
	err := bc.Validator.ValidateBlock(block)
	if err != nil {
		return err
	}

	bc.Lock.Lock()
	// add the block to the storage
	err = bc.Storage.Put(block, bc.Encoder)
	if err != nil {
		return err
	}
	// add the block to the blockchain headers
	bc.BlockHeaders = append(bc.BlockHeaders, block.Header)
	bc.Lock.Unlock()

	return nil
}

// HasBlock function compares the index of the block with the height of the blockchain.
func (bc *Blockchain) HasBlock(block *Block) bool {
	return block.Header.Index <= bc.GetHeight()
}

// GetHeaderByIndex returns header by block index
func (bc *Blockchain) GetHeaderByIndex(index uint32) (*BlockHeader, error) {
	if index > bc.GetHeight() {
		return nil, fmt.Errorf("block index is invalid, block index: %d, blockchain height: %d", index, bc.GetHeight())
	}

	bc.Lock.RLock()
	defer bc.Lock.RUnlock()

	return bc.BlockHeaders[index], nil
}
