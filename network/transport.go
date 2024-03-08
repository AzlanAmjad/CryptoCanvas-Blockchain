package network

// NetAddr is a string representing a network address.
type NetAddr string

// SendRPC is a struct containing the address to send to and the payload to send.
type SendRPC struct {
	To      NetAddr
	Payload []byte
}

// ReceiveRPC is a struct containing the address received from and the payload received.
type ReceiveRPC struct {
	From    NetAddr
	Payload []byte
}

// Transport is an interface for a network transport. It should be able to consume messages, connect to other transports, send messages, and get its address.
type Transport interface {
	Consume() <-chan ReceiveRPC
	SendToChannel(ReceiveRPC) error
	Connect(Transport) error
	SendMessageToPeer(SendRPC) error
	GetAddr() NetAddr
}
