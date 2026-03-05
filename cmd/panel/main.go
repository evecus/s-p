package main

import (
	"embed"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/singbox-panel/internal/api"
)

//go:embed all:dist
var webFS embed.FS

func main() {
	port := flag.Int("port", 8080, "Panel listen port")
	host := flag.String("host", "0.0.0.0", "Panel listen host")
	dataDir := flag.String("data", "/etc/singbox-panel", "Data directory")
	flag.Parse()

	if os.Geteuid() != 0 {
		fmt.Println("⚠️  Warning: Not running as root, firewall/service management may fail")
	}

	if err := os.MkdirAll(*dataDir, 0755); err != nil {
		log.Fatalf("Failed to create data dir: %v", err)
	}

	addr := fmt.Sprintf("%s:%d", *host, *port)
	fmt.Printf("🚀 Singbox Panel starting on http://%s\n", addr)

	server := api.NewServer(*dataDir, webFS)
	if err := server.Run(addr); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
