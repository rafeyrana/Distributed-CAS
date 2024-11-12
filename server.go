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
    case MessageGetFile:
		return s.handleMessageGetFile(from, v)
	}


	return nil
}


func (s *FileServer) handleMessageGetFile(from string, msg MessageGetFile) error {
	fmt.Printf("[%s] serving file over the network : %s", s.Transport.Addr(),msg.Key)
	if s.store.HasKey(msg.Key) {
		reader, err := s.store.Read(msg.Key)
		if err != nil {
			return err
		}
		peer , ok := s.peers[from]
		if !ok {
			return fmt.Errorf("peer  %s not found", peer)
		}


		peer.Send([]byte{p2p.IncomingStream})
		n , err := io.Copy(peer, reader)
		if err != nil {
			return err
		}
		fmt.Printf("\n [%s] written %d bytes over the network to  %s",s.Transport.Addr(), n ,from)



		return nil
	} else {
		return fmt.Errorf("[%s] Need to serve file : %s but File not found: ",s.Transport.Addr(), msg.Key)
	}

	return nil



}
func (s *FileServer) handleMessageStoreFile(from string, msg MessageStoreFile) error {

	peer , ok := s.peers[from]

	if !ok {
		fmt.Errorf("peer not found")
		panic("peer not found")
	}
	n, err := s.store.Write(msg.Key, io.LimitReader(peer,int64(msg.Size) ))
	if err != nil {
		return err
	}
	fmt.Println("wrote %d bytes to disk on address : [%s]", n, s.Transport.Addr())
	//peer.(*p2p.TCPPeer).Wg.Done()
	peer.CloseStream()
	return nil
	
}

func (s *FileServer) loop() {
	defer func(){
		log.Println("file server stopped due to error or user quit action")
		s.Transport.Close()
	}()

	for {
		select{
		case rpc := <- s.Transport.Consume():
			var msg Message
			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&msg); err != nil {
				log.Println("decoding error: ",err)
				return
			}
			if err := s.HandleMessage(rpc.From, &msg); err != nil {
				log.Println("handle message",err)
			}
		
		case <- s.quitchan:
			return 

		}
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
	msgBuffer := new(bytes.Buffer)
	if err := gob.NewEncoder(msgBuffer).Encode(msg); err != nil {
		return err
	}
	for _, peer := range s.peers {
		peer.Send([]byte{p2p.IncomingMessage})
		if err := peer.Send(msgBuffer.Bytes()); err != nil {
			return err
		}
	}
	return nil
}
type Message struct {
	Payload any
}


type MessageStoreFile struct {
	Key string
	Size int64

}


type MessageGetFile struct {
	Key string
}

func (s *FileServer) Get(key string) (io.Reader, error) {
	if s.store.HasKey(key) {
		fmt.Printf("[%s] serving file (%s) locally \n", s.Transport.Addr(), key)
		return s.store.Read(key)
	}

	fmt.Println("[%s] did not find file (%s) locally. Serving via network.......\n", s.Transport.Addr(), key)
	msg := Message{
		Payload: MessageGetFile{
			Key: key,
		},
	}
	if err := s.broadcast(&msg); err != nil {
		return nil , err
	}
	

	time.Sleep(500 * time.Millisecond)

	// now have to open up a stream and read from every peer
	for _, peer := range s.peers {
		fileBuffer := new(bytes.Buffer)
		n, err := io.CopyN(fileBuffer, peer, 15)
		
		if err != nil {
			return nil, err
		}
		fmt.Printf(" [%s] recieved bytes (%d) over the network from:  %s", s.Transport.Addr(), n, peer.RemoteAddr())
		fmt.Println(fileBuffer.String())

		peer.CloseStream()
	}


	select{}
	return nil, fmt.Errorf("key not found: %s", key)
}

func (s *FileServer) Store(key string, r io.Reader) error {
	// store this file in the disk
	// broadcast file to all known peers in the network
	var (
	fileBuf = new(bytes.Buffer)
	tee = io.TeeReader(r, fileBuf)
	)
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
	// send message to all peers
	if err:= s.broadcast(&msg); err != nil {
		return err
	}


	time.Sleep(1 * time.Millisecond)
	

	// todo : use a multiwriter here
	for _, peer := range s.peers {
		peer.Send([]byte{p2p.IncomingStream})
		n, err := io.Copy(peer, fileBuf)
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
	gob.Register(MessageGetFile{})

}