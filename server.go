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
	BootstrapNodes []string
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



func (s *FileServer)BootstrapNetwork() error{
	for _, node_address := range s.BootstrapNodes {
		go func (node_address string) {
			if err := s.Transport.Dial(node_address); err != nil {
				log.Println("error dialing node", node_address, err)
			}
		}(node_address)
	}
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