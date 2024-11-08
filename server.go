package main

import (
	"fmt"
	"log"

	"github.com/rafeyrana/Distributed-CAS/p2p"
)

type FileServerOpts struct {
	ListenAddress    string
	StorageRoot 	 string
	PathTransformFunc PathTransformFunc
	Transport p2p.Transport
}
type FileServer struct {
	FileServerOpts 
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


func (s *FileServer) Stop(){
	close(s.quitchan)
}


func (s *FileServer) Start() error {
	if err:= s.Transport.ListenAndAccept(); err != nil {
		return err
	}
	// can we block and start using go routine or not block or execute directly
	s.loop()
	return nil
}