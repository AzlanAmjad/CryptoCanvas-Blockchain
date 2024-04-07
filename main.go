package main

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"net"
	"net/http"
	"time"

	core "github.com/AzlanAmjad/DreamscapeCanvas-Blockchain/blockchain-core"
	crypto "github.com/AzlanAmjad/DreamscapeCanvas-Blockchain/cryptography"
	types "github.com/AzlanAmjad/DreamscapeCanvas-Blockchain/data-types"
	network "github.com/AzlanAmjad/DreamscapeCanvas-Blockchain/peer-to-peer-network"
	"github.com/sirupsen/logrus"
)

func main() {
	// entry point

	// ask user for node ID
	var id string
	logrus.Info("Enter the node ID: ")
	_, err := fmt.Scanln(&id)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to read node ID")
	}

	// ask user what port to listen on
	var port string
	logrus.Info("Enter the port to listen on: ")
	_, err = fmt.Scanln(&port)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to read port")
	}
	// create net.Addr for a TCP connection
	addr, err := net.ResolveTCPAddr("tcp", "localhost:"+port)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to resolve TCP address")
	}

	// ask user what port to listen on for the API
	var APIport string
	logrus.Info("Enter the port to listen on for the API: ")
	_, err = fmt.Scanln(&APIport)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to read port")
	}
	// create net.Addr for a TCP connection
	APIaddr, err := net.ResolveTCPAddr("tcp", "localhost:"+APIport)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to resolve TCP address")
	}

	// ask user if this node is a validator
	var isValidator string
	logrus.Info("Is this node a validator? (y/n): ")
	_, err = fmt.Scanln(&isValidator)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to read input")
	}

	var localNode *network.Server
	if isValidator == "y" {
		// generate private key
		privateKey := crypto.GeneratePrivateKey()
		// create local node
		localNode = makeServer(id, &privateKey, addr, APIaddr)
	} else {
		// create local node
		localNode = makeServer(id, nil, addr, APIaddr)
	}

	// start local node
	go localNode.Start()

	// start TCP tester
	go tcpTesterTransactionSender()

	select {}
}

func makeServer(id string, privateKey *crypto.PrivateKey, addr net.Addr, APIaddr net.Addr) *network.Server {
	// one seed node, hardcoded at port 8000
	seedAddr, err := net.ResolveTCPAddr("tcp", "localhost:8000")
	if err != nil {
		logrus.WithError(err).Fatal("Failed to resolve TCP address")
	}

	serverOptions := network.ServerOptions{
		Addr:      addr,
		APIAddr:   APIaddr,
		ID:        id,
		SeedNodes: []net.Addr{seedAddr},
	}
	if privateKey != nil {
		serverOptions.PrivateKey = privateKey
	}

	// We will create a new server with the server options.
	server, err := network.NewServer(serverOptions)
	if err != nil {
		panic(err)
	}
	return server
}

func tcpTesterTransactionSender() {
	// create a new http client to send transaction to nodes REST API
	client := http.DefaultClient
	privateKey := crypto.GeneratePrivateKey()

	msg, collection_hash := makeCollectionTransactionMessage(privateKey)
	_, err := client.Post("http://localhost:8080/transaction", "application/octet-stream", bytes.NewReader(msg))
	if err != nil {
		logrus.WithError(err).Fatal("Failed to send collection transaction to API")
	}

	// send the transaction to the API
	for i := 0; i < 20; i++ {
		msg := makeMintTransactionMessage(privateKey, collection_hash)
		_, err := client.Post("http://localhost:8080/transaction", "application/octet-stream", bytes.NewReader(msg))
		if err != nil {
			logrus.WithError(err).Fatal("Failed to send mint transaction to API")
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
