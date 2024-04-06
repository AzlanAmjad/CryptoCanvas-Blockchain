package main

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"net"
	"net/http"
	"time"

	core "github.com/AzlanAmjad/DreamscapeCanvas-Blockchain/blockchain-core"
	crypto "github.com/AzlanAmjad/DreamscapeCanvas-Blockchain/cryptography"
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

	// send the transaction to the API
	for {
		msg := makeTransactionMessage()
		_, err := client.Post("http://localhost:8080/transaction", "application/octet-stream", bytes.NewReader(msg))
		if err != nil {
			logrus.WithError(err).Fatal("Failed to send transaction to API")
		}
		// wait for a bit
		time.Sleep(500 * time.Millisecond)
	}
}

func makeTransactionMessage() []byte {
	// Generate a random byte slice
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to generate random bytes")
	}

	tx := core.NewTransaction(randomBytes)

	// generate private key
	privateKey := crypto.GeneratePrivateKey()
	// sign the transaction
	tx.Sign(&privateKey)

	// encode the transaction
	buf := bytes.Buffer{}
	enc := core.NewTransactionEncoder()
	tx.Encode(&buf, enc)

	// create a message
	msg := network.NewMessage(network.Transaction, buf.Bytes())
	return msg.Bytes()
}
