package core

import (
	"bytes"
	"math/big"
	"testing"

	crypto "github.com/AzlanAmjad/DreamscapeCanvas-Blockchain/cryptography"
	"github.com/stretchr/testify/assert"
)

func randomTransactionWithSignature(t *testing.T) *Transaction {
	// Create a new private key.
	privateKey := crypto.GeneratePrivateKey()

	// Create a new transaction.
	transaction := NewTransaction([]byte("Hello, world!"))

	// Sign the transaction with the private key.
	err := transaction.Sign(&privateKey)
	if err != nil {
		t.Fatal(err)
	}

	return transaction
}

func TestCryptoTransferTransactionEncodeAndDecode(t *testing.T) {
	// Create a new crypto transfer transaction.
	privKey := crypto.GeneratePrivateKey()
	transaction := &CryptoTransferTransaction{
		Fee:    100,
		Amount: 1000,
		To:     privKey.GetPublicKey(),
	}

	// Encode the transaction.
	buf := bytes.Buffer{}
	err := transaction.Encode(&buf, NewCryptoTransferTransactionEncoder())
	if err != nil {
		t.Fatal(err)
	}

	// Decode the transaction.
	decoded_transaction := &CryptoTransferTransaction{}
	err = decoded_transaction.Decode(&buf, NewCryptoTransferTransactionDecoder())
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, transaction.Fee, decoded_transaction.Fee)
	assert.Equal(t, transaction.Amount, decoded_transaction.Amount)
	assert.Equal(t, transaction.To, decoded_transaction.To)
	assert.Equal(t, transaction, decoded_transaction)
}

func TestMintTransactionEncodeAndDecode(t *testing.T) {
	// Create a new mint transaction.
	transaction := &MintTransaction{
		Fee:        100,
		NFT:        [32]byte{0x01},
		Collection: [32]byte{0x02},
		Metadata:   []byte("Hello, world!"),
	}

	// create a private key
	privateKey := crypto.GeneratePrivateKey()
	// sign the transaction
	transaction.Sign(&privateKey)

	// Encode the transaction.
	buf := bytes.Buffer{}
	err := transaction.Encode(&buf, NewMintTransactionEncoder())
	if err != nil {
		t.Fatal(err)
	}

	// Decode the transaction.
	decoded_transaction := &MintTransaction{}
	err = decoded_transaction.Decode(&buf, NewMintTransactionDecoder())
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, transaction.Fee, decoded_transaction.Fee)
	assert.Equal(t, transaction.NFT, decoded_transaction.NFT)
	assert.Equal(t, transaction.Collection, decoded_transaction.Collection)
	assert.Equal(t, transaction.Metadata, decoded_transaction.Metadata)
	assert.Equal(t, transaction, decoded_transaction)
}

func TestCollectionTransactionEncodeAndDecode(t *testing.T) {
	// Create a new collection transaction.
	transaction := &CollectionTransaction{
		Fee:      200,
		Name:     "Test Collection Transaction",
		Metadata: []byte("Hello, world!"),
	}

	// Encode the transaction.
	buf := bytes.Buffer{}
	err := transaction.Encode(&buf, NewCollectionTransactionEncoder())
	if err != nil {
		t.Fatal(err)
	}

	// Decode the transaction.
	decoded_transaction := &CollectionTransaction{}
	err = decoded_transaction.Decode(&buf, NewCollectionTransactionDecoder())
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, transaction.Fee, decoded_transaction.Fee)
	assert.Equal(t, transaction.Name, decoded_transaction.Name)
	assert.Equal(t, transaction.Metadata, decoded_transaction.Metadata)
	assert.Equal(t, transaction, decoded_transaction)
}

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
	transaction.From = otherPrivKey.GetPublicKey()

	// Verify the signature.
	verified, err := transaction.VerifySignature()
	assert.Equal(t, false, verified)
	assert.NoError(t, err)
}

func TestTransactionEncodeAndDecode(t *testing.T) {
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

	// Encode the transaction.
	buf := bytes.Buffer{}
	err = transaction.Encode(&buf, NewTransactionEncoder())
	if err != nil {
		t.Fatal(err)
	}

	// Decode the transaction.
	decoded_transaction := &Transaction{}
	err = decoded_transaction.Decode(&buf, NewTransactionDecoder())
	if err != nil {
		t.Fatal(err)
	}

	// Verify the signature.
	verified, err := decoded_transaction.VerifySignature()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, true, verified)
	assert.Equal(t, transaction.Data, decoded_transaction.Data)
	assert.Equal(t, transaction.From, decoded_transaction.From)
	assert.Equal(t, transaction.Signature, decoded_transaction.Signature)
}
