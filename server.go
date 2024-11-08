package main
import (
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
}

func NewFileServer(opts FileServerOpts) *FileServer {
	storeOpts := StoreOpts{
		Root: opts.StorageRoot,
		PathTransformFunc: opts.PathTransformFunc,
	}
	return &FileServer{
		FileServerOpts: opts,
		store : NewStore(storeOpts),
	}
}


func (s *FileServerOpts) Start() error {
	if err:= s.Transport.ListenAndAccept(); err != nil {
		return err
	}
	return nil

}