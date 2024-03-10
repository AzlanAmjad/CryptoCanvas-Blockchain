package main

import (
	"time"

	network "github.com/AzlanAmjad/DreamscapeCanvas-Blockchain/peer-to-peer-network"
)

func main() {
	// Start of the main function.

	// This is our local node, so we have a transport for our own node server.
	// We will have peers, which are remote peers, which will represent servers in the P2P network which are not our machine.
	transportLocal := network.NewLocalTransport("Local")

	// Remote peers will have their own transport, which will be different from the local transport.
	// We will have a transport for each peer.

	// Normally messages will go over RPC to the remote peers, but for now, we will just send messages to the remote transport.
	transportRemote := network.NewLocalTransport("Remote")

	// We will connect the local transport to the remote transport.
	// Note connect method for TCP and UDP will be different, than the local transport.
	transportLocal.Connect(transportRemote)
	transportRemote.Connect(transportLocal)

	// We will simulate the remote peer sending a message to the local peer, in a goroutine (thread).
	go func() {
		for {
			transportRemote.SendMessageToPeer(network.SendRPC{To: "Local", Payload: []byte("Hello from Remote!")})
			time.Sleep(1 * time.Second)
		}
	}()

	// We will have a server which will contain the transport. And any other transport we add in the future.
	serverOptions := network.ServerOptions{
		Transports: []network.Transport{transportLocal},
	}

	// We will create a new server with the server options.
	server := network.NewServer(serverOptions)
	server.Start()
}
