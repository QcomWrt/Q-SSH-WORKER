package main

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"

	"github.com/QcomWrt/Q-SSH-WORKER/config"
	"github.com/QcomWrt/Q-SSH-WORKER/logger" // Tersentralisasi ke logger
	"github.com/QcomWrt/Q-SSH-WORKER/network/dialer"
	"github.com/QcomWrt/Q-SSH-WORKER/version"
	"github.com/QcomWrt/Q-SSH-WORKER/worker"
)

func main() {
	if err := run(); err != nil {
		// Menangkap sisa error trakhir tanpa dobel print
		logger.Error(err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) < 2 {
		usage()
		return fmt.Errorf("missing command")
	}

	switch os.Args[1] {
	case "--version":
		showVersion()
		return nil
	case "--check":
		return checkConfig(requireConfig())
	case "--show-endpoint":
		return showEndpoint(requireConfig())
	case "--dial":
		cfg, err := config.Load(requireConfig())
		if err != nil {
			// Langsung return err mentah agar dihandle logger.Error(err) di main()
			return err 
		}
		return worker.StartWorker(cfg)
	default:
		usage()
		return fmt.Errorf("unknown command: %s", os.Args[1])
	}
}

func requireConfig() string {
	if len(os.Args) < 3 {
		logger.Error(fmt.Errorf("missing config file"))
		usage()
		os.Exit(1)
	}
	return os.Args[2]
}

func showVersion() {
	// Memanfaatkan global fallback logger.Error untuk print string biasa agar konsisten lewat emit
	logger.Error(fmt.Errorf("%s", version.Name))
	logger.Error(fmt.Errorf("Version : %s", version.Version))
	logger.Error(fmt.Errorf("Commit  : %s", version.Commit))
	logger.Error(fmt.Errorf("Build   : %s", version.BuildDate))
	logger.Error(fmt.Errorf("Go      : %s", runtime.Version()))
}

func checkConfig(file string) error {
	_, err := config.Load(file)
	if err != nil {
		return err
	}
	logger.Error(fmt.Errorf("Config OK"))
	return nil
}

func showEndpoint(file string) error {
	cfg, err := config.Load(file)
	if err != nil {
		return err
	}

	ips, err := dialer.Resolve(cfg.SSH.Host)
	if err != nil {
		return err
	}

	out := map[string]interface{}{
		"host": cfg.SSH.Host,
		"port": cfg.SSH.Port,
		"ips":  ips,
	}

	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return err
	}

	logger.Error(fmt.Errorf("%s", string(data)))
	return nil
}

func usage() {
	logger.Error(fmt.Errorf("%s\n\nUsage:\n  qtun-ssh-worker --version\n  qtun-ssh-worker --check <config.json>\n  qtun-ssh-worker --show-endpoint <config.json>\n  qtun-ssh-worker --dial <config.json>", version.Name))
}