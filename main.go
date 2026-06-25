package main

import (
	"fmt"

	"github.com/QcomWrt/Q-SSH-WORKER/version"
)

func main() {
	fmt.Printf("%s %s\n", version.Name, version.Version)
}