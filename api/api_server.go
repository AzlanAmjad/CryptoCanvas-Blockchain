package api

import (
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	core "github.com/AzlanAmjad/DreamscapeCanvas-Blockchain/blockchain-core"
	crypto "github.com/AzlanAmjad/DreamscapeCanvas-Blockchain/cryptography"
	types "github.com/AzlanAmjad/DreamscapeCanvas-Blockchain/data-types"
	"github.com/go-kit/log"
	"github.com/labstack/echo/v4"
)

// This is the API Server for the blockchain
// We use the Echo web framework to build our API here

// server configuration struct
type ServerConfig struct {
	// ListenAddr is the address the server listens on
	ListenAddr net.Addr
	Logger     log.Logger
}

// Server is the API Server for the blockchain
type Server struct {
	config *ServerConfig
	bc     *core.Blockchain
}

// transaction to send as JSON
type Transaction struct {
	Data      string
	From      string
	Signature *crypto.Signature
	FirstSeen time.Time
}

// transactions response
type Transactions struct {
	TxLength uint
	TxHashes []string
}

// block type to send as JSON
type Block struct {
	Version       uint32
	PrevBlockHash string
	DataHash      string
	Timestamp     time.Time
	Index         uint32
	Validator     string
	Signature     *crypto.Signature
	Transactions  Transactions
}

// NewServer creates a new Server
func NewServer(config *ServerConfig, bc *core.Blockchain) *Server {
	return &Server{
		config: config,
		bc:     bc,
	}
}

// Start starts the server
func (s *Server) Start() error {

	// create a new Echo instance
	e := echo.New()

	// routes for the API
	e.GET("/block/index/:index", s.getBlockByIndex)
	e.GET("/block/hash/:hash", s.getBlockByHash)
	e.GET("/transaction/hash/:hash", s.getTransactionByHash)

	// start the server
	err := e.Start(s.config.ListenAddr.String())
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) getTransactionByHash(c echo.Context) error {
	// get the hash from the URL
	hash := c.Param("hash")

	// convert the hash to a types.Hash
	hash_bytes, err := hex.DecodeString(hash)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{"error": "invalid hash"})
	}
	hash_type := types.Hash(hash_bytes)

	// get the transaction from the mempool
	tx, err := s.bc.GetTransactionByHash(hash_type)
	if err != nil {
		return c.JSON(http.StatusOK, map[string]any{"transaction": tx, "error": err.Error()})
	}

	txJSON := s.CreateJSONTransaction(tx)

	return c.JSON(http.StatusOK, txJSON)
}

// getBlockByIndex is the handler for the /block/index/:index route
func (s *Server) getBlockByIndex(c echo.Context) error {
	// get the index from the URL
	index := c.Param("index")

	// convert the index to an integer
	index_int, err := strconv.Atoi(index)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{"error": "invalid index"})
	}
	block, err := s.bc.GetBlockByIndex(uint32(index_int))
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]any{"message": "block not found", "error": err.Error()})
	}

	blockJSON := s.CreateJSONBlock(block)

	return c.JSON(http.StatusOK, blockJSON)
}

// getBlockByHash is the handler for the /block/hash/:hash route
func (s *Server) getBlockByHash(c echo.Context) error {
	// get the hash from the URL
	hash := c.Param("hash")
	fmt.Print(hash)

	// convert the hash to a types.Hash
	hash_bytes, err := hex.DecodeString(hash)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{"error": "invalid hash"})
	}
	hash_type := types.Hash(hash_bytes)

	block, err := s.bc.GetBlockByHash(hash_type)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]any{"message": "block not found", "error": err.Error()})
	}

	blockJSON := s.CreateJSONBlock(block)

	return c.JSON(http.StatusOK, blockJSON)
}

func (s *Server) CreateJSONBlock(block *core.Block) *Block {
	// create a new Block struct to send as JSON
	blockJSON := &Block{
		Version:       block.Header.Version,
		PrevBlockHash: block.Header.PrevBlockHash.String(),
		DataHash:      block.Header.DataHash.String(),
		Timestamp:     time.Unix(0, block.Header.Timestamp),
		Index:         block.Header.Index,
	}

	if block.Validator.Key != nil {
		blockJSON.Validator = block.Validator.String()
		blockJSON.Signature = block.Signature
	} else {
		blockJSON.Validator = ""
		blockJSON.Signature = nil
	}

	// add the transactions to the blockJSON
	blockJSON.Transactions.TxLength = uint(len(block.Transactions))
	blockJSON.Transactions.TxHashes = make([]string, len(block.Transactions))
	for i, tx := range block.Transactions {
		blockJSON.Transactions.TxHashes[i] = tx.GetHash(block.TransactionHasher).String()
	}

	return blockJSON
}

func (s *Server) CreateJSONTransaction(tx *core.Transaction) *Transaction {
	// create a new Transaction struct to send as JSON
	txJSON := &Transaction{
		Data:      hex.EncodeToString(tx.Data),
		From:      tx.From.String(),
		Signature: tx.Signature,
		FirstSeen: time.Unix(0, tx.FirstSeen),
	}

	return txJSON
}
