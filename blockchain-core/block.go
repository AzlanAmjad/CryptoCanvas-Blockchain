package core

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"io"

	types "github.com/AzlanAmjad/DreamscapeCanvas-Blockchain/data-types"
)

// BlockHeader is the header of a block in the blockchain.
type BlockHeader struct {
	Version       uint32
	PrevBlockHash types.Hash
	MerkleRoot    types.Hash
	Timestamp     int64
	Index         uint32
	Nonce         uint64
}

// Encode the block header to binary for any writer.
func (bh *BlockHeader) BinaryEncode(w io.Writer) error {
	err := binary.Write(w, binary.LittleEndian, bh.Version)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, bh.PrevBlockHash)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, bh.MerkleRoot)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, bh.Timestamp)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, bh.Index)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, bh.Nonce)
	if err != nil {
		return err
	}
	return nil
}

// Decode the block header from binary for any reader.
func (bh *BlockHeader) BinaryDecode(r io.Reader) error {
	err := binary.Read(r, binary.LittleEndian, &bh.Version)
	if err != nil {
		return err
	}
	err = binary.Read(r, binary.LittleEndian, &bh.PrevBlockHash)
	if err != nil {
		return err
	}
	err = binary.Read(r, binary.LittleEndian, &bh.MerkleRoot)
	if err != nil {
		return err
	}
	err = binary.Read(r, binary.LittleEndian, &bh.Timestamp)
	if err != nil {
		return err
	}
	err = binary.Read(r, binary.LittleEndian, &bh.Index)
	if err != nil {
		return err
	}
	err = binary.Read(r, binary.LittleEndian, &bh.Nonce)
	if err != nil {
		return err
	}
	return nil
}

// Block is a block in the blockchain.
type Block struct {
	Header       BlockHeader
	Transactions []*Transaction
	// cached block header hash
	hash types.Hash
}

// Hash returns the hash of the block.
func (b *Block) Hash() types.Hash {
	// If the hash is already cached, return it
	if !b.hash.IsZero() {
		return b.hash
	}

	buf := &bytes.Buffer{}
	b.Header.BinaryEncode(buf)
	b.hash = sha256.Sum256(buf.Bytes())
	return b.hash
}

// Encode the block to binary for any writer.
func (b *Block) BinaryEncode(w io.Writer) error {
	// Encode the block header
	err := b.Header.BinaryEncode(w)
	if err != nil {
		return err
	}
	// Encode the number of transactions
	err = binary.Write(w, binary.LittleEndian, uint32(len(b.Transactions)))
	if err != nil {
		return err
	}
	// Encode each transaction
	for _, tx := range b.Transactions {
		err = tx.BinaryEncode(w)
		if err != nil {
			return err
		}
	}
	return nil
}

// Decode the block from binary for any reader.
func (b *Block) BinaryDecode(r io.Reader) error {
	// Decode the block header
	err := b.Header.BinaryDecode(r)
	if err != nil {
		return err
	}
	// Decode the number of transactions
	var numTxs uint32
	err = binary.Read(r, binary.LittleEndian, &numTxs)
	if err != nil {
		return err
	}
	// Decode each transaction
	b.Transactions = make([]*Transaction, numTxs)
	for i := uint32(0); i < numTxs; i++ {
		tx := &Transaction{}
		err = tx.BinaryDecode(r)
		if err != nil {
			return err
		}
		b.Transactions[i] = tx
	}
	return nil
}
