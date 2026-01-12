package p2p

// always organise code from public to private
import (
	"errors"
	"fmt"
	"log"
	"net"
)

// TCPPeer represents the remote node over a TCP established connection.
type TCPPeer struct {
	// the underlying connection of the peer. which in this case is a TCP connection.
	net.Conn
	// if we dial a connection => outbound ==true
	// if we accept and  retrieve a connection => outbound ==false
	outbound bool
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		Conn:     conn,
		outbound: outbound,
	}
}

func (p *TCPPeer) Send(b []byte) error {
	_, err := p.Conn.Write(b)
	return err
}

type TCPTransportOpts struct {
	ListenAddr    string
	HandshakeFunc HandshakeFunc
	Decoder       Decoder
	OnPeer        func(Peer) error
}

type TCPTransport struct {
	listener net.Listener
	TCPTransportOpts
	rpcch chan RPC

	// // common practice put the mutex above the thing to lock
	// mu    sync.RWMutex
	// peers map[net.Addr]Peer
}

func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {
	return &TCPTransport{
		TCPTransportOpts: opts,
		rpcch:            make(chan RPC),
	}
}

// func (t *TCPTransport) ListenAddr() string {
// 	return t.listener.Addr().Network()
// }

// Consume implements the transport interface, which will return read-only channel for reading the incoming messages received from another peer in the network.
func (t *TCPTransport) Consume() <-chan RPC {
	return t.rpcch
}

// Close implements the Transport interface.
func (t *TCPTransport) Close() error {
	return t.listener.Close()
}

// Dial implements the transport interface.
func (t *TCPTransport) Dial(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Println("tcp dial error:", err)
		return err
	}
	fmt.Println(conn)
	go t.handleConn(conn, true) // outbound func
	return nil

}
func (t *TCPTransport) ListenAndAccept() error {

	var err error

	t.listener, err = net.Listen("tcp", t.ListenAddr)
	if err != nil {
		return err
	}
	go t.startAcceptLoop()
	log.Printf("TCP transport listening on port: %s\n", t.ListenAddr)
	return nil

}

func (t *TCPTransport) startAcceptLoop() {
	for {
		conn, err := t.listener.Accept()
		if errors.Is(err, net.ErrClosed) {
			return
		}
		if err != nil {
			fmt.Printf("TCP accept error: %s", err)
		}
		go t.handleConn(conn, false)
	}
}

// type Temp struct{}

func (t *TCPTransport) handleConn(conn net.Conn, outbound bool) {
	var err error
	defer func() {
		fmt.Printf("dropping peer connection: %s", err)
		conn.Close()
	}()
	peer := NewTCPPeer(conn, outbound)
	if err = t.HandshakeFunc(peer); err != nil { // perform handshake
		// conn.Close()
		// fmt.Printf("TCP handshake error: %s\n", err)
		return
	}
	if t.OnPeer != nil {
		if err = t.OnPeer(peer); err != nil { // onPeer is callback function whenever peer successfully connects to our server
			return
		}
	}
	for {
		rpc := RPC{}
		err = t.Decoder.Decode(conn, &rpc)
		// fmt.Println(reflect.TypeOf(err))
		// panic(err)
		// if err == net.OpError { //eg: TCP error: read tcp [::1]:3000->[::1]:52202: use of closed network connection
		// 	return
		// }
		if err != nil {
			// fmt.Printf("TCP read error: %s\n", err)
			// continue
			return
		}

		rpc.From = conn.RemoteAddr()
		fmt.Printf("message: %+v\n", rpc)
		t.rpcch <- RPC{}
	}
}
