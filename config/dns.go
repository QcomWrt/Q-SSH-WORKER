package config

type DNS struct {
	Nameserver   string `json:"nameserver"`    // IP DNS Resolver (contoh: "1.1.1.1:53")
	TargetDomain string `json:"target_domain"` // Domain NameServer (contoh: "ns-sg-melbi.toko-vpn.my.id")
	PubKey       string `json:"pubkey"`        // Public Key Hex 32-bytes (64 karakter Hex)
	QueryType    string `json:"query_type"`    // "NULL" (default) atau "TXT"
}

// Enabled menentukan apakah jalur SlowDNS dipakai.
func (d DNS) Enabled() bool {
	return d.TargetDomain != ""
}