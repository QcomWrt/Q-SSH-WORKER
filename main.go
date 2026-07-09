package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"runtime"

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
	)

	flag.StringVar(&dialPath, "dial", "", "Jalur ke file konfigurasi JSON untuk terhubung ke SSH")
	flag.StringVar(&checkPath, "check", "", "Hanya memvalidasi sintaks file konfigurasi JSON")
	flag.StringVar(&showEndpointPath, "show-endpoint", "", "Ambil detail IP Server VPS untuk keperluan routing bypass")
	flag.BoolVar(&showVersion, "version", false, "Menampilkan informasi versi biner Q-SSH-WORKER")
	flag.BoolVar(&forceDebug, "debug", false, "Memaksa mengaktifkan mode debug secara manual via CLI")
	
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

	if err := worker.StartWorker(cfg); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}