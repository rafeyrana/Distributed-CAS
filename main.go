package main

import (
    "bytes"

	"log"
	"time"

	"github.com/rafeyrana/Distributed-CAS/p2p"
)


func makeServer(listenAddr string, nodes ...string) *FileServer {
    tcpP2pTransportOpts := p2p.TCPTransportOpts{
		ListenAddress: listenAddr,
		HandShakeFunc: p2p.NOPHandShakeFunc,
		Decoder: p2p.DefaultDecoder{},
	}
    tcpTransport := p2p.NewTCPTransport(tcpP2pTransportOpts)
    fileServerOpts := FileServerOpts{
        EncKey: newEncryptionKey(),
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
    time.Sleep(1 * time.Second)
    go s2.Start()
    time.Sleep(1 * time.Second)
    data := bytes.NewReader([]byte("THIS LARGE FILE"))
    s2.Store("coolpicture.jpg", data)
    time.Sleep(1 * time.Second)

  
    // data := bytes.NewReader([]byte("THIS LARGE FILE"))pserver
    
    // s2.Store("myprivdata", data)

    // r, err := s2.Get("coolpicture.jpg")
    // if err != nil {
    //     log.Fatal(err)
    // }

    // b , err := ioutil.ReadAll(r)
    // if err != nil {
    //     log.Fatal(err)
    // }
    // fmt.Println(string(b))

    


}