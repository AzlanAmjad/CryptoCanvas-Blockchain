package core

import "fmt"

type Validator interface {
	ValidateBlock(block *Block) (int, error)
}

// default BlockValidator implementation
type BlockValidator struct {
	BC *Blockchain
}

func NewBlockValidator(bc *Blockchain) *BlockValidator {
	return &BlockValidator{BC: bc}
}

// function return error code, and error message
// Error Codes:
// 0: no error
// 1: block is nil
// 2: block header is nil
// 3: block already exists
// 4: block signature is invalid
// 5: block index is greater than blockchain height
// 6: block index is lower than blockchain height
// 7: previous hash is invalid
// 8: data hash is invalid
func (bv *BlockValidator) ValidateBlock(block *Block) (int, error) {
	// 1. check if the block is nil
	if block == nil {
		return 1, fmt.Errorf("block is nil")
	}
	// 2. check if the block header is nil
	if block.Header == nil {
		return 2, fmt.Errorf("block header is nil")
	}
	// 3. check if blockchain has block
	if bv.BC.HasBlock(block) {
		return 3, fmt.Errorf("block already exists, block index: %d", block.Header.Index)
	}
	// 4. verify the block signature with blocks public key
	if verified, _ := block.VerifySignature(); !verified {
		return 4, fmt.Errorf("block signature is invalid, block index: %d", block.Header.Index)
	}
	// 5. check if the block index is larger than the expected blockchain height
	if block.Header.Index > bv.BC.GetHeight()+1 {
		return 5, fmt.Errorf("block index is invalid, block index: %d, blockchain height: %d", block.Header.Index, bv.BC.GetHeight())
	}
	// 6. check if the block index is lower than the expected blockchain height
	if block.Header.Index < bv.BC.GetHeight()+1 {
		return 6, fmt.Errorf("block index is invalid, block index: %d, blockchain height: %d", block.Header.Index, bv.BC.GetHeight())
	}
	// 7. check if the block previous hash is valid
	if block.Header.Index > 0 { // if the block is not the genesis block
		prevBlockHeader, err := bv.BC.GetHeaderByIndex(block.Header.Index - 1)
		if err != nil {
			return 7, fmt.Errorf("failed to get previous block header, block index: %d, blockchain height: %d", block.Header.Index, bv.BC.GetHeight())
		}
		if prevBlockHeader == nil {
			return 7, fmt.Errorf("previous block header is nil, block index: %d, blockchain height: %d", block.Header.Index, bv.BC.GetHeight())
		}
		if bv.BC.BlockHeaderHasher.Hash(prevBlockHeader) != block.Header.PrevBlockHash {
			return 7, fmt.Errorf("previous hash is invalid, block index: %d, blockchain height: %d", block.Header.Index, bv.BC.GetHeight())
		}
	}
	// 8. check if the block data hash is valid
	if block.Header.Index > 0 { // if the block is not the genesis block
		dataHash, err := CalculateDataHash(block.Transactions, block.TransactionEncoder)
		if err != nil {
			return 8, fmt.Errorf("failed to calculate data hash, block index: %d, blockchain height: %d", block.Header.Index, bv.BC.GetHeight())
		}
		if dataHash != block.Header.DataHash {
			return 8, fmt.Errorf("data hash is invalid, block index: %d, blockchain height: %d", block.Header.Index, bv.BC.GetHeight())
		}
	}
	return 0, nil
}
