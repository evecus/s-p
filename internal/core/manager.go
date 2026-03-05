package core

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/singbox-panel/internal/config"
)

const (
	githubReleaseAPI = "https://api.github.com/repos/SagerNet/sing-box/releases/latest"
	binaryName       = "sing-box"
)

type DownloadProgress struct {
	Status   string  `json:"status"`  // idle / downloading / done / error
	Progress float64 `json:"progress"` // 0-100
	Message  string  `json:"message"`
}

type CoreInfo struct {
	Installed      bool   `json:"installed"`
	Version        string `json:"version"`
	Path           string `json:"path"`
	LatestVersion  string `json:"latest_version"`
	UpdateAvailable bool  `json:"update_available"`
	Running        bool   `json:"running"`
}

type StatusInfo struct {
	Running bool   `json:"running"`
	PID     int    `json:"pid"`
	Uptime  string `json:"uptime"`
}

type Manager struct {
	dataDir    string
	cfg        *config.Manager
	binaryPath string

	mu         sync.Mutex
	process    *exec.Cmd
	startTime  time.Time
	logBuffer  []string
	logMu      sync.RWMutex

	dlMu       sync.Mutex
	dlProgress DownloadProgress
}

func NewManager(dataDir string, cfg *config.Manager) *Manager {
	return &Manager{
		dataDir:    dataDir,
		cfg:        cfg,
		binaryPath: filepath.Join(dataDir, binaryName),
	}
}

func (m *Manager) appendLog(line string) {
	m.logMu.Lock()
	defer m.logMu.Unlock()
	m.logBuffer = append(m.logBuffer, line)
	if len(m.logBuffer) > 2000 {
		m.logBuffer = m.logBuffer[len(m.logBuffer)-2000:]
	}
}

func (m *Manager) GetLogs(linesStr string) []string {
	m.logMu.RLock()
	defer m.logMu.RUnlock()
	n := 100
	fmt.Sscanf(linesStr, "%d", &n)
	if n > len(m.logBuffer) {
		n = len(m.logBuffer)
	}
	return m.logBuffer[len(m.logBuffer)-n:]
}

func (m *Manager) binaryExists() bool {
	_, err := os.Stat(m.binaryPath)
	return err == nil
}

func (m *Manager) getInstalledVersion() string {
	if !m.binaryExists() {
		return ""
	}
	out, err := exec.Command(m.binaryPath, "version").Output()
	if err != nil {
		return "unknown"
	}
	// parse "sing-box version 1.9.0"
	parts := strings.Fields(string(out))
	for i, p := range parts {
		if p == "version" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return strings.TrimSpace(string(out))
}

func (m *Manager) getLatestVersion() string {
	resp, err := http.Get(githubReleaseAPI)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	var release struct {
		TagName string `json:"tag_name"`
	}
	json.NewDecoder(resp.Body).Decode(&release)
	return strings.TrimPrefix(release.TagName, "v")
}

func (m *Manager) Info() CoreInfo {
	installed := m.binaryExists()
	version := m.getInstalledVersion()
	latest := m.getLatestVersion()
	updateAvail := installed && latest != "" && version != latest

	m.mu.Lock()
	running := m.process != nil && m.process.Process != nil
	m.mu.Unlock()

	return CoreInfo{
		Installed:       installed,
		Version:         version,
		Path:            m.binaryPath,
		LatestVersion:   latest,
		UpdateAvailable: updateAvail,
		Running:         running,
	}
}

func (m *Manager) Status() StatusInfo {
	m.mu.Lock()
	defer m.mu.Unlock()
	running := m.process != nil && m.process.Process != nil
	pid := 0
	uptime := ""
	if running {
		pid = m.process.Process.Pid
		uptime = time.Since(m.startTime).Round(time.Second).String()
	}
	return StatusInfo{Running: running, PID: pid, Uptime: uptime}
}

func (m *Manager) detectArch(arch string) string {
	if arch != "" {
		return arch
	}
	switch runtime.GOARCH {
	case "amd64":
		return "amd64"
	case "arm64":
		return "arm64"
	case "arm":
		return "armv7"
	case "386":
		return "386"
	default:
		return "amd64"
	}
}

func (m *Manager) StartDownload(version, arch string) error {
	m.dlMu.Lock()
	if m.dlProgress.Status == "downloading" {
		m.dlMu.Unlock()
		return fmt.Errorf("download already in progress")
	}
	m.dlProgress = DownloadProgress{Status: "downloading", Progress: 0}
	m.dlMu.Unlock()

	go m.download(version, arch)
	return nil
}

func (m *Manager) download(version, arch string) {
	setProgress := func(p float64, msg string) {
		m.dlMu.Lock()
		m.dlProgress.Progress = p
		m.dlProgress.Message = msg
		m.dlMu.Unlock()
	}
	setError := func(msg string) {
		m.dlMu.Lock()
		m.dlProgress.Status = "error"
		m.dlProgress.Message = msg
		m.dlMu.Unlock()
	}
	setDone := func() {
		m.dlMu.Lock()
		m.dlProgress.Status = "done"
		m.dlProgress.Progress = 100
		m.dlProgress.Message = "Download complete"
		m.dlMu.Unlock()
	}

	// Resolve version
	if version == "" {
		setProgress(5, "Fetching latest version...")
		resp, err := http.Get(githubReleaseAPI)
		if err != nil {
			setError("Failed to fetch latest version: " + err.Error())
			return
		}
		var release struct {
			TagName string `json:"tag_name"`
		}
		json.NewDecoder(resp.Body).Decode(&release)
		resp.Body.Close()
		version = strings.TrimPrefix(release.TagName, "v")
	}

	detectedArch := m.detectArch(arch)
	os_name := "linux"
	fileName := fmt.Sprintf("sing-box-%s-%s-%s.tar.gz", version, os_name, detectedArch)
	downloadURL := fmt.Sprintf("https://github.com/SagerNet/sing-box/releases/download/v%s/%s", version, fileName)

	setProgress(10, fmt.Sprintf("Downloading sing-box v%s (%s)...", version, detectedArch))

	resp, err := http.Get(downloadURL)
	if err != nil {
		setError("Download failed: " + err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		setError(fmt.Sprintf("Download failed: HTTP %d", resp.StatusCode))
		return
	}

	total := resp.ContentLength
	tmpFile := filepath.Join(m.dataDir, "sing-box.tar.gz")
	f, err := os.Create(tmpFile)
	if err != nil {
		setError("Failed to create temp file: " + err.Error())
		return
	}

	reader := &progressReader{
		r:     resp.Body,
		total: total,
		onProgress: func(downloaded int64) {
			if total > 0 {
				pct := 10 + float64(downloaded)/float64(total)*70
				setProgress(pct, fmt.Sprintf("Downloading... %.1f%%", pct))
			}
		},
	}
	if _, err := io.Copy(f, reader); err != nil {
		f.Close()
		setError("Download error: " + err.Error())
		return
	}
	f.Close()

	setProgress(85, "Extracting...")

	// Extract tar.gz
	extractDir := filepath.Join(m.dataDir, "extract_tmp")
	os.MkdirAll(extractDir, 0755)
	defer os.RemoveAll(extractDir)

	if err := exec.Command("tar", "-xzf", tmpFile, "-C", extractDir).Run(); err != nil {
		setError("Extraction failed: " + err.Error())
		return
	}
	os.Remove(tmpFile)

	// Find the binary
	binaryInArchive := fmt.Sprintf("sing-box-%s-%s-%s/sing-box", version, os_name, detectedArch)
	extractedBin := filepath.Join(extractDir, binaryInArchive)
	if _, err := os.Stat(extractedBin); err != nil {
		// Try to find it
		extracted, _ := filepath.Glob(filepath.Join(extractDir, "*/sing-box"))
		if len(extracted) == 0 {
			setError("Binary not found in archive")
			return
		}
		extractedBin = extracted[0]
	}

	setProgress(95, "Installing...")

	// Stop running instance if needed
	m.Stop()

	if err := os.Rename(extractedBin, m.binaryPath); err != nil {
		// try copy
		src, _ := os.Open(extractedBin)
		dst, _ := os.Create(m.binaryPath)
		io.Copy(dst, src)
		src.Close()
		dst.Close()
	}
	os.Chmod(m.binaryPath, 0755)

	setDone()
}

func (m *Manager) DownloadProgress() DownloadProgress {
	m.dlMu.Lock()
	defer m.dlMu.Unlock()
	return m.dlProgress
}

func (m *Manager) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.process != nil && m.process.Process != nil {
		return fmt.Errorf("sing-box is already running")
	}
	if !m.binaryExists() {
		return fmt.Errorf("sing-box binary not found, please download first")
	}

	cmd := exec.Command(m.binaryPath, "run", "-c", m.cfg.ConfigPath())
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start sing-box: %w", err)
	}

	m.process = cmd
	m.startTime = time.Now()

	// Stream logs
	scanLogs := func(r io.Reader) {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			m.appendLog(scanner.Text())
		}
	}
	go scanLogs(stdout)
	go scanLogs(stderr)

	// Watch for exit
	go func() {
		cmd.Wait()
		m.mu.Lock()
		if m.process == cmd {
			m.process = nil
		}
		m.mu.Unlock()
		m.appendLog("[panel] sing-box process exited")
	}()

	return nil
}

func (m *Manager) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.process == nil || m.process.Process == nil {
		return nil
	}
	if err := m.process.Process.Kill(); err != nil {
		return err
	}
	m.process = nil
	return nil
}

func (m *Manager) Restart() error {
	if err := m.Stop(); err != nil {
		return err
	}
	time.Sleep(500 * time.Millisecond)
	return m.Start()
}

func (m *Manager) ValidateConfig() map[string]interface{} {
	if !m.binaryExists() {
		return map[string]interface{}{"valid": false, "error": "sing-box binary not found"}
	}
	out, err := exec.Command(m.binaryPath, "check", "-c", m.cfg.ConfigPath()).CombinedOutput()
	if err != nil {
		return map[string]interface{}{"valid": false, "error": string(out)}
	}
	return map[string]interface{}{"valid": true, "output": string(out)}
}

func (m *Manager) UpdateProvider(tag string) error {
	// Reload config in running sing-box (if using Clash API)
	return nil
}

func (m *Manager) UpdateRuleSets() error {
	return m.Restart()
}

// progressReader wraps an io.Reader and reports progress
type progressReader struct {
	r          io.Reader
	total      int64
	downloaded int64
	onProgress func(int64)
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.r.Read(p)
	pr.downloaded += int64(n)
	pr.onProgress(pr.downloaded)
	return n, err
}
