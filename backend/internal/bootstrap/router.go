package bootstrap

import (
	"genbi-go-backend/internal/config"
	"genbi-go-backend/internal/handler"
	"genbi-go-backend/internal/middleware"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RouterDeps 是装配路由所需的依赖集合。
type RouterDeps struct {
	Cfg          *config.Config
	Store        sessions.Store
	DB           *gorm.DB
	UserHandler  *handler.UserHandler
	ChartHandler *handler.ChartHandler
}

// NewRouter 装配 gin 引擎：全局中间件 + /api 路由组（health + user）。
func NewRouter(deps RouterDeps) *gin.Engine {
	r := gin.New()
	r.Use(middleware.Recovery(), middleware.Logger(), sessionMiddleware(deps.Store))

	auth := middleware.Auth(deps.DB)
	admin := middleware.Admin()

	api := r.Group("/api")
	api.GET("/health", handler.Health)

	// 用户路由
	u := api.Group("/user")
	{
		// 免登录
		u.POST("/register", deps.UserHandler.Register)
		u.POST("/login", deps.UserHandler.Login)
		// 需登录
		u.POST("/logout", auth, deps.UserHandler.Logout)
		u.GET("/get/login", auth, deps.UserHandler.GetLoginUser)
		u.POST("/update/my", auth, deps.UserHandler.UpdateMy)
		u.GET("/get/vo", auth, deps.UserHandler.GetVOByID)
		u.POST("/list/page/vo", auth, deps.UserHandler.PageVO)
		// 需管理员
		u.POST("/add", auth, admin, deps.UserHandler.Add)
		u.POST("/delete", auth, admin, deps.UserHandler.Delete)
		u.POST("/update", auth, admin, deps.UserHandler.Update)
		u.GET("/get", auth, admin, deps.UserHandler.GetByID)
		u.POST("/list/page", auth, admin, deps.UserHandler.Page)
	}

	// 图表路由（均需登录）
	ch := api.Group("/chart", auth)
	{
		ch.POST("/add", deps.ChartHandler.Add)
		ch.POST("/delete", deps.ChartHandler.Delete)
		ch.GET("/get", deps.ChartHandler.GetByID)
		ch.POST("/list/page", deps.ChartHandler.Page)
		ch.POST("/my/list/page", deps.ChartHandler.MyPage)
		ch.POST("/edit", deps.ChartHandler.Edit)
		ch.POST("/update", admin, deps.ChartHandler.Update)
		ch.POST("/gen", deps.ChartHandler.Gen)
		ch.POST("/gen/async", deps.ChartHandler.GenAsync)
	}

	return r
}
