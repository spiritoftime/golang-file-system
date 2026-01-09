package main

import (
	"file_system/p2p"
	"log"
	"time"
)

// func OnPeer(peer p2p.Peer) error {
// 	peer.Close()
// 	// fmt.Println("doing some logic with the peer outside of TCPTransport")
// 	return nil
// }

func main() {
	tcpTransportOpts := p2p.TCPTransportOpts{
		ListenAddr:    ":3000",
		HandshakeFunc: p2p.NOTHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
		// TODO: OnPeer func
		// OnPeer: ,
	}
	tcpTransport := p2p.NewTCPTransport(tcpTransportOpts)
	fileServerOpts := FileServerOpts{
		ListenAddr:        ":3000",
		StorageRoot:       "3000_network",
		PathTransformFunc: CASPathTransformFunc,
		Transport:         tcpTransport,
	}
	s := NewFileServer(fileServerOpts)
	go func() {
		time.Sleep(time.Second * 3)
		s.Stop()
	}()
	if err := s.Start(); err != nil {
		log.Fatal(err)
	}
	// tr := p2p.NewTCPTransport(":3000")
	// tcpOpts := p2p.TCPTransportOpts{
	// 	ListenAddr: ":3000",

	// 	HandshakeFunc: p2p.NOTHandshakeFunc,
	// 	Decoder:       p2p.DefaultDecoder{},
	// 	OnPeer:        OnPeer,
	// }
	// tr := p2p.NewTCPTransport(tcpOpts)
	// go func() {
	// 	for {
	// 		msg := <-tr.Consume()
	// 		fmt.Printf("%+v\n", msg)
	// 	}
	// }()
	// if err := tr.ListenAndAccept(); err != nil {
	// 	log.Fatal(err)
	// }
	// select {}

	// fmt.Println("me Gucci")
}
