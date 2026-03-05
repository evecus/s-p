package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// SingboxConfig is the full sing-box configuration structure
type SingboxConfig struct {
	Log          map[string]interface{}   `json:"log,omitempty"`
	DNS          map[string]interface{}   `json:"dns,omitempty"`
	NTP          map[string]interface{}   `json:"ntp,omitempty"`
	Inbounds     []map[string]interface{} `json:"inbounds,omitempty"`
	Outbounds    []map[string]interface{} `json:"outbounds,omitempty"`
	Route        map[string]interface{}   `json:"route,omitempty"`
	Experimental map[string]interface{}   `json:"experimental,omitempty"`
	Providers    []map[string]interface{} `json:"providers,omitempty"`
}

type Manager struct {
	dataDir    string
	configPath string
	mu         sync.RWMutex
}

func NewManager(dataDir string) *Manager {
	return &Manager{
		dataDir:    dataDir,
		configPath: filepath.Join(dataDir, "config.json"),
	}
}

func (m *Manager) ConfigPath() string {
	return m.configPath
}

func (m *Manager) load() (*SingboxConfig, error) {
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return defaultConfig(), nil
		}
		return nil, err
	}
	var cfg SingboxConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("invalid config JSON: %w", err)
	}
	return &cfg, nil
}

func (m *Manager) save(cfg *SingboxConfig) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	// Write to temp file first, then rename (atomic)
	tmp := m.configPath + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, m.configPath)
}

func (m *Manager) GetRaw() ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return json.MarshalIndent(defaultConfig(), "", "  ")
		}
		return nil, err
	}
	return data, nil
}

func (m *Manager) SetRaw(data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	// Validate JSON
	var test map[string]interface{}
	if err := json.Unmarshal(data, &test); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	tmp := m.configPath + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, m.configPath)
}

func (m *Manager) GetSections() (map[string]interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	cfg, err := m.load()
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"log":          cfg.Log,
		"dns":          cfg.DNS,
		"ntp":          cfg.NTP,
		"inbounds":     cfg.Inbounds,
		"outbounds":    cfg.Outbounds,
		"route":        cfg.Route,
		"experimental": cfg.Experimental,
		"providers":    cfg.Providers,
	}, nil
}

func (m *Manager) SetSection(section string, value interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cfg, err := m.load()
	if err != nil {
		return err
	}
	toMap := func(v interface{}) (map[string]interface{}, error) {
		b, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		var out map[string]interface{}
		return out, json.Unmarshal(b, &out)
	}
	toSlice := func(v interface{}) ([]map[string]interface{}, error) {
		b, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		var out []map[string]interface{}
		return out, json.Unmarshal(b, &out)
	}

	switch section {
	case "log":
		cfg.Log, err = toMap(value)
	case "dns":
		cfg.DNS, err = toMap(value)
	case "ntp":
		cfg.NTP, err = toMap(value)
	case "inbounds":
		cfg.Inbounds, err = toSlice(value)
	case "outbounds":
		cfg.Outbounds, err = toSlice(value)
	case "route":
		cfg.Route, err = toMap(value)
	case "experimental":
		cfg.Experimental, err = toMap(value)
	case "providers":
		cfg.Providers, err = toSlice(value)
	default:
		return fmt.Errorf("unknown section: %s", section)
	}
	if err != nil {
		return err
	}
	return m.save(cfg)
}

func (m *Manager) GetProviders() ([]map[string]interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	cfg, err := m.load()
	if err != nil {
		return nil, err
	}
	return cfg.Providers, nil
}

func (m *Manager) SetProviders(providers []interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cfg, err := m.load()
	if err != nil {
		return err
	}
	b, err := json.Marshal(providers)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, &cfg.Providers)
}

func (m *Manager) GetRuleSets() ([]map[string]interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	cfg, err := m.load()
	if err != nil {
		return nil, err
	}
	if cfg.Route == nil {
		return nil, nil
	}
	rs, _ := cfg.Route["rule_set"]
	if rs == nil {
		return nil, nil
	}
	b, _ := json.Marshal(rs)
	var out []map[string]interface{}
	json.Unmarshal(b, &out)
	return out, nil
}

func defaultConfig() *SingboxConfig {
	return &SingboxConfig{
		Log: map[string]interface{}{
			"disabled": false,
			"level":    "info",
		},
		DNS: map[string]interface{}{
			"servers": []interface{}{
				map[string]interface{}{
					"tag":    "ali-dns",
					"type":   "udp",
					"server": "223.5.5.5",
				},
				map[string]interface{}{
					"tag":    "google-dns",
					"type":   "https",
					"server": "8.8.8.8",
					"detour": "proxy",
				},
			},
			"rules":  []interface{}{},
			"final":  "google-dns",
			"strategy": "prefer_ipv4",
		},
		Inbounds: []map[string]interface{}{
			{
				"tag":    "tproxy-in",
				"type":   "tproxy",
				"listen": "::",
				"listen_port": 7899,
				"sniff": true,
				"sniff_override_destination": false,
			},
		},
		Outbounds: []map[string]interface{}{
			{
				"tag":  "proxy",
				"type": "selector",
				"outbounds": []string{"direct"},
			},
			{
				"tag":  "direct",
				"type": "direct",
			},
			{
				"tag":  "block",
				"type": "block",
			},
		},
		Route: map[string]interface{}{
			"rules":                 []interface{}{},
			"final":                 "proxy",
			"auto_detect_interface": true,
		},
		Experimental: map[string]interface{}{
			"cache_file": map[string]interface{}{
				"enabled": true,
				"path":    "/etc/singbox-panel/cache.db",
			},
			"clash_api": map[string]interface{}{
				"external_controller": "0.0.0.0:9090",
				"secret":              "changeme",
				"default_mode":        "rule",
			},
		},
	}
}
