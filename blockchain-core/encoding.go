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
	err := enc.Encode(b.Header.Index)
	if err != nil {
		return fmt.Errorf("failed to encode block header: %s", err)
	}

	// TODO (Azlan): Finish Encoding the block

	return nil
}

type BlockDecoder struct{}

func NewBlockDecoder() *BlockDecoder {
	return &BlockDecoder{}
}

func (d *BlockDecoder) Decode(r io.Reader, b *Block) error {
	dec := gob.NewDecoder(r)
	Index := b.Header.Index
	err := dec.Decode(&Index)
	if err != nil {
		return fmt.Errorf("failed to decode block header: %s", err)
	}

	// TODO (Azlan): Finish Decoding the block

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
	// gop encode the pemBytes
	enc.Encode(pemBytes)

	// encode the signature
	err = enc.Encode(t.Signature)

	return err
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

	return nil
}
