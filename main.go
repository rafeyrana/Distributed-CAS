package main

import (
	"fmt"
	"github.com/rafeyrana/Distributed-CAS/p2p"
	"log"

)

func main(){
	tr := p2p.NewTCPTransport(":3000")


	if err := tr.ListenAndAccept(); err != nil{
		log.Fatal(err)
	}


	select {}
	fmt.Println("Hello, World!")
}