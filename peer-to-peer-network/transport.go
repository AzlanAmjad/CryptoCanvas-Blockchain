package network

// NetAddr is a string representing a network address.
type NetAddr string

// Transport is an interface for a network transport. It should be able to consume messages, connect to other transports, send messages, and get its address.
type Transport interface {
	Consume() <-chan ReceiveRPC
	SendToChannel(ReceiveRPC) error
	Connect(Transport) error
	SendMessageToPeer(SendRPC) error
	GetAddr() NetAddr
	Broadcast([]byte) error
}
