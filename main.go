package main

import (
	"bytes"
	"encoding/gob"
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
	// log.Printf("wtf is %s, %s", listenAddr, strings.Split(listenAddr, ":")[0])
	tcpTransportOpts := p2p.TCPTransportOpts{
		ListenAddr:    listenAddr,
		HandshakeFunc: p2p.NOTHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
		// TODO: OnPeer func
		// OnPeer: ,
	}
	tcpTransport := p2p.NewTCPTransport(tcpTransportOpts)
	fileServerOpts := FileServerOpts{
		StorageRoot:       strings.TrimPrefix(listenAddr, ":") + "_network", // remove the :
		PathTransformFunc: CASPathTransformFunc,
		Transport:         tcpTransport,
		BootstrapNodes:    nodes,
	}
	s := NewFileServer(fileServerOpts)
	tcpTransport.OnPeer = s.OnPeer
	return s
}

// runs before main
func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	gob.Register(&DataMessage{})
}
func main() {

	s1 := makeServer(":3000", "")
	s2 := makeServer(":4000", ":3000")
	go func() {
		log.Fatal(s1.Start())
	}()
	time.Sleep(5 * time.Second)
	go s2.Start()
	time.Sleep(5 * time.Second)
	data := bytes.NewReader([]byte("my big data file here"))
	if err := s2.StoreData("myprivatedata", data); err != nil {
		fmt.Println(err)
	}
	time.Sleep(10 * time.Second)
	select {}
}
