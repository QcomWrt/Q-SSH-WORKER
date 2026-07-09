package internal

import (
	"io"
	"net"
)

// BufferedConn menggabungkan data yang terlanjur dibaca di buffer dengan socket asli
type BufferedConn struct {
	net.Conn
	r io.Reader
}

func NewBufferedConn(c net.Conn, r io.Reader) net.Conn {
	return &BufferedConn{
		Conn: c,
		r:    r,
	}
}

func (b *BufferedConn) Read(p []byte) (int, error) {
	return b.r.Read(p)
}