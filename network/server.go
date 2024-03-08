package network

import (
	"fmt"
	"time"
)

/*
	The server is a container which will contain every module of the server.
*/

type Server struct {
	ServerOptions ServerOptions
	rpcChannel    chan ReceiveRPC
	quitChannel   chan bool
}

// One node can have multiple transports, for example, a TCP transport and a UDP transport.
type ServerOptions struct {
	Transports []Transport
}

// NewServer creates a new server with the given options.
func NewServer(options ServerOptions) *Server {
	return &Server{
		ServerOptions: options,
		rpcChannel:    make(chan ReceiveRPC),
		quitChannel:   make(chan bool),
	}
}

// Start will start the server.
func (s *Server) Start() {
	s.initTransports()
	ticker := time.NewTicker(5 * time.Second)

	// Run the server in a loop forever.
	for {
		select {
		case message := <-s.rpcChannel:
			// Process the message.
			fmt.Printf("Received message from %s: %s\n", message.From, message.Payload)
		case <-s.quitChannel:
			// Quit the server.
			fmt.Println("Quitting the server.")
			return
		case <-ticker.C:
			// Do something every 5 seconds.
			fmt.Println("Doing something every 5 seconds.")
		}
	}
}

func (s *Server) Stop() {
	s.quitChannel <- true
}

// initTransports will initialize the transports.
func (s *Server) initTransports() {
	// For each transport, spin up a new goroutine (user level thread).
	for _, transport := range s.ServerOptions.Transports {
		go s.startTransport(transport)
	}
}

// startTransport will start the transport.
func (s *Server) startTransport(transport Transport) {
	// Keep reading the channel and process the messages forever.
	for message := range transport.Consume() {
		// Send ReceiveRPC message to the server's rpcChannel. For synchronous processing.
		s.rpcChannel <- message
	}
}
