package packet

import (
	"sync"
	"sync/atomic"
)

type PacketFilter struct {
	packetIncrement atomic.Uint32
	packetLowMutex  sync.Mutex
	packetLowMap    map[uint16]struct{}
	packetHighMutex sync.Mutex
	packetHighMap   map[uint16]struct{}
}

func NewPacketManager() *PacketFilter {
	return &PacketFilter{
		packetLowMap:  make(map[uint16]struct{}),
		packetHighMap: make(map[uint16]struct{}),
	}
}

func (pm *PacketFilter) NewPacketID() uint16 {
	nextID := pm.packetIncrement.Add(1)

	// clear map
	if uint16(nextID) == 0 {
		pm.packetLowMutex.Lock()
		pm.packetLowMap = make(map[uint16]struct{})
		pm.packetLowMutex.Unlock()
	} else if uint16(nextID) == 1<<15 {
		pm.packetHighMutex.Lock()
		pm.packetHighMap = make(map[uint16]struct{})
		pm.packetHighMutex.Unlock()
	}
	return uint16(nextID - 1)
}

func (pm *PacketFilter) CheckDuplicatePacketID(id uint16) bool {
	var ok bool
	if id < 1<<15 {
		pm.packetLowMutex.Lock()
		defer pm.packetLowMutex.Unlock()
		_, ok = pm.packetLowMap[id]
		if !ok {
			pm.packetLowMap[id] = struct{}{}
		}
	} else {
		pm.packetHighMutex.Lock()
		defer pm.packetHighMutex.Unlock()
		_, ok = pm.packetHighMap[id]
		if !ok {
			pm.packetHighMap[id] = struct{}{}
		}
	}
	return ok
}
