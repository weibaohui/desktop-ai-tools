package middleware

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"desktop-ai-tools/utils"
)

// ErrorResponse 统一错误响应结构
type ErrorResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message"`
	Error     string      `json:"error,omitempty"`
	Details   interface{} `json:"details,omitempty"`
	Timestamp string      `json:"timestamp"`
	Path      string      `json:"path"`
	Method    string      `json:"method"`
	Stack     []string    `json:"stack,omitempty"`
}

// ErrorHandlerConfig 错误处理配置
type ErrorHandlerConfig struct {
	// 是否显示详细错误信息（开发环境建议true，生产环境建议false）
	ShowDetails bool
	// 是否显示堆栈跟踪
	ShowStack bool
	// 是否记录错误日志
	LogErrors bool
}

// DefaultErrorHandlerConfig 默认错误处理配置
func DefaultErrorHandlerConfig() ErrorHandlerConfig {
	return ErrorHandlerConfig{
		ShowDetails: gin.Mode() != gin.ReleaseMode, // 非发布模式显示详细信息
		ShowStack:   gin.Mode() == gin.DebugMode,   // 调试模式显示堆栈
		LogErrors:   true,                          // 总是记录错误日志
	}
}

// ErrorHandler 创建错误处理中间件
func ErrorHandler(config ...ErrorHandlerConfig) gin.HandlerFunc {
	cfg := DefaultErrorHandlerConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	return func(c *gin.Context) {
		// 使用defer捕获panic
		defer func() {
			if err := recover(); err != nil {
				handlePanic(c, err, cfg)
			}
		}()

		// 继续处理请求
		c.Next()

		// 检查是否有错误
		if len(c.Errors) > 0 {
			handleErrors(c, c.Errors, cfg)
		}
	}
}

// handlePanic 处理panic错误
func handlePanic(c *gin.Context, err interface{}, cfg ErrorHandlerConfig) {
	// 获取堆栈信息
	stack := getStackTrace()
	
	// 记录错误日志
	if cfg.LogErrors {
		fmt.Printf("[ERROR] %s %s - Panic: %v\n", c.Request.Method, c.Request.URL.Path, err)
		if cfg.ShowStack {
			fmt.Printf("Stack trace:\n%s\n", strings.Join(stack, "\n"))
		}
	}
	
	// 使用统一的panic响应格式
	utils.PanicResponse(c, err)
}

// handleErrors 处理普通错误
func handleErrors(c *gin.Context, errors []*gin.Error, cfg ErrorHandlerConfig) {
	// 获取最后一个错误
	if len(errors) == 0 {
		return
	}
	lastError := errors[len(errors)-1]
	
	// 记录错误日志
	if cfg.LogErrors {
		fmt.Printf("[ERROR] %s %s - Error: %v\n", c.Request.Method, c.Request.URL.Path, lastError.Error())
	}

	// 如果已经写入了响应，则不再处理
	if c.Writer.Written() {
		return
	}

	// 构建错误响应
	response := ErrorResponse{
		Success:   false,
		Message:   "请求处理失败",
		Timestamp: time.Now().Format(time.RFC3339),
		Path:      c.Request.URL.Path,
		Method:    c.Request.Method,
	}

	// 根据配置决定是否显示详细信息
	if cfg.ShowDetails {
		response.Error = lastError.Error()
		response.Details = map[string]interface{}{
			"error_type":  fmt.Sprintf("%d", lastError.Type),
			"user_agent":  c.Request.UserAgent(),
			"remote_addr": c.ClientIP(),
		}
	}

	// 确定HTTP状态码
	statusCode := http.StatusInternalServerError
	if lastError.Type == gin.ErrorTypeBind {
		statusCode = http.StatusBadRequest
		response.Message = "请求参数错误"
	}

	// 设置响应头并返回JSON
	c.Header("Content-Type", "application/json")
	c.AbortWithStatusJSON(statusCode, response)
}

// getStackTrace 获取堆栈跟踪信息
func getStackTrace() []string {
	var stack []string
	
	// 跳过前几个调用栈帧（runtime相关的）
	for i := 3; i < 15; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		
		// 获取函数名
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}
		
		// 简化文件路径，只显示相对路径
		if idx := strings.LastIndex(file, "/"); idx >= 0 {
			file = file[idx+1:]
		}
		
		stack = append(stack, fmt.Sprintf("%s:%d %s", file, line, fn.Name()))
	}
	
	return stack
}

// RecoveryWithWriter 自定义恢复中间件，支持自定义日志输出
func RecoveryWithWriter() gin.HandlerFunc {
	return gin.CustomRecoveryWithWriter(nil, func(c *gin.Context, recovered interface{}) {
		// 使用我们的错误处理逻辑
		handlePanic(c, recovered, DefaultErrorHandlerConfig())
	})
}

// LogErrors 错误日志中间件
func LogErrors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		
		// 记录所有错误
		for _, err := range c.Errors {
			utils.PrintJSON(map[string]interface{}{
				"timestamp":   time.Now().Format(time.RFC3339),
				"method":      c.Request.Method,
				"path":        c.Request.URL.Path,
				"status_code": c.Writer.Status(),
				"error":       err.Error(),
				"error_type":  fmt.Sprintf("%d", err.Type),
				"client_ip":   c.ClientIP(),
				"user_agent":  c.Request.UserAgent(),
			}, "API错误日志")
		}
	}
}