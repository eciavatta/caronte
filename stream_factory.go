package main

import (
	"github.com/google/gopacket"
	"github.com/google/gopacket/tcpassembly"
)

type BiDirectionalStreamFactory struct {
	storage *Storage
	serverIp string
}

// httpStream will handle the actual decoding of http requests.
type uniDirectionalStream struct {
	net, transport gopacket.Flow
	r              StreamHandler
}

func (h *BiDirectionalStreamFactory) New(net, transport gopacket.Flow) tcpassembly.Stream {
	hstream := &uniDirectionalStream{
		net:       net,
		transport: transport,
		r:         NewStreamHandler(),
	}
	// go hstream.run() // Important... we must guarantee that data from the tcpreader stream is read.

	// StreamHandler implements tcpassembly.Stream, so we can return a pointer to it.
	return &hstream.r
}
