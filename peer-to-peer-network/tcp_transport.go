package network

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/go-kit/log"
)

// tcp buffer size const for reading from the connection
const TCP_BUFFER_SIZE = 8192
const FAIL_COUNT_LIMIT = 5

// TCPTransport is a transport that communicates with peers over TCP.
type TCPTransport struct {
	Addr        net.Addr
	listener    net.Listener
	peerChannel chan *TCPPeer
	Logger      log.Logger
}

type TCPPeer struct {
	conn net.Conn
	// tells us if the connection was established by us
	// (dialing / outgoing) or by another peer (accepting / incoming)
	Incoming bool
}

func (tp *TCPPeer) Send(payload []byte) error {
	_, err := tp.conn.Write(payload)
	if err != nil {
		fmt.Println("Error sending message to", tp.conn.RemoteAddr())
		return err
	}
	return nil
}

func (tp *TCPPeer) readLoop(rpcChannel chan ReceiveRPC, removePeerChannel chan *TCPPeer) {
	fail_count := 0

	for {
		buf := make([]byte, TCP_BUFFER_SIZE)
		n, err := tp.conn.Read(buf)
		if err != nil {
			if fail_count >= FAIL_COUNT_LIMIT {
				fmt.Println("Closing connection with", tp.conn.RemoteAddr())
				tp.conn.Close()

				// tell the node to remove this peer
				removePeerChannel <- tp

				// return to stop the loop
				return
			}
			fmt.Println("Error with connection:", tp.conn.RemoteAddr())
			fmt.Println("Error reading from connection:", err)
			// Wait 5 seconds before retrying
			time.Sleep(5 * time.Second)
			fail_count++
			continue
		}

		fmt.Println("Received message from", tp.conn.RemoteAddr())

		// create ReceiveRPC struct
		rpc := ReceiveRPC{From: tp.conn.RemoteAddr(), Payload: bytes.NewReader(buf[:n])}
		// send over the channel to the server
		rpcChannel <- rpc
	}
}

// NewTCPTransport creates a new instance of TCPTransport.
func NewTCPTransport(addr net.Addr, peerCh chan *TCPPeer, ID string) (*TCPTransport, error) {
	transport := &TCPTransport{
		Addr:        addr,
		listener:    nil,
		peerChannel: peerCh,
	}

	// set the default logger
	transport.Logger = log.NewLogfmtLogger(os.Stderr)
	transport.Logger = log.With(transport.Logger, "ID", ID)

	return transport, nil
}

// start the tcp transport
func (t *TCPTransport) Start() error {
	listener, err := net.Listen("tcp", t.Addr.String())
	if err != nil {
		return err
	}

	t.listener = listener

	t.Logger.Log("msg", "TCP transport listening", "addr", t.Addr.String())

	// forever loop to accept incoming connections
	go t.listen()

	return nil
}

// listen listens for incoming connections.
func (t *TCPTransport) listen() {
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		// handle incoming connection
		go t.handlePeerConn(conn)
	}
}

// handlePeerConn handles an incoming peer connection
// and adds it to the list of peers. there is a read loop in
// in this handlePeerConn function
// this is run in parallel using goroutines, each connection is handled in a separate goroutine
func (t *TCPTransport) handlePeerConn(conn net.Conn) {
	fmt.Println("Received connection from", conn.RemoteAddr())

	// create a new peer
	peer := &TCPPeer{
		conn:     conn,
		Incoming: true,
	}
	// send over the channel to the server
	t.peerChannel <- peer
}
