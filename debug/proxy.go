package debug

func Proxy(host string, port int, targetHost string, targetPort int) {
	Line("connected to socket %s:%d", host, port)
}