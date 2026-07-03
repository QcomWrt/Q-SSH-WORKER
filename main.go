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
	workerssh "github.com/QcomWrt/Q-SSH-WORKER/ssh"
	"github.com/QcomWrt/Q-SSH-WORKER/transport"
	"github.com/QcomWrt/Q-SSH-WORKER/version"
)

func main() {

	if err := run(); err != nil {
		log.Fatal(err)
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
		return dialTest(requireConfig())

	default:
		usage()
		return fmt.Errorf("unknown command: %s", os.Args[1])
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

func checkConfig(file string) error {

	_, err := config.Load(file)
	if err != nil {
		return fmt.Errorf("config error: %w", err)
	}

	fmt.Println("Config OK")

	return nil
}

func showEndpoint(file string) error {

	cfg, err := config.Load(file)
	if err != nil {
		return fmt.Errorf("config error: %w", err)
	}

	ips, err := dialer.Resolve(cfg.SSH.Host)
	if err != nil {
		return fmt.Errorf("resolve error: %w", err)
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

	fmt.Println(string(data))

	return nil
}

func dialTest(file string) error {

	cfg, err := config.Load(file)
	if err != nil {
		return fmt.Errorf("config error: %w", err)
	}

	n, err := network.New(cfg)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	// Network
	conn, err := n.Dial(ctx)
	if err != nil {
		return fmt.Errorf("dial error: %w", err)
	}

	fmt.Println("TCP Connected")

	// Transport
	conn, err = transport.Wrap(cfg, conn)
	if err != nil {
		conn.Close()
		return fmt.Errorf("transport error: %w", err)
	}

	fmt.Println("Transport Connected")

	// SSH
	client, err := workerssh.Dial(cfg, conn)
	if err != nil {
		conn.Close()
		return fmt.Errorf("ssh error: %w", err)
	}

	fmt.Println("SSH Connected")

	defer client.Close()
	defer conn.Close()

	fmt.Printf("Network : %s\n", cfg.Network.Type)
	fmt.Printf("Remote  : %s\n", conn.RemoteAddr())
	fmt.Printf("Local   : %s\n", conn.LocalAddr())

	return nil
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