package core

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/gob"
	"encoding/pem"
	"fmt"
	"io"
)

// Each encoder / decoder method to encode and decode the block
// will require a reader or writer, this makes the encoder and decoder
// dynamic, allowing us to use one encoder / decoder for multiple
// writer and reader types.

type Encoder[T any] interface {
	Encode(w io.Writer, v T) error
}

type Decoder[T any] interface {
	Decode(r io.Reader, v T) error
}

// Default block encoder and decoder

type BlockEncoder struct{}

func NewBlockEncoder() *BlockEncoder {
	return &BlockEncoder{}
}

func (e *BlockEncoder) Encode(w io.Writer, b *Block) error {
	enc := gob.NewEncoder(w)
	err := enc.Encode(b.Header)
	if err != nil {
		return fmt.Errorf("failed to encode block header: %s", err)
	}

	// encode number of transactions
	err = enc.Encode(len(b.Transactions))
	if err != nil {
		return fmt.Errorf("failed to encode number of transactions: %s", err)
	}

	// encode the transactions
	for _, tx := range b.Transactions {
		tx_enc := NewTransactionEncoder()
		err = tx.Encode(w, tx_enc)
		if err != nil {
			return fmt.Errorf("failed to encode block header: %s", err)
		}
	}

	// encode if there is a Validator
	validator_exists := b.Validator.Key != nil
	err = enc.Encode(validator_exists)
	if err != nil {
		return err
	}

	if validator_exists {
		// encode the Validator / Public Key of the block
		// Marshal the public key to ASN.1 DER format
		derBytes, err := x509.MarshalPKIXPublicKey(b.Validator.Key)
		if err != nil {
			return err
		}
		// Encode the DER bytes to PEM format
		pemBytes := pem.EncodeToMemory(&pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: derBytes,
		})
		// gob encode the pemBytes
		enc.Encode(pemBytes)
	}

	// encode if there is a signature
	signature_exists := b.Signature != nil
	err = enc.Encode(signature_exists)
	if err != nil {
		return err
	}

	// encode the signature
	if signature_exists {
		// encode the signature
		err = enc.Encode(b.Signature)
		if err != nil {
			return err
		}
	}

	return nil
}

type BlockDecoder struct{}

func NewBlockDecoder() *BlockDecoder {
	return &BlockDecoder{}
}

func (d *BlockDecoder) Decode(r io.Reader, b *Block) error {
	dec := gob.NewDecoder(r)
	err := dec.Decode(&b.Header)
	if err != nil {
		return fmt.Errorf("failed to decode block header: %s", err)
	}

	// decode number of transactions
	var numTransactions int
	err = dec.Decode(&numTransactions)
	if err != nil {
		return fmt.Errorf("failed to decode number of transactions: %s", err)
	}

	// decode the transactions
	for i := 0; i < numTransactions; i++ {
		tx_dec := NewTransactionDecoder()
		tx := Transaction{}
		err = tx.Decode(r, tx_dec)
		if err != nil {
			return fmt.Errorf("failed to decode transaction: %s", err)
		}
		b.Transactions = append(b.Transactions, &tx)
	}

	// decode if there is a Validator
	var validator_exists bool
	err = dec.Decode(&validator_exists)
	if err != nil {
		return err
	}

	if validator_exists {
		// Decode the Validator / Public Key of the block
		var pubKeyBytes []byte
		// Decode from gob
		err = dec.Decode(&pubKeyBytes)
		if err != nil {
			return err
		}
		// Parse the PEM-encoded public key
		block, _ := pem.Decode(pubKeyBytes)
		if block == nil {
			return fmt.Errorf("failed to decode PEM block containing public key")
		}
		pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return err
		}
		ecdsaPubKey, ok := pubKey.(*ecdsa.PublicKey)
		if !ok {
			return fmt.Errorf("failed to convert public key to ECDSA format")
		}
		b.Validator.Key = ecdsaPubKey
	}

	// decode if there is a signature
	var signature_exists bool
	err = dec.Decode(&signature_exists)
	if err != nil {
		return err
	}

	// Decode the signature
	if signature_exists {
		err = dec.Decode(&b.Signature)
		if err != nil {
			return err
		}
	}

	return nil
}

// Default transaction encoder and decoder
type TransactionEncoder struct{}

func NewTransactionEncoder() *TransactionEncoder {
	return &TransactionEncoder{}
}

func (e *TransactionEncoder) Encode(w io.Writer, t *Transaction) error {
	enc := gob.NewEncoder(w)
	err := enc.Encode(t.Data)
	if err != nil {
		return err
	}

	// Encode the public key
	// Marshal the public key to ASN.1 DER format
	derBytes, err := x509.MarshalPKIXPublicKey(t.From.Key)
	if err != nil {
		return err
	}
	// Encode the DER bytes to PEM format
	pemBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: derBytes,
	})
	// gob encode the pemBytes
	enc.Encode(pemBytes)

	// encode the signature
	err = enc.Encode(t.Signature)
	if err != nil {
		return err
	}

	// encode the first seen timestamp
	err = enc.Encode(t.FirstSeen)
	if err != nil {
		return err
	}

	return nil
}

type TransactionDecoder struct{}

func NewTransactionDecoder() *TransactionDecoder {
	return &TransactionDecoder{}
}

func (d *TransactionDecoder) Decode(r io.Reader, t *Transaction) error {
	dec := gob.NewDecoder(r)
	err := dec.Decode(&t.Data)
	if err != nil {
		return err
	}

	// Decode the public key
	var pubKeyBytes []byte
	// Decode from gob
	err = dec.Decode(&pubKeyBytes)
	if err != nil {
		return err
	}
	// Parse the PEM-encoded public key
	block, _ := pem.Decode(pubKeyBytes)
	if block == nil {
		return fmt.Errorf("failed to decode PEM block containing public key")
	}
	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return err
	}
	ecdsaPubKey, ok := pubKey.(*ecdsa.PublicKey)
	if !ok {
		return fmt.Errorf("failed to convert public key to ECDSA format")
	}
	t.From.Key = ecdsaPubKey

	// Decode the signature
	err = dec.Decode(&t.Signature)
	if err != nil {
		return err
	}

	// Decode the first seen timestamp
	err = dec.Decode(&t.FirstSeen)
	if err != nil {
		return err
	}

	return nil
}
