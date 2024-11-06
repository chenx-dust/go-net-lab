package main

import (
	"log"
	"net"
	"sync"
)

func udpForwarder(wg *sync.WaitGroup, localAddr, remoteAddr string) {
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

	udpAddrConnMap := make(map[string]*net.UDPConn)
	for {
		bufferPtr := bufferPool.Get().(*[]byte)
		buffer := *bufferPtr
		n, addr, err := listener.ReadFrom(buffer)
		if err != nil {
			log.Fatalf("Failed to read from UDP listener: %v", err)
		}

		udpConn, ok := udpAddrConnMap[addr.String()]
		if !ok {
			udpConn, err = net.DialUDP("udp", nil, remoteUDPAddr)
			if err != nil {
				log.Printf("Failed to dial UDP: %v", err)
				continue
			}
			udpConn.SetReadBuffer(bufferSize)
			udpConn.SetWriteBuffer(bufferSize)
			udpAddrConnMap[addr.String()] = udpConn
			go handleUDPConnection(udpConn, listener, addr)
		}

		go func() {
			defer bufferPool.Put(bufferPtr)
			_, err = udpConn.Write(buffer[:n])
			if err != nil {
				log.Printf("Failed to write to UDP: %v", err)
			}
		}()
	}
}

func handleUDPConnection(conn *net.UDPConn, listener net.PacketConn, destAddr net.Addr) {
	for {
		bufferPtr := bufferPool.Get().(*[]byte)
		buffer := *bufferPtr
		n, err := conn.Read(buffer)
		if err != nil {
			log.Printf("Failed to read from UDP: %v", err)
			continue
		}

		go func() {
			defer bufferPool.Put(bufferPtr)
			_, err = listener.WriteTo(buffer[:n], destAddr)
			if err != nil {
				log.Printf("Failed to write to UDP: %v", err)
			}
		}()
	}
}
