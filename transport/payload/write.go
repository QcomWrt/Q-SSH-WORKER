package payload

import "net"

func Write(conn net.Conn, data string) error {

	_, err := conn.Write([]byte(data))
	if err != nil {
		return err
	}

	return nil
}