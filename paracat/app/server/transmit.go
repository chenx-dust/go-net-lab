package server

import (
	"log"
	"net"
	"sync"

	"github.com/chenx-dust/go-net-lab/paracat/packet"
)

func (server *Server) forward(buf []byte, connID uint16) {
	server.forwardMutex.Lock()
	conn, ok := server.forwardConns[connID]
	if !ok {
		remoteAddr, err := net.ResolveUDPAddr("udp", server.cfg.RemoteAddr)
		if err != nil {
			log.Fatalln("error resolving remote addr:", err)
		}
		conn, err = net.DialUDP("udp", nil, remoteAddr)
		if err != nil {
			log.Fatalln("error dialing relay:", err)
		}
		server.forwardConns[connID] = conn
		go server.handleReverse(conn, connID)
	}
	server.forwardMutex.Unlock()

	conn.Write(buf)
}

func (server *Server) handleReverse(conn *net.UDPConn, connID uint16) {
	for {
		buf := make([]byte, server.cfg.BufferSize)
		n, err := conn.Read(buf)
		if err != nil {
			log.Fatalln("error reading from udp:", err)
		}

		go func() {
			packetID := server.packetFilter.NewPacketID()

			wg := sync.WaitGroup{}
			server.sourceMutex.RLock()
			for _, sourceConn := range server.sourceTCPConns {
				wg.Add(1)
				go func() {
					defer wg.Done()
					packet.WritePacket(sourceConn, buf[:n], connID, packetID)
				}()
			}
			udpPacked := packet.Pack(buf[:n], connID, packetID)
			for _, sourceConn := range server.sourceUDPAddrs {
				wg.Add(1)
				go func() {
					defer wg.Done()
					conn.WriteToUDP(udpPacked, sourceConn)
				}()
			}
			server.sourceMutex.RUnlock()
			wg.Wait()
		}()
	}
}
