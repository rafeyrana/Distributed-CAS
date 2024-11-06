package p2p



// Message represents any arbitrary data that is sent between two nodes
type Message struct {
	Payload []byte
}