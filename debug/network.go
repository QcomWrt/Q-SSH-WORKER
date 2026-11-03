package debug

import "net"

func SSHNetworkDetails(netType string, remote string, proxy net.Addr, local net.Addr) {
	Line("Server connected: %s (%s via %s)", remote, netType, proxy.String())
	Line("Local address: %s", local.String())
}