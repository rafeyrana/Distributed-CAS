package p2p


import {
	"testing"

	"github.com/stretchr/testify/assert"}
func TestTCPTransport(t *testing.T) {
	tr:= NewTCPTransport("127.0.0.1:0")
	assert.Equal(t, "127.0.0.1:0", tr.listenAddress)




	// for example if this is a web server
}