package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/singbox-panel/internal/firewall"
)

func (s *Server) proxyModeGet(c *gin.Context) {
	mode, err := s.fwMgr.GetMode()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, mode)
}

func (s *Server) proxyApply(c *gin.Context) {
	var req firewall.ProxyConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := s.fwMgr.Apply(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Proxy rules applied"})
}

func (s *Server) proxyStop(c *gin.Context) {
	if err := s.fwMgr.Stop(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Proxy rules cleared"})
}

func (s *Server) proxyStatus(c *gin.Context) {
	status := s.fwMgr.Status()
	c.JSON(http.StatusOK, status)
}
