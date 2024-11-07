package server

import (
	"log"

	"github.com/chenx-dust/go-net-lab/paracat/packet"
)

func (server *Server) handleUDP() {
	for {
		buf := server.bufferPool.Get().([]byte)
		n, err := server.udpListener.Read(buf)
		if err != nil {
			log.Println("error reading packet:", err)
			continue
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
			defer server.bufferPool.Put(&buf)
			server.forward(data, connID)
		}()
	}
}
