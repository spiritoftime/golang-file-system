package main

import (
	"bytes"
	"encoding/gob"
	"file_system/p2p"
	"fmt"
	"io"
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
	encKey, err := newEncryptionKey()
	if err != nil {
		fmt.Println(err)
	}
	fileServerOpts := FileServerOpts{
		EncKey:            encKey,
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
	// gob interface anything in any type needs to be declared
	gob.Register(MessageStoreFile{})
	gob.Register(MessageGetFile{})
}
func main() {
	s1 := makeServer(":3000", "")
	s2 := makeServer(":7000", ":3000")
	// todo: implement automatic peer discovery
	s3 := makeServer(":5000", ":3000", ":7000")

	go func() {
		log.Fatal(s1.Start())
	}()
	go func() { //need separate go routines because .Start has a accept loop so it will never run

		log.Fatal(s2.Start())
	}()
	time.Sleep(4 * time.Second)
	go s3.Start()
	time.Sleep(2 * time.Second) //if no sleep store will run and broadcast will be attempted before server even starts

	for i := 0; i < 20; i++ {

		// key := "coolPicture.jpg"
		key := fmt.Sprintf("picture_%d.png", i)
		data := bytes.NewReader([]byte("my big data file here"))
		s3.Store(key, data)
		time.Sleep(200 * time.Millisecond)
		if err := s3.store.Delete(s3.ID, key); err != nil {
			log.Fatal(err)
		}
		// -----------get file
		r, err := s3.Get(key)
		if err != nil {
			log.Fatal(err)
		}
		b, err := io.ReadAll(r)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(b))
	}
}
