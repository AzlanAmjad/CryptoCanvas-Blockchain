package core

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"

	crypto "github.com/AzlanAmjad/DreamscapeCanvas-Blockchain/cryptography"
	types "github.com/AzlanAmjad/DreamscapeCanvas-Blockchain/data-types"
)

// BlockHeader is the header of a block in the blockchain.
type BlockHeader struct {
	Version       uint32
	PrevBlockHash types.Hash
	MerkleRoot    types.Hash // transactions merkle root hash, used to verify the integrity of the transactions
	Timestamp     int64
	Index         uint32
}

// GetBytes returns the bytes of the block header.
func (bh *BlockHeader) GetBytes() []byte {
	buf := &bytes.Buffer{}
	encoder := gob.NewEncoder(buf)
	err := encoder.Encode(bh)
	if err != nil {
		panic(err)
	}
	return buf.Bytes()
}

// Block is a block in the blockchain.
type Block struct {
	Header       *BlockHeader
	Transactions []*Transaction
	Validator    crypto.PublicKey
	Signature    *crypto.Signature

	// cached block header hash
	hash types.Hash
}

// Add transaction to block
func (b *Block) AddTransaction(tx *Transaction) {
	b.Transactions = append(b.Transactions, tx)
}

// Sign the block with a private key.
func (b *Block) Sign(privateKey *crypto.PrivateKey) error {
	signature, err := privateKey.Sign(b.Header.GetBytes())
	if err != nil {
		return err
	}

	b.Validator = privateKey.GetPublicKey()
	b.Signature = signature

	return nil
}

// VerifySignature verifies the signature of the block.
func (b *Block) VerifySignature() (bool, error) {
	if b.Signature == nil {
		return false, fmt.Errorf("no signature to verify")
	}

	// Verify the block transactions
	for _, tx := range b.Transactions {
		if valid, err := tx.VerifySignature(); !valid {
			return false, err
		}
	}

	// Verify the block signature
	return b.Validator.Verify(b.Header.GetBytes(), b.Signature), nil
}

// Decode decodes a block from a reader, in a modular fashion.
func (b *Block) Decode(r io.Reader, dec Decoder[*Block]) error {
	return dec.Decode(r, b)
}

// Encode encodes a block to a writer, in a modular fashion.
func (b *Block) Encode(w io.Writer, enc Encoder[*Block]) error {
	return enc.Encode(w, b)
}

func (b *Block) GetHash(hasher Hasher[*BlockHeader]) types.Hash {
	// if the hash is already cached, return it
	if !b.hash.IsZero() {
		return b.hash
	}

	return hasher.Hash(b.Header)
}
