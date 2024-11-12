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
	wg *sync.WaitGroup // used to wait for all goroutines to finish
}


// implements the transport interface returning the address the transport is accpeting connecitons
func (t *TCPTransport) Addr() string {
	return t.ListenAddress
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
		wg: &sync.WaitGroup{},
	}
}
func (p *TCPPeer) CloseStream() {
	p.wg.Done()
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
		rpcch: make(chan RPC, 1024),
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


	// Read Loop
	for {
		rpc := RPC{}
		err := t.Decoder.Decode(conn, &rpc)
		if errors.Is(err, net.ErrClosed) {
			fmt.Printf(" peer closed connection: %s\n", err)
			return
		}
		rpc.From = conn.RemoteAddr().String()
		if rpc.Stream {
            fmt.Println("Received stream message from ", rpc.From)
            peer.wg.Add(1)
			fmt.Printf("[%s] incoming stream ..... \n", conn.RemoteAddr().String())
			peer.wg.Wait()
			fmt.Printf("Stream closed. Resuming with read loop \n")
			continue
        }
		t.rpcch <- rpc

		fmt.Println("Continuing...")
	
	}
}