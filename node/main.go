package main

import (
	"fmt"
	"net/http"
	"net/rpc"
)

type Block struct {
	Index     int
	Timestamp string
	Data      int
	Hash      string
	PrevHash  string
}

type BlockChain struct {
	Chain []Block
}

type BlockChainService struct{}

var blockChain BlockChain

type AddBlockArgs struct {
	Data int
}

type AddBlockReply struct {
	Index int
	Hash  string
}

func (b *BlockChainService) AddBlock(args *AddBlockArgs, reply *AddBlockReply) error {
	fmt.Println("AddBlock called")

	// Get the previous block
	prevBlock := blockChain.Chain[len(blockChain.Chain)-1]

	// Create the new block
	newBlock := Block{Index: prevBlock.Index + 1, Timestamp: "2017-01-01 00:00:00", Data: args.Data, PrevHash: prevBlock.Hash}

	// Calculate the hash of the new block
	newBlock.Hash = "0"

	// Add the new block to the block chain
	blockChain.Chain = append(blockChain.Chain, newBlock)

	// Return the new block's index and hash
	reply.Index = newBlock.Index
	reply.Hash = newBlock.Hash

	return nil
}

type GetBlockchainArgs struct{}

type GetBlockchainReply struct {
	Chain []Block
}

func (b *BlockChainService) GetBlockchain(args *GetBlockchainArgs, reply *GetBlockchainReply) error {
	fmt.Println("GetBlockchain called")

	// Return the entire block chain
	reply.Chain = blockChain.Chain

	return nil
}

func (b *BlockChainService) Ping(args *struct{}, reply *string) error {
	*reply = "Pong"
	return nil
}

func main() {
	// Initialize the block chain with the genesis block
	blockChain = BlockChain{Chain: []Block{{Index: 0, Timestamp: "2017-01-01 00:00:00", Data: 0, Hash: "0", PrevHash: "0"}}}

	// user input to request port number
	var port string
	fmt.Println("Enter the port number to start the server")
	fmt.Scanln(&port)

	// New RPC server
	server := rpc.NewServer()

	// Register the BlockChainService with the RPC server
	server.Register(&BlockChainService{})

	// Handle RPC requests
	server.HandleHTTP(rpc.DefaultRPCPath, rpc.DefaultDebugPath)
	// Start the HTTP server
	http.ListenAndServe(":"+port, nil)
}
