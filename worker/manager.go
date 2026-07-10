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

// StartWorker mengelola inisialisasi tunggal dengan kendali reconnect internal di awal dial.
// Jika terjadi kegagalan fatal di tengah jalan saat terowongan aktif, ia akan keluar
// agar siklus recovery diambil alih penuh oleh watchdog.sh OpenWrt.
func StartWorker(cfg *config.Config) error {
	// Inisialisasi kebijakan jeda koneksi ulang untuk mengamankan proses dial awal
	reconnectPolicy := NewReconnectPolicy(2*time.Second, 30*time.Second)

	n, err := network.New(cfg)
	if err != nil {
		return fmt.Errorf("network init failed: %w", err)
	}

	var conn net.Conn

	// ======================================================================
	// 🟢 CETAK LOG HANYA 1 KALI SAAT START DI LUAR LOOP
	// ======================================================================
	if cfg.Proxy.Host != "" {
		logger.ProxyConnecting() // Mengeluarkan: "Connecting Proxy..."
	} else {
		logger.TCPConnecting()   // Mengeluarkan: "Connecting TCP..."
	}

	// Loop khusus pemicu dial awal sampai sukses terhubung
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

		// 1. Dial Connection (Bisa ke Proxy IP atau langsung ke SSH IP)
		conn, err = n.Dial(ctx)
		cancel() // Amankan context sesegera mungkin
		if err != nil {
			delay := reconnectPolicy.GetDelay()
			logger.StatusError(fmt.Sprintf("Connection failed: %v. Retrying in %v...", err, delay.Round(time.Second)))
			time.Sleep(delay)
			continue // Ulangi proses dial dengan interval backoff
		}

		// Keluar dari loop jika koneksi soket dasar berhasil terbentuk
		break
	}

	// Reset hitungan kegagalan backoff karena dial dasar sukses
	reconnectPolicy.Reset()

	// 2. Bungkus koneksi dengan observer statistik Rx/Tx
	workerStats := &TrafficStats{}
	conn = NewObservedConn(conn, workerStats)

	// ======================================================================
	// 🟢 3. TRANSPORT LAYER INJECTION (DIAMANKAN DARI NIL POINTER PANIC)
	// ======================================================================
	wrappedConn, err := transport.Wrap(cfg, conn)
	if err != nil {
		// Jika handshake payload gagal/ditutup CDN, pastikan socket dasar (conn) ditutup
		if conn != nil {
			conn.Close()
		}
		logger.ProxyError(err) // Mengirim log [PROXY ERROR] ke sistem emit tanpa crash
		return err
	}
	// Salin koneksi yang berhasil dimanipulasi ke variabel utama
	conn = wrappedConn

	// ======================================================================
	// AMAN DARI LOG GANDA: DICETAK HANYA SETELAH JABAT TANGAN PAYLOAD SUKSES
	// ======================================================================
	if cfg.Proxy.Host != "" {
		// Menggunakan fungsi debug.Proxy asli bawaan proyekmu yang sudah patuh pada debug.Enable
		debug.Proxy(cfg.Proxy.Host, cfg.Proxy.Port, cfg.SSH.Host, cfg.SSH.Port)
		logger.ProxyConnected()
	}

	logger.SSHConnecting()

	// 4. Jabat Tangan / Handshake Protokol SSH
	client, err := workerssh.Dial(cfg, conn)
	if err != nil {
		if conn != nil {
			conn.Close()
		}
		
		// Cetak satu baris error kustom ringkas penanda kegagalan auth akun
		logger.SSHError(err) 
		
		// Langsung matikan program secara paksa agar main.go tidak ikut memuntahkan string error panjang
		os.Exit(1) 
	}

	logger.SSHConnected()

	// Ambil IP SSH untuk keperluan cetak log debug
	remoteIP := cfg.SSH.Host
	if ips, err := net.LookupIP(cfg.SSH.Host); err == nil && len(ips) > 0 {
		for _, ip := range ips {
			if ip.To4() != nil {
				remoteIP = ip.String()
				break
			}
		}
	}
	remoteAddrStr := fmt.Sprintf("%s:%d", remoteIP, cfg.SSH.Port)

	debug.SSHNetworkDetails(cfg.Network.Type, remoteAddrStr, conn.RemoteAddr(), conn.LocalAddr())
	logger.StatusConnected()

	// 5. Jalankan Server SOCKS5 Inbound secara Asinkronus (Goroutine)
	socksErrChan := make(chan error, 1)
	go func() {
		socksErrChan <- socks.ListenAndServe(cfg, client)
	}()

	// 6. Jalankan Liveness Check via MonitorHealth
	healthCtx, cancelHealth := context.WithCancel(context.Background())
	healthFailChan := make(chan bool, 1)
	go MonitorHealth(healthCtx, client, healthFailChan)

	// Menahan proses tetap hidup melayani data. Jika salah satu pemicu aktif, matikan worker.
	var fatalErr error
	select {
	case err := <-socksErrChan:
		fatalErr = fmt.Errorf("socks5 server stopped: %v", err)
	case <-healthFailChan:
		fatalErr = fmt.Errorf("koneksi internet mati gantung dideteksi oleh health monitor")
	}

	// Bersihkan resource sebelum exit
	cancelHealth()
	client.Close()
	conn.Close()

	// Kembalikan error agar ditangkap main.go untuk memicu os.Exit(1)
	return fatalErr
}