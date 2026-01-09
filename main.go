package main

import (
	"file_system/p2p"
	"fmt"
	"log"
)

func OnPeer(peer p2p.Peer) error {
	peer.Close()
	// fmt.Println("doing some logic with the peer outside of TCPTransport")
	return nil
}

func main() {

	// tr := p2p.NewTCPTransport(":3000")
	tcpOpts := p2p.TCPTransportOpts{
		ListenAddr: ":3000",

		HandshakeFunc: p2p.NOTHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
		OnPeer:        OnPeer,
	}
	tr := p2p.NewTCPTransport(tcpOpts)
	go func() {
		for {
			msg := <-tr.Consume()
			fmt.Printf("%+v\n", msg)
		}
	}()
	if err := tr.ListenAndAccept(); err != nil {
		log.Fatal(err)
	}
	select {}

	// fmt.Println("me Gucci")
}
