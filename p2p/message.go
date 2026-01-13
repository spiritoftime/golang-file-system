package p2p

const (
	IncomingMessage = 0x1 // hex for 1
	IncomingStream  = 0x2 //hex for 2
)

// RPC holds any arbitrary data that is being sent over the
// each transport between two nodes in the network.
type RPC struct {
	From    string
	Payload []byte
	Stream  bool
}
