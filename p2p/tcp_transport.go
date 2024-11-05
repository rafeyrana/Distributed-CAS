package p2p
import (
	"net"
	"sync"
	"fmt"
)

// represents the remor node over a TCp established connecion
type TCPPeer struct {
	conn net.Conn
	outbound bool // outbound peer if we are the one who initiated the connection (true) but if we accept it is an inbound peer
}


func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		conn: conn,
		outbound: outbound,
	}
}


type TCPTransport struct {
	TCPTransportOpts
	listener net.Listener
	rpcch    chan RPC
}
type TCPTransport struct {
	listenAddress string
	listener net.Listener

	mu sync.RWMutex // protects peers
	peers map[net.Addr]Peer

}





func NewTCPTransport(listenAddress string) *TCPTransport {
	return &TCPTransport{
		listenAddress: listenAddress,
		peers: make(map[net.Addr]Peer),
	}
}



func (t *TCPTransport) ListenAndAccept()  error{
	var err error
	t.listener , err = net.Listen("tcp", t.listenAddress)
	if err != nil {
		return err
	}
	go t.startAcceptLoop()

	return nil


}

func (t *TCPTransport) startAcceptLoop() {
	for {
		conn, err := t.listener.Accept()
		if errors.Is(err, net.ErrClosed) {
			return
		}
		if err != nil {
			fmt.Printf("TCP accept error: %s\n", err)
		}

		go t.handleConn(conn)
	}
}


func (t *TCPTransport) handleConn(conn net.Conn) {
	fmt.Println("new incoming connection:", conn)
	// peer := NewTCPPeer(conn)
	// t.addPeer(peer)
	// return nil
}