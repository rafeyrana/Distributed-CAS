package main

import (
    "fmt"
    "log"
    "github.com/rafeyrana/Distributed-CAS/p2p"
)

func main() {
    tcpOpts := p2p.TCPTransportOpts{
        ListenAddress: ":3000",
        HandShakeFunc: p2p.NOPHandShakeFunc,
        Decoder: &p2p.DefaultDecoder{},
    }
    tr := p2p.NewTCPTransport(tcpOpts)

    if err := tr.ListenAndAccept(); err != nil {
        log.Fatal(err)
    }

    fmt.Printf("TCP transport listening on %s\n", tcpOpts.ListenAddress)
    select {} // Block forever
}