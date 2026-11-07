package logger

func SlowDNSConnecting(nameserver, targetDomain string) {
	Status("Connecting SlowDNS via %s (NS: %s)...", nameserver, targetDomain)
}

func SlowDNSConnected() {
	Status("SlowDNS Tunnel Established!")
}

func SlowDNSError(err error) {
	Status("[SLOWDNS ERROR] %v", err)
}

func SlowDNSResolverError(nameserver string, err error) {
	Status("[SLOWDNS ERROR] Gagal dial UDP resolver %s: %v", nameserver, err)
}

func SlowDNSConfigError(msg string) {
	Status("[SLOWDNS ERROR] Konfigurasi tidak valid: %s", msg)
}