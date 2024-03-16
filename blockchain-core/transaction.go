package core

import (
	"fmt"
	"io"

	crypto "github.com/AzlanAmjad/DreamscapeCanvas-Blockchain/cryptography"
	types "github.com/AzlanAmjad/DreamscapeCanvas-Blockchain/data-types"
)

type Transaction struct {
	Data []byte

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
	return &Transaction{
		Data: data,
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
		return false, fmt.Errorf("no signature to verify")
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
