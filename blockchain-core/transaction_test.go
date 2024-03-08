package core

import (
	"math/big"
	"testing"

	crypto "github.com/AzlanAmjad/DreamscapeCanvas-Blockchain/cryptography"
	"github.com/stretchr/testify/assert"
)

func TestTransactionSignAndVerifyPass(t *testing.T) {
	// Create a new private key.
	privateKey := crypto.GeneratePrivateKey()

	// Create a new transaction.
	transaction := &Transaction{
		Data: []byte("Hello, world!"),
	}

	// Sign the transaction with the private key.
	err := transaction.Sign(&privateKey)
	if err != nil {
		t.Fatal(err)
	}

	// Verify the signature.
	verified, err := transaction.VerifySignature()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, true, verified)
}

func TestTransactionSignAndVerifyFailChangedData(t *testing.T) {
	// Create a new private key.
	privateKey := crypto.GeneratePrivateKey()

	// Create a new transaction.
	transaction := &Transaction{
		Data: []byte("Hello, world!"),
	}

	// Sign the transaction with the private key.
	err := transaction.Sign(&privateKey)
	if err != nil {
		t.Fatal(err)
	}

	// Modify the transaction data.
	transaction.Data = []byte("Goodbye, world!")

	// Verify the signature.
	verified, err := transaction.VerifySignature()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, false, verified)
}

func TestTransactionSignAndVerifyFailNoSignature(t *testing.T) {
	// Create a new transaction.
	transaction := &Transaction{
		Data: []byte("Hello, world!"),
	}

	// Verify the signature.
	verified, err := transaction.VerifySignature()
	assert.Equal(t, false, verified)
	assert.Error(t, err)
}

func TestTransactionSignAndVerifyFailInvalidSignature(t *testing.T) {
	// Create a new private key.
	privateKey := crypto.GeneratePrivateKey()

	// Create a new transaction.
	transaction := &Transaction{
		Data: []byte("Hello, world!"),
	}

	// Sign the transaction with the private key.
	err := transaction.Sign(&privateKey)
	if err != nil {
		t.Fatal(err)
	}

	// Modify the signature.
	transaction.Signature = &crypto.Signature{
		R: new(big.Int),
		S: new(big.Int),
	}

	// Verify the signature.
	verified, err := transaction.VerifySignature()
	assert.Equal(t, false, verified)
	assert.NoError(t, err)
}

func TestTransactionSignAndVerifyFailInvalidPublicKey(t *testing.T) {
	// Create a new private key.
	privateKey := crypto.GeneratePrivateKey()

	// Create a new transaction.
	transaction := &Transaction{
		Data: []byte("Hello, world!"),
	}

	// Sign the transaction with the private key.
	err := transaction.Sign(&privateKey)
	if err != nil {
		t.Fatal(err)
	}

	// Modify the public key.
	otherPrivKey := crypto.GeneratePrivateKey()
	transaction.PublicKey = otherPrivKey.GetPublicKey()

	// Verify the signature.
	verified, err := transaction.VerifySignature()
	assert.Equal(t, false, verified)
	assert.NoError(t, err)
}
