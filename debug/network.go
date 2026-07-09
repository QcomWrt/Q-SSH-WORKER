package debug

import (
	"net"
)

// SSHNetworkDetails mencetak info socket network secara mendalam jika debug aktif
func SSHNetworkDetails(netType string, remote string, proxy net.Addr, local net.Addr) {
	if !Enable {
		return
	}

	Println("SSH Connected")
	Printf("Network : %s\n", netType)
	Printf("Remote  : %s\n", remote)
	Printf("Proxy   : %s\n", proxy.String())
	Printf("Local   : %s\n", local.String())
	Println() // Spasi pemisah bawah
}