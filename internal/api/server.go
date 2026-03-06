package api

import (
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/singbox-panel/internal/config"
	"github.com/singbox-panel/internal/core"
	"github.com/singbox-panel/internal/firewall"
)

type Server struct {
	dataDir   string
	staticDir string // 静态文件目录路径，空字符串表示无前端
	cfg       *config.Manager
	coreMgr   *core.Manager
	fwMgr     *firewall.Manager
}

func NewServer(dataDir string, staticDir string) *Server {
	cfgMgr := config.NewManager(dataDir)
	return &Server{
		dataDir:   dataDir,
		staticDir: staticDir,
		cfg:       cfgMgr,
		coreMgr:   core.NewManager(dataDir, cfgMgr),
		fwMgr:     firewall.NewManager(dataDir),
	}
}

func (s *Server) Run(addr string) error {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	// API routes
	api := r.Group("/api")
	{
		// Auth
		auth := api.Group("/auth")
		auth.GET("/status", s.authStatus)
		auth.POST("/setup", s.authSetup)
		auth.POST("/login", s.authLogin)

		// Protected routes
		protected := api.Group("/")
		protected.Use(s.jwtMiddleware())
		{
			// System
			protected.GET("/system/info", s.systemInfo)
			protected.GET("/system/status", s.systemStatus)

			// Core (sing-box binary)
			protected.GET("/core/info", s.coreInfo)
			protected.POST("/core/download", s.coreDownload)
			protected.GET("/core/download/progress", s.coreDownloadProgress)
			protected.POST("/core/start", s.coreStart)
			protected.POST("/core/stop", s.coreStop)
			protected.POST("/core/restart", s.coreRestart)
			protected.GET("/core/logs", s.coreLogs)

			// Config
			protected.GET("/config/raw", s.configGetRaw)
			protected.PUT("/config/raw", s.configSetRaw)
			protected.GET("/config/sections", s.configGetSections)
			protected.PUT("/config/sections/:section", s.configSetSection)
			protected.POST("/config/validate", s.configValidate)

			// Providers
			protected.GET("/providers", s.providersGet)
			protected.PUT("/providers", s.providersSet)
			protected.POST("/providers/:tag/update", s.providerUpdate)

			// Firewall / Proxy Mode
			protected.GET("/proxy/mode", s.proxyModeGet)
			protected.POST("/proxy/apply", s.proxyApply)
			protected.POST("/proxy/stop", s.proxyStop)
			protected.GET("/proxy/status", s.proxyStatus)

			// Rule sets
			protected.GET("/rulesets", s.rulesetsGet)
			protected.POST("/rulesets/update", s.rulesetsUpdate)
		}
	}

	// 服务前端静态文件
	if s.staticDir != "" {
		fileServer := http.FileServer(http.Dir(s.staticDir))
		r.NoRoute(func(c *gin.Context) {
			path := c.Request.URL.Path

			// 检查文件是否存在
			fullPath := s.staticDir + path
			if info, err := os.Stat(fullPath); err == nil && !info.IsDir() {
				// 文件存在，直接服务
				fileServer.ServeHTTP(c.Writer, c.Request)
				return
			}

			// 文件不存在，回退到 index.html（SPA 路由）
			indexPath := s.staticDir + "/index.html"
			if _, err := os.Stat(indexPath); err == nil {
				http.ServeFile(c.Writer, c.Request, indexPath)
			} else {
				c.String(http.StatusNotFound, "index.html not found")
			}
		})
	} else {
		// 没有前端文件时，返回提示页面而不是 301 循环
		r.NoRoute(func(c *gin.Context) {
			if c.Request.URL.Path == "/" || c.Request.URL.Path == "" {
				c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(`<!DOCTYPE html>
<html lang="zh-CN">
<head><meta charset="UTF-8"><title>Singbox Panel</title>
<style>body{font-family:sans-serif;background:#0f172a;color:#94a3b8;display:flex;align-items:center;justify-content:center;min-height:100vh;margin:0;}
.box{text-align:center;padding:2rem;background:#1e293b;border-radius:1rem;border:1px solid #334155;}
h1{color:#818cf8;margin-bottom:1rem;}code{background:#0f172a;padding:.2em .6em;border-radius:.3rem;color:#7dd3fc;}</style>
</head>
<body><div class="box">
<h1>🚀 Singbox Panel</h1>
<p>API 服务运行正常，但前端文件未找到。</p>
<p>请先构建前端：</p>
<pre><code>cd web && npm install && npm run build</code></pre>
<p>然后重启程序。</p>
</div></body></html>`))
			} else {
				c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			}
		})
	}

	return r.Run(addr)
}
