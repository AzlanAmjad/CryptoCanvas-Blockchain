package network

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
)

/*
Implement the Transport interface here
LocalTransport is a transport that is used for communication between modules in the same process.
LocalTransport is not going over the network (TCP or UDP), like other transports might, it is just
sending messages to a channel within the same process.
*/

type LocalTransport struct {
	Addr      NetAddr
	Peers     map[NetAddr]Transport
	Lock      sync.RWMutex
	ConsumeCh chan ReceiveRPC
}

func NewLocalTransport(addr NetAddr) *LocalTransport {
	return &LocalTransport{
		Addr:      addr,
		Peers:     make(map[NetAddr]Transport),
		ConsumeCh: make(chan ReceiveRPC),
	}
}

// method to consume from the channel, used by the transport itself
func (t *LocalTransport) Consume() <-chan ReceiveRPC {
	return t.ConsumeCh
}

// method to send to the channel, used by other peers
func (t *LocalTransport) SendToChannel(rpc ReceiveRPC) error {
	t.ConsumeCh <- rpc
	return nil
}

// method to connect to another transport, used by transport itself
func (t *LocalTransport) Connect(other Transport) error {
	t.Lock.Lock()
	defer t.Lock.Unlock()
	otherAddr := other.GetAddr()
	// make other pointer to LocalTransport
	t.Peers[otherAddr] = other
	return nil
}

// method to send a message, sends to another peers channel, used by transport itself
func (t *LocalTransport) SendMessageToPeer(rpc SendRPC) error {
	t.Lock.RLock()
	defer t.Lock.RUnlock()
	peer, ok := t.Peers[rpc.To]
	if ok {
		peer.SendToChannel(ReceiveRPC{From: t.GetAddr(), Payload: rpc.Payload})
	} else {
		keys := make([]string, 0, len(t.Peers))
		for key := range t.Peers {
			keys = append(keys, string(key))
		}
		return fmt.Errorf("peer %s not found, peers keys: %s", rpc.To, keys)
	}
	return nil
}

// method to get the address of the transport, used by transport itself
func (t *LocalTransport) GetAddr() NetAddr {
	return t.Addr
}

// method to broadcast a message to all peers, used by transport itself
func (t *LocalTransport) Broadcast(payload []byte) error {
	for _, peer := range t.Peers {
		err := t.SendMessageToPeer(SendRPC{To: peer.GetAddr(), Payload: bytes.NewReader(payload)})
		if err != nil {
			logrus.WithError(err).Error("error sending message to peer")
			return err
		}
	}
	return nil
}
