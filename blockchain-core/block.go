package core

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"io"
	"time"

	crypto "github.com/AzlanAmjad/DreamscapeCanvas-Blockchain/cryptography"
	types "github.com/AzlanAmjad/DreamscapeCanvas-Blockchain/data-types"
)

// BlockHeader is the header of a block in the blockchain.
type BlockHeader struct {
	Version       uint32
	PrevBlockHash types.Hash
	DataHash      types.Hash // transactions merkle root hash, used to verify the integrity of the transactions
	Timestamp     int64
	Index         uint32
	Nonce         uint64
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

	// everything below should not be encoded
	// created on NewBlock()

	// cached block header hash
	hash types.Hash

	// transaction encoder and decoder
	TransactionEncoder Encoder[*Transaction]
	TransactionDecoder Decoder[*Transaction]

	// transaction hasher
	TransactionHasher Hasher[*Transaction]
}

// NewBlock creates a new block. Returns a default block to use.
// This API should always be used instead of explicitly creating a block.
func NewBlock() *Block {
	return &Block{
		Header: &BlockHeader{
			Version:       1,
			PrevBlockHash: types.Hash{},
			DataHash:      types.Hash{},
			Timestamp:     time.Now().UnixNano(),
			Index:         0,
			Nonce:         0,
		},
		Transactions: []*Transaction{},
		Validator:    crypto.PublicKey{},

		// default transaction encoder and decoder
		TransactionEncoder: NewTransactionEncoder(),
		TransactionDecoder: NewTransactionDecoder(),
		// default transaction hasher
		TransactionHasher: NewTransactionHasher(),
	}
}

func NewBlockWithTransactions(transactions []*Transaction) *Block {
	b := NewBlock()
	b.Transactions = transactions

	// calculate data hash
	dataHash, err := CalculateDataHash(b.Transactions, b.TransactionHasher)
	if err != nil {
		panic(err)
	}
	b.Header.DataHash = dataHash

	return b
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
		return false, fmt.Errorf("no block signature to verify")
	}

	// Verify the block transactions
	for _, tx := range b.Transactions {
		// invalid signature, and transaction is not from coinbase account
		// coinbase account initiated transactions are not signed, therefor not verified
		// TODO: this is not a secure method of creating coinbase transactions, must find a better way
		pub_key, _ := GetCoinbaseAccount()
		if valid, err := tx.VerifySignature(); !valid && tx.From != *pub_key {
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

// GetHash returns the hash of the block.
func (b *Block) GetHash(hasher Hasher[*BlockHeader]) types.Hash {
	// if the hash is already cached, return it
	if !b.hash.IsZero() {
		return b.hash
	}

	return hasher.Hash(b.Header)
}

// Calculate DataHash from transactions
func CalculateDataHash(transactions []*Transaction, hasher Hasher[*Transaction]) (types.Hash, error) {
	final_buf := []byte{}

	for _, tx := range transactions {
		hash := tx.GetHash(hasher)
		// append
		final_buf = append(final_buf, hash[:]...)
	}

	hash := sha256.Sum256(final_buf)

	return types.Hash(hash), nil
}
