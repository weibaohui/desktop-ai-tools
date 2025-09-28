package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"desktop-ai-tools/database"
	"desktop-ai-tools/middleware"
	"desktop-ai-tools/models"
	"desktop-ai-tools/services"
	"desktop-ai-tools/utils"
)

// App struct
type App struct {
	ctx              context.Context
	router           *gin.Engine
	mcpServerService *services.MCPServerService
	mcpToolService   *services.MCPToolService
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

	// 初始化数据库
	if err := database.InitDatabase(); err != nil {
		fmt.Printf("数据库初始化失败: %v\n", err)
		panic(err)
	}

	// 初始化种子数据
	if err := database.SeedData(); err != nil {
		fmt.Printf("种子数据初始化失败: %v\n", err)
	}

	// 初始化服务
	app.mcpServerService = services.NewMCPServerService()
	app.mcpToolService = services.NewMCPToolService(database.GetDB())

	app.setupRouter()
	return app
}

// setupRouter 设置Gin路由
func (a *App) setupRouter() {
	// 设置Gin为开发模式（用于测试详细错误信息）
	gin.SetMode(gin.DebugMode)

	a.router = gin.Default()

	// 添加错误处理中间件
	a.router.Use(middleware.ErrorHandler())
	a.router.Use(middleware.LogErrors())

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

		// 测试错误处理的端点
		api.GET("/test-error", a.handleTestError)
		api.GET("/test-panic", a.handleTestPanic)

		// MCP Server 相关路由
		mcpServers := api.Group("/mcp-servers")
		{
			mcpServers.GET("", a.handleGetMCPServers)
			mcpServers.POST("", a.handleCreateMCPServer)
			mcpServers.GET("/:id", a.handleGetMCPServer)
			mcpServers.PUT("/:id", a.handleUpdateMCPServer)
			mcpServers.DELETE("/:id", a.handleDeleteMCPServer)
			mcpServers.PUT("/:id/status", a.handleUpdateMCPServerStatus)
			mcpServers.PUT("/:id/toggle", a.handleToggleMCPServer)
			mcpServers.GET("/tags", a.handleGetMCPServerTags)

			// 工具发现路由
			mcpServers.POST("/:id/discover-tools", a.handleDiscoverTools)
		}

		// MCP Tools 相关路由
		mcpTools := api.Group("/mcp-tools")
		{
			mcpTools.GET("", a.handleGetMCPTools)
			mcpTools.PUT("/:id", a.handleUpdateMCPTool)
			mcpTools.PUT("/batch", a.handleBatchUpdateMCPTools)
			mcpTools.GET("/categories", a.handleGetMCPToolCategories)
			mcpTools.POST("/refresh/:serverID", a.handleRefreshTools)
		}
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
		"status":  "ok",
		"message": "Server is running",
	})
}

// handleGetMCPServers 获取MCP服务器列表
func (a *App) handleGetMCPServers(c *gin.Context) {
	var req models.MCPServerListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid query parameters",
			"success": false,
		})
		return
	}

	// 设置默认值
	if req.Page == 0 {
		req.Page = 1
	}
	if req.Size == 0 {
		req.Size = 10
	}
	if req.OrderBy == "" {
		req.OrderBy = "created_at"
	}
	if req.OrderDir == "" {
		req.OrderDir = "desc"
	}

	result, err := a.mcpServerService.GetList(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// handleCreateMCPServer 创建MCP服务器
func (a *App) handleCreateMCPServer(c *gin.Context) {
	var req models.MCPServerCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"success": false,
		})
		return
	}

	server, err := a.mcpServerService.Create(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    server,
		"message": "MCP服务器创建成功",
	})
}

// handleGetMCPServer 获取单个MCP服务器
func (a *App) handleGetMCPServer(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid server ID",
			"success": false,
		})
		return
	}

	server, err := a.mcpServerService.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    server,
	})
}

// handleUpdateMCPServer 更新MCP服务器
func (a *App) handleUpdateMCPServer(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid server ID",
			"success": false,
		})
		return
	}

	var req models.MCPServerUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"success": false,
		})
		return
	}

	server, err := a.mcpServerService.Update(uint(id), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    server,
		"message": "MCP服务器更新成功",
	})
}

// handleDeleteMCPServer 删除MCP服务器
func (a *App) handleDeleteMCPServer(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid server ID",
			"success": false,
		})
		return
	}

	err = a.mcpServerService.Delete(uint(id))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "MCP服务器删除成功",
	})
}

// handleUpdateMCPServerStatus 更新MCP服务器状态
func (a *App) handleUpdateMCPServerStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid server ID",
			"success": false,
		})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required,oneof=active inactive error"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"success": false,
		})
		return
	}

	err = a.mcpServerService.UpdateStatus(uint(id), req.Status)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "服务器状态更新成功",
	})
}

// handleToggleMCPServer 切换MCP服务器启用状态
func (a *App) handleToggleMCPServer(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid server ID",
			"success": false,
		})
		return
	}

	server, err := a.mcpServerService.ToggleEnabled(uint(id))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    server,
		"message": "服务器状态切换成功",
	})
}

// handleGetMCPServerTags 获取所有标签
func (a *App) handleGetMCPServerTags(c *gin.Context) {
	tags, err := a.mcpServerService.GetTags()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tags,
	})
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// handleDiscoverTools 处理工具发现请求
func (a *App) handleDiscoverTools(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的服务器ID",
		})
		return
	}

	response, err := a.mcpToolService.DiscoverTools(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": fmt.Sprintf("工具发现失败: %v", err),
		})
		return
	}

	if response.Success {
		c.JSON(http.StatusOK, response)
	} else {
		c.JSON(http.StatusBadRequest, response)
	}
}

// handleGetMCPTools 处理获取工具列表请求
func (a *App) handleGetMCPTools(c *gin.Context) {
	var req models.MCPToolListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	response, err := a.mcpToolService.GetToolsByServer(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取工具列表失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// handleUpdateMCPTool 处理更新工具请求
func (a *App) handleUpdateMCPTool(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的工具ID",
		})
		return
	}

	var req models.MCPToolUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	if err := a.mcpToolService.UpdateTool(uint(id), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "更新工具失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "工具更新成功",
	})
}

// handleBatchUpdateMCPTools 处理批量更新工具请求
func (a *App) handleBatchUpdateMCPTools(c *gin.Context) {
	var req models.MCPToolBatchUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	if err := a.mcpToolService.BatchUpdateTools(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "批量更新工具失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("成功更新 %d 个工具", len(req.ToolIDs)),
	})
}

// handleGetMCPToolCategories 处理获取工具分类请求
func (a *App) handleGetMCPToolCategories(c *gin.Context) {
	serverIDStr := c.Query("server_id")
	var serverID uint
	if serverIDStr != "" {
		id, err := strconv.ParseUint(serverIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "无效的服务器ID",
			})
			return
		}
		serverID = uint(id)
	}

	categories, err := a.mcpToolService.GetToolCategories(serverID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取工具分类失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    categories,
	})
}

// handleRefreshTools 处理刷新指定服务器的工具列表
func (a *App) handleRefreshTools(c *gin.Context) {
	serverIDStr := c.Param("serverID")
	serverID, err := strconv.ParseUint(serverIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的服务器ID",
		})
		return
	}
	fmt.Printf("刷新服务器 %d 的工具列表\n", uint(serverID))
	response, err := a.mcpToolService.RefreshAllTools(uint(serverID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "刷新工具失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// handleTestError 测试错误处理的端点
func (a *App) handleTestError(c *gin.Context) {
	errorType := c.Query("type")

	switch errorType {
	case "400":
		utils.BadRequestResponse(c, fmt.Errorf("这是一个400错误测试"), "请求参数错误测试")
	case "404":
		utils.NotFoundResponse(c, fmt.Errorf("这是一个404错误测试"), "资源未找到测试")
	case "500":
		utils.InternalServerErrorResponse(c, fmt.Errorf("这是一个500错误测试"), "服务器内部错误测试")
	default:
		utils.SuccessResponse(c, gin.H{
			"message": "错误测试端点正常工作",
			"usage":   "使用 ?type=400|404|500 来测试不同类型的错误",
		}, "测试成功")
	}
}

// handleTestPanic 测试panic处理的端点
func (a *App) handleTestPanic(c *gin.Context) {
	panic("这是一个测试panic")
}
