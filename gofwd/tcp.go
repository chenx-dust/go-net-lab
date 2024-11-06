package main

import (
	"log"
	"net"
	"sync"
)

func tcpForwarder(wg *sync.WaitGroup, localAddr, remoteAddr string) {
	defer wg.Done()
	listener, err := net.Listen("tcp", localAddr)
	if err != nil {
		log.Fatalf("Failed to start listener: %v", err)
	}
	defer listener.Close()

	log.Printf("Forwarding TCP from %s to %s", localAddr, remoteAddr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}
		go handleTCPConnection(conn, remoteAddr)
	}
}

func handleTCPConnection(local net.Conn, remoteAddr string) {
	defer local.Close()

	remote, err := net.Dial("tcp", remoteAddr)
	if err != nil {
		log.Printf("Failed to connect to remote address: %v", err)
		return
	}
	defer remote.Close()

	var wg sync.WaitGroup
	wg.Add(2)

	// Forward in both directions
	go func() {
		defer wg.Done()
		forward(local, remote)
	}()

	go func() {
		defer wg.Done()
		forward(remote, local)
	}()

	wg.Wait()
}

func forward(dst, src net.Conn) {
	// Get a buffer from the pool
	buffer_ptr := bufferPool.Get().(*[]byte)
	// Return the buffer to the pool when done
	defer bufferPool.Put(buffer_ptr)
	buffer := *buffer_ptr

	for {
		n, err := src.Read(buffer)
		if err != nil {
			return
		}
		if n > 0 {
			_, err := dst.Write(buffer[:n])
			if err != nil {
				return
			}
		}
	}
}
