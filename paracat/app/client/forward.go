package client

import (
	"log"
	"sync"

	"github.com/chenx-dust/go-net-lab/paracat/packet"
)

func (client *Client) handleForward() {
	for {
		buf := client.bufferPool.Get().([]byte)
		n, addr, err := client.udpListener.ReadFromUDP(buf)
		if err != nil {
			log.Println("error reading from udp conn:", err)
			continue
		}
		go func() {
			defer client.bufferPool.Put(&buf)
			connID, ok := client.connAddrIDMap[addr.String()]
			if !ok {
				connID = uint16(client.connIncrement.Add(1) - 1)
				client.connMutex.Lock()
				client.connIDAddrMap[connID] = addr
				client.connAddrIDMap[addr.String()] = connID
				client.connMutex.Unlock()
			}
			packetID := client.packetFilter.NewPacketID()

			wg := sync.WaitGroup{}
			for _, relay := range client.tcpRelays {
				wg.Add(1)
				go func() {
					defer wg.Done()
					packet.WritePacket(relay, buf[:n], connID, packetID)
				}()
			}
			udpPacked := packet.Pack(buf[:n], connID, packetID)
			for _, relay := range client.udpRelays {
				wg.Add(1)
				go func() {
					defer wg.Done()
					relay.Write(udpPacked)
				}()
			}
			wg.Wait()
		}()
	}
}
