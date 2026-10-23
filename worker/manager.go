package worker

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/QcomWrt/Q-SSH-WORKER/config"
	"github.com/QcomWrt/Q-SSH-WORKER/debug"
	"github.com/QcomWrt/Q-SSH-WORKER/logger"
	"github.com/QcomWrt/Q-SSH-WORKER/network"
	"github.com/QcomWrt/Q-SSH-WORKER/socks"
	workerssh "github.com/QcomWrt/Q-SSH-WORKER/ssh"
	"github.com/QcomWrt/Q-SSH-WORKER/transport"
)

// API lama tetap tersedia untuk kompatibilitas OpenWrt.
func StartWorker(cfg *config.Config) error {
	return StartWorkerContext(context.Background(), cfg)
}

// StartWorkerContext memungkinkan pemanggil menghentikan worker.
func StartWorkerContext(ctx context.Context, cfg *config.Config) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	reconnectPolicy := NewReconnectPolicy(
		2*time.Second,
		30*time.Second,
	)

	n, err := network.New(cfg)
	if err != nil {
		return fmt.Errorf("network init failed: %w", err)
	}

	if cfg.Proxy.Host != "" {
		logger.ProxyConnecting()
	} else {
		logger.TCPConnecting()
	}

	// 1. Dial awal yang dapat dibatalkan.
	var conn net.Conn

	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		dialCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		conn, err = n.Dial(dialCtx)
		cancel()

		if err == nil {
			break
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}

		delay := reconnectPolicy.GetDelay()
		logger.StatusError(fmt.Sprintf(
			"Connection failed: %v. Retrying in %v...",
			err,
			delay.Round(time.Second),
		))

		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			return ctx.Err()

		case <-timer.C:
		}
	}

	reconnectPolicy.Reset()

	// Tutup socket dasar jika context dibatalkan, termasuk saat
	// transport atau SSH handshake sedang berlangsung.
	stopBaseClose := context.AfterFunc(ctx, func() {
		_ = conn.Close()
	})
	defer stopBaseClose()

	workerStats := &TrafficStats{}
	conn = NewObservedConn(conn, workerStats)

	wrappedConn, err := transport.Wrap(cfg, conn)
	if err != nil {
		_ = conn.Close()
		logger.ProxyError(err)
		return err
	}
	conn = wrappedConn

	if cfg.Proxy.Host != "" {
		debug.Proxy(
			cfg.Proxy.Host,
			cfg.Proxy.Port,
			cfg.SSH.Host,
			cfg.SSH.Port,
		)
		logger.ProxyConnected()
	}

	if err := ctx.Err(); err != nil {
		_ = conn.Close()
		return err
	}

	logger.SSHConnecting()

	client, err := workerssh.Dial(cfg, conn)
	if err != nil {
		_ = conn.Close()

		if ctx.Err() != nil {
			return ctx.Err()
		}

		logger.SSHError(err)

		return fmt.Errorf("SSH connection failed: %w", err)
	}

	// Cleanup tetap berjalan meskipun worker dihentikan.
	defer client.Close()
	defer conn.Close()

	if err := ctx.Err(); err != nil {
		return err
	}

	logger.SSHConnected()

	remoteIP := cfg.SSH.Host
	if ips, lookupErr := net.LookupIP(cfg.SSH.Host); lookupErr == nil {
		for _, ip := range ips {
			if ip.To4() != nil {
				remoteIP = ip.String()
				break
			}
		}
	}

	remoteAddrStr := fmt.Sprintf("%s:%d", remoteIP, cfg.SSH.Port)
	debug.SSHNetworkDetails(
		cfg.Network.Type,
		remoteAddrStr,
		conn.RemoteAddr(),
		conn.LocalAddr(),
	)
	logger.StatusConnected()

	// Pertahankan dukungan port dinamis dari Master Manager OpenWrt.
	if envPort := os.Getenv("QTUN_TARGET_PORT"); envPort != "" {
		var dynamicPort int
		if _, parseErr := fmt.Sscanf(envPort, "%d", &dynamicPort); parseErr == nil &&
			dynamicPort > 0 && dynamicPort <= 65535 {
			cfg.Listen.Port = dynamicPort
		}
	}

	socksErrChan := make(chan error, 1)
	go func() {
		socksErrChan <- socks.ListenAndServeContext(ctx, cfg, client)
	}()

	healthCtx, cancelHealth := context.WithCancel(ctx)
	defer cancelHealth()

	healthFailChan := make(chan bool, 1)
	go MonitorHealth(healthCtx, client, healthFailChan)

	select {
	case <-ctx.Done():
		return ctx.Err()

	case err := <-socksErrChan:
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return fmt.Errorf("socks5 server stopped: %v", err)

	case <-healthFailChan:
		return fmt.Errorf(
			"koneksi internet mati gantung dideteksi oleh health monitor",
		)
	}
}
