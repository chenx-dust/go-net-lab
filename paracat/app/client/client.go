package client

import (
	"log"
	"net"
	"sync"
	"sync/atomic"

	"github.com/chenx-dust/go-net-lab/paracat/config"
	"github.com/chenx-dust/go-net-lab/paracat/packet"
)

type Client struct {
	cfg         *config.Config
	udpListener *net.UDPConn
	tcpRelays   []*net.TCPConn
	udpRelays   []*net.UDPConn

	connMutex     sync.RWMutex
	connIncrement atomic.Uint32
	connIDAddrMap map[uint16]*net.UDPAddr
	connAddrIDMap map[string]uint16

	packetFilter *packet.PacketFilter

	bufferPool sync.Pool
}

func NewClient(cfg *config.Config) *Client {
	return &Client{cfg: cfg, packetFilter: packet.NewPacketManager(), bufferPool: sync.Pool{
		New: func() any {
			return make([]byte, cfg.BufferSize)
		},
	}}
}

func (client *Client) Run() error {
	log.Println("running client")

	udpAddr, err := net.ResolveUDPAddr("udp", client.cfg.ListenAddr)
	if err != nil {
		return err
	}
	client.udpListener, err = net.ListenUDP("udp", udpAddr)
	if err != nil {
		return err
	}
	log.Println("listening on", client.cfg.ListenAddr)

	for _, relay := range client.cfg.RelayServers {
		for i := 0; i < relay.Weight; i++ {
			switch relay.ConnType {
			case config.TCPConnectionType:
				tcpAddr, err := net.ResolveTCPAddr("tcp", relay.Address)
				if err != nil {
					return err
				}
				conn, err := net.DialTCP("tcp", nil, tcpAddr)
				if err != nil {
					return err
				}
				client.tcpRelays = append(client.tcpRelays, conn)
				log.Println("connected to tcp relay", relay.Address)
			case config.UDPConnectionType:
				udpAddr, err := net.ResolveUDPAddr("udp", relay.Address)
				if err != nil {
					return err
				}
				conn, err := net.DialUDP("udp", nil, udpAddr)
				if err != nil {
					return err
				}
				client.udpRelays = append(client.udpRelays, conn)
				log.Println("connected to udp relay", relay.Address)
			default:
				log.Fatalf("invalid connection type")
			}
		}
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		client.handleForward()
	}()
	for _, relay := range client.tcpRelays {
		wg.Add(1)
		go func() {
			defer wg.Done()
			client.handleTCPReverse(relay)
		}()
	}
	for _, relay := range client.udpRelays {
		wg.Add(1)
		go func() {
			defer wg.Done()
			client.handleUDPReverse(relay)
		}()
	}
	wg.Wait()

	return nil
}
