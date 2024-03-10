package network

import (
	"bytes"
	"fmt"
	"time"

	core "github.com/AzlanAmjad/DreamscapeCanvas-Blockchain/blockchain-core"
	crypto "github.com/AzlanAmjad/DreamscapeCanvas-Blockchain/cryptography"
	"github.com/sirupsen/logrus"
)

// The server is a container which will contain every module of the server.
// Nodes / servers can BlockExplorers, Wallets, and other nodes, to strengthen the network.
// Nodes / servers can can also be validators, that participate in the consensus algorithm. They
// can be elected to create new blocks, and validate transactions, and propose blocks into the network.

var defaultBlockTime = 5 * time.Second

// One node can have multiple transports, for example, a TCP transport and a UDP transport.
type ServerOptions struct {
	RPCDecodeFunc RPCDecodeFunc
	RPCProcessor  RPCProcessor
	Transports    []Transport
	PrivateKey    *crypto.PrivateKey
	// if a node / server is elect as a validator, server needs to know when its time to consume its mempool and create a new block.
	BlockTime time.Duration
}

type Server struct {
	ServerOptions ServerOptions
	isValidator   bool
	// holds transactions that are not yet included in a block.
	memPool     *TxPool
	rpcChannel  chan ReceiveRPC
	quitChannel chan bool
}

// NewServer creates a new server with the given options.
func NewServer(options ServerOptions) *Server {
	if options.BlockTime == time.Duration(0) {
		options.BlockTime = defaultBlockTime
	}

	s := &Server{
		ServerOptions: options,
		// Validators will have private keys, so they can sign blocks.
		// Note: only one validator can be elected, one source of truth, one miner in the whole network.
		isValidator: options.PrivateKey != nil,
		memPool:     NewTxPool(),
		rpcChannel:  make(chan ReceiveRPC),
		quitChannel: make(chan bool),
	}

	// Set the default RPC decoder.
	if s.ServerOptions.RPCDecodeFunc == nil {
		s.ServerOptions.RPCDecodeFunc = DefaultRPCDecoder
	}

	// Set the default RPC processor.
	if s.ServerOptions.RPCProcessor == nil {
		s.ServerOptions.RPCProcessor = s
	}

	return s
}

// Start will start the server.
// This is the core of the server, where the server will listen for messages from the transports.
// We make calls to helper functions for processing from here.
func (s *Server) Start() {
	s.initTransports()
	ticker := time.NewTicker(s.ServerOptions.BlockTime)

	// Run the server in a loop forever.
	for {
		select {
		case message := <-s.rpcChannel:
			// Decode the message.
			decodedMessage, err := s.ServerOptions.RPCDecodeFunc(message)
			if err != nil {
				logrus.WithError(err).Error("Failed to decode message")
				continue
			}
			// Process the message.
			err = s.ServerOptions.RPCProcessor.ProcessMessage(decodedMessage)
			if err != nil {
				logrus.WithError(err).Error("Failed to process message")
			}
		case <-s.quitChannel:
			// Quit the server.
			fmt.Println("Quitting the server.")
			return
		case <-ticker.C:
			// Here we will include CONSENSUS LOGIC / LEADER ELECTION LOGIC / BLOCK CREATION LOGIC.
			if s.isValidator {
				// If the server is validator, create a new block.
				s.createNewBlock()
			}
		}
	}
}

func (s *Server) ProcessMessage(decodedMessage *DecodedMessage) error {
	logrus.WithFields(
		logrus.Fields{
			"type": decodedMessage.Header,
			"from": decodedMessage.From,
		},
	).Info("Processing message")

	switch decodedMessage.Header {
	case Transaction:
		tx, ok := decodedMessage.Message.(core.Transaction)
		if !ok {
			return fmt.Errorf("failed to cast message to transaction")
		}
		return s.processTransaction(&tx)
	default:
		return fmt.Errorf("unknown message type: %d", decodedMessage.Header)
	}
}

func (s *Server) broadcast(payload []byte) error {
	for _, transport := range s.ServerOptions.Transports {
		err := transport.Broadcast(payload)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) broadcastTx(tx *core.Transaction) error {
	// encode the transaction, and put it in a message, and broadcast it to the network.
	transactionBytes := bytes.Buffer{}
	enc := core.NewTransactionEncoder()
	tx.Encode(&transactionBytes, enc)
	msg := NewMessage(Transaction, transactionBytes.Bytes())
	return s.broadcast(msg.Bytes())
}

// handling transactions coming into the blockchain
// two ways:
// 1. through a wallet (client), HTTP API will receive the transaction, and send it to the server, server will add it to the mempool.
// 2. through another node, the node will send the transaction to the server, server will add it to the mempool.
func (s *Server) processTransaction(tx *core.Transaction) error {
	transaction_hash := tx.GetHash(s.memPool.TransactionHasher)

	logrus.WithFields(
		logrus.Fields{
			"hash": transaction_hash,
		},
	).Info("Received new transaction")

	// check if transaction exists
	if s.memPool.Has(tx.GetHash(s.memPool.TransactionHasher)) {
		logrus.WithFields(
			logrus.Fields{
				"hash": transaction_hash,
			},
		).Warn("Transaction already exists in the mempool")
		return nil
	}

	// verify the transaction signature
	verified, err := tx.VerifySignature()
	if err != nil {
		return err
	}
	if !verified {
		return fmt.Errorf("transaction signature is invalid")
	}

	// set first seen time if not set yet
	// that means we are the first to see this transaction
	// i.e. it has come form an external source, and not
	// been broadcasted to us from another node.
	if tx.GetFirstSeen() == 0 {
		tx.SetFirstSeen(time.Now().UnixNano())
	}

	// broadcast the transaction to the network
	go s.broadcastTx(tx)

	logrus.WithFields(
		logrus.Fields{
			"hash":        transaction_hash,
			"memPoolSize": s.memPool.Len(),
		},
	).Info("Adding transaction to mempool")

	// add transaction to mempool
	return s.memPool.Add(tx)
}

// createNewBlock will create a new block.
func (s *Server) createNewBlock() error {
	// Create a new block and broadcast it to the network.
	fmt.Println("Creating a new block.")
	return nil
}

// Stop will stop the server.
func (s *Server) Stop() {
	s.quitChannel <- true
}

// initTransports will initialize the transports.
func (s *Server) initTransports() {
	// For each transport, spin up a new goroutine (user level thread).
	for _, transport := range s.ServerOptions.Transports {
		go s.startTransport(transport)
	}
}

// startTransport will start the transport.
func (s *Server) startTransport(transport Transport) {
	// Keep reading the channel and process the messages forever.
	for message := range transport.Consume() {
		// Send ReceiveRPC message to the server's rpcChannel. For synchronous processing.
		s.rpcChannel <- message
	}
}
