package main

import (
	"log"
	"net"
	"sync"
)

func udpSinglePortForwarder(wg *sync.WaitGroup, localAddr, remoteAddr string) {
	defer wg.Done()
	listener, err := net.ListenPacket("udp", localAddr)
	if err != nil {
		log.Fatalf("Failed to start listener: %v", err)
	}
	defer listener.Close()

	remoteUDPAddr, err := net.ResolveUDPAddr("udp", remoteAddr)
	if err != nil {
		log.Fatalf("Failed to resolve UDP address: %v", err)
	}

	log.Printf("Forwarding UDP from %s to %s", localAddr, remoteAddr)

	var reqAddr *net.UDPAddr

	for {
		bufferPtr := bufferPool.Get().(*[]byte)
		buffer := *bufferPtr
		readN, addr, err := listener.ReadFrom(buffer)
		if err != nil {
			log.Printf("Failed to read from listener: %v", err)
			continue
		}
		udpAddr, err := net.ResolveUDPAddr(addr.Network(), addr.String())
		if err != nil {
			log.Printf("Failed to resolve UDP address: %v", err)
			continue
		}
		if udpAddr.String() == remoteAddr {
			if reqAddr == nil {
				log.Printf("Response came before request")
				continue
			}
			go send(listener, reqAddr, bufferPtr, readN)
		} else {
			reqAddr = udpAddr
			go send(listener, remoteUDPAddr, bufferPtr, readN)
		}
	}
}

func send(conn net.PacketConn, remoteUDPAddr net.Addr, bufferPtr *[]byte, n int) {
	defer bufferPool.Put(bufferPtr)
	buffer := *bufferPtr
	_, err := conn.WriteTo(buffer[:n], remoteUDPAddr)
	if err != nil {
		log.Printf("Failed to write to connection: %v", err)
	}
}
