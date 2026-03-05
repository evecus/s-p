package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Server) coreInfo(c *gin.Context) {
	info := s.coreMgr.Info()
	c.JSON(http.StatusOK, info)
}

func (s *Server) coreDownload(c *gin.Context) {
	var req struct {
		Version string `json:"version"` // empty = latest
		Arch    string `json:"arch"`    // empty = auto-detect
	}
	c.ShouldBindJSON(&req)

	if err := s.coreMgr.StartDownload(req.Version, req.Arch); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Download started"})
}

func (s *Server) coreDownloadProgress(c *gin.Context) {
	progress := s.coreMgr.DownloadProgress()
	c.JSON(http.StatusOK, progress)
}

func (s *Server) coreStart(c *gin.Context) {
	if err := s.coreMgr.Start(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "sing-box started"})
}

func (s *Server) coreStop(c *gin.Context) {
	if err := s.coreMgr.Stop(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "sing-box stopped"})
}

func (s *Server) coreRestart(c *gin.Context) {
	if err := s.coreMgr.Restart(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "sing-box restarted"})
}

func (s *Server) coreLogs(c *gin.Context) {
	lines := c.DefaultQuery("lines", "100")
	logs := s.coreMgr.GetLogs(lines)
	c.JSON(http.StatusOK, gin.H{"logs": logs})
}
