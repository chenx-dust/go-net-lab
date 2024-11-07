package server

import (
	"log"
	"net"

	"github.com/chenx-dust/go-net-lab/paracat/packet"
)

func (server *Server) handleTCP() {
	for {
		conn, err := server.tcpListener.AcceptTCP()
		if err != nil {
			log.Fatalln("error accepting tcp connection:", err)
		}
		server.sourceMutex.Lock()
		server.sourceTCPConns = append(server.sourceTCPConns, conn)
		server.sourceMutex.Unlock()
		go server.handleTCPConn(conn)
	}
}

func (server *Server) handleTCPConn(conn *net.TCPConn) {
	defer conn.Close()
	for {
		buf := make([]byte, server.cfg.BufferSize)
		n, connID, packetID, err := packet.ReadPacket(conn, buf)
		if err != nil {
			log.Fatalln("error reading packet:", err)
		}

		isDuplicate := server.packetFilter.CheckDuplicatePacketID(packetID)
		if isDuplicate {
			continue
		}

		go func() {
			server.forward(buf[:n], connID)
		}()
	}
}
