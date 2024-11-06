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


type TCPTransport struct {
	TCPTransportOpts
	listener net.Listener
	mu sync.RWMutex // protects peers
	peers map[net.Addr]Peer

}


func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {
	return &TCPTransport{
		TCPTransportOpts: opts,
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
		fmt.Println("new incoming connection: %s\n", conn)
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

	msg := &Message{}
	// Read Loop
	for {
		if err := t.Decoder.Decode(conn, msg); err!= nil {
			fmt.Printf("tcp error in decoding : failed to read from peer: %s\n", err)
			continue
		}
	}
	fmt.Printf("message received from peer: %s\n", msg)


	fmt.Println("new incoming connection:", conn)
	// peer := NewTCPPeer(conn)
	// t.addPeer(peer)
	// return nil
}