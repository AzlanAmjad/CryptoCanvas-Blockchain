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
	// 5. check if the block index is valid
	if block.Header.Index != bv.BC.GetHeight()+1 {
		return fmt.Errorf("block index is invalid, block index: %d, blockchain height: %d", block.Header.Index, bv.BC.GetHeight())
	}
	// 6. check if the block previous hash is valid
	if block.Header.Index > 0 { // if the block is not the genesis block
		prevBlockHeader, err := bv.BC.GetHeaderByIndex(block.Header.Index - 1)
		if err != nil {
			return fmt.Errorf("failed to get previous block header, block index: %d, blockchain height: %d", block.Header.Index, bv.BC.GetHeight())
		}
		if prevBlockHeader == nil {
			return fmt.Errorf("previous block header is nil, block index: %d, blockchain height: %d", block.Header.Index, bv.BC.GetHeight())
		}
		if bv.BC.BlockHeaderHasher.Hash(prevBlockHeader) != block.Header.PrevBlockHash {
			return fmt.Errorf("previous hash is invalid, block index: %d, blockchain height: %d", block.Header.Index, bv.BC.GetHeight())
		}
	}
	// 7. check if the block data hash is valid
	if block.Header.Index > 0 { // if the block is not the genesis block
		dataHash, err := CalculateDataHash(block.Transactions, block.transactionEncoder)
		if err != nil {
			return fmt.Errorf("failed to calculate data hash, block index: %d, blockchain height: %d", block.Header.Index, bv.BC.GetHeight())
		}
		if dataHash != block.Header.DataHash {
			return fmt.Errorf("data hash is invalid, block index: %d, blockchain height: %d", block.Header.Index, bv.BC.GetHeight())
		}
	}
	return nil
}
