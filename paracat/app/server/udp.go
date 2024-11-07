package server

import (
	"log"

	"github.com/chenx-dust/go-net-lab/paracat/packet"
)

func (server *Server) handleUDP() {
	for {
		buf := make([]byte, server.cfg.BufferSize)
		n, err := server.udpListener.Read(buf)
		if err != nil {
			log.Fatalln("error reading packet:", err)
		}

		connID, packetID, data, err := packet.Unpack(buf[:n])
		if err != nil {
			log.Println("error unpacking packet:", err)
			continue
		}

		isDuplicate := server.packetFilter.CheckDuplicatePacketID(packetID)
		if isDuplicate {
			continue
		}

		go func() {
			server.forward(data, connID)
		}()
	}
}
