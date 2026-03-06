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

// responseCapture 拦截 http.FileServer 的响应状态码
type responseCapture struct {
	http.ResponseWriter
	status int
}

func (r *responseCapture) WriteHeader(status int) {
	r.status = status
	// 先不真正写，等我们决定是否替换
}

func (r *responseCapture) Written() bool {
	return r.status != 0
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

	// 直接用标准库的 StripPrefix + FileServer
	// 让它自己处理路径，不做任何手动路径拼接
	fileServer := http.FileServer(http.FS(distFS))

	// 读取 index.html 内容，用于 SPA 回退
	indexHTML, err := fs.ReadFile(distFS, "index.html")
	if err != nil {
		panic("embedded dist/index.html not found")
	}

	spaHandler := func(c *gin.Context) {
		path := c.Request.URL.Path

		// 根路径直接返回 index.html，避免任何重定向
		if path == "/" || path == "" {
			c.Data(http.StatusOK, "text/html; charset=utf-8", indexHTML)
			return
		}

		// 其他路径：用 http.FileServer 尝试服务静态文件
		// 用一个自定义 ResponseWriter 来检测 404
		cw := &captureWriter{ResponseWriter: c.Writer, code: 0}
		fileServer.ServeHTTP(cw, c.Request)

		// 如果 FileServer 找不到文件（404），回退到 index.html（Vue Router路由）
		if cw.code == http.StatusNotFound {
			c.Data(http.StatusOK, "text/html; charset=utf-8", indexHTML)
		}
	}

	r.GET("/", spaHandler)
	r.NoRoute(spaHandler)

	return r.Run(addr)
}

// captureWriter 捕获 http.FileServer 写出的状态码
// 如果是 404 则丢弃，让我们返回 index.html
type captureWriter struct {
	gin.ResponseWriter
	code    int
	written bool
}

func (cw *captureWriter) WriteHeader(code int) {
	cw.code = code
	if code != http.StatusNotFound {
		cw.ResponseWriter.WriteHeader(code)
		cw.written = true
	}
}

func (cw *captureWriter) Write(b []byte) (int, error) {
	if cw.code == http.StatusNotFound {
		// 丢弃 404 的响应体
		return len(b), nil
	}
	cw.written = true
	return cw.ResponseWriter.Write(b)
}
