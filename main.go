package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/QcomWrt/Q-SSH-WORKER/config"
	"github.com/QcomWrt/Q-SSH-WORKER/debug"
	"github.com/QcomWrt/Q-SSH-WORKER/logger"
	"github.com/QcomWrt/Q-SSH-WORKER/version"
	"github.com/QcomWrt/Q-SSH-WORKER/worker"
)

func main() {
	var (
		dialPath         string
		checkPath        string
		showEndpointPath string
		showVersion      bool
		forceDebug       bool
		isChild          bool
	)

	// var logPath string
	// flag.StringVar(&logPath, "log", "", "Jalur file log (opsional)")
	flag.StringVar(&dialPath, "dial", "", "Jalur ke file konfigurasi JSON untuk terhubung ke SSH")
	flag.StringVar(&checkPath, "check", "", "Hanya memvalidasi sintaks file konfigurasi JSON")
	flag.StringVar(&showEndpointPath, "show-endpoint", "", "Ambil detail IP Server VPS untuk keperluan routing bypass")
	flag.BoolVar(&showVersion, "version", false, "Menampilkan informasi versi biner Q-SSH-WORKER")
	flag.BoolVar(&forceDebug, "debug", false, "Memaksa mengaktifkan mode debug secara manual via CLI")
	flag.BoolVar(&isChild, "child", false, "Flag internal penanda proses child-worker")
	flag.Parse()

	// if logPath != "" {
	// 	if err := logger.OpenFile(logPath); err != nil {
	// 		fmt.Printf("Gagal membuka file log: %v\n", err)
	// 		os.Exit(1)
	// 	}
	// }

	if showVersion {
		fmt.Printf("Q-SSH-WORKER\n")
		fmt.Printf("Version : %s\n", version.Version)
		fmt.Printf("Commit  : %s\n", version.Commit)
		fmt.Printf("Build   : %s\n", version.BuildDate)
		fmt.Printf("Go      : %s\n", runtime.Version())
		os.Exit(0)
	}

	if checkPath != "" {
		if _, err := config.Load(checkPath); err != nil {
			fmt.Printf("[CHECK ERROR] File konfigurasi kotor/invalid: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("[SUCCESS] File konfigurasi valid.")
		os.Exit(0)
	}

	if showEndpointPath != "" {
		cfg, err := config.Load(showEndpointPath)
		if err != nil {
			fmt.Printf("[ERROR] Gagal memuat config: %v\n", err)
			os.Exit(1)
		}

		ips, err := net.LookupIP(cfg.SSH.Host)
		if err != nil {
			fmt.Printf("[ERROR] Gagal resolve DNS host %s: %v\n", cfg.SSH.Host, err)
			os.Exit(1)
		}

		var targetIP string
		for _, ip := range ips {
			if ip.To4() != nil {
				targetIP = ip.String()
				break
			}
		}
		if targetIP == "" {
			fmt.Println("[ERROR] IP IPv4 tidak ditemukan untuk host tersebut.")
			os.Exit(1)
		}

		fmt.Printf("%s\n", targetIP)
		os.Exit(0)
	}

	if dialPath == "" {
		fmt.Println("Gunakan perintah:")
		fmt.Println("  ./Q-SSH-WORKER --dial <file.json>")
		fmt.Println("  ./Q-SSH-WORKER --show-endpoint <file.json>")
		os.Exit(1)
	}

	cfg, err := config.Load(dialPath)
	if err != nil {
		fmt.Printf("Gagal memuat konfigurasi: %v\n", err)
		os.Exit(1)
	}

	if forceDebug {
		debug.Enable = true
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Jalur A: satu worker di proses ini (child, atau concurrency nonaktif).
	if isChild || !cfg.Worker.Enable {
		if env := os.Getenv("QTUN_TARGET_PORT"); env != "" {
			var p int
			if _, err := fmt.Sscanf(env, "%d", &p); err == nil && p > 0 && p <= 65535 {
				cfg.Listen.Port = p
			}
		}
		port := cfg.Listen.Port

		err := worker.Supervise(ctx, cfg, worker.SuperviseOptions{
			MaxReconnect: 0,
			OnStatus: func(healthy bool, msg string) {
				logger.Status("[STATUS] Port: %d (%s)", port, msg)
			},
			Logf: func(format string, args ...any) {
				logger.Status("[SUP] "+format, args...)
			},
		})
		if err != nil {
			fmt.Printf("Worker Error: %v\n", err)
			if strings.Contains(err.Error(), "Auth Failed") {
				os.Exit(5)
			}
			os.Exit(1)
		}
		return
	}

	// Jalur B: master manager, spawn satu child per port.
	fmt.Printf("👑 Q-SSH-WORKER bertindak sebagai Master Manager (Menjaga %d Workers)\n", cfg.Worker.Workers)

	binPath, err := os.Executable()
	if err != nil {
		fmt.Printf("Gagal mendeteksi executable path biner: %v\n", err)
		os.Exit(1)
	}

	for i := 0; i < cfg.Worker.Workers; i++ {
		port := cfg.Worker.StartPort + i
		go superviseChild(ctx, binPath, dialPath, forceDebug, port)

		// Jeda antar spawn awal agar jabat tangan SSH mengantre tertib.
		time.Sleep(3 * time.Second)
	}

	<-ctx.Done()
}

// superviseChild menjalankan satu child. Exit code 5 (auth salah) menghentikan
// seluruh master. Exit lain dijalankan ulang setelah jeda, sebagai pengaman
// kalau child sendiri berhenti.
func superviseChild(ctx context.Context, binPath, dialPath string, forceDebug bool, port int) {
	for ctx.Err() == nil {
		fmt.Printf("[MASTER] Spawning Child Worker untuk mendengarkan port %d...\n", port)

		args := []string{"--dial", dialPath, "--child"}
		if forceDebug {
			args = append(args, "--debug")
		}

		cmd := exec.CommandContext(ctx, binPath, args...)
		cmd.Env = append(os.Environ(), fmt.Sprintf("QTUN_TARGET_PORT=%d", port))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err := cmd.Run()
		if ctx.Err() != nil {
			return
		}

		if exitErr, ok := err.(*exec.ExitError); ok {
			if status, ok := exitErr.Sys().(interface{ ExitStatus() int }); ok && status.ExitStatus() == 5 {
				fmt.Printf("\n❌ [MASTER] Menangkap kode 5. Menghentikan seluruh Master Manager karena kredensial expired/salah!\n")
				os.Exit(1)
			}
		}

		fmt.Printf("⚠️ Worker port %d berhenti, dijalankan ulang dalam 3 detik...\n", port)
		select {
		case <-ctx.Done():
			return
		case <-time.After(3 * time.Second):
		}
	}
}
