package main

import (
	"bytes"
	"encoding/gob"
	"file_system/p2p"
	"fmt"
	"io"
	"log"
	"sync"
)

type FileServerOpts struct {
	ListenAddr        string
	StorageRoot       string
	PathTransformFunc PathTransformFunc
	Transport         p2p.Transport
	BootstrapNodes    []string
}
type FileServer struct {
	FileServerOpts
	peerLock sync.Mutex
	peers    map[string]p2p.Peer
	store    *Store
	quitch   chan struct{} // quit channel
}

func NewFileServer(opts FileServerOpts) *FileServer {
	storeOpts := StoreOpts{
		Root:              opts.StorageRoot,
		PathTransformFunc: opts.PathTransformFunc,
	}
	return &FileServer{
		store:          NewStore(storeOpts),
		FileServerOpts: opts,
		quitch:         make(chan struct{}),
		peers:          make(map[string]p2p.Peer),
	}
}

func (s *FileServer) broadcast(msg *Message) error {
	peers := []io.Writer{}

	for _, peer := range s.peers {
		peers = append(peers, peer)
	}
	mw := io.MultiWriter(peers...)
	return gob.NewEncoder(mw).Encode(msg)
}

type Message struct {
	From    string
	Payload any
}

type DataMessage struct {
	Key  string
	Data []byte
}

func (s *FileServer) StoreData(key string, r io.Reader) error {
	//  1. Store this file to disk
	// 2. broadcast this file to all known peers in the network

	buf := new(bytes.Buffer)

	tee := io.TeeReader(r, buf)
	fmt.Println("i am writing some thing")
	if err := s.store.Write(key, tee); err != nil {
		fmt.Println("i am under the water")

		return err
	}

	p := &DataMessage{
		Key:  key,
		Data: buf.Bytes(),
	}

	return s.broadcast(&Message{
		From:    "todo",
		Payload: p,
	})

}

func (s *FileServer) Stop() {
	close(s.quitch)
}

func (s *FileServer) OnPeer(peer p2p.Peer) error {
	s.peerLock.Lock()
	defer s.peerLock.Unlock()
	s.peers[peer.RemoteAddr().String()] = peer
	fmt.Printf("connected with remote %s", peer.RemoteAddr().String())
	return nil
}

func (s *FileServer) loop() {

	defer func() {
		log.Println("file server stopped due to user quit action")
		s.Transport.Close()

	}()

	for {
		select {
		case msg := <-s.Transport.Consume(): // receive and use the value
			var m Message
			if err := gob.NewDecoder(bytes.NewReader(msg.Payload)).Decode(&m); err != nil {
				log.Println(err)
			}
			if err := s.handleMessage(&m); err != nil {

				log.Println(err)

			}
			// fmt.Printf("%+v\n", string(m.Payload.Data))
		case <-s.quitch: // receive but ignore the value
			return
		}
	}
}

func (s *FileServer) handleMessage(msg *Message) error {

	switch v := msg.Payload.(type) {
	case *DataMessage:

		fmt.Printf("received data %+v\n", v)
	}
	return nil
}

// set up a peer to peer network
// A peer is an individual node/participant in the P2P network. Each peer can:

// Send data to other peers
// Receive data from other peers
// Store data locally
// Forward requests to other peers
// Traditional Client-Server vs P2P
// CLIENT-SERVER:
// Client → Server ← Client
// Client → Server ← Client
// (all data flows through central server)

// P2P:
// Peer ↔ Peer
//
//	↕       ↕
//
// Peer ↔ Peer
// (peers connect directly to each other)
func (s *FileServer) bootstrapNetwork() error {
	for _, addr := range s.BootstrapNodes {
		if len(addr) == 0 { // someone passed in ""
			continue
		}
		go func(addr string) {

			if err := s.Transport.Dial(addr); err != nil {
				log.Println("dial error: ", err)

			}
		}(addr)
	}
	return nil
}

func (s *FileServer) Start() error {
	if err := s.Transport.ListenAndAccept(); err != nil {
		return err
	}
	if len(s.BootstrapNodes) != 0 {
		s.bootstrapNetwork()
	}
	s.loop()
	return nil
}

func (s *FileServer) Store(key string, r io.Reader) error {
	return s.store.Write(key, r)
}
