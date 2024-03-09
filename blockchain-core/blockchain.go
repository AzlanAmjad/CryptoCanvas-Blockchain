package core

import "fmt"

type Blockchain struct {
	BlockHeaders []*BlockHeader
	Storage      Storage
	Validator    Validator
	// block encoder and decoder
	Encoder Encoder[*Block]
	Decoder Decoder[*Block]
}

// NewBlockchain creates a new empty blockchain. With the default block validator.
func NewBlockchain(storage Storage, genesis *Block) (*Blockchain, error) {
	bc := &Blockchain{
		BlockHeaders: make([]*BlockHeader, 0),
		Storage:      storage,
	}
	// set the default block validator
	bc.SetValidator(NewBlockValidator(bc))
	// add the genesis block to the blockchain
	err := bc.addBlockWithoutValidation(genesis)
	// add default block encoder and decoder
	bc.Encoder = NewBlockEncoder()
	bc.Decoder = NewBlockDecoder()

	return bc, err
}

// addBlockWithoutValidation adds a block to the blockchain without validation.
func (bc *Blockchain) addBlockWithoutValidation(genesis *Block) error {
	// add the genesis block to the storage
	err := bc.Storage.Put(genesis, bc.Encoder)
	if err != nil {
		panic(err)
	}
	// add the genesis block to the blockchain headers.
	bc.BlockHeaders = append(bc.BlockHeaders, genesis.Header)
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
	// add the block to the storage
	err = bc.Storage.Put(block, bc.Encoder)
	if err != nil {
		return err
	}
	// add the block to the blockchain headers
	bc.BlockHeaders = append(bc.BlockHeaders, block.Header)

	return nil
}

// HasBlock function compares the index of the block with the height of the blockchain.
func (bc *Blockchain) HasBlock(block *Block) bool {
	return block.Header.Index <= bc.GetHeight()
}
