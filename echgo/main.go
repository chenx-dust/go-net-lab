package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
)

func main() {
	port := flag.Int("port", 9000, "port to listen on")
	flag.Parse()

	tcpListener, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("Failed to listen tcp port %d: %v", *port, err)
	}
	defer tcpListener.Close()

	udpListener, err := net.ListenUDP("udp", &net.UDPAddr{Port: *port})
	if err != nil {
		log.Fatalf("Failed to listen udp port %d: %v", *port, err)
	}
	defer udpListener.Close()

	log.Printf("Listening tcp and udp on port %d\n", *port)

	wg := sync.WaitGroup{}
	wg.Add(2)

	go handleTcp(tcpListener, &wg)
	go handleUdp(udpListener, &wg)

	wg.Wait()
}

func handleTcp(listener net.Listener, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept tcp connection: %v", err)
			continue
		}

		go func() {
			defer conn.Close()
			_, err := io.Copy(conn, conn)
			if err != nil {
				log.Printf("Failed to echo tcp connection: %v", err)
			}
		}()
	}
}

func handleUdp(listener *net.UDPConn, wg *sync.WaitGroup) {
	defer wg.Done()

	buf := make([]byte, 9000)
	for {
		n, addr, err := listener.ReadFromUDP(buf)
		if err != nil {
			log.Printf("Failed to read from udp connection: %v", err)
			continue
		}

		go func() {
			_, err := listener.WriteToUDP(buf[:n], addr)
			if err != nil {
				log.Printf("Failed to echo udp connection: %v", err)
			}
		}()
	}
}
