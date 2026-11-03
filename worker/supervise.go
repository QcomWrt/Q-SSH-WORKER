package worker

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strconv"
	"time"

	"github.com/QcomWrt/Q-SSH-WORKER/config"
	"github.com/QcomWrt/Q-SSH-WORKER/debug"
)

// SuperviseOptions mengatur reconnect dan probe. Nilai nol memakai default.
type SuperviseOptions struct {
	// MaxReconnect: jumlah reconnect berturut-turut sebelum worker menyerah.
	// 0 berarti tanpa batas.
	MaxReconnect int

	// ReadyTimeout: batas waktu worker harus lolos probe setelah dijalankan.
	ReadyTimeout time.Duration

	// ProbeInterval: jeda antar probe saat worker sudah sehat.
	ProbeInterval time.Duration

	// ProbeTimeout: batas satu probe SOCKS.
	ProbeTimeout time.Duration

	// MaxProbeFails: probe gagal berturut-turut sebelum sesi dianggap mati.
	MaxProbeFails int

	// StableAfter: sesi yang sehat selama ini mereset hitungan reconnect
	// dan backoff.
	StableAfter time.Duration

	BackoffMin time.Duration
	BackoffMax time.Duration

	// OnStatus dipanggil hanya saat status berubah (sehat / tidak sehat).
	OnStatus func(healthy bool, msg string)

	// Logf opsional, untuk pesan supervisor.
	Logf func(format string, args ...any)
}

func (o SuperviseOptions) withDefaults() SuperviseOptions {
	if o.ReadyTimeout <= 0 {
		o.ReadyTimeout = 20 * time.Second
	}
	if o.ProbeInterval <= 0 {
		o.ProbeInterval = 5 * time.Second
	}
	if o.ProbeTimeout <= 0 {
		o.ProbeTimeout = 10 * time.Second
	}
	if o.MaxProbeFails <= 0 {
		o.MaxProbeFails = 3
	}
	if o.StableAfter <= 0 {
		o.StableAfter = 60 * time.Second
	}
	if o.BackoffMin <= 0 {
		o.BackoffMin = 1 * time.Second
	}
	if o.BackoffMax <= 0 {
		o.BackoffMax = 15 * time.Second
	}
	if o.OnStatus == nil {
		o.OnStatus = func(bool, string) {}
	}
	if o.Logf == nil {
		o.Logf = func(string, ...any) {}
	}
	return o
}

// Supervise menjalankan satu worker dan menghidupkannya lagi setiap kali
// sesi putus, dengan port yang sama. Berhenti saat ctx dibatalkan (nil),
// atau saat reconnect sudah mencapai MaxReconnect (error).
func Supervise(ctx context.Context, cfg *config.Config, o SuperviseOptions) error {
	o = o.withDefaults()
	port := cfg.Listen.Port

	failures := 0
	backoff := o.BackoffMin

	for ctx.Err() == nil {
		healthy, started, err := runWithProbe(ctx, cfg, port, o)
		if ctx.Err() != nil {
			return nil
		}

		if healthy && time.Since(started) >= o.StableAfter {
			failures = 0
			backoff = o.BackoffMin
		}
		failures++

		o.Logf("worker %d berhenti: %v (reconnect ke-%d)", port, err, failures)

		if o.MaxReconnect > 0 && failures > o.MaxReconnect {
			return fmt.Errorf(
				"worker %d menyerah setelah %d reconnect: %v",
				port, o.MaxReconnect, err,
			)
		}

		o.Logf("worker %d reconnect dalam %v", port, backoff)
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(backoff):
		}

		backoff *= 2
		if backoff > o.BackoffMax {
			backoff = o.BackoffMax
		}
	}
	return nil
}

// runWithProbe menjalankan satu sesi dan memantaunya dengan probe SOCKS.
// Mengembalikan apakah sesi sempat sehat, waktu mulai, dan penyebab berhenti.
func runWithProbe(
	ctx context.Context,
	cfg *config.Config,
	port int,
	o SuperviseOptions,
) (bool, time.Time, error) {
	started := time.Now()

	sctx, cancel := context.WithCancel(ctx)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- runSession(sctx, cfg)
	}()

	// Tunggu sampai probe lolos, sesi berhenti, atau batas waktu habis.
	deadline := time.After(o.ReadyTimeout)
	tick := time.NewTicker(time.Second)
	defer tick.Stop()

readyLoop:
	for {
		select {
		case err := <-done:
			return false, started, err

		case <-deadline:
			cancel()
			<-done
			return false, started, fmt.Errorf("belum siap dalam %v", o.ReadyTimeout)

		case <-tick.C:
			if ProbeSOCKS(port, o.ProbeTimeout) == nil {
				break readyLoop
			}
		}
	}

	o.OnStatus(true, "terhubung")
	defer o.OnStatus(false, "terputus")

	// Pantau sesi yang sudah sehat.
	t := time.NewTicker(o.ProbeInterval)
	defer t.Stop()

	fails := 0
	for {
		select {
		case err := <-done:
			return true, started, err

		case <-t.C:
			if err := ProbeSOCKS(port, o.ProbeTimeout); err != nil {
				fails++
				debug.Printf("worker %d probe gagal (%d/%d): %v",
					port, fails, o.MaxProbeFails, err)
				if fails >= o.MaxProbeFails {
					cancel()
					<-done
					return true, started, fmt.Errorf(
						"probe gagal %d kali berturut-turut: %v", fails, err)
				}
				continue
			}
			fails = 0
		}
	}
}

// ProbeSOCKS mengecek SOCKS5 lokal: greeting lalu CONNECT ke example.com:443.
// Mengembalikan error yang menyebut tahap yang gagal.
func ProbeSOCKS(port int, timeout time.Duration) error {
	addr := net.JoinHostPort("127.0.0.1", strconv.Itoa(port))
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}
	defer conn.Close()

	_ = conn.SetDeadline(time.Now().Add(timeout))

	if _, err = conn.Write([]byte{0x05, 0x01, 0x00}); err != nil {
		return fmt.Errorf("greeting tulis: %w", err)
	}
	var greeting [2]byte
	if _, err = io.ReadFull(conn, greeting[:]); err != nil {
		return fmt.Errorf("greeting baca: %w", err)
	}
	if greeting[0] != 0x05 || greeting[1] != 0x00 {
		return fmt.Errorf("greeting ditolak: %v", greeting)
	}

	host := "example.com"
	req := []byte{0x05, 0x01, 0x00, 0x03, byte(len(host))}
	req = append(req, host...)
	var pb [2]byte
	binary.BigEndian.PutUint16(pb[:], 443)
	req = append(req, pb[:]...)
	if _, err = conn.Write(req); err != nil {
		return fmt.Errorf("connect tulis: %w", err)
	}

	var rep [4]byte
	if _, err = io.ReadFull(conn, rep[:]); err != nil {
		return fmt.Errorf("connect baca: %w", err)
	}
	if rep[0] != 0x05 {
		return fmt.Errorf("versi balasan salah: %d", rep[0])
	}
	if rep[1] != 0x00 {
		return fmt.Errorf("connect ditolak server, reply=%d", rep[1])
	}

	switch rep[3] {
	case 0x01:
		_, err = io.CopyN(io.Discard, conn, 4+2)
	case 0x04:
		_, err = io.CopyN(io.Discard, conn, 16+2)
	case 0x03:
		var n [1]byte
		if _, err = io.ReadFull(conn, n[:]); err == nil {
			_, err = io.CopyN(io.Discard, conn, int64(n[0])+2)
		}
	default:
		return fmt.Errorf("tipe alamat balasan tidak dikenal: %d", rep[3])
	}
	if err != nil {
		return fmt.Errorf("baca alamat balasan: %w", err)
	}
	return nil
}