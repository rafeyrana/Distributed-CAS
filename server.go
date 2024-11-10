package main

import (
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/rafeyrana/Distributed-CAS/p2p"
)

type FileServerOpts struct {
	ListenAddress    string
	StorageRoot 	 string
	PathTransformFunc PathTransformFunc
	Transport p2p.Transport
	BootstrapNodes []string
}
type FileServer struct {
	FileServerOpts 
	peerLock sync.Mutex
	peers map[string]p2p.Peer
	store *Store

	quitchan chan struct{}
}

func NewFileServer(opts FileServerOpts) *FileServer {
	storeOpts := StoreOpts{
		Root: opts.StorageRoot,
		PathTransformFunc: opts.PathTransformFunc,
	}
	return &FileServer{
		FileServerOpts: opts,
		store : NewStore(storeOpts),
		quitchan : make(chan struct{}),
		peers : make(map[string]p2p.Peer),
	}
}

func (s *FileServer) loop() {
	defer func(){
		log.Println("file server stopped")
		s.Transport.Close()
	}()

	for {
		select{
		case msg := <- s.Transport.Consume():
			fmt.Printf("received message %s", msg.Payload)
		case <- s.quitchan:
			return 

		}
	}
}

type Payload struct{
	Key string
	Data []byte
}

func (s *FileServer) Broadcast (p Payload) error {
	peers := []io.Writer{}
	for _, peer := range s.peers {
		peers = append(peers, peer)

		}
	mw := io.MultiWriter(peers...)

	}
	


func (s *FileServer) StoreData(key string, r io.Reader) error {
	// store this file in the disk
	// broadcast file to all known peers in the network

	return nil
	
}

func (s *FileServer) Stop(){
	close(s.quitchan)
}

func (s *FileServer)BootstrapNetwork() error{
	for _, node_address := range s.BootstrapNodes {
		if len(node_address) == 0 {
			continue
		}
		go func (node_address string) {
			fmt.Printf("[%s] attemping to connect with remote %s\n", node_address, node_address)
			if err := s.Transport.Dial(node_address); err != nil {
				log.Println("error dialing node", node_address, err)
			}
		}(node_address)
	}
	return nil


}


func (s *FileServer) OnPeer(p p2p.Peer) error {
	fmt.Printf("connected to remote peer %s", p.RemoteAddr().String())
    s.peerLock.Lock()
    defer s.peerLock.Unlock()
    s.peers[p.RemoteAddr().String()] = p
    fmt.Printf("connected to remote peer %s", p.RemoteAddr().String())
    return nil
}

func (s *FileServer) Start() error {
	if err:= s.Transport.ListenAndAccept(); err != nil {
		return err
	}

	if err := s.BootstrapNetwork(); err != nil {
		return err
	}

	// can we block and start using go routine or not block or execute directly
	s.loop()
	return nil
}