package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/QcomWrt/Q-SSH-WORKER/config"
	"github.com/QcomWrt/Q-SSH-WORKER/network"
	"github.com/QcomWrt/Q-SSH-WORKER/network/dialer"
	"github.com/QcomWrt/Q-SSH-WORKER/version"
)

func main() {

	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	switch os.Args[1] {

	case "--version":
		showVersion()

	case "--check":
		checkConfig(requireConfig())

	case "--show-endpoint":
		showEndpoint(requireConfig())

	case "--dial":
		dialTest(requireConfig())

	default:
		usage()
		os.Exit(1)
	}
}

func requireConfig() string {

	if len(os.Args) < 3 {
		fmt.Println("Missing config file")
		fmt.Println()
		usage()
		os.Exit(1)
	}

	return os.Args[2]
}

func showVersion() {

	fmt.Println(version.Name)
	fmt.Println()

	fmt.Printf("Version : %s\n", version.Version)
	fmt.Printf("Commit  : %s\n", version.Commit)
	fmt.Printf("Build   : %s\n", version.BuildDate)
	fmt.Printf("Go      : %s\n", runtime.Version())
}

func checkConfig(file string) {

	_, err := config.Load(file)
	if err != nil {
		log.Fatalf("Config Error: %v", err)
	}

	fmt.Println("Config OK")
}

func showEndpoint(file string) {

	cfg, err := config.Load(file)
	if err != nil {
		log.Fatalf("Config Error: %v", err)
	}

	ips, err := dialer.Resolve(cfg.SSH.Host)
	if err != nil {
		log.Fatalf("Resolve Error: %v", err)
	}

	out := map[string]interface{}{
		"host": cfg.SSH.Host,
		"port": cfg.SSH.Port,
		"ips":  ips,
	}

	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(data))
}

func dialTest(file string) {

	cfg, err := config.Load(file)
	if err != nil {
		log.Fatalf("Config Error: %v", err)
	}

	n, err := network.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := n.Dial(ctx)
	if err != nil {
		log.Fatalf("Dial Error: %v", err)
	}
	defer conn.Close()

	fmt.Println("Network Connected")
	fmt.Printf("Network : %s\n", cfg.Network.Type)
	fmt.Printf("Remote  : %s\n", conn.RemoteAddr())
	fmt.Printf("Local   : %s\n", conn.LocalAddr())
}

func usage() {

	fmt.Println(version.Name)
	fmt.Println()

	fmt.Println("Usage:")
	fmt.Println("  qtun-ssh-worker --version")
	fmt.Println("  qtun-ssh-worker --check <config.json>")
	fmt.Println("  qtun-ssh-worker --show-endpoint <config.json>")
	fmt.Println("  qtun-ssh-worker --dial <config.json>")
}