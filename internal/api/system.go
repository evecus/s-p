package api

import (
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var startTime = time.Now()

func (s *Server) systemInfo(c *gin.Context) {
	hostname, _ := exec.Command("hostname").Output()
	kernelBytes, _ := exec.Command("uname", "-r").Output()

	c.JSON(http.StatusOK, gin.H{
		"hostname":   strings.TrimSpace(string(hostname)),
		"kernel":     strings.TrimSpace(string(kernelBytes)),
		"os":         runtime.GOOS,
		"arch":       runtime.GOARCH,
		"go_version": runtime.Version(),
		"uptime":     time.Since(startTime).String(),
	})
}

func (s *Server) systemStatus(c *gin.Context) {
	coreStatus := s.coreMgr.Status()
	proxyStatus := s.fwMgr.Status()

	c.JSON(http.StatusOK, gin.H{
		"core":    coreStatus,
		"proxy":   proxyStatus,
		"time":    time.Now().Format(time.RFC3339),
		"uptime":  time.Since(startTime).Seconds(),
	})
}
