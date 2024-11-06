package p2p

import "net"


// Message represents any arbitrary data that is sent between two nodes
type Message struct {
	Payload []byte
	From net.Addr

}