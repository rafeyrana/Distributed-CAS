package p2p
import (
	"net"
	"sync"
	"fmt"
	"errors"
)


// represents the remote node over a TCp established connecion
type TCPPeer struct {
	conn net.Conn
	outbound bool // outbound peer if we are the one who initiated the connection (true) but if we accept it is an inbound peer
}




type TCPTransportOpts struct {

	ListenAddress string
	HandShakeFunc HandShakeFunc
	Decoder Decoder


}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		conn: conn,
		outbound: outbound,
	}
}


// peer interface implementation
func (p *TCPPeer) Close() error {
	return p.conn.Close()
}
// Consume is the implmenetation for the transport interface which will return a read only channel for reading the incoming messages
func (t *TCPTransport) Consume() <-chan RPC {
	return t.rpcch
}

type TCPTransport struct {
	TCPTransportOpts
	listener net.Listener

	rpcch chan RPC
	mu sync.RWMutex // protects peers
	peers map[net.Addr]Peer

}


func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {
	return &TCPTransport{
		TCPTransportOpts: opts,
		rpcch: make(chan RPC),
	}
}



func (t *TCPTransport) ListenAndAccept()  error{
	var err error
	t.listener , err = net.Listen("tcp", t.ListenAddress)
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
		fmt.Printf("new incoming connection: %s\n", conn)
		go t.handleConn(conn)
	}
}



func (t *TCPTransport) handleConn(conn net.Conn) {
	peer := NewTCPPeer(conn, true)
	if err:= t.HandShakeFunc(peer); err != nil {
		fmt.Printf("failed handshake with peer: %s\n", err)

		// here we should close the connection if the handshake fails
		fmt.Println("closed connection because of handshake failiure")
		conn.Close()
		return 
	}

	rpc := RPC{}
	// Read Loop
	for {
		if err := t.Decoder.Decode(conn, &rpc); err!= nil {
			fmt.Printf("tcp error in decoding : failed to read from peer: %s\n", err)
			continue
		}


		rpc.From = conn.RemoteAddr()
		t.rpcch <- rpc
		// fmt.Printf("message received from peer: %+v\n", rpc)
	
	}


	// fmt.Println("new incoming connection:", conn)
	// peer := NewTCPPeer(conn)
	// t.addPeer(peer)
	// return nil
}