package main

import (
	"bytes"
	"file_system/p2p"
	"fmt"
	"log"
	"strings"
	"time"
)

// func OnPeer(peer p2p.Peer) error {
// 	peer.Close()
// 	// fmt.Println("doing some logic with the peer outside of TCPTransport")
// 	return nil
// }

func makeServer(listenAddr string, nodes ...string) *FileServer {
	tcpTransportOpts := p2p.TCPTransportOpts{
		ListenAddr:    listenAddr,
		HandshakeFunc: p2p.NOTHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
		// TODO: OnPeer func
		// OnPeer: ,
	}
	tcpTransport := p2p.NewTCPTransport(tcpTransportOpts)
	fileServerOpts := FileServerOpts{
		StorageRoot:       strings.Split(listenAddr, ":")[0] + "_network",
		PathTransformFunc: CASPathTransformFunc,
		Transport:         tcpTransport,
		BootstrapNodes:    nodes,
	}
	s := NewFileServer(fileServerOpts)
	tcpTransport.OnPeer = s.OnPeer
	return s
}

func main() {
	s1 := makeServer(":3000", "")
	s2 := makeServer(":4000", ":3000")
	go func() {
		log.Fatal(s1.Start())
	}()
	go s2.Start()
	time.Sleep(1 * time.Second)
	data := bytes.NewReader([]byte("my big data file here"))
	if err := s1.StoreData("myprivatedata", data); err != nil {
		fmt.Println(err)
	}
	time.Sleep(5 * time.Second)
}
