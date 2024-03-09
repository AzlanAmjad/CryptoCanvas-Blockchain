package core

import (
	"fmt"

	crypto "github.com/AzlanAmjad/DreamscapeCanvas-Blockchain/cryptography"
	types "github.com/AzlanAmjad/DreamscapeCanvas-Blockchain/data-types"
)

type Transaction struct {
	Data []byte

	From      crypto.PublicKey
	Signature *crypto.Signature

	// cached hash of the transaction data.
	hash types.Hash
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
