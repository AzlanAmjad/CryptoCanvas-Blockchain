package network

import (
	"bytes"
	"fmt"
	"os"
	"time"

	core "github.com/AzlanAmjad/DreamscapeCanvas-Blockchain/blockchain-core"
	crypto "github.com/AzlanAmjad/DreamscapeCanvas-Blockchain/cryptography"
	"github.com/go-kit/log"
	"github.com/sirupsen/logrus"
)

// The server is a container which will contain every module of the server.
// 1. Nodes / servers can be BlockExplorers, Wallets, and other nodes, to strengthen the network.
// 1. Nodes / servers can also be validators, that participate in the consensus algorithm. They
// can be elected to create new blocks, and validate transactions, and propose blocks into the network.

var defaultBlockTime = 5 * time.Second

// One node can have multiple transports, for example, a TCP transport and a UDP transport.
type ServerOptions struct {
	ID            string
	Logger        log.Logger
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
	chain       *core.Blockchain
}

// NewServer creates a new server with the given options.
func NewServer(options ServerOptions) (*Server, error) {
	if options.BlockTime == time.Duration(0) {
		options.BlockTime = defaultBlockTime
	}

	if options.Logger == nil {
		options.Logger = log.NewLogfmtLogger(os.Stderr)
		options.Logger = log.With(options.Logger, "ID", options.ID)
	}

	// create the default LevelDB storage
	dbPath := fmt.Sprintf("./leveldb/%s/blockchain", options.ID)
	storage, err := core.NewLevelDBStorage(dbPath)
	if err != nil {
		return nil, err
	}
	// create new blockchain with the LevelDB storage
	bc, err := core.NewBlockchain(storage, genesisBlock(), options.ID)
	if err != nil {
		return nil, err
	}

	s := &Server{
		ServerOptions: options,
		// Validators will have private keys, so they can sign blocks.
		// Note: only one validator can be elected, one source of truth, one miner in the whole network.
		isValidator: options.PrivateKey != nil,
		memPool:     NewTxPool(),
		rpcChannel:  make(chan ReceiveRPC),
		quitChannel: make(chan bool),
		chain:       bc,
	}

	// Set the default RPC decoder.
	if s.ServerOptions.RPCDecodeFunc == nil {
		s.ServerOptions.RPCDecodeFunc = DefaultRPCDecoder
	}

	// Set the default RPC processor.
	if s.ServerOptions.RPCProcessor == nil {
		s.ServerOptions.RPCProcessor = s
	}

	// Goroutine (Thread) to process creation of blocks if the node is a validator.
	if s.isValidator {
		go s.validatorLoop()
	}

	// log server creation
	s.ServerOptions.Logger.Log(
		"msg", "server created",
		"server_id", s.ServerOptions.ID,
		"validator", s.isValidator,
	)

	return s, nil
}

// Start will start the server.
// This is the core of the server, where the server will listen for messages from the transports.
// We make calls to helper functions for processing from here.
func (s *Server) Start() {
	err := s.initTransports()
	if err != nil {
		logrus.WithError(err).Error("Failed to initialize transports")
		return
	}

	// Run the server in a loop forever.
	s.ServerOptions.Logger.Log("msg", "Starting main server loop", "server_id", s.ServerOptions.ID)

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
			// log the quit signal
			s.ServerOptions.Logger.Log("msg", "Received quit signal", "server_id", s.ServerOptions.ID)
			// Quit the server.
			fmt.Println("Quitting the server.")
			s.handleQuit()
			return
		}
	}
}

// validatorLoop will create a new block every block time.
// validatorLoop is only started if the server is a validator.
func (s *Server) validatorLoop() {
	s.ServerOptions.Logger.Log("msg", "Starting validator loop", "server_id", s.ServerOptions.ID)

	// Create a ticker to create a new block every block time.
	ticker := time.NewTicker(s.ServerOptions.BlockTime)
	for {
		select {
		case <-s.quitChannel:
			return
		case <-ticker.C:
			// Here we will include CONSENSUS LOGIC / LEADER ELECTION LOGIC / BLOCK CREATION LOGIC.
			// If the server is the elected validator, create a new block.
			err := s.createNewBlock()
			if err != nil {
				logrus.WithError(err).Error("Failed to create new block")
			}
		}
	}
}

func (s *Server) ProcessMessage(decodedMessage *DecodedMessage) error {
	//s.ServerOptions.Logger.Log("msg", "Processing message", "type", decodedMessage.Header, "from", decodedMessage.From)

	switch decodedMessage.Header {
	case Transaction:
		tx, ok := decodedMessage.Message.(core.Transaction)
		if !ok {
			return fmt.Errorf("failed to cast message to transaction")
		}
		return s.processTransaction(&tx)
	case Block:
		block, ok := decodedMessage.Message.(core.Block)
		if !ok {
			print(decodedMessage.Message)
			return fmt.Errorf("failed to cast message to block")
		}
		return s.processBlock(&block)
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

// broadcast the transaction to all known peers, eventually consistency model
func (s *Server) broadcastTx(tx *core.Transaction) {
	//s.ServerOptions.Logger.Log("msg", "Broadcasting transaction", "hash", tx.GetHash(s.memPool.TransactionHasher))

	// encode the transaction, and put it in a message, and broadcast it to the network.
	transactionBytes := bytes.Buffer{}
	enc := core.NewTransactionEncoder()
	err := tx.Encode(&transactionBytes, enc)
	if err != nil {
		logrus.WithError(err).Error("Failed to encode transaction")
		return
	}
	msg := NewMessage(Transaction, transactionBytes.Bytes())
	err = s.broadcast(msg.Bytes())
	if err != nil {
		logrus.WithError(err).Error("Failed to broadcast transaction")
	}
}

// broadcast the block to all known peers, eventually consistency model
func (s *Server) broadcastBlock(block *core.Block) {
	s.ServerOptions.Logger.Log("msg", "Broadcasting block", "hash", block.GetHash(s.chain.BlockHeaderHasher))
	fmt.Println("block_index", block.Header.Index)

	// encode the block, and put it in a message, and broadcast it to the network.
	blockBytes := bytes.Buffer{}
	enc := core.NewBlockEncoder()
	err := block.Encode(&blockBytes, enc)
	if err != nil {
		logrus.WithError(err).Error("Failed to encode block")
		return
	}
	msg := NewMessage(Block, blockBytes.Bytes())
	err = s.broadcast(msg.Bytes())
	if err != nil {
		logrus.WithError(err).Error("Failed to broadcast block")
	}
}

// handling blocks coming into the blockchain
// two ways:
// 1. through another block that is forwarding the block to known peers
// 2. the leader node chose through consensus has created a new block, and is broadcasting it to us
func (s *Server) processBlock(block *core.Block) error {
	blockHash := block.GetHash(s.chain.BlockHeaderHasher)
	s.ServerOptions.Logger.Log("msg", "Received new block", "hash", blockHash)

	// add the block to the blockchain, this includes validation
	// of the block before addition
	err := s.chain.AddBlock(block)
	if err != nil {
		return err
	}

	// broadcast the block to the network (all peers)
	go s.broadcastBlock(block)

	s.ServerOptions.Logger.Log("msg", "Added block to blockchain", "hash", blockHash, "blockchainHeight", s.chain.GetHeight())

	return nil
}

// handling transactions coming into the blockchain
// two ways:
// 1. through a wallet (client), HTTP API will receive the transaction, and send it to the server, server will add it to the mempool.
// 2. through another node, the node will send the transaction to the server, server will add it to the mempool.
func (s *Server) processTransaction(tx *core.Transaction) error {
	transaction_hash := tx.GetHash(s.memPool.TransactionHasher)

	s.ServerOptions.Logger.Log("msg", "Received new transaction", "hash", transaction_hash)

	// check if transaction exists
	if s.memPool.Has(tx.GetHash(s.memPool.TransactionHasher)) {
		s.ServerOptions.Logger.Log("msg", "Transaction already exists in the mempool", "hash", transaction_hash)
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
	// i.e. it has come from an external source, and not
	// been broadcasted to us from another node.
	// if another node running this same software has already seen this transaction
	// then it will have set the first seen time. giving us total ordering of transactions.
	// WE DO NOT USE LOGICAL TIMESTAMPS OR SYNCHRONIZATION MECHANISMS HERE BECAUSE TRANSACTIONS
	// ARE EXTERNAL TO THE NETWORK, AND CAN BE SENT FROM ANYWHERE.
	if tx.GetFirstSeen() == 0 {
		tx.SetFirstSeen(time.Now().UnixNano())
	}

	// broadcast the transaction to the network (all peers)
	go s.broadcastTx(tx)

	s.ServerOptions.Logger.Log("msg", "Adding transaction to mempool", "hash", transaction_hash, "memPoolSize", s.memPool.Len())

	// add transaction to mempool
	return s.memPool.Add(tx)
}

// Stop will stop the server.
func (s *Server) Stop() {
	s.quitChannel <- true
}

// handleQuit will handle the quit signal sent by Stop()
func (s *Server) handleQuit() {
	// shutdown the storage
	s.chain.Storage.Shutdown()
}

// initTransports will initialize the transports.
func (s *Server) initTransports() error {
	if s.ServerOptions.Transports == nil {
		return fmt.Errorf("no transports to initialize")
	}

	// log init of transports
	s.ServerOptions.Logger.Log(
		"msg", "Initializing transports",
		"server_id", s.ServerOptions.ID,
		"transports", len(s.ServerOptions.Transports),
	)

	// For each transport, spin up a new goroutine (user level thread).
	for _, transport := range s.ServerOptions.Transports {
		if transport == nil {
			return fmt.Errorf("transport is nil")
		}
		// print concrete type of transport variable
		go s.startTransport(transport)
	}

	return nil
}

// startTransport will start the transport.
func (s *Server) startTransport(transport Transport) {
	// Keep reading the channel and process the messages forever.
	for message := range transport.Consume() {
		// Send ReceiveRPC message to the server's rpcChannel. For synchronous processing.
		s.rpcChannel <- message
	}
}

// createNewBlock will create a new block.
func (s *Server) createNewBlock() error {
	// get the previous block hash
	prevBlockHash, err := s.chain.GetBlockHash(s.chain.GetHeight())
	if err != nil {
		return err
	}

	// create a new block, with all mempool transactions
	// normally we might just include a specific number of transactions in each block
	// but here we are including all transactions in the mempool.
	// once we know how our transactions will be structured
	// we can include a specific number of transactions in each block.
	transactions := s.memPool.GetTransactions()
	// flush mempool since we got all the transactions
	// IMPORTANT: we flush right away because while we execute
	// the block creation logic, new transactions might come in, since
	// createBlock() runs in the validatorLoop() goroutine, it is running
	// parallel to the main server loop Start().
	s.memPool.Flush()
	// create the block
	block := core.NewBlockWithTransactions(transactions)
	// link block to previous block
	block.Header.PrevBlockHash = prevBlockHash
	// update the block index
	block.Header.Index = s.chain.GetHeight() + 1
	// TODO (Azlan): update other block parameters so that this block is not marked as invalid.

	// sign the block as a validator of this block
	err = block.Sign(s.ServerOptions.PrivateKey)
	if err != nil {
		return err
	}

	// add the block to the blockchain
	err = s.chain.AddBlock(block)
	if err != nil {
		// something ent wrong, we need to re-add the transactions to the mempool
		// so that they can be included in the next block.
		for _, tx := range transactions {
			s.memPool.Add(tx)
		}

		return err
	}

	// as the leader node chosen by consensus for block addition
	// we were able to create a new block successfully,
	// broadcast it to all our peers, to ensure eventual consistency
	go s.broadcastBlock(block)

	return nil
}

// genesisBlock creates the first block in the blockchain.
func genesisBlock() *core.Block {
	b := core.NewBlock()
	b.Header.Timestamp = 0000000000 // setting to zero for all genesis blocks created across all nodes
	b.Header.Index = 0
	b.Header.Version = 1
	return b
}
