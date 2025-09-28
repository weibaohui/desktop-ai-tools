package models

import (
	"time"
	"gorm.io/gorm"
)

// MCPServer MCP服务器数据模型
type MCPServer struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"not null;size:100" binding:"required"`
	Description string         `json:"description" gorm:"size:500"`
	URL         string         `json:"url" gorm:"not null;size:255" binding:"required,url"`
	AuthType    string         `json:"auth_type" gorm:"size:50;default:'none'"` // none, bearer, basic, api_key
	AuthConfig  string         `json:"auth_config" gorm:"type:text"`            // JSON格式的认证配置
	Status      string         `json:"status" gorm:"size:20;default:'inactive'"` // active, inactive, error
	IsEnabled   bool           `json:"is_enabled" gorm:"default:true"`
	Tags        string         `json:"tags" gorm:"size:255"`                    // 逗号分隔的标签
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`
	
	// 关联的工具
	Tools []MCPTool `json:"tools,omitempty" gorm:"foreignKey:ServerID"`
}

// MCPTool MCP工具数据模型
type MCPTool struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	ServerID    uint           `json:"server_id" gorm:"not null"`
	Name        string         `json:"name" gorm:"not null;size:100"`
	Description string         `json:"description" gorm:"size:500"`
	Category    string         `json:"category" gorm:"size:50"`
	Parameters  string         `json:"parameters" gorm:"type:text"` // JSON格式的参数定义
	IsEnabled   bool           `json:"is_enabled" gorm:"default:true"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`
	
	// 关联的服务器
	Server MCPServer `json:"server,omitempty" gorm:"foreignKey:ServerID"`
}

// MCPServerCreateRequest 创建MCP服务器请求结构
type MCPServerCreateRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=100"`
	Description string `json:"description" binding:"max=500"`
	URL         string `json:"url" binding:"required,url"`
	AuthType    string `json:"auth_type" binding:"oneof=none bearer basic api_key"`
	AuthConfig  string `json:"auth_config"`
	Tags        string `json:"tags" binding:"max=255"`
}

// MCPServerUpdateRequest 更新MCP服务器请求结构
type MCPServerUpdateRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=100"`
	Description string `json:"description" binding:"max=500"`
	URL         string `json:"url" binding:"required,url"`
	AuthType    string `json:"auth_type" binding:"oneof=none bearer basic api_key"`
	AuthConfig  string `json:"auth_config"`
	IsEnabled   *bool  `json:"is_enabled"`
	Tags        string `json:"tags" binding:"max=255"`
}

// MCPServerListResponse 服务器列表响应结构
type MCPServerListResponse struct {
	Total   int64       `json:"total"`
	Page    int         `json:"page"`
	Size    int         `json:"size"`
	Servers []MCPServer `json:"servers"`
}

// MCPServerResponse 单个服务器响应结构
type MCPServerResponse struct {
	Success bool      `json:"success"`
	Message string    `json:"message"`
	Data    MCPServer `json:"data,omitempty"`
}

// MCPServerListRequest 服务器列表查询请求
type MCPServerListRequest struct {
	Page     int    `form:"page,default=1" binding:"min=1"`
	Size     int    `form:"size,default=10" binding:"min=1,max=100"`
	Search   string `form:"search"`
	Status   string `form:"status" binding:"omitempty,oneof=active inactive error"`
	Enabled  *bool  `form:"enabled"`
	OrderBy  string `form:"order_by,default=created_at" binding:"oneof=created_at updated_at name"`
	OrderDir string `form:"order_dir,default=desc" binding:"oneof=asc desc"`
}

// TableName 指定表名
func (MCPServer) TableName() string {
	return "mcp_servers"
}

// TableName 指定表名
func (MCPTool) TableName() string {
	return "mcp_tools"
}

// BeforeCreate 创建前的钩子函数
func (m *MCPServer) BeforeCreate(tx *gorm.DB) error {
	if m.AuthType == "" {
		m.AuthType = "none"
	}
	if m.Status == "" {
		m.Status = "inactive"
	}
	return nil
}

// GetTagList 获取标签列表
func (m *MCPServer) GetTagList() []string {
	if m.Tags == "" {
		return []string{}
	}
	// 这里可以实现标签分割逻辑
	return []string{m.Tags}
}