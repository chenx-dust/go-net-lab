package client

import (
	"log"
	"net"

	"github.com/chenx-dust/go-net-lab/paracat/packet"
)

func (client *Client) handleTCPReverse(conn *net.TCPConn) error {
	for {
		buf := make([]byte, client.cfg.BufferSize)
		n, connID, packetID, err := packet.ReadPacket(conn, buf)
		if err != nil {
			log.Fatalln("error reading from reverse conn:", err)
		}

		isDuplicate := client.packetFilter.CheckDuplicatePacketID(packetID)
		if isDuplicate {
			continue
		}

		go client.sendReverse(buf[:n], n, connID)
	}
}

func (client *Client) handleUDPReverse(conn *net.UDPConn) error {
	for {
		buf := make([]byte, client.cfg.BufferSize)
		n, err := conn.Read(buf)
		if err != nil {
			log.Fatalln("error reading from reverse conn:", err)
		}

		connID, packetID, data, err := packet.Unpack(buf[:n])
		if err != nil {
			log.Println("error unpacking packet:", err)
			continue
		}

		isDuplicate := client.packetFilter.CheckDuplicatePacketID(packetID)
		if isDuplicate {
			continue
		}

		go client.sendReverse(data, len(data), connID)
	}
}

func (client *Client) sendReverse(buf []byte, length int, connID uint16) {
	client.connMutex.RLock()
	udpAddr, ok := client.connIDAddrMap[connID]
	client.connMutex.RUnlock()
	if !ok {
		log.Println("conn not found")
		return
	}
	client.packetStat.Reverse.CountPacket(uint32(length))
	n, err := client.udpListener.WriteToUDP(buf[:length], udpAddr)
	if err != nil {
		log.Println("error writing to udp:", err)
	}
	if n != length {
		log.Println("error writing to udp: wrote", n, "bytes instead of", length)
	}
}
