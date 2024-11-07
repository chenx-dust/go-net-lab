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
			log.Println("error accepting tcp connection:", err)
			continue
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
		buf := server.bufferPool.Get().([]byte)
		n, connID, packetID, err := packet.ReadPacket(conn, buf)
		if err != nil {
			log.Println("error reading packet:", err)
			continue
		}

		isDuplicate := server.packetFilter.CheckDuplicatePacketID(packetID)
		if isDuplicate {
			continue
		}

		go func() {
			defer server.bufferPool.Put(&buf)
			server.forward(buf[:n], connID)
		}()
	}
}
