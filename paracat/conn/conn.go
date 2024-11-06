package conn

type PacketConn interface {
	ReadPacket(p []byte) (n int, id uint16, err error)
	WritePacket(p []byte, id uint16) (n int, err error)
	Close() error
}
