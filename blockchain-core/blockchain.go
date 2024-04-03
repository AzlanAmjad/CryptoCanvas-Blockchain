package core

import (
	"fmt"
	"os"
	"sync"

	types "github.com/AzlanAmjad/DreamscapeCanvas-Blockchain/data-types"
	"github.com/go-kit/log"
)

type Blockchain struct {
	ID string
	// should never be used outside a blockchain function
	// (i.e, should never be used as Blockchain.Lock outside a blockchain function)
	Lock         sync.RWMutex
	BlockHeaders []*BlockHeader
	Storage      Storage

	Validator Validator
	// block encoder and decoder
	BlockEncoder Encoder[*Block]
	BlockDecoder Decoder[*Block]
	// block header hasher
	BlockHeaderHasher Hasher[*BlockHeader]
	Logger            log.Logger
}

// NewBlockchain creates a new empty blockchain. With the default block validator.
func NewBlockchain(storage Storage, genesis *Block, ID string) (*Blockchain, error) {
	bc := &Blockchain{
		BlockHeaders: make([]*BlockHeader, 0),
		Storage:      storage,
		ID:           ID,
	}

	// add default block encoder and decoder
	bc.BlockEncoder = NewBlockEncoder()
	bc.BlockDecoder = NewBlockDecoder()
	// add default block header hasher
	bc.BlockHeaderHasher = NewBlockHeaderHasher()
	// set the default block validator
	bc.SetValidator(NewBlockValidator(bc))
	// set the default logger
	bc.Logger = log.NewLogfmtLogger(os.Stderr)
	bc.Logger = log.With(bc.Logger, "ID", bc.ID)

	// add the genesis block to the blockchain
	err := bc.addBlockWithoutValidation(genesis)

	// log the creation of the blockchain
	bc.Logger.Log(
		"msg", "blockchain created",
		"blockchain_id", bc.ID,
		"genesis_block_hash", genesis.GetHash(bc.BlockHeaderHasher),
	)

	return bc, err
}

// addBlockWithoutValidation adds a block to the blockchain without validation.
func (bc *Blockchain) addBlockWithoutValidation(block *Block) error {
	index := block.Header.Index
	hash := block.GetHash(bc.BlockHeaderHasher)

	bc.Lock.Lock()
	// add the block to the storage
	err := bc.Storage.Put(block, bc.BlockEncoder)
	if err != nil {
		panic(err)
	}
	// add the block to the blockchain headers.
	bc.BlockHeaders = append(bc.BlockHeaders, block.Header)
	bc.Lock.Unlock()

	bc.Logger.Log(
		"msg", "genesis block added to the blockchain",
		"block_index", index,
		"block_hash", hash,
	)

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
// int return value is only for validate block error code, every other error will have a 0 error code
func (bc *Blockchain) AddBlock(block *Block) (int, error) {
	if bc.Validator == nil {
		return 0, fmt.Errorf("no validator to validate the block")
	}

	// validate the block before adding it to the blockchain
	error_code, err := bc.Validator.ValidateBlock(block)
	if err != nil {
		return error_code, err
	}

	bc.Lock.Lock()
	// add the block to the storage
	err = bc.Storage.Put(block, bc.BlockEncoder)
	if err != nil {
		return 0, err
	}
	// add the block to the blockchain headers
	bc.BlockHeaders = append(bc.BlockHeaders, block.Header)
	bc.Lock.Unlock()

	bc.Logger.Log(
		"msg", "Block added to the blockchain",
		"block_index", block.Header.Index,
		"block_hash", block.GetHash(bc.BlockHeaderHasher),
		"data_hash", block.Header.DataHash,
		"transactions", len(block.Transactions),
		"blockchain_height", bc.GetHeight(),
		"block validator", string(block.Validator.ToBytes()),
		"block signature R", string(block.Signature.R.Bytes()),
		"block signature S", string(block.Signature.S.Bytes()),
	)

	return 0, nil
}

// get multiple blocks in a specified range
// start is inclusive
// end is not inclusive
func (bc *Blockchain) GetBlocks(start, end uint32) ([]*Block, error) {
	if start >= end {
		return nil, fmt.Errorf("start index is greater than or equal to end index")
	}

	bc.Lock.RLock() // lock because we are reading
	defer bc.Lock.RUnlock()
	return bc.Storage.GetBlocks(start, end, bc.BlockDecoder)
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

func (bc *Blockchain) GetBlockHash(index uint32) (types.Hash, error) {
	header, err := bc.GetHeaderByIndex(index)
	if err != nil {
		return types.Hash{}, err
	}
	return bc.BlockHeaderHasher.Hash(header), nil
}
