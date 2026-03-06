package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Server) configGetRaw(c *gin.Context) {
	raw, err := s.cfg.GetRaw()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"config": string(raw)})
}

func (s *Server) configSetRaw(c *gin.Context) {
	var req struct {
		Config string `json:"config" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := s.cfg.SetRaw([]byte(req.Config)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Config saved"})
}

func (s *Server) configGetSections(c *gin.Context) {
	sections, err := s.cfg.GetSections()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, sections)
}

func (s *Server) configSetSection(c *gin.Context) {
	section := c.Param("section")
	// 用 interface{} 接收，兼容对象（route/dns等）和数组（inbounds/outbounds等）
	var body interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := s.cfg.SetSection(section, body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Section saved"})
}

func (s *Server) configValidate(c *gin.Context) {
	result := s.coreMgr.ValidateConfig()
	c.JSON(http.StatusOK, result)
}

func (s *Server) providersGet(c *gin.Context) {
	providers, err := s.cfg.GetProviders()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"providers": providers})
}

func (s *Server) providersSet(c *gin.Context) {
	var req struct {
		Providers []interface{} `json:"providers"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := s.cfg.SetProviders(req.Providers); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Providers saved"})
}

func (s *Server) providerUpdate(c *gin.Context) {
	tag := c.Param("tag")
	if err := s.coreMgr.UpdateProvider(tag); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Provider updated"})
}

func (s *Server) rulesetsGet(c *gin.Context) {
	rulesets, err := s.cfg.GetRuleSets()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"rule_sets": rulesets})
}

func (s *Server) rulesetsUpdate(c *gin.Context) {
	if err := s.coreMgr.UpdateRuleSets(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Rule sets updated"})
}
