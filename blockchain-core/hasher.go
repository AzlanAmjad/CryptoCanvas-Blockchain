package core

import (
	"crypto/sha256"

	types "github.com/AzlanAmjad/DreamscapeCanvas-Blockchain/data-types"
)

type Hasher[T any] interface {
	Hash(T) types.Hash
}

// BlockHeaderHasher is a hasher for block headers. Implementation of the Hasher interface.
type BlockHasher struct{}

func (h *BlockHasher) Hash(b *Block) types.Hash {
	blockHeaderHash := sha256.Sum256(b.GetHeaderBytes())
	return types.Hash(blockHeaderHash[:])
}
