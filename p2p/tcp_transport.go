package p2p
import (
	"net"
	"fmt"
	"errors"
	"log"
	"sync"
)


// represents the remote node over a TCp established connecion
type TCPPeer struct {
	// the underlying connection of the peer which is the tcp connection in this case
	 net.Conn
	outbound bool // outbound peer if we are the one who initiated the connection (true) but if we accept it is an inbound peer
	Wg *sync.WaitGroup // used to wait for all goroutines to finish
}




type TCPTransportOpts struct {

	ListenAddress string
	HandShakeFunc HandShakeFunc
	Decoder Decoder
	OnPeer func(Peer) error


}



func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		Conn :conn,
		outbound: outbound,
		Wg: &sync.WaitGroup{},
	}
}


func (p *TCPPeer) Send(msg []byte) error {
	_, err := p.Conn.Write(msg)
	return err
}

// Consume is the implmenetation for the transport interface which will return a read only channel for reading the incoming messages
func (t *TCPTransport) Consume() <-chan RPC {
	return t.rpcch
}

type TCPTransport struct {
	TCPTransportOpts
	listener net.Listener
	rpcch chan RPC
}


func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {
	return &TCPTransport{
		TCPTransportOpts: opts,
		rpcch: make(chan RPC),
	}
}

func (t *TCPTransport) Close() error {
	return t.listener.Close()
}




// Impelementation of the transport interface
func (t *TCPTransport) Dial(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	go t.handleConn(conn, true)
	return nil
}


func (t *TCPTransport) ListenAndAccept() error {
	var err error
	t.listener, err = net.Listen("tcp", t.ListenAddress)
	if err != nil {
		return err
	}
	go t.startAcceptLoop()
	log.Printf("TCP transport listening on port: %s\n", t.ListenAddress)
	return nil
}

func (t *TCPTransport) startAcceptLoop() {
	for {
		conn, err := t.listener.Accept()
		if errors.Is(err, net.ErrClosed) {
			return 
		}
		if errors.Is(err, net.ErrClosed) {
			return
		}
		if err != nil {
			fmt.Printf("TCP accept error: %s\n", err)
		}
		fmt.Printf("new incoming connection: %s\n", conn)
		go t.handleConn(conn, false)
	}
}



func (t *TCPTransport) handleConn(conn net.Conn, outbound bool) {
	var err error
	defer func(){
		fmt.Printf("dropping peer connection: %s\n", err)
		conn.Close()

		}()
	peer := NewTCPPeer(conn, outbound)


	if err:= t.HandShakeFunc(peer); err != nil {
		fmt.Printf("failed handshake with peer: %s\n", err)

		// here we should close the connection if the handshake fails
		fmt.Println("closed connection because of handshake failiure")
		conn.Close()
		return 
	}

	if t.OnPeer != nil {
		if err := t.OnPeer(peer); err != nil {
			return 
		}
	}

	rpc := RPC{}
	// Read Loop
	for {
		err := t.Decoder.Decode(conn, &rpc)
		if errors.Is(err, net.ErrClosed) {
			fmt.Printf(" peer closed connection: %s\n", err)
			return
		}
		
		if err != nil {
			fmt.Printf("tcp error in decoding : failed to read from peer: %s\n", err) 
			continue // this is what was causing the infinite loop but how do we deal with any decode errors, we can choose to drop the connection as well so we just seperate the logic and implement checks seperately for dropped connection from peer
		}


		rpc.From = conn.RemoteAddr().String()
		peer.Wg.Add(1)
		fmt.Println("Waiting till stream is done...")
		t.rpcch <- rpc
		peer.Wg.Wait()
		fmt.Println("Continuing...")
		// fmt.Printf("message received from peer: %+v\n", rpc)
	
	}


	// fmt.Println("new incoming connection:", conn)
	// peer := NewTCPPeer(conn)
	// t.addPeer(peer)
	// return nil
}