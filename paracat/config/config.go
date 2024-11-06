package config

type AppMode int
type ConnectionType int

const (
	NotDefined AppMode = iota
	ClientMode
	RelayMode // optional for udp-to-tcp or tcp-to-udp
	ServerMode
)

const (
	NotDefinedConnectionType ConnectionType = iota
	TCPConnectionType
	UDPConnectionType
	BothConnectionType
)

type Config struct {
	Mode         AppMode
	ListenAddr   string
	RemoteAddr   string        // not necessary in ClientMode
	RelayServers []RelayServer // only used in ClientMode
	RelayType    RelayType     // only used in RelayMode
}

type RelayServer struct {
	addr     string
	connType ConnectionType
}

type RelayType struct {
	listenType  ConnectionType
	forwardType ConnectionType
}
