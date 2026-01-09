package p2p

// Peer is an interface that represents the remote node
type Peer interface {
	Close() error
	// Consume() <-chan RPC
}

// Transport is anything that handles the communication
// between the nodes in the network. this can be of the form (TCP,UDP, websockets, ...)
// 3 or 4 functions on a interface should be the maximum
type Transport interface {
	ListenAndAccept() error
	Consume() <-chan RPC
	Close() error
}
