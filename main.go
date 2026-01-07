package main

import (
	"file_system/p2p"
	"log"
)

func main() {

	// tr := p2p.NewTCPTransport(":3000")
	tcpOpts := p2p.TCPTransportOpts{
		ListenAddr: ":3000",

		HandshakeFunc: p2p.NOTHandshakeFunc,
		Decoder:       p2p.GOBDecoder{},
	}
	tr := p2p.NewTCPTransport(tcpOpts)
	if err := tr.ListenAndAccept(); err != nil {
		log.Fatal(err)
	}
	select {}

	// fmt.Println("me Gucci")
}
