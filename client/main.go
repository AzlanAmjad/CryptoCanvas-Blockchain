package main

import (
	"fmt"
	"os"

	"net/rpc"
)

type Block struct {
	Index     int
	Timestamp string
	Data      int
	Hash      string
	PrevHash  string
}

type AddBlockArgs struct {
	Data int
}

type AddBlockReply struct {
	Index int
	Hash  string
}

type GetBlockchainArgs struct{}

type GetBlockchainReply struct {
	Chain []Block
}

func main() {
	// get port number for server to connect to from user
	var port int
	fmt.Print("Enter the port number for the server to connect to: ")
	fmt.Scanln(&port)

	// create server address
	serverAddress := fmt.Sprintf("localhost:%d", port)
	client, err := rpc.Dial("tcp", serverAddress)
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		os.Exit(1)
	}
	defer client.Close()

	// create args and reply structs
	var add_block_args = AddBlockArgs{Data: 42}
	var add_block_reply AddBlockReply

	var get_blockchain_args = GetBlockchainArgs{}
	var get_blockchain_reply GetBlockchainReply

	// Call the Ping method
	var reply string
	err = client.Call("BlockChainService.Ping", struct{}{}, &reply)
	if err != nil {
		fmt.Println("Error calling Ping:", err)
	} else {
		fmt.Println("Ping result:", reply)
	}

	// continuous user interaction loop, ask user to select blockchain operation
	for {
		// display menu
		fmt.Println("1. Add a block")
		fmt.Println("2. View the blockchain")
		fmt.Println("3. Exit")
		fmt.Print("Enter a selection: ")

		// get user input
		var selection int
		fmt.Scanln(&selection)

		// handle user input
		switch selection {
		case 1:
			// add a block
			err = client.Call("BlockChainService.AddBlock", add_block_args, &add_block_reply)
			if err != nil {
				if rpcErr, ok := err.(*rpc.ServerError); ok {
					fmt.Println("RPC error:", rpcErr.Error())
				} else {
					// print raw response from server
					fmt.Println("Raw response:", add_block_reply)
					fmt.Println("Error adding block:", err)
				}
			} else {
				fmt.Println("New block added to blockchain with index", add_block_reply.Index, "and hash", add_block_reply.Hash)
			}
		case 2:
			// view the blockchain
			err = client.Call("BlockChainService.GetBlockchain", get_blockchain_args, &get_blockchain_reply)
			if err != nil {
				if rpcErr, ok := err.(*rpc.ServerError); ok {
					fmt.Println("RPC error:", rpcErr.Error())
				} else {
					// print raw response from server
					fmt.Println("Raw response:", get_blockchain_reply)
					fmt.Println("Error getting blockchain:", err)
				}
			} else {
				fmt.Println("Blockchain:")
				for _, block := range get_blockchain_reply.Chain {
					fmt.Printf("Index: %d\n", block.Index)
					fmt.Printf("Timestamp: %s\n", block.Timestamp)
					fmt.Printf("Data: %d\n", block.Data)
					fmt.Printf("Hash: %s\n", block.Hash)
					fmt.Printf("PrevHash: %s\n", block.PrevHash)
					fmt.Println()
				}
			}
		case 3:
			// exit the program
			os.Exit(0)
		default:
			// invalid selection
			fmt.Println("Invalid selection")
		}
	}
}
