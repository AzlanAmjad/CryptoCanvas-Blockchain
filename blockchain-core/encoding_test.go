package core

import (
	"bytes"
	"encoding/gob"
	"math/big"
	"testing"

	crypto "github.com/AzlanAmjad/DreamscapeCanvas-Blockchain/cryptography"
	"github.com/stretchr/testify/assert"
)

func TestPublicKeyEncodeDecode(t *testing.T) {
	// generate a public key
	privateKey := crypto.GeneratePrivateKey()
	publicKey := privateKey.GetPublicKey()

	// encode the public key
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	err := EncodePublicKey(enc, publicKey.Key)
	if err != nil {
		t.Fatalf("failed to encode public key: %s", err)
	}

	// decode the public key
	dec := gob.NewDecoder(buf)
	decodedPublicKey, err := DecodePublicKey(dec)
	if err != nil {
		t.Fatalf("failed to decode public key: %s", err)
	}

	// compare the public keys
	assert.Equal(t, &publicKey.Key, &decodedPublicKey)
}

func TestEncodePublicKeys(t *testing.T) {
	// encode the same public key twice, but change the address
	// and compare the bytes of both encodings
	privateKey := crypto.GeneratePrivateKey()
	publicKey := privateKey.GetPublicKey()
	publicKey2 := privateKey.GetPublicKey()

	// encode the public keys
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	err := EncodePublicKey(enc, publicKey.Key)
	if err != nil {
		t.Fatalf("failed to encode public key: %s", err)
	}

	buf2 := new(bytes.Buffer)
	enc2 := gob.NewEncoder(buf2)
	err = EncodePublicKey(enc2, publicKey2.Key)
	if err != nil {
		t.Fatalf("failed to encode public key: %s", err)
	}

	// compare the bytes
	assert.Equal(t, buf.Bytes(), buf2.Bytes())
}

func TestEncodeTransactionsByEncodeDecode(t *testing.T) {
	// create one transaction, sign it, copy it over to another address
	// encode both transactions, and compare the bytes
	privateKey := crypto.GeneratePrivateKey()

	// create a transaction
	tx := NewTransaction([]byte("hello"))
	tx.Sign(&privateKey)

	// encode the tx
	buf := new(bytes.Buffer)
	err := tx.Encode(buf, NewTransactionEncoder())
	if err != nil {
		t.Fatalf("failed to encode transaction: %s", err)
	}

	// decode tx into tx2
	tx2 := NewTransaction([]byte("hello"))
	err = tx2.Decode(buf, NewTransactionDecoder())
	if err != nil {
		t.Fatalf("failed to decode transaction: %s", err)
	}

	// encode tx2
	buf2 := new(bytes.Buffer)
	err = tx2.Encode(buf2, NewTransactionEncoder())
	if err != nil {
		t.Fatalf("failed to encode transaction: %s", err)
	}

	// encode tx
	buf3 := new(bytes.Buffer)
	err = tx.Encode(buf3, NewTransactionEncoder())
	if err != nil {
		t.Fatalf("failed to encode transaction: %s", err)
	}

	// compare the bytes
	assert.Equal(t, buf3.Bytes(), buf2.Bytes())
	assert.Equal(t, tx, tx2)
}

func TestEncodeTransactionsByManualAssignment(t *testing.T) {
	// private key
	privateKey := crypto.GeneratePrivateKey()

	// create a transaction
	tx := Transaction{
		Data: []byte("hello"),
		From: privateKey.GetPublicKey(),
		Signature: &crypto.Signature{
			R: &big.Int{},
			S: &big.Int{},
		},
	}

	// encode the tx
	buf := new(bytes.Buffer)
	err := tx.Encode(buf, NewTransactionEncoder())
	if err != nil {
		t.Fatalf("failed to encode transaction: %s", err)
	}

	// create a new transaction
	tx2 := Transaction{
		Data: []byte("hello"),
		From: privateKey.GetPublicKey(),
		Signature: &crypto.Signature{
			R: &big.Int{},
			S: &big.Int{},
		},
	}

	// encode the transaction
	buf2 := new(bytes.Buffer)
	err = tx2.Encode(buf2, NewTransactionEncoder())
	if err != nil {
		t.Fatalf("failed to encode transaction: %s", err)
	}

	// compare the bytes
	assert.Equal(t, buf.Bytes(), buf2.Bytes())
	assert.Equal(t, tx.Data, tx2.Data)
	assert.Equal(t, tx.From, tx2.From)
	assert.Equal(t, tx.Signature, tx2.Signature)
}
