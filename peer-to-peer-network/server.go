package network

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net"
	"os"
	"reflect"
	"sync"
	"time"

	"github.com/AzlanAmjad/DreamscapeCanvas-Blockchain/api"
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

const BLOCKS_TOO_HIGH_THRESHOLD = 5

// One node can have multiple transports, for example, a TCP transport and a UDP transport.
type ServerOptions struct {
	SeedNodes     []net.Addr
	Addr          net.Addr
	APIAddr       net.Addr
	ID            string
	Logger        log.Logger
	RPCDecodeFunc RPCDecodeFunc
	RPCProcessor  RPCProcessor
	PrivateKey    *crypto.PrivateKey
	// if a node / server is elect as a validator, server needs to know when its time to consume its mempool and create a new block.
	BlockTime      time.Duration
	MaxMemPoolSize int
}

type Server struct {
	TCPTransport *TCPTransport

	mu    sync.RWMutex
	Peers map[net.Addr]*TCPPeer

	ServerOptions ServerOptions
	isValidator   bool
	// holds transactions that are not yet included in a block.
	memPool     *core.TxPool
	rpcChannel  chan ReceiveRPC
	peerChannel chan *TCPPeer
	quitChannel chan bool
	Chain       *core.Blockchain
	// count of blocks that are sent to us that are too high for our chain
	// used for syncing the blockchain with the peer in processBlock
	blocks_too_high_count int
}

// NewServer creates a new server with the given options.
func NewServer(options ServerOptions) (*Server, error) {
	// setting default values if none are specified
	if options.BlockTime == time.Duration(0) {
		options.BlockTime = defaultBlockTime
	}
	if options.Logger == nil {
		options.Logger = log.NewLogfmtLogger(os.Stderr)
		options.Logger = log.With(options.Logger, "ID", options.ID)
	}
	if options.MaxMemPoolSize == 0 {
		options.MaxMemPoolSize = 100
	}

	// create server mempool
	memPool := core.NewTxPool(options.MaxMemPoolSize)

	// create the default LevelDB storage
	dbPath := fmt.Sprintf("./leveldb/%s/blockchain", options.ID)
	storage, err := core.NewLevelDBStorage(dbPath)
	if err != nil {
		return nil, err
	}
	// create new blockchain with the LevelDB storage
	bc, err := core.NewBlockchain(storage, genesisBlock(), options.ID, memPool)
	if err != nil {
		return nil, err
	}

	// create a channel to receive peers from the transport
	peerCh := make(chan *TCPPeer)
	// create the TCP transport
	tcpTransport, err := NewTCPTransport(options.Addr, peerCh, options.ID)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to create TCP transport")
	}

	s := &Server{
		TCPTransport:  tcpTransport,
		Peers:         make(map[net.Addr]*TCPPeer),
		ServerOptions: options,
		// Validators will have private keys, so they can sign blocks.
		// Note: only one validator can be elected, one source of truth, one miner in the whole network.
		isValidator:           options.PrivateKey != nil,
		memPool:               memPool,
		rpcChannel:            make(chan ReceiveRPC),
		peerChannel:           peerCh,
		quitChannel:           make(chan bool),
		Chain:                 bc,
		blocks_too_high_count: 0,
	}

	// Set the default RPC decoder.
	if s.ServerOptions.RPCDecodeFunc == nil {
		s.ServerOptions.RPCDecodeFunc = DefaultRPCDecoder
	}

	// Set the default RPC processor.
	if s.ServerOptions.RPCProcessor == nil {
		s.ServerOptions.RPCProcessor = s
	}

	// boot up API server
	if s.ServerOptions.APIAddr != nil {
		go s.startAPIServer()
		// log API server start
		s.ServerOptions.Logger.Log("msg", "API server started", "address", s.ServerOptions.APIAddr.String())
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

// startAPIServer will start the API server.
func (s *Server) startAPIServer() {
	// API server config
	apiServerConfig := &api.ServerConfig{
		ListenAddr: s.ServerOptions.APIAddr,
		Logger:     s.ServerOptions.Logger,
	}

	// create a new API server
	apiServer := api.NewServer(apiServerConfig, s.Chain)

	// start the API server
	err := apiServer.Start()
	if err != nil {
		logrus.WithError(err).Fatal("Failed to start API server")
	}
}

// dial the seed nodes, and add them to the list of peers
func (s *Server) bootstrapNetwork() {
	// dial the seed nodes
	for _, seed := range s.ServerOptions.SeedNodes {
		// make sure the peer is not ourselves
		if s.TCPTransport.Addr.String() != seed.String() {
			// dial the seed node
			conn, err := net.Dial("tcp", seed.String())
			if err != nil {
				logrus.WithError(err).Error("Failed to dial seed node")
				continue
			}

			fmt.Println("Dialed seed node", seed.String())

			// create a new peer
			peer := &TCPPeer{conn: conn, Incoming: false} // we initiated this connection so Incoming is false
			// send peer to server peer channel
			s.peerChannel <- peer
		} else {
			// log that we are not dialing ourselves
			s.ServerOptions.Logger.Log("msg", "We are a seed node, not dialing ourselves", "seed", seed)
		}
	}
}

// Start will start the server.
// This is the core of the server, where the server will listen for messages from the transports.
// We make calls to helper functions for processing from here.
func (s *Server) Start() error {
	s.ServerOptions.Logger.Log("msg", "Starting main server loop", "server_id", s.ServerOptions.ID)
	// start the transport
	s.TCPTransport.Start()
	// bootstrap the network
	go s.bootstrapNetwork()

	// Run the server in a loop forever.
	for {
		select {
		// receive a new peer, we either connected via
		// 1. Seed node dialing
		// 2. Another peer dialed us
		case peer := <-s.peerChannel:
			// TODO (Azlan): Add mutex to protect the Peers map.
			// Add the peer to the list of peers.
			s.Peers[peer.conn.RemoteAddr()] = peer
			// print that peer was added
			s.ServerOptions.Logger.Log("msg", "Peer added", "us", peer.conn.LocalAddr(), "peer", peer.conn.RemoteAddr())
			// start reading from the peer
			go peer.readLoop(s.rpcChannel)

			// synchronize with the new peer, we do not broadcast here due to problems
			// with the underlying algorithm, it will result in us having a corrupted blockchain
			// if we begin to synchronize with multiple peers at the same time that have different state
			err := s.sendGetStatus(peer)
			if err != nil {
				logrus.WithError(err).Error("Failed to send getStatus to peer")
			}
		// receive RPC message to process
		case rpc := <-s.rpcChannel:
			// Decode the message.
			decodedMessage, err := s.ServerOptions.RPCDecodeFunc(rpc)
			if err != nil {
				logrus.WithError(err).Error("Failed to decode message")
				continue
			}
			// Process the message.
			err = s.ServerOptions.RPCProcessor.ProcessMessage(rpc.From, decodedMessage)
			if err != nil {
				logrus.WithError(err).Error("Failed to process message")
			}
		case <-s.quitChannel:
			// log the quit signal
			s.ServerOptions.Logger.Log("msg", "Received quit signal", "server_id", s.ServerOptions.ID)
			// Quit the server.
			fmt.Println("Quitting the server.")
			s.handleQuit()
			return nil
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

// function to fulfill the RPCProcessor interface
// this function will process the message received from the transport
func (s *Server) ProcessMessage(from net.Addr, decodedMessage *DecodedMessage) error {
	s.ServerOptions.Logger.Log("msg", "Processing message", "type", decodedMessage.Header, "from", decodedMessage.From)

	switch decodedMessage.Header {
	case Transaction:
		tx, ok := decodedMessage.Message.(core.Transaction)
		if !ok {
			return fmt.Errorf("failed to cast message to transaction")
		}
		return s.processTransaction(from, &tx)
	case Block: // single block
		block, ok := decodedMessage.Message.(core.Block)
		if !ok {
			return fmt.Errorf("failed to cast message to block")
		}
		return s.processBlock(from, &block)
	case Blocks: // multiple blocks
		blocks, ok := decodedMessage.Message.(core.BlocksMessage)
		if !ok {
			return fmt.Errorf("failed to cast message to blocks")
		}
		return s.processBlocks(from, &blocks)
	case GetStatus:
		getStatus, ok := decodedMessage.Message.(core.GetStatusMessage)
		if !ok {
			fmt.Printf("error: %s", getStatus)
			return fmt.Errorf("failed to cast message to get status")
		}
		return s.processGetStatus(from, &getStatus)
	case Status:
		status, ok := decodedMessage.Message.(core.StatusMessage)
		if !ok {
			return fmt.Errorf("failed to cast message to status")
		}
		return s.processStatus(from, &status)
	case GetBlocks:
		blocks, ok := decodedMessage.Message.(core.GetBlocksMessage)
		if !ok {
			return fmt.Errorf("failed to cast message to get blocks")
		}
		return s.processGetBlocks(from, &blocks)
	default:
		return fmt.Errorf("unknown message type: %d", decodedMessage.Header)
	}
}

// generic function to broadcast to all peers
func (s *Server) broadcast(from net.Addr, payload []byte) error {
	s.mu.RLock() // make sure the peers map is not being modified
	defer s.mu.RUnlock()
	s.ServerOptions.Logger.Log("msg", "broadcasting to peers", "number of peers", len(s.Peers))
	for _, peer := range s.Peers {
		// don't send the message back to the peer that sent it to us
		// nil condition is checked because this data might not be from anyone
		// but from the node itself, for example, when a new block is created.
		if from == nil || peer.conn.RemoteAddr().String() != from.String() {
			err := peer.Send(payload)
			if err != nil {
				logrus.WithError(err).Error("Failed to send message to peer")
			}
		}
	}
	return nil
}

// broadcast the transaction to all known peers, eventually consistency model
func (s *Server) broadcastTx(from net.Addr, tx *core.Transaction) {
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
	err = s.broadcast(from, msg.Bytes())
	if err != nil {
		logrus.WithError(err).Error("Failed to broadcast transaction")
	}
}

// broadcast the block to all known peers, eventually consistency model
func (s *Server) broadcastBlock(from net.Addr, block *core.Block) {
	s.ServerOptions.Logger.Log("msg", "Broadcasting block", "hash", block.GetHash(s.Chain.BlockHeaderHasher))

	// encode the block, and put it in a message, and broadcast it to the network.
	blockBytes := bytes.Buffer{}
	err := block.Encode(&blockBytes, s.Chain.BlockEncoder)
	if err != nil {
		logrus.WithError(err).Error("Failed to encode block")
		return
	}
	msg := NewMessage(Block, blockBytes.Bytes())
	err = s.broadcast(from, msg.Bytes())
	if err != nil {
		logrus.WithError(err).Error("Failed to broadcast block")
	}
}

// initiate the process of possibly syncing with a peer
// by asking for its status (Blockchain Height)
func (s *Server) sendGetStatus(peer *TCPPeer) error {
	// encode the get status message, and put it in a message, and broadcast it to the network.
	getStatusBytes := bytes.Buffer{}
	getStatusMessage := new(core.GetStatusMessage)
	if err := gob.NewEncoder(&getStatusBytes).Encode(getStatusMessage); err != nil {
		return err
	}
	msg := NewMessage(GetStatus, getStatusBytes.Bytes())

	s.ServerOptions.Logger.Log("msg", "sending get status")

	// send the message to the peer
	return peer.Send(msg.Bytes())
}

// handling blocks coming into the blockchain
// two ways:
// 1. through another block that is forwarding the block to known peers
// 2. the leader node chosen through consensus has created a new block, and is broadcasting it to us
func (s *Server) processBlock(from net.Addr, block *core.Block) error {
	blockHash := block.GetHash(s.Chain.BlockHeaderHasher)
	s.ServerOptions.Logger.Log("msg", "Received new block", "hash", blockHash)

	// add the block to the blockchain, this includes validation
	// of the block before addition
	error_code, err := s.Chain.AddBlock(block)

	// if the block height is too high for our blockchain
	// indicated by error code 5, we need to sync with the peer
	// we also check if we have had BLOCKS_TOO_HIGH_THRESHOLD invalid block additions that have had
	// the same error code, this is to prevent spamming the peer with getStatus messages

	if error_code == 5 && s.blocks_too_high_count == BLOCKS_TOO_HIGH_THRESHOLD {
		// send a get status message to start the syncing process, with the peer
		err_sendGetStatus := s.sendGetStatus(s.Peers[from])
		if err_sendGetStatus != nil {
			logrus.WithError(err).Error("Failed to send getStatus to peer, failed to start syncing process")
		}
		s.blocks_too_high_count = 0
	}

	if err != nil {
		if error_code == 5 {
			s.blocks_too_high_count++
		}
		// we return here so we don't broadcast to our peers
		// this is because if we have already added the block
		// which is possibly why we are in this error condition, then we
		// have already broadcasted the block to our peers previously.
		return err
	}

	// broadcast the block to the network (all peers)
	go s.broadcastBlock(from, block)

	return nil
}

// handling transactions coming into the blockchain
// two ways:
// 1. through a wallet (client), HTTP API will receive the transaction, and send it to the server, server will add it to the mempool.
// 2. through another node, the node will send the transaction to the server, server will add it to the mempool.
func (s *Server) processTransaction(from net.Addr, tx *core.Transaction) error {
	transaction_hash := tx.GetHash(s.memPool.Pending.TransactionHasher)
	s.ServerOptions.Logger.Log("msg", "Received new transaction", "hash", transaction_hash)

	// check if transaction exists
	if s.memPool.PendingHas(tx.GetHash(s.memPool.Pending.TransactionHasher)) {
		logrus.WithField("hash", transaction_hash).Warn("Transaction already exists in the mempool")
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
	go s.broadcastTx(from, tx)

	s.ServerOptions.Logger.Log("msg", "Adding transaction to mempool", "hash", transaction_hash, "memPoolSize", s.memPool.PendingLen())

	// add transaction to mempool
	return s.memPool.Add(tx)
}

// process a get status message
// this will happen when a new node joins or either some node wants to
// know the status of the network
// some node has asked for our status
func (s *Server) processGetStatus(from net.Addr, _ *core.GetStatusMessage) error {
	s.ServerOptions.Logger.Log(
		"msg", "received get status message",
		"from", from.String(),
	)

	// prepare a status message to send back to whoever requested our status
	statusMessage := &core.StatusMessage{
		BlockchainHeight: s.Chain.GetHeight(),
		ID:               s.ServerOptions.ID,
	}
	// encode the message using GOB
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(statusMessage); err != nil {
		return err
	}
	// wrap status message in a message
	msg := NewMessage(Status, buf.Bytes())

	s.ServerOptions.Logger.Log(
		"msg", "sending status message",
		"blockchain_height", statusMessage.BlockchainHeight,
		"peer_address", from.String(),
	)

	// Get TCP Peer who sent the message, send status to that peer
	s.mu.RLock() // we are about to read from peers map
	defer s.mu.RUnlock()
	// check if peer exists
	if _, ok := s.Peers[from]; !ok {
		return fmt.Errorf("peer who sent get status not found")
	}
	s.Peers[from].Send(msg.Bytes())

	return nil
}

// we will receive this message when we have requested the status of some or all peer nodes
// this will function handles the status message received from the peer that we requested the status from
func (s *Server) processStatus(from net.Addr, peerStatus *core.StatusMessage) error {
	s.ServerOptions.Logger.Log("msg", "received status message", peerStatus)
	// we check if the blockchain height of the peer is greater than ours
	// if the peers blockchain height is less than or equal to ours, we are unable to sync
	// because we as a node need to sync with someone who might be ahead of us in the blockchain
	if peerStatus.BlockchainHeight <= s.Chain.GetHeight() {
		s.ServerOptions.Logger.Log("msg", "cannot sync", "our height", s.Chain.GetHeight(), "peers height", peerStatus.BlockchainHeight, "peer address", from.String())
		return nil
	} else {
		s.ServerOptions.Logger.Log("msg", "can sync", "our height", s.Chain.GetHeight(), "peers height", peerStatus.BlockchainHeight, "peer address", from.String())
	}

	// if we can sync we need to request the blocks from this peer node
	// create a get blocks message, encode it, and wrap it with a message
	getBlocksMsg := &core.GetBlocksMessage{
		FromIndex: s.Chain.GetHeight() + 1, // from is inclusive, we already have the block at this height so we do + 1
		ToIndex:   0,                       // 0 specifies that we want all blocks from the peer starting from "FromIndex"
	}

	s.ServerOptions.Logger.Log(
		"msg", "requesting blocks from peer",
		"peer address", from.String(),
		"from index", getBlocksMsg.FromIndex,
		"to index", getBlocksMsg.ToIndex,
	)

	buffer := new(bytes.Buffer)
	if err := gob.NewEncoder(buffer).Encode(getBlocksMsg); err != nil {
		return err
	}
	msg := NewMessage(GetBlocks, buffer.Bytes())

	// send to the the peer that sent us the status message
	s.mu.RLock() // we are about to read from peers map
	defer s.mu.RUnlock()
	// check if the peer exists
	if _, ok := s.Peers[from]; !ok {
		return fmt.Errorf("peer who sent status not found")
	}
	return s.Peers[from].Send(msg.Bytes())
}

// function to handle get blocks message, a node has requested blocks from us
// this function gets the requested blocks and returns them to the requesting node
// this requires a query into our blockchain LevelDB storage
func (s *Server) processGetBlocks(from net.Addr, getBlocks *core.GetBlocksMessage) error {
	s.ServerOptions.Logger.Log("msg", "received get blocks message")
	s.ServerOptions.Logger.Log("msg", "peer requested blocks", "peer address", from.String())

	// if getBlocks is 0 set it to the height of the blockchain + 1 since end is not inclusive
	if getBlocks.ToIndex == 0 {
		getBlocks.ToIndex = s.Chain.GetHeight() + 1
	}
	// slice of blocks to retrieve
	blocks, err := s.Chain.GetBlocks(getBlocks.FromIndex, getBlocks.ToIndex)
	if err != nil {
		return err
	}

	// create a list of block messages to send, note each can not be greater
	// than TCP_BUFFER_SIZE
	blocksMessages := make([]core.BlocksMessage, 0)

	blocksMessage := &core.BlocksMessage{} // initial blocks message to start loop
	// size of the blocks message to start loop, blocks message + the new message that will be constructed
	blocksMessageSize := int(reflect.TypeOf(*blocksMessage).Size()) + int(reflect.TypeOf(Message{}).Size())

	// loop through the blocks, encode them, and append them to the blocks message
	// we start a new blocksMessage if we reach the TCP_BUFFER_SIZE
	for _, block := range blocks {
		blockBytes := bytes.Buffer{}
		err := block.Encode(&blockBytes, s.Chain.BlockEncoder)
		if err != nil {
			return err
		}

		// check if the block can fit in the blocks message
		// if it can not fit, appends blocksMessage and reset
		if blocksMessageSize+len(blockBytes.Bytes()) >= TCP_BUFFER_SIZE {
			blocksMessages = append(blocksMessages, *blocksMessage)
			blocksMessage = &core.BlocksMessage{}
			blocksMessageSize = int(reflect.TypeOf(*blocksMessage).Size()) + int(reflect.TypeOf(Message{}).Size())
		}

		// append the block to the blocks message
		blocksMessage.Blocks = append(blocksMessage.Blocks, blockBytes.Bytes())
		blocksMessageSize += len(blockBytes.Bytes())
	} // for after last loop iteration
	blocksMessages = append(blocksMessages, *blocksMessage)

	// log length of blocks messages
	s.ServerOptions.Logger.Log("msg", "BLOCKS MESSAGES LENGTH", "length", len(blocksMessages))

	// for each new blocksMessage we created
	for _, blocksMessage := range blocksMessages {
		// encode the blocks message
		blocksMessageBytes := new(bytes.Buffer)
		if err := gob.NewEncoder(blocksMessageBytes).Encode(blocksMessage); err != nil {
			return err
		}
		// create a new message
		msg := NewMessage(Blocks, blocksMessageBytes.Bytes())
		s.ServerOptions.Logger.Log(
			"msg", "sending blocks message to peer",
			"peer address", from.String(),
			// message length can not be more than TCP buffer size TCP_BUFFER_SIZE
			"message length", len(msg.Bytes()), // used to check byte length of encoded message
		)

		// wait a bit before sending the message
		// this is to prevent the peer from being overwhelmed and its TCP buffer filling up too fast
		time.Sleep(100 * time.Millisecond)

		// send the message to the peer
		s.mu.RLock() // we are about to read from peers map
		defer s.mu.RUnlock()
		// check if the peer exists
		if _, ok := s.Peers[from]; !ok {
			return fmt.Errorf("peer who sent get blocks not found")
		}
		err := s.Peers[from].Send(msg.Bytes())
		if err != nil {
			return err
		}
	}

	return nil
}

// process blocks, function to process multiple blocks
func (s *Server) processBlocks(from net.Addr, blocksMsg *core.BlocksMessage) error {
	s.ServerOptions.Logger.Log(
		"msg", "received blocks message",
		"length", len(blocksMsg.Blocks), // encoded blocks
	)
	s.ServerOptions.Logger.Log("msg", "peer sent blocks", "peer address", from.String())

	// loop the blocks and decode them, and add them to the blockchain
	for _, blockBytes := range blocksMsg.Blocks {
		block := core.NewBlock()
		err := block.Decode(bytes.NewReader(blockBytes), s.Chain.BlockDecoder)
		if err != nil {
			return err
		}
		blockHash := block.GetHash(s.Chain.BlockHeaderHasher)
		s.ServerOptions.Logger.Log("msg", "Received new block", "hash", blockHash)
		// add the block to the blockchain, this includes validation
		// of the block before addition
		_, err = s.Chain.AddBlock(block)
		if err != nil {
			return err // if addition of a block fails we will return, no need to continue
		}
	}

	return nil
}

// Stop will stop the server.
func (s *Server) Stop() {
	s.quitChannel <- true
}

// handleQuit will handle the quit signal sent by Stop()
func (s *Server) handleQuit() {
	// shutdown the storage
	s.Chain.Storage.Shutdown()
}

// createNewBlock will create a new block.
func (s *Server) createNewBlock() error {
	// get the previous block hash
	prevBlockHash, err := s.Chain.GetBlockHash(s.Chain.GetHeight())
	if err != nil {
		return err
	}

	// create a new block, with all mempool transactions
	// normally we might just include a specific number of transactions in each block
	// but here we are including all transactions in the mempool.
	// once we know how our transactions will be structured
	// we can include a specific number of transactions in each block.
	transactions := s.memPool.GetPendingTransactions()
	// flush mempool since we got all the transactions
	// IMPORTANT: we flush right away because while we execute
	// the block creation logic, new transactions might come in, since
	// createBlock() runs in the validatorLoop() goroutine, it is running
	// parallel to the main server loop Start().
	s.memPool.FlushPending()
	// create the block
	block := core.NewBlockWithTransactions(transactions)
	// link block to previous block
	block.Header.PrevBlockHash = prevBlockHash
	// update the block index
	block.Header.Index = s.Chain.GetHeight() + 1
	// TODO (Azlan): update other block parameters so that this block is not marked as invalid.

	// sign the block as a validator of this block
	err = block.Sign(s.ServerOptions.PrivateKey)
	if err != nil {
		return err
	}

	// add the block to the blockchain
	_, err = s.Chain.AddBlock(block)
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
	go s.broadcastBlock(nil, block)

	return nil
}

// genesisBlock creates the first block in the blockchain.
func genesisBlock() *core.Block {
	b := core.NewBlock()
	b.Header.Version = 1
	b.Header.PrevBlockHash = [32]byte{} // initialize with an empty byte array
	b.Header.DataHash = [32]byte{}      // initialize with an empty byte array
	b.Header.Timestamp = 0000000000     // setting to zero for all genesis blocks created across all nodes
	b.Header.Index = 0
	return b
}
