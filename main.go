package main

import (
	"fmt"
	"log"

	"github.com/rafeyrana/Distributed-CAS/p2p"
)


func OnPeer(peer p2p.Peer) error {
    peer.Close() // this will close the peer from the peer itself but will send the TCP transport into a decode infinite loop because it wont have closed on this end

    fmt.Printf("Logic Check")
    return nil
}
func main() {
    tcpOpts := p2p.TCPTransportOpts{
        ListenAddress: ":3000",
        HandShakeFunc: p2p.NOPHandShakeFunc,
        Decoder: &p2p.DefaultDecoder{},
        OnPeer:    OnPeer,
    }
    tr := p2p.NewTCPTransport(tcpOpts)


    go func(){
		for {
			msg := <- tr.Consume()
			fmt.Println("%+v\n", msg)
		}
	}()

    if err := tr.ListenAndAccept(); err != nil {
        log.Fatal(err)
    }

    fmt.Printf("TCP transport listening on %s\n", tcpOpts.ListenAddress)
    select {} // Block forever
}