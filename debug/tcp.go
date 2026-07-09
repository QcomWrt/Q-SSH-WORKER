package debug

import "net"

func TCP(
	dial string,
	conn net.Conn,
) {

	if !Enable {
		return
	}

	Separator()

	Println("============= TCP =============")

	Printf("Dial   : %s\n", dial)
	Printf("Local  : %s\n", conn.LocalAddr())
	Printf("Remote : %s\n", conn.RemoteAddr())

	Println("===============================")
}