package p2p

type HandshakeFunc func(any) error


func NOPHandshakeFunc(Peer) error { return nil } // NOP function does nothing
