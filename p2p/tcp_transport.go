package p2p

// always organise code from public to private
import (
	"fmt"
	"log"
	"net"
)

// TCPPeer represents the remote node over a TCP established connection.
type TCPPeer struct {
	// conn is the underlying connection of the peer.
	conn net.Conn
	// if we dial a connection => outbound ==true
	// if we accept and  retrieve a connection => outbound ==false
	outbound bool
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		conn:     conn,
		outbound: outbound,
	}
}

// Close implements the Peer interface
func (p *TCPPeer) Close() error {
	return p.conn.Close()
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

// Consume implements the transport interface, which will return read-only channel for reading the incoming messages received from another peer in the network.
func (t *TCPTransport) Consume() <-chan RPC {
	return t.rpcch
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
		if err != nil {
			fmt.Printf("TCP accept error: %s", err)
		}
		fmt.Printf("new incoming connection %+v\n", conn)
		go t.handleConn(conn)
	}
}

// type Temp struct{}

func (t *TCPTransport) handleConn(conn net.Conn) {
	var err error
	defer func() {
		fmt.Printf("dropping peer connection: %s", err)
		conn.Close()
	}()
	peer := NewTCPPeer(conn, true)
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
