package debug

func Proxy(
	host string,
	port int,
	targetHost string,
	targetPort int,
) {
	if !Enable {
		return
	}

	Separator()

	Println("============ PROXY ============")
	Printf("Proxy  : %s:%d\n", host, port)
	Printf("Target : %s:%d\n", targetHost, targetPort)
	Println("===============================")
}