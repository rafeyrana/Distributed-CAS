package main

import (
	"log"
    "bytes"
    "time"
	"github.com/rafeyrana/Distributed-CAS/p2p"
)


func makeServer(listenAddr string, nodes ...string) *FileServer {
    tcpP2pTransportOpts := p2p.TCPTransportOpts{
		ListenAddress: listenAddr,
		HandShakeFunc: p2p.NOPHandShakeFunc,
		Decoder: p2p.DefaultDecoder{},
        // TOOD: implment the on peer functions
	}
    tcpTransport := p2p.NewTCPTransport(tcpP2pTransportOpts)
    fileServerOpts := FileServerOpts{
		StorageRoot: listenAddr + "_network", // for multiple roots for different networks
		PathTransformFunc: CASPathTransformFunc,
        Transport: tcpTransport,
        BootstrapNodes: nodes,
	}
    s := NewFileServer(fileServerOpts)
    
    tcpTransport.OnPeer = s.OnPeer
    return s
    
}
func main() {

   s1 := makeServer(":3000", "")


   go func(){
    log.Fatal(s1.Start())
    }()
    s2 := makeServer(":4000", ":3000")
    
    go s2.Start()
    time.Sleep(1 * time.Second)

    data := bytes.NewReader([]byte("hello world this si my big data"))
    s2.StoreData("myprivdata", data)
    select {}


}