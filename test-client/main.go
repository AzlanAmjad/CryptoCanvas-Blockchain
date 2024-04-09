package main

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"net/http"
	"time"

	core "github.com/AzlanAmjad/DreamscapeCanvas-Blockchain/blockchain-core"
	crypto "github.com/AzlanAmjad/DreamscapeCanvas-Blockchain/cryptography"
	types "github.com/AzlanAmjad/DreamscapeCanvas-Blockchain/data-types"
	network "github.com/AzlanAmjad/DreamscapeCanvas-Blockchain/peer-to-peer-network"
	"github.com/sirupsen/logrus"
)

func main() {
	go tcpTesterTransactionSender()

	select {}
}

func tcpTesterTransactionSender() {
	// create a new http client to send transaction to nodes REST API
	client := http.DefaultClient
	privateKey := crypto.GeneratePrivateKey()
	toPrivateKey := crypto.GeneratePrivateKey()

	msg, collection_hash := makeCollectionTransactionMessage(privateKey)
	_, err := client.Post("http://localhost:8080/transaction", "application/octet-stream", bytes.NewReader(msg))
	if err != nil {
		logrus.WithError(err).Fatal("Failed to send collection transaction to API")
	}

	// send mint transactions to the API
	for i := 0; i < 20; i++ {
		msg := makeMintTransactionMessage(privateKey, collection_hash)
		_, err := client.Post("http://localhost:8080/transaction", "application/octet-stream", bytes.NewReader(msg))
		if err != nil {
			logrus.WithError(err).Fatal("Failed to send mint transaction to API")
		}
		// wait for a bit
		time.Sleep(1000 * time.Millisecond)
	}

	// send crypto transfer transaction to the API
	for i := 0; i < 20; i++ {
		msg := makeCryptoTransferTransactionMessage(privateKey, toPrivateKey)
		_, err := client.Post("http://localhost:8080/transaction", "application/octet-stream", bytes.NewReader(msg))
		if err != nil {
			logrus.WithError(err).Fatal("Failed to send crypto transfer transaction to API")
		}
		// wait for a bit
		time.Sleep(1000 * time.Millisecond)
	}
}

func makeCollectionTransactionMessage(privateKey crypto.PrivateKey) ([]byte, types.Hash) {
	// make collection transaction and encode it
	collection_tx := &core.CollectionTransaction{
		Fee:      200,
		Name:     "Test Collection",
		Metadata: []byte("Test Metadata"),
	}

	buf := bytes.Buffer{}
	err := collection_tx.Encode(&buf, core.NewCollectionTransactionEncoder())
	if err != nil {
		logrus.WithError(err).Fatal("Failed to encode collection transaction")
	}

	tx := core.NewTransaction(buf.Bytes())
	tx.Type = core.TxCollection

	// sign the transaction
	tx.Sign(&privateKey)

	// encode the transaction
	buf = bytes.Buffer{}
	err = tx.Encode(&buf, core.NewTransactionEncoder())
	if err != nil {
		logrus.WithError(err).Fatal("Failed to encode transaction")
	}

	hash := tx.GetHash(core.NewTransactionHasher())
	// create a message
	msg := network.NewMessage(network.Transaction, buf.Bytes())
	return msg.Bytes(), hash
}

func makeMintTransactionMessage(privateKey crypto.PrivateKey, collection types.Hash) []byte {
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to generate random bytes")
	}

	// make mint transaction and encode it
	mint_tx := &core.MintTransaction{
		Fee:        100,
		Metadata:   []byte("Test Metadata"),
		Collection: collection,
		NFT:        sha256.Sum256(randomBytes),
	}
	mint_tx.Sign(&privateKey)

	buf := bytes.Buffer{}
	err = mint_tx.Encode(&buf, core.NewMintTransactionEncoder())
	if err != nil {
		logrus.WithError(err).Fatal("Failed to encode mint transaction")
	}

	tx := core.NewTransaction(buf.Bytes())
	tx.Type = core.TxMint

	// sign the transaction
	tx.Sign(&privateKey)

	// encode the transaction
	buf = bytes.Buffer{}
	err = tx.Encode(&buf, core.NewTransactionEncoder())
	if err != nil {
		logrus.WithError(err).Fatal("Failed to encode transaction")
	}

	// create a message
	msg := network.NewMessage(network.Transaction, buf.Bytes())
	return msg.Bytes()
}

func makeCryptoTransferTransactionMessage(privateKey crypto.PrivateKey, toPrivateKey crypto.PrivateKey) []byte {
	// make crypto transfer transaction and encode it
	crypto_transfer_tx := &core.CryptoTransferTransaction{
		Fee:    100,
		To:     toPrivateKey.GetPublicKey(),
		Amount: 100,
	}

	buf := bytes.Buffer{}
	err := crypto_transfer_tx.Encode(&buf, core.NewCryptoTransferTransactionEncoder())
	if err != nil {
		logrus.WithError(err).Fatal("Failed to encode crypto transfer transaction")
	}

	tx := core.NewTransaction(buf.Bytes())
	tx.Type = core.TxCryptoTransfer

	// sign the transaction
	tx.Sign(&privateKey)

	// encode the transaction
	buf = bytes.Buffer{}
	err = tx.Encode(&buf, core.NewTransactionEncoder())
	if err != nil {
		logrus.WithError(err).Fatal("Failed to encode transaction")
	}

	// create a message
	msg := network.NewMessage(network.Transaction, buf.Bytes())
	return msg.Bytes()
}
