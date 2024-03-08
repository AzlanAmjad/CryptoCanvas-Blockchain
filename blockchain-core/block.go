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

// Block is a block in the blockchain.
type Block struct {
	Header       *BlockHeader
	Transactions []*Transaction
	PublicKey    crypto.PublicKey
	Signature    *crypto.Signature

	// cached block header hash
	hash types.Hash
}

// Sign the block with a private key.
func (b *Block) Sign(privateKey *crypto.PrivateKey) error {
	signature, err := privateKey.Sign(b.GetHeaderBytes())
	if err != nil {
		return err
	}

	b.PublicKey = privateKey.GetPublicKey()
	b.Signature = signature

	return nil
}

// VerifySignature verifies the signature of the block.
func (b *Block) VerifySignature() (bool, error) {
	if b.Signature == nil {
		return false, fmt.Errorf("no signature to verify")
	}

	return b.PublicKey.Verify(b.GetHeaderBytes(), b.Signature), nil
}

// Decode decodes a block from a reader, in a modular fashion.
func (b *Block) Decode(r io.Reader, dec Decoder[*Block]) error {
	return dec.Decode(r, b)
}

// Encode encodes a block to a writer, in a modular fashion.
func (b *Block) Encode(w io.Writer, enc Encoder[*Block]) error {
	return enc.Encode(w, b)
}

func (b *Block) GetHash(hasher Hasher[*Block]) types.Hash {
	// if the hash is already cached, return it
	if !b.hash.IsZero() {
		return b.hash
	}

	return hasher.Hash(b)
}

// GetHeaderBytes returns the bytes of the block header.
func (b *Block) GetHeaderBytes() []byte {
	buf := &bytes.Buffer{}
	encoder := gob.NewEncoder(buf)
	err := encoder.Encode(b.Header)
	if err != nil {
		panic(err)
	}
	return buf.Bytes()
}
