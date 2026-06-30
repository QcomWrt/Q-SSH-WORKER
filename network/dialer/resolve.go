package dialer

import (
	"net"
	"sort"
)

func Resolve(host string) ([]string, error) {

	ips, err := net.LookupHost(host)
	if err != nil {
		return nil, err
	}

	sort.Strings(ips)

	return ips, nil
}