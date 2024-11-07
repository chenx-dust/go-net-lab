package server

import (
	"log"
	"net"
	"sync"

	"github.com/chenx-dust/go-net-lab/paracat/config"
	"github.com/chenx-dust/go-net-lab/paracat/packet"
)

type Server struct {
	cfg         *config.Config
	tcpListener *net.TCPListener
	udpListener *net.UDPConn

	sourceMutex    sync.RWMutex
	sourceTCPConns []*net.TCPConn
	sourceUDPAddrs []*net.UDPAddr

	forwardMutex sync.Mutex
	forwardConns map[uint16]*net.UDPConn

	packetFilter *packet.PacketFilter

	bufferPool sync.Pool
}

func NewServer(cfg *config.Config) *Server {
	return &Server{cfg: cfg, packetFilter: packet.NewPacketManager(), bufferPool: sync.Pool{
		New: func() any {
			return make([]byte, cfg.BufferSize)
		},
	}}
}

func (server *Server) Run() error {
	log.Println("running server")

	tcpAddr, err := net.ResolveTCPAddr("tcp", server.cfg.ListenAddr)
	if err != nil {
		return err
	}
	server.tcpListener, err = net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return err
	}

	udpAddr, err := net.ResolveUDPAddr("udp", server.cfg.ListenAddr)
	if err != nil {
		return err
	}
	server.udpListener, err = net.ListenUDP("udp", udpAddr)
	if err != nil {
		return err
	}
	log.Println("listening on", server.cfg.ListenAddr)

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		server.handleTCP()
	}()
	go func() {
		defer wg.Done()
		server.handleUDP()
	}()
	wg.Wait()

	return nil
}
