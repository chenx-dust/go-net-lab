package client

import (
	"log"

	"github.com/chenx-dust/go-net-lab/paracat/packet"
)

func (client *Client) handleForward() {
	for {
		buf := make([]byte, client.cfg.BufferSize)
		n, addr, err := client.udpListener.ReadFromUDP(buf)
		if err != nil {
			log.Fatalln("error reading from udp conn:", err)
		}
		go func() {
			connID, ok := client.connAddrIDMap[addr.String()]
			if !ok {
				connID = uint16(client.connIncrement.Add(1) - 1)
				client.connMutex.Lock()
				client.connIDAddrMap[connID] = addr
				client.connAddrIDMap[addr.String()] = connID
				client.connMutex.Unlock()
			}
			packetID := client.packetFilter.NewPacketID()

			client.packetStat.Forward.CountPacket(uint32(n))

			for _, relay := range client.tcpRelays {
				go func() {
					n, err := packet.WritePacket(relay, buf[:n], connID, packetID)
					if err != nil {
						log.Println("error writing to tcp:", err)
					}
					if n != len(buf) {
						log.Println("error writing to tcp: wrote", n, "bytes instead of", len(buf))
					}
				}()
			}
			udpPacked := packet.Pack(buf[:n], connID, packetID)
			for _, relay := range client.udpRelays {
				go func() {
					n, err := relay.Write(udpPacked)
					if err != nil {
						log.Println("error writing to udp:", err)
					}
					if n != len(udpPacked) {
						log.Println("error writing to udp: wrote", n, "bytes instead of", len(udpPacked))
					}
				}()
			}
		}()
	}
}
