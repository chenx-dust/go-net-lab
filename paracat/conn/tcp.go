package conn

import (
	"net"
)

type TCPConn struct {
	net.Conn
}

func (conn *TCPConn) Read(p []byte) (n int, err error) {
	return ReadPacket(conn, p)
}

func (conn *TCPConn) Write(p []byte) (n int, err error) {
	return WritePacket(conn, p)
}

func (conn *TCPConn) Close() error {
	return conn.Conn.Close()
}
