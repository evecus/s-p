package firewall

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ProxyScope: "self" = local machine only, "gateway" = transparent gateway for LAN
// ProxyMode: "tproxy", "redir", "tun"
// ProxyIP: "ipv4", "ipv6", "both"

type ProxyConfig struct {
	Enabled    bool   `json:"enabled"`
	Scope      string `json:"scope"`       // "self" | "gateway"
	Mode       string `json:"mode"`        // "tproxy" | "redir" | "tun"
	IPVersion  string `json:"ip_version"`  // "ipv4" | "ipv6" | "both"
	TProxyPort int    `json:"tproxy_port"` // default 7899
	RedirPort  int    `json:"redir_port"`  // default 7898
	DNSPort    int    `json:"dns_port"`    // default 1053
	FWMark     int    `json:"fwmark"`      // default 1
	RouteTable int    `json:"route_table"` // default 100
	Interface  string `json:"interface"`   // auto-detect if empty
}

type StatusInfo struct {
	Enabled   bool   `json:"enabled"`
	Mode      string `json:"mode"`
	Scope     string `json:"scope"`
	IPVersion string `json:"ip_version"`
}

type Manager struct {
	dataDir    string
	modePath   string
	currentCfg *ProxyConfig
}

func NewManager(dataDir string) *Manager {
	return &Manager{
		dataDir:  dataDir,
		modePath: filepath.Join(dataDir, "proxy_mode.json"),
	}
}

func (m *Manager) GetMode() (*ProxyConfig, error) {
	data, err := os.ReadFile(m.modePath)
	if err != nil {
		if os.IsNotExist(err) {
			return defaultProxyConfig(), nil
		}
		return nil, err
	}
	var cfg ProxyConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (m *Manager) saveMode(cfg *ProxyConfig) error {
	data, _ := json.MarshalIndent(cfg, "", "  ")
	return os.WriteFile(m.modePath, data, 0644)
}

func (m *Manager) Status() StatusInfo {
	cfg, err := m.GetMode()
	if err != nil || cfg == nil {
		return StatusInfo{}
	}
	return StatusInfo{
		Enabled:   cfg.Enabled,
		Mode:      cfg.Mode,
		Scope:     cfg.Scope,
		IPVersion: cfg.IPVersion,
	}
}

func (m *Manager) Apply(cfg *ProxyConfig) error {
	if cfg.TProxyPort == 0 {
		cfg.TProxyPort = 7899
	}
	if cfg.RedirPort == 0 {
		cfg.RedirPort = 7898
	}
	if cfg.DNSPort == 0 {
		cfg.DNSPort = 1053
	}
	if cfg.FWMark == 0 {
		cfg.FWMark = 1
	}
	if cfg.RouteTable == 0 {
		cfg.RouteTable = 100
	}

	// Auto-detect interface
	if cfg.Interface == "" {
		iface, err := getDefaultInterface()
		if err == nil {
			cfg.Interface = iface
		}
	}

	// Clear existing rules first
	m.clearRules()

	if !cfg.Enabled {
		cfg.Enabled = false
		return m.saveMode(cfg)
	}

	var err error
	switch cfg.Mode {
	case "tproxy":
		err = m.applyTProxy(cfg)
	case "redir":
		err = m.applyRedir(cfg)
	case "tun":
		err = m.applyTUN(cfg)
	default:
		return fmt.Errorf("unknown proxy mode: %s", cfg.Mode)
	}

	if err != nil {
		return err
	}
	return m.saveMode(cfg)
}

func (m *Manager) Stop() error {
	m.clearRules()
	cfg, _ := m.GetMode()
	if cfg != nil {
		cfg.Enabled = false
		m.saveMode(cfg)
	}
	return nil
}

// ─── TProxy Mode ────────────────────────────────────────────────────────────

func (m *Manager) applyTProxy(cfg *ProxyConfig) error {
	useV4 := cfg.IPVersion == "ipv4" || cfg.IPVersion == "both"
	useV6 := cfg.IPVersion == "ipv6" || cfg.IPVersion == "both"

	// Build nftables config
	var sb strings.Builder
	sb.WriteString("#!/usr/sbin/nft -f\n\n")
	sb.WriteString("table inet singbox {\n")

	// ── Prerouting (intercept incoming / forwarded traffic) ──
	sb.WriteString("    chain prerouting_tproxy {\n")
	sb.WriteString("        type filter hook prerouting priority mangle; policy accept;\n")

	// DNS hijack
	if useV4 {
		sb.WriteString(fmt.Sprintf("        ip  daddr != 127.0.0.0/8  meta l4proto { tcp, udp } th dport 53 tproxy to 127.0.0.1:%d accept comment \"DNS hijack IPv4\"\n", cfg.DNSPort))
	}
	if useV6 {
		sb.WriteString(fmt.Sprintf("        ip6 daddr != ::1/128      meta l4proto { tcp, udp } th dport 53 tproxy to [::1]:%d  accept comment \"DNS hijack IPv6\"\n", cfg.DNSPort))
	}

	// Anti-loop
	sb.WriteString(fmt.Sprintf("        fib daddr type local meta l4proto { tcp, udp } th dport %d reject with icmpx type host-unreachable comment \"anti-loop\"\n", cfg.TProxyPort))
	sb.WriteString("        fib daddr type local accept comment \"allow local\"\n")

	// Bypass private/reserved IPv4
	if useV4 {
		sb.WriteString("        ip daddr { 127.0.0.0/8, 10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16, 100.64.0.0/10, 169.254.0.0/16, 224.0.0.0/4, 240.0.0.0/4, 255.255.255.255/32 } accept comment \"bypass IPv4 private\"\n")
	}
	// Bypass private/reserved IPv6
	if useV6 {
		sb.WriteString("        ip6 daddr { ::1/128, fc00::/7, fe80::/10, ff00::/8, 64:ff9b::/96 } accept comment \"bypass IPv6 private\"\n")
	}

	// TProxy redirect
	if useV4 {
		sb.WriteString(fmt.Sprintf("        ip  meta l4proto { tcp, udp } tproxy to 127.0.0.1:%d meta mark set %d comment \"tproxy IPv4\"\n", cfg.TProxyPort, cfg.FWMark))
	}
	if useV6 {
		sb.WriteString(fmt.Sprintf("        ip6 meta l4proto { tcp, udp } tproxy to [::1]:%d meta mark set %d comment \"tproxy IPv6\"\n", cfg.TProxyPort, cfg.FWMark))
	}
	sb.WriteString("    }\n\n")

	// ── Output (intercept local-originated traffic) ──
	sb.WriteString("    chain output_tproxy {\n")
	sb.WriteString("        type route hook output priority mangle; policy accept;\n")
	if cfg.Scope == "self" {
		// Only intercept traffic from this machine
		sb.WriteString(fmt.Sprintf("        oifname != \"%s\" accept\n", cfg.Interface))
	}
	sb.WriteString(fmt.Sprintf("        meta mark %d accept comment \"bypass marked\"\n", cfg.FWMark))

	// DNS mark
	if useV4 {
		sb.WriteString(fmt.Sprintf("        ip  meta l4proto { tcp, udp } th dport 53 meta mark set %d accept comment \"DNS mark IPv4\"\n", cfg.FWMark))
	}
	if useV6 {
		sb.WriteString(fmt.Sprintf("        ip6 meta l4proto { tcp, udp } th dport 53 meta mark set %d accept comment \"DNS mark IPv6\"\n", cfg.FWMark))
	}

	// Bypass private
	if useV4 {
		sb.WriteString("        ip daddr { 127.0.0.0/8, 10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16, 100.64.0.0/10, 169.254.0.0/16, 224.0.0.0/4, 240.0.0.0/4 } accept\n")
	}
	if useV6 {
		sb.WriteString("        ip6 daddr { ::1/128, fc00::/7, fe80::/10, ff00::/8 } accept\n")
	}

	if useV4 {
		sb.WriteString(fmt.Sprintf("        ip  meta l4proto { tcp, udp } meta mark set %d comment \"route mark IPv4\"\n", cfg.FWMark))
	}
	if useV6 {
		sb.WriteString(fmt.Sprintf("        ip6 meta l4proto { tcp, udp } meta mark set %d comment \"route mark IPv6\"\n", cfg.FWMark))
	}
	sb.WriteString("    }\n")
	sb.WriteString("}\n")

	// Write and apply nftables config
	nftCfg := filepath.Join(m.dataDir, "nftables.conf")
	if err := os.WriteFile(nftCfg, []byte(sb.String()), 0644); err != nil {
		return err
	}
	if err := runCmd("nft", "-f", nftCfg); err != nil {
		return fmt.Errorf("nft apply failed: %w", err)
	}

	// Policy routing
	if useV4 {
		runCmd("ip", "rule", "add", "fwmark", fmt.Sprintf("%d", cfg.FWMark), "table", fmt.Sprintf("%d", cfg.RouteTable), "priority", "100")
		runCmd("ip", "route", "add", "local", "default", "dev", "lo", "table", fmt.Sprintf("%d", cfg.RouteTable))
	}
	if useV6 {
		runCmd("ip", "-6", "rule", "add", "fwmark", fmt.Sprintf("%d", cfg.FWMark), "table", fmt.Sprintf("%d", cfg.RouteTable), "priority", "100")
		runCmd("ip", "-6", "route", "add", "local", "default", "dev", "lo", "table", fmt.Sprintf("%d", cfg.RouteTable))
	}

	// Enable IP forwarding for gateway mode
	if cfg.Scope == "gateway" {
		if useV4 {
			runCmd("sysctl", "-w", "net.ipv4.ip_forward=1")
		}
		if useV6 {
			runCmd("sysctl", "-w", "net.ipv6.conf.all.forwarding=1")
		}
	}

	return nil
}

// ─── Redirect Mode ───────────────────────────────────────────────────────────

func (m *Manager) applyRedir(cfg *ProxyConfig) error {
	useV4 := cfg.IPVersion == "ipv4" || cfg.IPVersion == "both"
	useV6 := cfg.IPVersion == "ipv6" || cfg.IPVersion == "both"

	var sb strings.Builder
	sb.WriteString("#!/usr/sbin/nft -f\n\n")
	sb.WriteString("table inet singbox {\n")

	sb.WriteString("    chain prerouting_redir {\n")
	sb.WriteString("        type nat hook prerouting priority dstnat; policy accept;\n")
	if useV4 {
		sb.WriteString(fmt.Sprintf("        ip  daddr != 127.0.0.0/8  meta l4proto { tcp, udp } th dport 53 redirect to :%d comment \"DNS IPv4\"\n", cfg.DNSPort))
		sb.WriteString("        ip  daddr { 127.0.0.0/8, 10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16, 100.64.0.0/10, 169.254.0.0/16, 224.0.0.0/4, 240.0.0.0/4 } accept\n")
		sb.WriteString(fmt.Sprintf("        ip  meta l4proto tcp redirect to :%d comment \"TCP redirect IPv4\"\n", cfg.RedirPort))
	}
	if useV6 {
		sb.WriteString(fmt.Sprintf("        ip6 daddr != ::1/128      meta l4proto { tcp, udp } th dport 53 redirect to :%d comment \"DNS IPv6\"\n", cfg.DNSPort))
		sb.WriteString("        ip6 daddr { ::1/128, fc00::/7, fe80::/10, ff00::/8 } accept\n")
		sb.WriteString(fmt.Sprintf("        ip6 meta l4proto tcp redirect to :%d comment \"TCP redirect IPv6\"\n", cfg.RedirPort))
	}
	sb.WriteString("    }\n")

	sb.WriteString("    chain output_redir {\n")
	sb.WriteString("        type nat hook output priority dstnat; policy accept;\n")
	sb.WriteString(fmt.Sprintf("        meta mark %d accept\n", cfg.FWMark))
	if useV4 {
		sb.WriteString("        ip  daddr { 127.0.0.0/8, 10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16, 100.64.0.0/10, 169.254.0.0/16, 224.0.0.0/4, 240.0.0.0/4 } accept\n")
		sb.WriteString(fmt.Sprintf("        ip  meta l4proto tcp meta mark set %d redirect to :%d\n", cfg.FWMark, cfg.RedirPort))
	}
	if useV6 {
		sb.WriteString("        ip6 daddr { ::1/128, fc00::/7, fe80::/10, ff00::/8 } accept\n")
		sb.WriteString(fmt.Sprintf("        ip6 meta l4proto tcp meta mark set %d redirect to :%d\n", cfg.FWMark, cfg.RedirPort))
	}
	sb.WriteString("    }\n")
	sb.WriteString("}\n")

	nftCfg := filepath.Join(m.dataDir, "nftables.conf")
	if err := os.WriteFile(nftCfg, []byte(sb.String()), 0644); err != nil {
		return err
	}
	if err := runCmd("nft", "-f", nftCfg); err != nil {
		return fmt.Errorf("nft apply failed: %w", err)
	}

	if cfg.Scope == "gateway" {
		if useV4 {
			runCmd("sysctl", "-w", "net.ipv4.ip_forward=1")
		}
		if useV6 {
			runCmd("sysctl", "-w", "net.ipv6.conf.all.forwarding=1")
		}
	}
	return nil
}

// ─── TUN Mode ────────────────────────────────────────────────────────────────

func (m *Manager) applyTUN(cfg *ProxyConfig) error {
	// TUN mode is handled by sing-box itself via tun inbound
	// We just need to enable IP forwarding for gateway mode
	if cfg.Scope == "gateway" {
		useV4 := cfg.IPVersion == "ipv4" || cfg.IPVersion == "both"
		useV6 := cfg.IPVersion == "ipv6" || cfg.IPVersion == "both"
		if useV4 {
			runCmd("sysctl", "-w", "net.ipv4.ip_forward=1")
		}
		if useV6 {
			runCmd("sysctl", "-w", "net.ipv6.conf.all.forwarding=1")
		}
	}
	return nil
}

// ─── Clear Rules ─────────────────────────────────────────────────────────────

func (m *Manager) clearRules() {
	// Remove nftables table
	runCmd("nft", "delete", "table", "inet", "singbox")

	// Remove policy routing rules
	for i := 0; i < 3; i++ {
		runCmd("ip", "rule", "del", "fwmark", "1", "table", "100")
		runCmd("ip", "-6", "rule", "del", "fwmark", "1", "table", "100")
	}
	runCmd("ip", "route", "flush", "table", "100")
	runCmd("ip", "-6", "route", "flush", "table", "100")
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func getDefaultInterface() (string, error) {
	out, err := exec.Command("ip", "route", "show", "default").Output()
	if err != nil {
		return "", err
	}
	fields := strings.Fields(string(out))
	for i, f := range fields {
		if f == "dev" && i+1 < len(fields) {
			return fields[i+1], nil
		}
	}
	return "", fmt.Errorf("no default interface found")
}

func runCmd(name string, args ...string) error {
	return exec.Command(name, args...).Run()
}

func defaultProxyConfig() *ProxyConfig {
	return &ProxyConfig{
		Enabled:    false,
		Scope:      "self",
		Mode:       "tproxy",
		IPVersion:  "ipv4",
		TProxyPort: 7899,
		RedirPort:  7898,
		DNSPort:    1053,
		FWMark:     1,
		RouteTable: 100,
	}
}
