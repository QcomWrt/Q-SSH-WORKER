package dns

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/xtaci/kcp-go/v5"
	"github.com/xtaci/smux"

	"github.com/QcomWrt/Q-SSH-WORKER/config"
	qdebug "github.com/QcomWrt/Q-SSH-WORKER/debug"
	"github.com/QcomWrt/Q-SSH-WORKER/logger"
	dnstt "github.com/QcomWrt/Q-SSH-WORKER/transport/dns/internal/dns"
	"github.com/QcomWrt/Q-SSH-WORKER/transport/dns/noise"
)

const (
	smuxKeepAlive = 2 * time.Minute
	smuxMaxBuffer = 1 * 1024 * 1024
	watchInterval = 2 * time.Second
)

// Dial membuka sesi SlowDNS: UDP ke resolver, KCP, Noise NK, lalu satu
// stream smux yang dibungkus sebagai net.Conn.
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
	qdebug.SlowDNSConnectingDetails(nameserver, cfg.DNS.TargetDomain)

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

	pconn := NewDNSPacketConn(udpConn, remoteAddr, domain, cfg.DNS.QueryType)

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

	stopAbort := context.AfterFunc(ctx, func() {
		kcpConn.Close()
		pconn.Close()
	})

	rw, err := noise.NewClient(loggedConn{kcpConn}, pubkey)
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

	go watchClosed("smux", func() bool { return sess.IsClosed() })

	return &streamConn{
		stream:  stream,
		sess:    sess,
		kcpConn: kcpConn,
		pconn:   pconn,
		local:   udpConn.LocalAddr(),
		remote:  remoteAddr,
	}, nil
}

// loggedConn membungkus UDPSession dan mencatat error Read/Write dari KCP (mengabaikan closed pipe biasa).
type loggedConn struct {
	*kcp.UDPSession
}

func (l loggedConn) Read(b []byte) (int, error) {
	n, err := l.UDPSession.Read(b)
	if err != nil && !errors.Is(err, io.ErrClosedPipe) && !strings.Contains(err.Error(), "closed pipe") {
		qdebug.Printf("[SLOWDNS DEBUG] kcp Read error: %v", err)
	}
	return n, err
}

func (l loggedConn) Write(b []byte) (int, error) {
	n, err := l.UDPSession.Write(b)
	if err != nil && !errors.Is(err, io.ErrClosedPipe) && !strings.Contains(err.Error(), "closed pipe") {
		qdebug.Printf("[SLOWDNS DEBUG] kcp Write error: %v", err)
	}
	return n, err
}

// watchClosed memantau satu layer dan mencatat saat layer itu tertutup.
func watchClosed(name string, closed func() bool) {
	t := time.NewTicker(watchInterval)
	defer t.Stop()
	for range t.C {
		if closed() {
			qdebug.Printf("[SLOWDNS DEBUG] layer %s tertutup", name)
			return
		}
	}
}

// configureKCP menyamakan parameter KCP dengan dnstt-client dan dioptimalkan untuk jaringan seluler.
func configureKCP(c *kcp.UDPSession, mtu int) error {
	c.SetStreamMode(true)

	// OPTIMASI JARINGAN SELULER:
	// - nodelay = 1 (aktifkan mode fast retransmit)
	// - interval = 20ms (update clock KCP cepat)
	// - resend = 2 (trigger fast retransmit setelah 2x duplicate ACK)
	// - nc = 1 (non-congestion control agar throughput stabil)
	c.SetNoDelay(1, 20, 2, 1)

	// REKOMENDASI TAHAN BANTING SELULER:
	// - Send Window (32): Melindungi resolver DNS dari rate-limiting query upstream.
	// - Receive Window (128): Menampung jitter & paket out-of-order akibat latensi seluler yang melompat.
	c.SetWindowSize(32, 128)
	// c.SetWindowSize(64, 256)

	// Batasi MTU maksimal 470 byte untuk mencegah fragmentasi paket IP di BTS seluler.
	if mtu > 470 {
		mtu = 470
	}

	if ok := c.SetMtu(mtu); !ok {
		return fmt.Errorf("gagal set MTU KCP ke %d", mtu)
	}
	return nil
}

// mtuFor menghitung ukuran payload per paket, sama dengan dnstt-client.
func mtuFor(domain string) int {
	return nameCapacity(domain) - 8 - 1 - numPadding - 1
}

// nameCapacity: jumlah byte data yang muat di nama DNS setelah domain.
func nameCapacity(domain string) int {
	capacity := 255 - 1
	for _, label := range strings.Split(domain, ".") {
		if label == "" {
			continue
		}
		capacity -= len(label) + 1
	}
	capacity = capacity * 63 / 64
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

// streamConn membungkus stream smux menjadi net.Conn untuk SSH.
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
	// qdebug.Printf("[SLOWDNS DEBUG] streamConn.Close dipanggil")
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
