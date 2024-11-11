package main

import (
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"
	"time"
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
// func (s *FileServer) HandleMessage(msg *Message) error {
// 	switch v := msg.Payload.(type){
// 	case *DataMessage:
// 		fmt.Println("received data message: %s\n", v.Data)
// 	}
// 	return nil
// }


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
			}


			fmt.Printf("\n received message in the loop: %s \n", string(msg.Payload.([]byte)))
			peer, ok := s.peers[rpc.From]
			if !ok {
				panic("peer not found")
			}
			fmt.Printf("%v", peer)
			b := make([]byte, 1000)
			if _, err:= peer.Read(b); err != nil {
				panic(err)
			}
			
			fmt.Printf("\n received data in the loop: %s", string(b))

			// if err := s.HandleMessage(&m); err != nil {
			// 	log.Println(err) 
				
			// }
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

func (s *FileServer) StoreData(key string, r io.Reader) error {
	// store this file in the disk
	// broadcast file to all known peers in the network
	// create a message
	msg := Message{
		Payload: []byte("strage key"),
	}
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		return err
	}
	for _, peer := range s.peers {
		if err := peer.Send(buf.Bytes()); err != nil {
			return err
		}
	}
	time.Sleep(3 * time.Second)
	payload := []byte("THIS IS A LARGE FILEEEEEEE")	
	for _, peer := range s.peers {
		if err := peer.Send(payload); err != nil {
			return err
		}
	}
	return nil
	
	// buf := new(bytes.Buffer)
	// tee := io.TeeReader(r, buf)
	// // store it to our own disk
	// if err := s.store.Write(key, tee); err != nil {
	// 	return err
	// }
	// // now the reader is empty because it has been read
   
	// p := &DataMessage{
	// 	Key: key,
	// 	Data: buf.Bytes(),
	// }
	// fmt.Println("%s", buf.Bytes())


	// return s.Broadcast(&Message{
	// 	From: "todo",
	// 	Payload: p,
	// })
	
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