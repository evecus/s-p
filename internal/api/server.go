package api

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/singbox-panel/internal/config"
	"github.com/singbox-panel/internal/core"
	"github.com/singbox-panel/internal/firewall"
)

type Server struct {
	dataDir string
	webFS   embed.FS
	cfg     *config.Manager
	coreMgr *core.Manager
	fwMgr   *firewall.Manager
}

func NewServer(dataDir string, webFS embed.FS) *Server {
	cfgMgr := config.NewManager(dataDir)
	return &Server{
		dataDir: dataDir,
		webFS:   webFS,
		cfg:     cfgMgr,
		coreMgr: core.NewManager(dataDir, cfgMgr),
		fwMgr:   firewall.NewManager(dataDir),
	}
}

func (s *Server) Run(addr string) error {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// 必须关闭，否则 Gin 路由引擎会在 NoRoute 执行前
	// 对未注册路径发出 301，导致死循环
	r.RedirectTrailingSlash = false
	r.RedirectFixedPath = false

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

	// 服务内嵌的 Vue SPA
	distFS, err := fs.Sub(s.webFS, "dist")
	if err != nil {
		panic("embedded dist/ not found, please build frontend first: cd web && npm run build")
	}
	httpFS := http.FS(distFS)

	// 静态文件 handler：先找真实文件，找不到就返回 index.html（SPA 路由）
	spaHandler := func(c *gin.Context) {
		cleanPath := c.Request.URL.Path[1:] // 去掉开头的 /
		if cleanPath == "" {
			cleanPath = "index.html"
		}
		f, err := distFS.Open(cleanPath)
		if err == nil {
			f.Close()
			c.FileFromFS(cleanPath, httpFS)
			return
		}
		c.FileFromFS("index.html", httpFS)
	}

	// 必须显式注册 /，否则 Gin 不会走 NoRoute 处理根路径
	r.GET("/", spaHandler)
	// 其余所有未匹配路径（/login /dashboard /assets/xxx 等）
	r.NoRoute(spaHandler)

	return r.Run(addr)
}
