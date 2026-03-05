package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("singbox-panel-secret-change-me")

type authConfig struct {
	PasswordHash string `json:"password_hash"`
	SetupDone    bool   `json:"setup_done"`
}

func (s *Server) authConfigPath() string {
	return filepath.Join(s.dataDir, "auth.json")
}

func (s *Server) loadAuthConfig() (*authConfig, error) {
	data, err := os.ReadFile(s.authConfigPath())
	if err != nil {
		if os.IsNotExist(err) {
			return &authConfig{}, nil
		}
		return nil, err
	}
	var cfg authConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (s *Server) saveAuthConfig(cfg *authConfig) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.authConfigPath(), data, 0600)
}

func hashPassword(pwd string) string {
	h := sha256.New()
	h.Write([]byte(pwd + "singbox-panel-salt"))
	return hex.EncodeToString(h.Sum(nil))
}

func (s *Server) authStatus(c *gin.Context) {
	cfg, err := s.loadAuthConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"setup_done": cfg.SetupDone})
}

func (s *Server) authSetup(c *gin.Context) {
	cfg, err := s.loadAuthConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if cfg.SetupDone {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Already setup"})
		return
	}
	var req struct {
		Password string `json:"password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cfg.PasswordHash = hashPassword(req.Password)
	cfg.SetupDone = true
	if err := s.saveAuthConfig(cfg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	token := s.generateToken()
	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (s *Server) authLogin(c *gin.Context) {
	var req struct {
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cfg, err := s.loadAuthConfig()
	if err != nil || !cfg.SetupDone {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not setup"})
		return
	}
	if hashPassword(req.Password) != cfg.PasswordHash {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid password"})
		return
	}
	token := s.generateToken()
	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (s *Server) generateToken() string {
	claims := jwt.MapClaims{
		"sub": "admin",
		"exp": time.Now().Add(24 * 7 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString(jwtSecret)
	return signed
}

func (s *Server) jwtMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "No token"})
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
			return
		}
		token, err := jwt.Parse(parts[1], func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return jwtSecret, nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}
		c.Next()
	}
}
