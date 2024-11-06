package p2p



// Peer is an interface that represents the remote connected node
type Peer interface{
	Close() error


}

// Transport is anything that handles the communication between nodes in the network
// This can be TCP / UDP / Web sockets and so on
type Transport interface{
	ListenAndAccept() error
	Consume() <-chan RPC
}