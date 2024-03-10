package network

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"

	core "github.com/AzlanAmjad/DreamscapeCanvas-Blockchain/blockchain-core"
)

// SendRPC is a struct containing the address to send to and the payload to send.
type SendRPC struct {
	To      NetAddr
	Payload io.Reader
}

// ReceiveRPC is a struct containing the address received from and the payload received.
type ReceiveRPC struct {
	From    NetAddr
	Payload io.Reader
}

// RPC message types
type MessageType byte

const (
	Transaction MessageType = iota
	Block
)

// Message is a struct containing the message type and the payload.
type Message struct {
	Header  MessageType
	Payload []byte
}

func NewMessage(header MessageType, payload []byte) Message {
	return Message{Header: header, Payload: payload}
}

func (m *Message) Bytes() []byte {
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(m)
	if err != nil {
		panic(err)
	}
	return buf.Bytes()
}

// RPCHandler is an interface for handling RPCs.
type RPCHandler interface {
	HandleRPC(rpc ReceiveRPC) error
}

// DefaultRPCHandler is the default implementation of RPCHandler.
type DefaultRPCHandler struct {
	Processor RPCProcessor
}

func NewDefaultRPCHandler(processor RPCProcessor) *DefaultRPCHandler {
	return &DefaultRPCHandler{Processor: processor}
}

// HandleRPC processes the RPC.
func (h *DefaultRPCHandler) HandleRPC(rpc ReceiveRPC) error {
	msg := Message{}

	// expecting the payload to be a gob encoded message
	err := gob.NewDecoder(rpc.Payload).Decode(&msg)
	if err != nil {
		return fmt.Errorf("error decoding message from %s: %s", rpc.From, err)
	}

	switch msg.Header {
	case Transaction:
		tx := core.Transaction{}
		err = tx.Decode(bytes.NewReader(msg.Payload), core.NewTransactionDecoder())
		if err != nil {
			return err
		}
		return h.Processor.ProcessTransaction(rpc.From, &tx)
	default:
		return fmt.Errorf("unknown message type: %d", msg.Header)
	}
}

// RPCProcessor is an interface for processing RPCs.
type RPCProcessor interface {
	ProcessTransaction(NetAddr, *core.Transaction) error
}
