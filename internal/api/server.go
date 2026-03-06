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
	// 用 http.FileServer + fs.Sub，绝对不用 FileFromFS
	// FileFromFS 内部修改 URL.Path 后交给 http.FileServer，
	// 而 http.FileServer 对不以 / 开头的路径会发 301，导致死循环
	distFS, err := fs.Sub(s.webFS, "dist")
	if err != nil {
		panic("embedded dist/ not found")
	}

	fileServer := http.FileServer(http.FS(distFS))

	spaHandler := func(c *gin.Context) {
		path := c.Request.URL.Path[1:] // 去掉开头的 /

		// 检查文件是否存在于 embed.FS
		_, err := distFS.Open(path)
		if err != nil {
			// 文件不存在（Vue Router 路由），改写路径到 index.html
			c.Request.URL.Path = "/"
		}

		// 用标准 http.FileServer 服务，它会正确处理 Content-Type、缓存头等
		// URL.Path 以 / 开头，http.FileServer 不会再发 301
		fileServer.ServeHTTP(c.Writer, c.Request)
	}

	r.GET("/", spaHandler)
	r.NoRoute(spaHandler)

	return r.Run(addr)
}
