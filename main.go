package main

import (
	"log"
	"time"

	"github.com/rafeyrana/Distributed-CAS/p2p"
)
func main() {
    tcpP2pTransportOpts := p2p.TCPTransportOpts{
		ListenAddress: ":3000",
		HandShakeFunc: p2p.NOPHandShakeFunc,
		Decoder: p2p.DefaultDecoder{},
        // TOOD: implment the on peer functions
	}
    tcpTransport := p2p.NewTCPTransport(tcpP2pTransportOpts)

   s := NewFileServer(FileServerOpts{
		StorageRoot: "3000_files", // for multiple roots for different networks
		PathTransformFunc: CASPathTransformFunc,
        Transport: tcpTransport,
	})
    go func(){
        time.Sleep(time.Second * 3)
        s.Stop()
    }()

    if err := s.Start(); err != nil {
        log.Fatal(err)
    }


}