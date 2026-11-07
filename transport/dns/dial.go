package dns

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/xtaci/kcp-go/v5"
	"github.com/xtaci/smux"

	"github.com/QcomWrt/Q-SSH-WORKER/config"
	"github.com/QcomWrt/Q-SSH-WORKER/debug"
	"github.com/QcomWrt/Q-SSH-WORKER/logger"
	dnstt "github.com/QcomWrt/Q-SSH-WORKER/transport/dns/internal/dns"
	turbotunnel "github.com/QcomWrt/Q-SSH-WORKER/transport/dns/internal/turbotunnel"
	"github.com/QcomWrt/Q-SSH-WORKER/transport/dns/noise"
)

const (
	smuxKeepAlive = 2 * time.Minute
	smuxMaxBuffer = 1 * 1024 * 1024
)

// Dial membuka sesi SlowDNS: UDP ke resolver, KCP, Noise NK, lalu satu
// stream smux yang dibungkus sebagai net.Conn. Context membatasi proses
// dial (termasuk handshake). Setelah berhasil, koneksi tidak lagi
// terikat context dial.
func Dial(ctx context.Context, cfg *config.Config) (net.Conn, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if cfg.DNS.TargetDomain == "" || cfg.DNS.PubKey == "" {
		err := fmt.Errorf("target_domain atau pubkey kosong")
		logger.SlowDNSConfigError(err.Error())
		return nil, err
	}

	pubkey, err := noise.DecodeKey(cfg.DNS.PubKey)
	if err != nil {
		logger.SlowDNSError(fmt.Errorf("pubkey tidak valid: %w", err))
		return nil, err
	}

	domain, err := dnstt.ParseName(cfg.DNS.TargetDomain)
	if err != nil {
		logger.SlowDNSError(fmt.Errorf("target_domain tidak valid: %w", err))
		return nil, err
	}

	nameserver := normalizeResolver(cfg.DNS.Nameserver)
	debug.SlowDNSConnectingDetails(nameserver, cfg.DNS.TargetDomain)

	mtu := mtuFor(cfg.DNS.TargetDomain)
	if mtu < 80 {
		err := fmt.Errorf("domain %s hanya menyisakan %d byte untuk payload", cfg.DNS.TargetDomain, mtu)
		logger.SlowDNSError(err)
		return nil, err
	}

	remoteAddr, err := net.ResolveUDPAddr("udp", nameserver)
	if err != nil {
		logger.SlowDNSResolverError(nameserver, err)
		return nil, err
	}

	udpConn, err := net.ListenUDP("udp", nil)
	if err != nil {
		logger.SlowDNSResolverError(nameserver, err)
		return nil, err
	}

	pconn := NewDNSPacketConn(udpConn, remoteAddr, domain)

	kcpConn, err := kcp.NewConn2(remoteAddr, nil, 0, 0, pconn)
	if err != nil {
		pconn.Close()
		logger.SlowDNSError(fmt.Errorf("membuka KCP: %w", err))
		return nil, err
	}
	if err := configureKCP(kcpConn, mtu); err != nil {
		kcpConn.Close()
		pconn.Close()
		logger.SlowDNSError(err)
		return nil, err
	}

	// Handshake Noise berjalan di dalam stream KCP dan memblokir. Jika ctx
	// dibatalkan (misalnya timeout dial), KCP ditutup supaya handshake berhenti.
	stopAbort := context.AfterFunc(ctx, func() {
		kcpConn.Close()
		pconn.Close()
	})

	rw, err := noise.NewClient(kcpConn, pubkey)
	stillActive := stopAbort()
	if !stillActive || err != nil {
		kcpConn.Close()
		pconn.Close()
		if ctxErr := ctx.Err(); ctxErr != nil {
			return nil, ctxErr
		}
		if err == nil {
			err = errors.New("dial SlowDNS dibatalkan")
		}
		logger.SlowDNSError(fmt.Errorf("handshake DNSTT gagal: %w", err))
		return nil, err
	}

	smuxCfg := smux.DefaultConfig()
	smuxCfg.Version = 2
	smuxCfg.KeepAliveTimeout = smuxKeepAlive
	smuxCfg.MaxStreamBuffer = smuxMaxBuffer

	sess, err := smux.Client(rw, smuxCfg)
	if err != nil {
		kcpConn.Close()
		pconn.Close()
		logger.SlowDNSError(fmt.Errorf("membuka smux: %w", err))
		return nil, err
	}

	stream, err := sess.OpenStream()
	if err != nil {
		sess.Close()
		kcpConn.Close()
		pconn.Close()
		logger.SlowDNSError(fmt.Errorf("membuka stream smux: %w", err))
		return nil, err
	}

	return &streamConn{
		stream:  stream,
		sess:    sess,
		kcpConn: kcpConn,
		pconn:   pconn,
		local:   udpConn.LocalAddr(),
		remote:  remoteAddr,
	}, nil
}

// configureKCP menyamakan parameter KCP dengan dnstt-client (main.go).
func configureKCP(c *kcp.UDPSession, mtu int) error {
	c.SetStreamMode(true)
	// nc=1: congestion window dinonaktifkan, hanya dibatasi window statis.
	c.SetNoDelay(0, 0, 0, 1)
	c.SetWindowSize(turbotunnel.QueueSize/2, turbotunnel.QueueSize/2)
	if ok := c.SetMtu(mtu); !ok {
		return fmt.Errorf("gagal set MTU KCP ke %d", mtu)
	}
	return nil
}

// mtuFor menghitung ukuran payload per paket, sama dengan dnstt-client:
// kapasitas nama DNS - clientID(8) - panjang padding(1) - padding - panjang data(1).
func mtuFor(domain string) int {
	return nameCapacity(domain) - 8 - 1 - numPadding - 1
}

// nameCapacity: jumlah byte data yang muat di nama DNS setelah domain.
func nameCapacity(domain string) int {
	// Nama DNS maksimal 255 byte, termasuk null terminator.
	capacity := 255 - 1
	for _, label := range strings.Split(domain, ".") {
		if label == "" {
			continue
		}
		// Panjang label + 1 byte length octet.
		capacity -= len(label) + 1
	}
	// Setiap label maksimal 63 byte dan memakai 64 byte.
	capacity = capacity * 63 / 64
	// Base32: 5 byte menjadi 8 karakter.
	capacity = capacity * 5 / 8
	return capacity
}

func normalizeResolver(addr string) string {
	if addr == "" {
		return "1.1.1.1:53"
	}
	if !strings.Contains(addr, ":") {
		return addr + ":53"
	}
	return addr
}

// streamConn membungkus stream smux menjadi net.Conn untuk SSH, dan
// menutup seluruh lapisan saat Close dipanggil.
type streamConn struct {
	stream  *smux.Stream
	sess    *smux.Session
	kcpConn *kcp.UDPSession
	pconn   net.PacketConn
	local   net.Addr
	remote  net.Addr
}

func (c *streamConn) Read(b []byte) (int, error)  { return c.stream.Read(b) }
func (c *streamConn) Write(b []byte) (int, error) { return c.stream.Write(b) }

func (c *streamConn) Close() error {
	c.stream.Close()
	c.sess.Close()
	c.kcpConn.Close()
	return c.pconn.Close()
}

func (c *streamConn) LocalAddr() net.Addr  { return c.local }
func (c *streamConn) RemoteAddr() net.Addr { return c.remote }

func (c *streamConn) SetDeadline(t time.Time) error {
	if err := c.stream.SetReadDeadline(t); err != nil {
		return err
	}
	return c.stream.SetWriteDeadline(t)
}

func (c *streamConn) SetReadDeadline(t time.Time) error  { return c.stream.SetReadDeadline(t) }
func (c *streamConn) SetWriteDeadline(t time.Time) error { return c.stream.SetWriteDeadline(t) }
