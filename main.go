package main

import "github.com/AzlanAmjad/DreamscapeCanvas-Blockchain/network"

func main() {
	// Start of the main function.

	// This is our local node, so we have a transport for our own node server.
	// We will have peers, which are remote peers, which will represent servers in the P2P network which are not our machine.
	transportLocal := network.NewLocalTransport("Local")

	// We will have a server which will contain the transport. And any other transport we add in the future.
	serverOptions := network.ServerOptions{
		Transports: []network.Transport{transportLocal},
	}

	// We will create a new server with the server options.
	server := network.NewServer(serverOptions)
	server.Start()
}
