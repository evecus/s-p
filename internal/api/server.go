package api

import (
    "embed"
    "io/fs"
    "net/http"
    "strings"

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
		auth := api.Group("/auth")
		auth.GET("/status", s.authStatus)
		auth.POST("/setup", s.authSetup)
		auth.POST("/login", s.authLogin)

		protected := api.Group("/")
		protected.Use(s.jwtMiddleware())
		{
			protected.GET("/system/info", s.systemInfo)
			protected.GET("/system/status", s.systemStatus)
			protected.GET("/core/info", s.coreInfo)
			protected.POST("/core/download", s.coreDownload)
			protected.GET("/core/download/progress", s.coreDownloadProgress)
			protected.POST("/core/start", s.coreStart)
			protected.POST("/core/stop", s.coreStop)
			protected.POST("/core/restart", s.coreRestart)
			protected.GET("/core/logs", s.coreLogs)
			protected.GET("/config/raw", s.configGetRaw)
			protected.PUT("/config/raw", s.configSetRaw)
			protected.GET("/config/sections", s.configGetSections)
			protected.PUT("/config/sections/:section", s.configSetSection)
			protected.POST("/config/validate", s.configValidate)
			protected.GET("/providers", s.providersGet)
			protected.PUT("/providers", s.providersSet)
			protected.POST("/providers/:tag/update", s.providerUpdate)
			protected.GET("/proxy/mode", s.proxyModeGet)
			protected.POST("/proxy/apply", s.proxyApply)
			protected.POST("/proxy/stop", s.proxyStop)
			protected.GET("/proxy/status", s.proxyStatus)
			protected.GET("/rulesets", s.rulesetsGet)
			protected.POST("/rulesets/update", s.rulesetsUpdate)
		}
	}

	// 服务内嵌的 Vue SPA
	distFS, err := fs.Sub(s.webFS, "dist")
	if err != nil {
		panic("embedded dist/ not found")
	}

	fileServer := http.FileServer(http.FS(distFS))

	// 预读 index.html，用于 SPA 路由回退
	indexHTML, err := fs.ReadFile(distFS, "index.html")
	if err != nil {
		panic("embedded dist/index.html not found")
	}

	// 判断路径是否是真实静态文件（以 /assets/ 开头或是已知静态资源后缀）
	isStaticFile := func(path string) bool {
		if strings.HasPrefix(path, "/assets/") {
			return true
		}
		staticExts := []string{".js", ".css", ".ico", ".png", ".jpg", ".svg", ".woff", ".woff2", ".ttf"}
		for _, ext := range staticExts {
			if strings.HasSuffix(path, ext) {
				return true
			}
		}
		return false
	}

	spaHandler := func(c *gin.Context) {
		path := c.Request.URL.Path

		if isStaticFile(path) {
			// 静态资源：直接交给 fileServer，它能正确处理 Content-Type
			fileServer.ServeHTTP(c.Writer, c.Request)
			return
		}

		// 其他路径（/, /dashboard, /login 等）：直接返回 index.html
		// 不经过任何 fileServer，避免 Content-Type 被污染
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.Status(http.StatusOK)
		_, _ = c.Writer.Write(indexHTML)
	}

	r.GET("/", spaHandler)
	r.NoRoute(spaHandler)

	return r.Run(addr)
}
