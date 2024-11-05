package p2p
import (
	"net"
)
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
	t.listener , err := net.Listen("tcp", t.listenAddress)
	if err != nil {
		return err
	}

}