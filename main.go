package main

import (
	"html/template"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"

	"fool_blog_go/internal/config"
	"fool_blog_go/internal/database"
	"fool_blog_go/internal/handler"
	"fool_blog_go/internal/middleware"
)

func main() {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	db, err := database.Init(cfg)
	if err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}

	r := gin.New()
	r.Use(gin.Recovery(), middleware.RequestLogger())

	// Session
	store := cookie.NewStore([]byte(cfg.Security.SessionKey))
	store.Options(sessions.Options{Path: "/", MaxAge: 7 * 24 * 3600, HttpOnly: true})
	r.Use(sessions.Sessions("foolblog", store))
	r.Use(middleware.LoadCurrentUser(db))
	r.Use(middleware.TrackVisitor(db))

	r.SetFuncMap(template.FuncMap{
		"safehtml":   func(s string) template.HTML { return template.HTML(s) },
		"safeurl":    func(s string) template.URL { return template.URL(s) },
		"formatTime": func(t time.Time) string { return t.Format("2006-01-02 15:04") },
		"formatInputTime": func(t *time.Time) string {
			if t == nil || t.IsZero() {
				return ""
			}
			return t.Format("2006-01-02T15:04")
		},
		"formatDateOnly": func(t time.Time) string { return t.Format("01-02") },
		"add":            func(a, b int) int { return a + b },
		"sub":            func(a, b int) int { return a - b },
		"firstChar": func(s string) string {
			if s == "" {
				return "?"
			}
			runes := []rune(s)
			return string(runes[0])
		},
	})
	loadTemplates(r)

	r.Static("/static", "static")
	r.Static("/uploads", "uploads")

	h := handler.New(db, cfg)

	// 前台
	r.GET("/", h.Index())
	r.GET("/article/:slug", h.ArticleDetail())
	r.GET("/search", h.Search())
	r.GET("/tags", h.TagList())
	r.GET("/tag/:id", h.TagDetail())
	r.GET("/archive", h.Archive())
	r.POST("/article/:id/like", h.LikeArticle())
	r.POST("/article/:id/comment", middleware.AuthRequired(), h.PostComment())
	r.GET("/api/ai-review/:id", h.AIReview())

	// 登录注册
	r.GET("/login", h.LoginPage())
	r.POST("/login", h.LoginSubmit())
	r.GET("/register", h.RegisterPage())
	r.POST("/register", h.RegisterSubmit())
	r.POST("/register/send-code", h.SendVerifyCode())
	r.POST("/logout", h.Logout())

	// 个人中心
	me := r.Group("/me", middleware.AuthRequired())
	{
		me.GET("", h.ProfilePage())
		me.POST("", h.ProfileUpdate())
		me.POST("/avatar", h.UploadAvatar())
		me.POST("/avatar/random", h.RandomAvatar())
	}

	// 后台
	admin := r.Group("/admin", middleware.AdminRequired())
	{
		admin.GET("", h.Dashboard())

		admin.GET("/articles", h.AdminArticles())
		admin.GET("/articles/new", h.AdminArticleNew())
		admin.POST("/articles", h.AdminArticleCreate())
		admin.GET("/articles/:id/edit", h.AdminArticleEdit())
		admin.POST("/articles/:id", h.AdminArticleUpdate())
		admin.POST("/articles/:id/delete", h.AdminArticleDelete())
		admin.POST("/articles/import", h.AdminArticleImport())
		admin.GET("/articles/:id/export", h.AdminArticleExport())
		admin.POST("/articles/export/zip", h.AdminArticleExportZip())

		admin.GET("/tags", h.AdminTags())
		admin.POST("/tags", h.AdminTagCreate())
		admin.POST("/tags/:id", h.AdminTagUpdate())
		admin.POST("/tags/:id/delete", h.AdminTagDelete())

		admin.GET("/comments", h.AdminComments())
		admin.POST("/comments/:id/delete", h.AdminCommentDelete())

		admin.GET("/users", h.AdminUsers())
		admin.POST("/users/:id/role", h.AdminUserRole())
		admin.POST("/users/:id/status", h.AdminUserStatus())
		admin.POST("/users/:id/delete", h.AdminUserDelete())

		admin.GET("/site", h.AdminSite())
		admin.POST("/site", h.AdminSiteUpdate())

		admin.GET("/logs", h.AdminLogs())

		admin.POST("/upload/article", h.UploadArticleImage())
		admin.POST("/upload/cover", h.UploadCover())
	}

	log.Printf("[Fool Blog] 监听 %s", cfg.App.Addr)
	if err := r.Run(cfg.App.Addr); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
	_ = db
}

func loadTemplates(r *gin.Engine) {
	patterns := []string{
		"templates/*.html",
		"templates/admin/*.html",
	}
	tmpl := template.New("").Funcs(r.FuncMap)
	for _, p := range patterns {
		files, err := filepath.Glob(p)
		if err != nil {
			log.Fatalf("模板扫描失败: %v", err)
		}
		for _, f := range files {
			b, err := os.ReadFile(f)
			if err != nil {
				log.Fatalf("读取模板失败 %s: %v", f, err)
			}
			tmpl, err = tmpl.Parse(string(b))
			if err != nil {
				log.Fatalf("解析模板失败 %s: %v", f, err)
			}
		}
	}
	r.SetHTMLTemplate(tmpl)
}
