package p2p

import "errors"

// ErrInvalidHandshake is returned if handshake between local and remote node could not be established
var ErrInvalidHandshake = errors.New("invalid handshake")

// HandshakeFunc... ?
type HandshakeFunc func(Peer) error

func NOTHandshakeFunc(Peer) error { return nil }
