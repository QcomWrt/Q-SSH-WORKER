package debug

import "strings"

func SlowDNSConnectingDetails(nameserver, targetDomain string) {
	if !Enable {
		return
	}

	Printf("[SLOWDNS DEBUG] Menghubungkan ke Resolver: %s | NS Target: %s", nameserver, targetDomain)
}

func SlowDNSQueryDetails(queryType, domain string, bytesSent int) {
	if !Enable {
		return
	}

	Printf("============ SLOWDNS DEBUG ============")
	Printf("Query Type : %s", strings.ToUpper(queryType))
	Printf("DNS Target : %s", domain)
	Printf("Sent Bytes : %d bytes", bytesSent)
	Printf("=======================================")
}