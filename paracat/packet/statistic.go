package packet

import (
	"log"
	"sync/atomic"
	"time"
)

type PacketStatistic struct {
	packetCount atomic.Uint32
	bandwidth   atomic.Uint64
}

type BiPacketStatistic struct {
	Forward PacketStatistic
	Reverse PacketStatistic
}

func NewPacketStatistic() *PacketStatistic {
	return &PacketStatistic{}
}

func NewBiPacketStatistic() *BiPacketStatistic {
	return &BiPacketStatistic{}
}

func (ps *PacketStatistic) CountPacket(size uint32) {
	ps.packetCount.Add(1)
	ps.bandwidth.Add(uint64(size))
}

func (ps *PacketStatistic) GetAndReset() (count uint32, bandwidth uint64) {
	count = ps.packetCount.Swap(0)
	bandwidth = ps.bandwidth.Swap(0)
	return
}

func (bps *BiPacketStatistic) Print(period time.Duration) {
	sendCount, sendBandwidth := bps.Forward.GetAndReset()
	recvCount, recvBandwidth := bps.Reverse.GetAndReset()
	log.Printf("forward: %d packets, %d bytes in %s, %.2f bytes/s", sendCount, sendBandwidth, period, float64(sendBandwidth)/period.Seconds())
	log.Printf("reverse: %d packets, %d bytes in %s, %.2f bytes/s", recvCount, recvBandwidth, period, float64(recvBandwidth)/period.Seconds())
}
