package core

import (
	"fmt"
	"io"

	crypto "github.com/AzlanAmjad/DreamscapeCanvas-Blockchain/cryptography"
	types "github.com/AzlanAmjad/DreamscapeCanvas-Blockchain/data-types"
	"golang.org/x/exp/rand"
)

// native NFT implementation

type TransactionType byte

const (
	TxCollection TransactionType = iota
	TxMint
	TxCryptoTransfer
	TxCommon // placeholder for future transaction types, so we can handle arbitrary transaction types.
)

type CryptoTransferTransaction struct {
	Fee    uint64               // fee for the transaction
	To     crypto.PublicKey     // who we are sending to
	Amount types.CurrencyAmount // how much we are sending
}

// Encode encodes a crypto transfer transaction to a writer, in a modular fashion.
func (t *CryptoTransferTransaction) Encode(w io.Writer, enc Encoder[*CryptoTransferTransaction]) error {
	return enc.Encode(w, t)
}

// Decode decodes a crypto transfer transaction from a reader, in a modular fashion.
func (t *CryptoTransferTransaction) Decode(r io.Reader, dec Decoder[*CryptoTransferTransaction]) error {
	return dec.Decode(r, t)
}

type CollectionTransaction struct {
	Fee      uint64
	Name     string
	Metadata []byte
}

// Encode encodes a collection transaction to a writer, in a modular fashion.
func (t *CollectionTransaction) Encode(w io.Writer, enc Encoder[*CollectionTransaction]) error {
	return enc.Encode(w, t)
}

// Decode decodes a collection transaction from a reader, in a modular fashion.
func (t *CollectionTransaction) Decode(r io.Reader, dec Decoder[*CollectionTransaction]) error {
	return dec.Decode(r, t)
}

type MintTransaction struct {
	Fee               uint64
	NFT               types.Hash // Non-fungible token, the hash of the NFT, can be any real world tangible or digital asset.
	Collection        types.Hash // The hash of the collection to which the NFT belongs.
	CollectionCreator crypto.PublicKey
	Signature         *crypto.Signature
	Metadata          []byte
}

// Function to sign the mint transaction.
func (t *MintTransaction) Sign(privateKey *crypto.PrivateKey) error {
	signature, err := privateKey.Sign(t.NFT[:])
	if err != nil {
		return err
	}

	t.Signature = signature
	t.CollectionCreator = privateKey.GetPublicKey()

	return nil
}

// Function to verify the signature of the mint transaction.
func (t *MintTransaction) VerifySignature() (bool, error) {
	if t.Signature == nil {
		return false, fmt.Errorf("no signature to verify")
	}

	return t.CollectionCreator.Verify(t.NFT[:], t.Signature), nil
}

// Encode encodes a mint transaction to a writer, in a modular fashion.
func (t *MintTransaction) Encode(w io.Writer, enc Encoder[*MintTransaction]) error {
	return enc.Encode(w, t)
}

// Decode decodes a mint transaction from a reader, in a modular fashion.
func (t *MintTransaction) Decode(r io.Reader, dec Decoder[*MintTransaction]) error {
	return dec.Decode(r, t)
}

// Transaction struct to hold types of transactions.
type Transaction struct {
	Type  TransactionType
	Data  []byte // holds the encoded transaction data or any arbitrary data.
	Nonce uint64

	From      crypto.PublicKey
	Signature *crypto.Signature

	// cached hash of the transaction data.
	hash types.Hash
	/*
		firstSeen is the time when the transaction was first seen locally by a node.
		The firstSeen timestamp is used to order transactions in the mempool based on when they were
		first added. This approach simplifies the ordering logic and ensures that transactions are
		processed in the order they were received, without the need for complex synchronization
		mechanisms like logical timestamps or totally ordered broadcast, which are more suitable for
		ordering events across different nodes in a distributed system.
	*/
	FirstSeen int64
}

// NewTransaction creates a new transaction.
func NewTransaction(data []byte) *Transaction {
	random_int := rand.Intn(10000000000000)

	return &Transaction{
		Type:  TxCommon,
		Data:  data,
		Nonce: uint64(random_int),
	}
}

// GetHash returns the hash of the transaction.
func (t *Transaction) GetHash(hasher Hasher[*Transaction]) types.Hash {
	if !t.hash.IsZero() {
		return t.hash
	}

	return hasher.Hash(t)
}

// Sign signs the transaction with the given private key.
func (t *Transaction) Sign(privateKey *crypto.PrivateKey) error {
	signature, err := privateKey.Sign(t.Data)
	if err != nil {
		return err
	}

	t.Signature = signature
	t.From = privateKey.GetPublicKey()

	return nil
}

// VerifySignature verifies the signature of the transaction.
func (t *Transaction) VerifySignature() (bool, error) {
	if t.Signature == nil {
		return false, fmt.Errorf("no transaction signature to verify")
	}

	return t.From.Verify(t.Data, t.Signature), nil
}

// SetFirstSeen sets the firstSeen timestamp of the transaction.
func (t *Transaction) SetFirstSeen(timestamp int64) {
	t.FirstSeen = timestamp
}

// GetFirstSeen returns the firstSeen timestamp of the transaction.
func (t *Transaction) GetFirstSeen() int64 {
	return t.FirstSeen
}

// Encode encodes a transaction to a writer, in a modular fashion.
func (t *Transaction) Encode(w io.Writer, enc Encoder[*Transaction]) error {
	return enc.Encode(w, t)
}

// Decode decodes a transaction from a reader, in a modular fashion.
func (t *Transaction) Decode(r io.Reader, dec Decoder[*Transaction]) error {
	return dec.Decode(r, t)
}
