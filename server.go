package main

import (
	"file_system/p2p"
	"fmt"
	"io"
	"log"
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
	store  *Store
	quitch chan struct{} // quit channel
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
	}
}

func (s *FileServer) Stop() {
	close(s.quitch)
}

func (s *FileServer) loop() {

	defer func() {
		log.Println("file server stopped due to user quit action")
		s.Transport.Close()

	}()

	for {
		select {
		case msg := <-s.Transport.Consume(): // receive and use the value
			fmt.Println(msg)
		case <-s.quitch: // receive but ignore the value
			return
		}
	}
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
