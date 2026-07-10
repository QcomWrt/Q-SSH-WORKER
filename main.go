package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/QcomWrt/Q-SSH-WORKER/config"
	"github.com/QcomWrt/Q-SSH-WORKER/debug"
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
		isChild          bool // 🟢 Flag internal rahasia untuk memisahkan Master & Child
	)

	flag.StringVar(&dialPath, "dial", "", "Jalur ke file konfigurasi JSON untuk terhubung ke SSH")
	flag.StringVar(&checkPath, "check", "", "Hanya memvalidasi sintaks file konfigurasi JSON")
	flag.StringVar(&showEndpointPath, "show-endpoint", "", "Ambil detail IP Server VPS untuk keperluan routing bypass")
	flag.BoolVar(&showVersion, "version", false, "Menampilkan informasi versi biner Q-SSH-WORKER")
	flag.BoolVar(&forceDebug, "debug", false, "Memaksa mengaktifkan mode debug secara manual via CLI")
	flag.BoolVar(&isChild, "child", false, "Flag internal penanda proses child-worker")
	
	flag.Parse()

	// 1. HANDLER: --version
	if showVersion {
		fmt.Printf("Q-SSH-WORKER\n")
		fmt.Printf("Version : %s\n", version.Version)
		fmt.Printf("Commit  : %s\n", version.Commit)
		fmt.Printf("Build   : %s\n", version.BuildDate)
		fmt.Printf("Go      : %s\n", runtime.Version())
		os.Exit(0)
	}

	// 2. HANDLER: --check
	if checkPath != "" {
		_, err := config.Load(checkPath)
		if err != nil {
			fmt.Printf("[CHECK ERROR] File konfigurasi kotor/invalid: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("[SUCCESS] File konfigurasi valid.")
		os.Exit(0)
	}

	// ======================================================================
	// 🟢 HANDLER UTAMA: --show-endpoint (UNTUK KEBUTUHAN IP ROUTING BYPASS)
	// ======================================================================
	if showEndpointPath != "" {
		cfg, err := config.Load(showEndpointPath)
		if err != nil {
			fmt.Printf("[ERROR] Gagal memuat config: %v\n", err)
			os.Exit(1)
		}

		// Cari IP asli dari Domain SSH Host via DNS Lookup internal
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

		// Cetak string bersih mentah agar mudah di-grep / di-parse oleh script routing LuCI
		fmt.Printf("%s\n", targetIP)
		os.Exit(0) // Langsung exit aman tanpa dial ke network
	}

	// 3. HANDLER: --dial (Proses Normal Kerja Core)
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

	// ======================================================================
	// 🟢 LOGIKA AUTONOMOUS WORKER MANAGEMENT (SINGLE VS MULTI)
	// ======================================================================
	
	// Jalur A: Jika bertindak sebagai Child, ATAU user menonaktifkan fitur concurrency di JSON
	// 🔄 UBAH: dari cfg.Concurrency.Enable menjadi cfg.Worker.Enable
	if isChild || !cfg.Worker.Enable {
		if err := worker.StartWorker(cfg); err != nil {
			fmt.Printf("Worker Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Jalur B: Jika bertindak sebagai Master Manager (Concurrency aktif & --child tidak dipanggil)
	// 🔄 UBAH: dari cfg.Concurrency.Workers menjadi cfg.Worker.Workers
	fmt.Printf("👑 Q-SSH-WORKER bertindak sebagai Master Manager (Menjaga %d Workers)\n", cfg.Worker.Workers)

	binPath, err := os.Executable()
	if err != nil {
		fmt.Printf("Gagal mendeteksi executable path biner: %v\n", err)
		os.Exit(1)
	}

	// 🔄 UBAH: dari cfg.Concurrency.Workers menjadi cfg.Worker.Workers
	for i := 0; i < cfg.Worker.Workers; i++ {
		// 🔄 UBAH: dari cfg.Concurrency.StartPort menjadi cfg.Worker.StartPort
		targetPort := cfg.Worker.StartPort + i

		// Eksekusi monitoring pararel per port
		go func(port int) {
			for {
				fmt.Printf("[MASTER] Spawning Child Worker untuk mendengarkan port %d...\n", port)

				args := []string{"--dial", dialPath, "--child"}
				if forceDebug {
					args = append(args, "--debug")
				}
				
				cmd := exec.Command(binPath, args...)
				cmd.Env = append(os.Environ(), fmt.Sprintf("QTUN_TARGET_PORT=%d", port))
				
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr

				_ = cmd.Run()

				fmt.Printf("⚠️ Worker port %d terputus gantung (EOF/Mati)! Membangunkan ulang dalam 3 detik...\n", port)
				time.Sleep(3 * time.Second) // Jeda napas anti-looper sebelum spawn ulang
			}
		}(targetPort)

		// Beri jeda antar spawn awal agar jabat tangan SSH ke VPS mengantre tertib
		time.Sleep(3 * time.Second)
	}

	// Menahan proses Master utama agar tetap hidup mengawal anak-anaknya di background
	select {}
}