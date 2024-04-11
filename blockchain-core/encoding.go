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

func EncodePublicKey(enc *gob.Encoder, key *ecdsa.PublicKey) error {
	// encode the Validator / Public Key of the block
	// Marshal the public key to ASN.1 DER format
	derBytes, err := x509.MarshalPKIXPublicKey(key)
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
	return nil
}

func DecodePublicKey(dec *gob.Decoder) (*ecdsa.PublicKey, error) {
	// Decode the Validator / Public Key of the block
	var pubKeyBytes []byte
	// Decode from gob
	err := dec.Decode(&pubKeyBytes)
	if err != nil {
		return nil, err
	}
	// Parse the PEM-encoded public key
	block, _ := pem.Decode(pubKeyBytes)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block containing public key")
	}
	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	ecdsaPubKey, ok := pubKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("failed to convert public key to ECDSA format")
	}
	return ecdsaPubKey, nil
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
		// encode the public key
		err = EncodePublicKey(enc, b.Validator.Key)
		if err != nil {
			return err
		}
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
		tx := NewTransaction(nil)
		err = tx.Decode(r, tx_dec)
		if err != nil {
			return fmt.Errorf("failed to decode transaction: %s", err)
		}
		b.Transactions = append(b.Transactions, tx)
	}

	// decode if there is a Validator
	var validator_exists bool
	err = dec.Decode(&validator_exists)
	if err != nil {
		return err
	}

	if validator_exists {
		// decode the public key
		b.Validator.Key, err = DecodePublicKey(dec)
		if err != nil {
			return err
		}
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

	// Encode the transaction type
	err := enc.Encode(t.Type)
	if err != nil {
		return err
	}

	// Encode the data
	err = enc.Encode(t.Data)
	if err != nil {
		return err
	}

	// Encode the Nonce
	err = enc.Encode(t.Nonce)
	if err != nil {
		return err
	}

	// Encode the public key
	err = EncodePublicKey(enc, t.From.Key)
	if err != nil {
		return err
	}

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

	// Decode the transaction type
	err := dec.Decode(&t.Type)
	if err != nil {
		return err
	}

	// Decode the data
	err = dec.Decode(&t.Data)
	if err != nil {
		return err
	}

	// Decode the Nonce
	err = dec.Decode(&t.Nonce)
	if err != nil {
		return err
	}

	// Decode the public key
	t.From.Key, err = DecodePublicKey(dec)
	if err != nil {
		return err
	}

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

	fmt.Print("SUCCESS 8")

	return nil
}

// Default collection transaction encoder and decoder
type CollectionTransactionEncoder struct{}

func NewCollectionTransactionEncoder() *CollectionTransactionEncoder {
	return &CollectionTransactionEncoder{}
}

func (e *CollectionTransactionEncoder) Encode(w io.Writer, t *CollectionTransaction) error {
	enc := gob.NewEncoder(w)

	// encode the fee
	err := enc.Encode(t.Fee)
	if err != nil {
		return err
	}

	// encode the name
	err = enc.Encode(t.Name)
	if err != nil {
		return err
	}

	// encode the metadata
	err = enc.Encode(t.Metadata)
	if err != nil {
		return err
	}

	return nil
}

type CollectionTransactionDecoder struct{}

func NewCollectionTransactionDecoder() *CollectionTransactionDecoder {
	return &CollectionTransactionDecoder{}
}

func (d *CollectionTransactionDecoder) Decode(r io.Reader, t *CollectionTransaction) error {
	dec := gob.NewDecoder(r)

	// decode the fee
	err := dec.Decode(&t.Fee)
	if err != nil {
		return err
	}

	// decode the name
	err = dec.Decode(&t.Name)
	if err != nil {
		return err
	}

	// decode the metadata
	err = dec.Decode(&t.Metadata)
	if err != nil {
		return err
	}

	return nil
}

// Default mint transaction encoder and decoder
type MintTransactionEncoder struct{}

func NewMintTransactionEncoder() *MintTransactionEncoder {
	return &MintTransactionEncoder{}
}

func (e *MintTransactionEncoder) Encode(w io.Writer, t *MintTransaction) error {
	enc := gob.NewEncoder(w)

	// encode the fee
	err := enc.Encode(t.Fee)
	if err != nil {
		return err
	}

	// encode the NFT
	err = enc.Encode(t.NFT)
	if err != nil {
		return err
	}

	// encode the collection
	err = enc.Encode(t.Collection)
	if err != nil {
		return err
	}

	// encode if there is a collection creator
	collection_creator_exists := t.CollectionCreator.Key != nil
	err = enc.Encode(collection_creator_exists)
	if err != nil {
		return err
	}

	if collection_creator_exists {
		// encode the collection creator / public key
		err = EncodePublicKey(enc, t.CollectionCreator.Key)
		if err != nil {
			return err
		}
	}

	// encode if there is a signature
	signature_exists := t.Signature != nil
	err = enc.Encode(signature_exists)
	if err != nil {
		return err
	}

	if signature_exists {
		// encode the signature
		err = enc.Encode(t.Signature)
		if err != nil {
			return err
		}
	}

	// encode the metadata
	err = enc.Encode(t.Metadata)
	if err != nil {
		return err
	}

	return nil
}

type MintTransactionDecoder struct{}

func NewMintTransactionDecoder() *MintTransactionDecoder {
	return &MintTransactionDecoder{}
}

func (d *MintTransactionDecoder) Decode(r io.Reader, t *MintTransaction) error {
	dec := gob.NewDecoder(r)

	// decode the fee
	err := dec.Decode(&t.Fee)
	if err != nil {
		return err
	}

	// decode the NFT
	err = dec.Decode(&t.NFT)
	if err != nil {
		return err
	}

	// decode the collection
	err = dec.Decode(&t.Collection)
	if err != nil {
		return err
	}

	// decode if there is a collection creator
	var collection_creator_exists bool
	err = dec.Decode(&collection_creator_exists)
	if err != nil {
		return err
	}

	if collection_creator_exists {
		// decode the collection creator / public key
		t.CollectionCreator.Key, err = DecodePublicKey(dec)
		if err != nil {
			return err
		}
	}

	// decode if there is a signature
	var signature_exists bool
	err = dec.Decode(&signature_exists)
	if err != nil {
		return err
	}

	if signature_exists {
		// decode the signature
		err = dec.Decode(&t.Signature)
		if err != nil {
			return err
		}
	}

	// decode the metadata
	err = dec.Decode(&t.Metadata)
	if err != nil {
		return err
	}

	return nil
}

// Default transfer transaction encoder and decoder
type CryptoTransferTransactionEncoder struct{}

func NewCryptoTransferTransactionEncoder() *CryptoTransferTransactionEncoder {
	return &CryptoTransferTransactionEncoder{}
}

func (e *CryptoTransferTransactionEncoder) Encode(w io.Writer, t *CryptoTransferTransaction) error {
	enc := gob.NewEncoder(w)

	// encode the fee
	err := enc.Encode(t.Fee)
	if err != nil {
		return err
	}

	// encode if the to / public key exists
	to_exists := t.To.Key != nil
	err = enc.Encode(to_exists)
	if err != nil {
		return err
	}

	if to_exists {
		// encode the to / public key
		err = EncodePublicKey(enc, t.To.Key)
		if err != nil {
			return err
		}
	}

	// encode the amount
	err = enc.Encode(t.Amount)
	if err != nil {
		return err
	}

	return nil
}

type CryptoTransferTransactionDecoder struct{}

func NewCryptoTransferTransactionDecoder() *CryptoTransferTransactionDecoder {
	return &CryptoTransferTransactionDecoder{}
}

func (d *CryptoTransferTransactionDecoder) Decode(r io.Reader, t *CryptoTransferTransaction) error {
	dec := gob.NewDecoder(r)

	// decode the fee
	err := dec.Decode(&t.Fee)
	if err != nil {
		return err
	}

	// decode if the to / public key exists
	var to_exists bool
	err = dec.Decode(&to_exists)
	if err != nil {
		return err
	}

	if to_exists {
		// decode the to / public key
		t.To.Key, err = DecodePublicKey(dec)
		if err != nil {
			return err
		}
	}

	// decode the amount
	err = dec.Decode(&t.Amount)
	if err != nil {
		return err
	}

	return nil
}
