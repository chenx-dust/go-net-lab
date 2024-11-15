package server

import (
	"log"
	"net"

	"github.com/chenx-dust/go-net-lab/paracat/packet"
)

func (server *Server) forward(buf []byte, connID uint16) {
	server.forwardMutex.RLock()
	conn, ok := server.forwardConns[connID]
	server.forwardMutex.RUnlock()
	if !ok {
		server.forwardMutex.Lock()
		conn, ok = server.forwardConns[connID]
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
	}

	server.packetStat.Reverse.CountPacket(uint32(len(buf)))

	n, err := conn.Write(buf)
	if err != nil {
		log.Println("error writing to udp:", err)
	}
	if n != len(buf) {
		log.Println("error writing to udp: wrote", n, "bytes instead of", len(buf))
	}
}

func (server *Server) handleReverse(conn *net.UDPConn, connID uint16) {
	for {
		buf := make([]byte, server.cfg.BufferSize)
		n, err := conn.Read(buf)
		if err != nil {
			log.Fatalln("error reading from udp:", err)
		}
		server.packetStat.Forward.CountPacket(uint32(n))

		go func() {
			packetID := server.packetFilter.NewPacketID()

			server.sourceMutex.RLock()
			for _, sourceConn := range server.sourceTCPConns {
				go func() {
					n, err := packet.WritePacket(sourceConn, buf[:n], connID, packetID)
					if err != nil {
						log.Println("error writing to tcp:", err)
					}
					if n != len(buf) {
						log.Println("error writing to tcp: wrote", n, "bytes instead of", len(buf))
					}
				}()
			}
			udpPacked := packet.Pack(buf[:n], connID, packetID)
			for sourceStr := range server.sourceUDPAddrs {
				go func() {
					sourceAddr, err := net.ResolveUDPAddr("udp", sourceStr)
					if err != nil {
						log.Fatalln("error resolving source addr:", err)
					}
					n, err := server.udpListener.WriteToUDP(udpPacked, sourceAddr)
					if err != nil {
						log.Println("error writing to udp:", err)
					}
					if n != len(udpPacked) {
						log.Println("error writing to udp: wrote", n, "bytes instead of", len(udpPacked))
					}
				}()
			}
			server.sourceMutex.RUnlock()
		}()
	}
}
