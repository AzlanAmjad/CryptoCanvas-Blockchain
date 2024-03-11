package main

import (
	"bytes"
	"crypto/rand"
	"time"

	core "github.com/AzlanAmjad/DreamscapeCanvas-Blockchain/blockchain-core"
	crypto "github.com/AzlanAmjad/DreamscapeCanvas-Blockchain/cryptography"
	network "github.com/AzlanAmjad/DreamscapeCanvas-Blockchain/peer-to-peer-network"
	"github.com/sirupsen/logrus"
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
			// We will simulate the remote peer sending a message to the local peer.
			if err := sendTransaction(transportRemote, transportLocal.GetAddr()); err != nil {
				logrus.WithError(err).Error("Failed to send transaction")
			}
			time.Sleep(1 * time.Second)
		}
	}()

	// Generate private key for server
	privateKey := crypto.GeneratePrivateKey()
	// We will have a server which will contain the transport. And any other transport we add in the future.
	serverOptions := network.ServerOptions{
		PrivateKey: &privateKey,
		ID:         "LocalServer",
		Transports: []network.Transport{transportLocal},
	}

	// We will create a new server with the server options.
	server, _ := network.NewServer(serverOptions)
	server.Start()
}

func sendTransaction(tr network.Transport, to network.NetAddr) error {
	// Generate a random byte slice
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return err
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

	// send the message
	tr.SendMessageToPeer(network.SendRPC{To: to, Payload: bytes.NewReader(msg.Bytes())})

	return nil
}
