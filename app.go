package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// App struct
type App struct {
	ctx    context.Context
	router *gin.Engine
}

// HelloRequest 请求结构体
type HelloRequest struct {
	Message string `json:"message" binding:"required"`
}

// HelloResponse 响应结构体
type HelloResponse struct {
	Response string `json:"response"`
	Success  bool   `json:"success"`
}

// NewApp creates a new App application struct
func NewApp() *App {
	app := &App{}
	app.setupRouter()
	return app
}

// setupRouter 设置Gin路由
func (a *App) setupRouter() {
	// 设置Gin为发布模式
	gin.SetMode(gin.ReleaseMode)
	
	a.router = gin.Default()
	
	// 配置CORS
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	a.router.Use(cors.New(config))
	
	// 设置API路由
	api := a.router.Group("/api")
	{
		api.POST("/hello", a.handleHello)
		api.GET("/health", a.handleHealth)
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	
	// 启动Gin服务器
	go func() {
		if err := a.router.Run(":8080"); err != nil {
			fmt.Printf("Failed to start server: %v\n", err)
		}
	}()
}

// handleHello 处理Hello请求的API接口
func (a *App) handleHello(c *gin.Context) {
	var req HelloRequest
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"success": false,
		})
		return
	}
	
	// 处理Hello逻辑
	response := fmt.Sprintf("Hello from backend! You sent: %s", req.Message)
	
	c.JSON(http.StatusOK, HelloResponse{
		Response: response,
		Success:  true,
	})
}

// handleHealth 健康检查接口
func (a *App) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"message": "Server is running",
	})
}

// Greet returns a greeting for the given name (保留原有的Wails方法)
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}
