package payload

import (
	"net"

	"github.com/QcomWrt/Q-SSH-WORKER/debug"
)

func Write(conn net.Conn, data string) error {

	n, err := conn.Write([]byte(data))

	debug.Bytes(n)

	return err
}