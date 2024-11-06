package main

import (
	"flag"
	"sync"
)

var bufferSize int

// Create a pool of reusable buffers
var bufferPool = sync.Pool{
	New: func() interface{} {
		b := make([]byte, bufferSize)
		return &b // Return pointer to buffer
	},
}

func main() {
	localAddrPtr := flag.String("listen", ":9000", "local address")
	remoteAddrPtr := flag.String("remote", "127.0.0.1:9001", "remote address")
	enableTcp := flag.Bool("tcp", false, "enable tcp listener")
	enableUdp := flag.Bool("udp", false, "enable udp listener")
	bufferSizePtr := flag.Int("size", 32*1024, "buffer size")
	flag.Parse()

	localAddr := *localAddrPtr
	remoteAddr := *remoteAddrPtr
	bufferSize = *bufferSizePtr

	wg := sync.WaitGroup{}
	if *enableTcp {
		wg.Add(1)
		go tcpForwarder(&wg, localAddr, remoteAddr)
	}
	if *enableUdp {
		wg.Add(1)
		go udpForwarder(&wg, localAddr, remoteAddr)
	}
	wg.Wait()
}
