package core

import "fmt"

type Validator interface {
	ValidateBlock(block *Block) error
}

// default BlockValidator implementation
type BlockValidator struct {
	BC *Blockchain
}

func NewBlockValidator(bc *Blockchain) *BlockValidator {
	return &BlockValidator{BC: bc}
}

func (bv *BlockValidator) ValidateBlock(block *Block) error {
	// 1. check if the block is nil
	if block == nil {
		return fmt.Errorf("block is nil")
	}
	// 2. check if the block header is nil
	if block.Header == nil {
		return fmt.Errorf("block header is nil")
	}
	// 3. check if blockchain has block
	if bv.BC.HasBlock(block) {
		return fmt.Errorf("block already exists, block index: %d", block.Header.Index)
	}
	// 4. verify the block signature with blocks public key
	if verified, _ := block.VerifySignature(); !verified {
		return fmt.Errorf("block signature is invalid, block index: %d", block.Header.Index)
	}
	return nil
}
