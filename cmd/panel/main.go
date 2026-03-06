package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/singbox-panel/internal/api"
)

func main() {
	port := flag.Int("port", 8080, "Panel listen port")
	host := flag.String("host", "0.0.0.0", "Panel listen host")
	dataDir := flag.String("data", "/etc/singbox-panel", "Data directory")
	webDir := flag.String("web", "", "Web static files directory (default: dist/ next to binary)")
	flag.Parse()

	if os.Geteuid() != 0 {
		fmt.Println("⚠️  Warning: Not running as root, firewall/service management may fail")
	}

	if err := os.MkdirAll(*dataDir, 0755); err != nil {
		log.Fatalf("Failed to create data dir: %v", err)
	}

	// 自动探测 web 静态文件目录
	staticDir := *webDir
	if staticDir == "" {
		// 优先级：1) 可执行文件旁的 dist/  2) 当前目录的 dist/  3) 源码目录的 dist/
		exe, err := os.Executable()
		if err == nil {
			candidate := filepath.Join(filepath.Dir(exe), "dist")
			if info, err := os.Stat(candidate); err == nil && info.IsDir() {
				staticDir = candidate
			}
		}
		if staticDir == "" {
			candidate := "dist"
			if info, err := os.Stat(candidate); err == nil && info.IsDir() {
				staticDir = candidate
			}
		}
		if staticDir == "" {
			// 开发模式：源码目录
			_, file, _, ok := runtime.Caller(0)
			if ok {
				candidate := filepath.Join(filepath.Dir(file), "dist")
				if info, err := os.Stat(candidate); err == nil && info.IsDir() {
					staticDir = candidate
				}
			}
		}
	}

	if staticDir == "" {
		fmt.Println("⚠️  Warning: Web static files (dist/) not found.")
		fmt.Println("   Please build the frontend first:")
		fmt.Println("   cd web && npm install && npm run build")
		fmt.Println("   Or specify --web /path/to/dist")
	} else {
		fmt.Printf("📁 Serving web files from: %s\n", staticDir)
	}

	addr := fmt.Sprintf("%s:%d", *host, *port)
	fmt.Printf("🚀 Singbox Panel starting on http://%s\n", addr)

	server := api.NewServer(*dataDir, staticDir)
	if err := server.Run(addr); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
