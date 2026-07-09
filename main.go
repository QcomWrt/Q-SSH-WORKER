package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/QcomWrt/Q-SSH-WORKER/config"
	"github.com/QcomWrt/Q-SSH-WORKER/debug"
	"github.com/QcomWrt/Q-SSH-WORKER/worker"
	"github.com/QcomWrt/Q-SSH-WORKER/version"
)

func main() {
	// 1. Deklarasi CLI Flags menggunakan package flag standar
	var (
		dialPath    string
		showVersion bool
		forceDebug  bool
	)

	flag.StringVar(&dialPath, "dial", "", "Jalur ke file konfigurasi JSON (contoh: examples/config.json)")
	flag.BoolVar(&showVersion, "version", false, "Menampilkan informasi versi biner Q-SSH-WORKER")
	flag.BoolVar(&forceDebug, "debug", false, "Memaksa mengaktifkan mode debug secara manual via CLI")
	
	// Parsing argumen yang masuk dari terminal
	flag.Parse()

	// ======================================================================
	// 🟢 FITUR 1: HANDLER VERSION (BERSIH DARI TEKS "Error:")
	// ======================================================================
	if showVersion {
		fmt.Printf("Q-SSH-WORKER\n")
		fmt.Printf("Version : %s\n", version.Version)
		fmt.Printf("Commit  : %s\n", version.Commit)
		fmt.Printf("Build   : %s\n", version.BuildDate)
		fmt.Printf("Go      : %s\n", runtime.Version())
		os.Exit(0) // Keluar dengan kode sukses (0)
	}

	// Cek apakah argumen `--dial` kosong
	if dialPath == "" {
		fmt.Println("Gunakan perintah: ./q-ssh-worker --dial <file.json>")
		fmt.Println("Atau ketik: ./q-ssh-worker --help untuk bantuan.")
		os.Exit(1)
	}

	// 2. Load konfigurasi dari file JSON
	cfg, err := config.Load(dialPath)
	if err != nil {
		fmt.Printf("Gagal memuat konfigurasi: %v\n", err)
		os.Exit(1)
	}

	// ======================================================================
	// 🟢 FITUR 2: SAKELAR OVERRIDE MODE DEBUG VIA CLI
	// ======================================================================
	if forceDebug {
		debug.Enable = true
	} else {
		// Jika tidak dipaksa lewat CLI, ikuti config global (jika ada field debug di json)
		// debug.Enable = cfg.Debug 
	}

	// 3. Jalankan Worker Utama
	if err := worker.StartWorker(cfg); err != nil {
		// Error fatal di jalan (seperti auth failed / connection lost)
		// Catatan: worker.StartWorker sudah menangani os.Exit(1) internal jika terjadi auth error
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}