package main

import (
	"bytes"
	"encoding/gob"
	"file_system/p2p"
	"fmt"
	"io"
	"log"
	"sync"
	"time"
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

func (s *FileServer) stream(msg *Message) error {
	peers := []io.Writer{}

	for _, peer := range s.peers {
		peers = append(peers, peer)
	}
	mw := io.MultiWriter(peers...)
	return gob.NewEncoder(mw).Encode(msg)
}

func (s *FileServer) broadcast(msg *Message) error {
	buf := new(bytes.Buffer) // use for gob -encoded data
	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		return err
	}
	for _, peer := range s.peers {
		if err := peer.Send(buf.Bytes()); err != nil {
			return err
		}
	}
	return nil
}

type Message struct {
	Payload any
}

type MessageStoreFile struct {
	// store file based on key
	Key  string
	Size int64
}

type MessageGetFile struct {
	Key string
}

func (s *FileServer) Get(key string) (io.Reader, error) {
	if s.store.Has(key) {
		return s.store.Read(key)
	}
	fmt.Printf("dont have file (%s) locally, fetching from network\n", key)

	msg := Message{
		Payload: MessageGetFile{
			Key: key,
		},
	}
	// fetch from all peers to see if they have
	if err := s.broadcast(&msg); err != nil {
		return nil, err
	}
	time.Sleep(time.Second * 3)
	for _, peer := range s.peers {
		fmt.Println("receiving stream from peer: ", peer.RemoteAddr())
		fileBuffer := new(bytes.Buffer)
		n, err := io.CopyN(fileBuffer, peer, 22)
		if err != nil {
			return nil, err
		}
		fmt.Println("received bytes over the network: ", n)
		fmt.Println(fileBuffer.String())
	}
	select {}
	// return nil, nil
}

func (s *FileServer) Store(key string, r io.Reader) error {
	//  1. Store this file to disk
	// 2. broadcast this file to all known peers in the network
	var (
		fileBuffer = new(bytes.Buffer) // use for file content
		tee        = io.TeeReader(r, fileBuffer)
	)
	size, err := s.store.Write(key, tee)
	if err != nil {
		return err
	}
	msg := Message{
		Payload: MessageStoreFile{
			Key:  key,
			Size: size,
		},
	}
	if err := s.broadcast(&msg); err != nil {
		return err
	}

	time.Sleep(3 * time.Second)

	// TODO: use a multiwriter here
	for _, peer := range s.peers {
		n, err := io.Copy(peer, fileBuffer)
		if err != nil {
			return err
		}
		fmt.Println("received and written bytes to disk:", n)
	}

	return nil

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
		log.Println("file server stopped due to error or user quit action")
		s.Transport.Close()

	}()

	for {
		select {
		case rpc := <-s.Transport.Consume(): // receive and use the value
			var msg Message
			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&msg); err != nil {
				log.Println("decoding error: ", err)

			}
			if err := s.handleMessage(rpc.From, &msg); err != nil {
				log.Println("handle message error: ", err)
			}

		case <-s.quitch: // receive but ignore the value
			return
		}
	}
}

func (s *FileServer) handleMessage(from string, msg *Message) error {

	switch v := msg.Payload.(type) {
	case MessageStoreFile:
		return s.handleMessageStoreFile(from, v)
	case MessageGetFile:
		return s.handleMessageGetFile(from, v)
	}
	return nil
}

func (s *FileServer) handleMessageGetFile(from string, msg MessageGetFile) error {
	if !s.store.Has(msg.Key) {
		return fmt.Errorf("need to serve  file (%s) but it does not exist on disk", msg.Key)
	}
	fmt.Printf("serving file (%s) serving over the network\n", msg.Key)
	r, err := s.store.Read(msg.Key)
	if err != nil {
		return err
	}

	peer, ok := s.peers[from]
	if !ok {
		return fmt.Errorf("peer %s not in map", from)
	}
	// copy file to peer
	n, err := io.Copy(peer, r)
	if err != nil {
		return err
	}
	fmt.Printf("written %d bytes over the network to %s\n", n, from)
	return nil
}

func (s *FileServer) handleMessageStoreFile(from string, msg MessageStoreFile) error {
	peer, ok := s.peers[from]
	if !ok {
		return fmt.Errorf("peer (%s) could not be found in the peer list", from)
	}
	n, err := s.store.Write(msg.Key, io.LimitReader(peer, msg.Size))
	if err != nil {
		return err
	}
	fmt.Printf("written %d bytes to disk \n", n)
	peer.(*p2p.TCPPeer).Wg.Done()
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

// func (s *FileServer) Store(key string, r io.Reader) (int64, error) {
// 	return s.store.Write(key, r)
// }
