package main

import (
	"encoding/gob"
	"fmt"
	"io"
	"time"
	"log"
	"sync"
	"bytes"
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
func (s *FileServer) HandleMessage(from string, msg *Message) error {
	switch v := msg.Payload.(type){
	case MessageStoreFile:
		return s.handleMessageStoreFile(from, v)
	}
	return nil
}
func (s *FileServer) handleMessageStoreFile(from string, msg MessageStoreFile) error {

	peer , ok := s.peers[from]

	if !ok {
		fmt.Errorf("peer not found")
		panic("peer not found")
	}
	if _, err := s.store.Write(msg.Key, io.LimitReader(peer,int64(msg.Size) )); err != nil {
		return err
	}
	peer.(*p2p.TCPPeer).Wg.Done()
	return nil
	
}

func (s *FileServer) loop() {
	defer func(){
		log.Println("file server stopped")
		s.Transport.Close()
	}()

	for {
		select{
		case rpc := <- s.Transport.Consume():
			var msg Message
			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&msg); err != nil {
				log.Println(err)
				return
			}
			if err := s.HandleMessage(rpc.From, &msg); err != nil {
				log.Println(err)
				return
			}
		
		case <- s.quitchan:
			return 

		}
	}
}
func (s *FileServer) Broadcast(msg *Message) error {
	peers := []io.Writer{}
	for _, peer := range s.peers {
		peers = append(peers, peer)
	}
	mw := io.MultiWriter(peers...)
	return gob.NewEncoder(mw).Encode(msg)
}	

type Message struct {
	Payload any
}


type MessageStoreFile struct {
	Key string
	Size int64

}

func (s *FileServer) StoreData(key string, r io.Reader) error {
	// store this file in the disk
	// broadcast file to all known peers in the network


	buf := new(bytes.Buffer)
	tee := io.TeeReader(r, buf)
	// // store it to our own disk
	size, err := s.store.Write(key, tee);
	if err != nil {
		return err
	}

	// create a message
	msg := Message{
		Payload: MessageStoreFile{
			Key: key,
			Size: size,
		},
	}
	fmt.Printf("Payload : %v", msg)
	msgBuffer := new(bytes.Buffer)
	if err := gob.NewEncoder(msgBuffer).Encode(msg); err != nil {
		return err
	}
	for _, peer := range s.peers {
		if err := peer.Send(msgBuffer.Bytes()); err != nil {
			return err
		}
	}

	time.Sleep(3 * time.Second)
	
	
	for _, peer := range s.peers {
		n, err := io.Copy(peer, buf)
		if err != nil {
			return err
		}
		fmt.Println("received and written bytes to disk: ", n)
	}

	
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
			fmt.Println("[%s] attemping to connect with remote %s\n", node_address, node_address)
			if err := s.Transport.Dial(node_address); err != nil {
				log.Println("error dialing node", node_address, err)
			}
		}(node_address)
	}
	return nil


}


func (s *FileServer) OnPeer(p p2p.Peer) error {
	fmt.Println("connected to remote peer %s", p.RemoteAddr().String())
    s.peerLock.Lock()
    defer s.peerLock.Unlock()
    s.peers[p.RemoteAddr().String()] = p
    fmt.Println("connected to remote peer %s", p.RemoteAddr().String())
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


func init() {
	gob.Register(MessageStoreFile{})

}