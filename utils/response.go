package utils

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// APIResponse 统一API响应结构
type APIResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
	Timestamp string      `json:"timestamp"`
}

// SuccessResponse 成功响应
func SuccessResponse(c *gin.Context, data interface{}, message ...string) {
	response := APIResponse{
		Success:   true,
		Data:      data,
		Timestamp: time.Now().Format(time.RFC3339),
	}
	
	if len(message) > 0 {
		response.Message = message[0]
	}
	
	c.JSON(http.StatusOK, response)
}

// ErrorResponse 错误响应
func ErrorResponse(c *gin.Context, statusCode int, err error, message ...string) {
	response := APIResponse{
		Success:   false,
		Error:     err.Error(),
		Timestamp: time.Now().Format(time.RFC3339),
	}
	
	if len(message) > 0 {
		response.Message = message[0]
	} else {
		// 根据状态码设置默认消息
		switch statusCode {
		case http.StatusBadRequest:
			response.Message = "请求参数错误"
		case http.StatusNotFound:
			response.Message = "资源未找到"
		case http.StatusInternalServerError:
			response.Message = "服务器内部错误"
		default:
			response.Message = "请求失败"
		}
	}
	
	// 使用Gin的Error方法记录错误，这样中间件可以捕获到
	c.Error(err)
	c.JSON(statusCode, response)
}

// BadRequestResponse 400错误响应
func BadRequestResponse(c *gin.Context, err error, message ...string) {
	ErrorResponse(c, http.StatusBadRequest, err, message...)
}

// NotFoundResponse 404错误响应
func NotFoundResponse(c *gin.Context, err error, message ...string) {
	ErrorResponse(c, http.StatusNotFound, err, message...)
}

// InternalServerErrorResponse 500错误响应
func InternalServerErrorResponse(c *gin.Context, err error, message ...string) {
	ErrorResponse(c, http.StatusInternalServerError, err, message...)
}

// ValidationErrorResponse 参数验证错误响应
func ValidationErrorResponse(c *gin.Context, err error) {
	BadRequestResponse(c, err, "请求参数验证失败")
}

// PanicResponse 处理panic的响应
func PanicResponse(c *gin.Context, recovered interface{}) {
	response := APIResponse{
		Success:   false,
		Message:   "服务器内部错误",
		Error:     "系统发生了意外错误，请稍后重试",
		Timestamp: time.Now().Format(time.RFC3339),
	}
	
	// 在开发环境下显示更多信息
	if gin.Mode() != gin.ReleaseMode {
		response.Data = gin.H{
			"panic": recovered,
			"debug_message": "系统发生panic错误",
		}
		response.Error = "系统发生panic错误，详细信息请查看data字段"
	}
	
	c.AbortWithStatusJSON(http.StatusInternalServerError, response)
}