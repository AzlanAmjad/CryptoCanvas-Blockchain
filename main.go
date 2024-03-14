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
	transportRemoteA := network.NewLocalTransport("RemoteA")
	transportRemoteB := network.NewLocalTransport("RemoteB")
	transportRemoteC := network.NewLocalTransport("RemoteC")

	// mock client transport that will send transactions to the local transport
	transportMockClient := network.NewLocalTransport(("MockClient"))
	transportMockClient.Connect(transportLocal)

	// Connect one transport to another, not all are connected, blockchain will use the gossip architecture to
	// propagate transactions and blocks across the simulated network.
	// Note connect method for TCP and UDP will be different, than the local transport.
	transportLocal.Connect(transportRemoteA)
	transportRemoteA.Connect(transportRemoteB)
	transportRemoteB.Connect(transportRemoteC)

	// We will simulate the remote peer sending a message to the local peer, in a goroutine (thread).
	go func() {
		for {
			// We will simulate the mock client sending a message to the local.
			if err := sendTransaction(transportMockClient, transportLocal.GetAddr()); err != nil {
				logrus.WithError(err).Error("Failed to send transaction")
			}
			time.Sleep(3 * time.Second)
		}
	}()

	// Generate private key for server
	privateKey := crypto.GeneratePrivateKey()

	// We will create a remote server with the remote transports, to simulate the remote peers.
	remoteServerA := makeServer("RemoteServerA", transportRemoteA, nil)
	remoteServerB := makeServer("RemoteServerB", transportRemoteB, nil)
	remoteServerC := makeServer("RemoteServerC", transportRemoteC, nil)

	// separate go routine so we don't block the main thread
	go remoteServerA.Start()
	go remoteServerB.Start()
	go remoteServerC.Start()

	// We will create a local server with the local transport.
	// only local server will be the validator here in this simulated environment.
	localServer := makeServer("LocalServer", transportLocal, &privateKey)

	// Start the local server
	localServer.Start()
}

func makeServer(id string, tr network.Transport, privateKey *crypto.PrivateKey) *network.Server {
	// We will have a server which will contain the transport. And any other transport we add in the future.
	serverOptions := network.ServerOptions{
		ID:         id,
		Transports: []network.Transport{tr},
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
	err = tr.SendMessageToPeer(network.SendRPC{To: to, Payload: bytes.NewReader(msg.Bytes())})
	if err != nil {
		return err
	}

	return nil
}
