package models

import (
	"encoding/json"
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

// MCPToolDiscoveryRequest 工具发现请求
type MCPToolDiscoveryRequest struct {
	ServerID uint `json:"server_id" binding:"required"`
}

// MCPToolDiscoveryResponse 工具发现响应
type MCPToolDiscoveryResponse struct {
	Success bool      `json:"success"`
	Message string    `json:"message"`
	Tools   []MCPTool `json:"tools,omitempty"`
}

// MCPToolListRequest 工具列表查询请求
type MCPToolListRequest struct {
	ServerID uint   `form:"server_id"`
	Category string `form:"category"`
	Enabled  *bool  `form:"enabled"`
	Search   string `form:"search"`
	Page     int    `form:"page,default=1" binding:"min=1"`
	Size     int    `form:"size,default=50" binding:"min=1,max=100"`
}

// MCPToolListResponse 工具列表响应
type MCPToolListResponse struct {
	Total int64     `json:"total"`
	Page  int       `json:"page"`
	Size  int       `json:"size"`
	Tools []MCPTool `json:"tools"`
}

// MCPToolUpdateRequest 工具更新请求
type MCPToolUpdateRequest struct {
	IsEnabled *bool  `json:"is_enabled"`
	Category  string `json:"category" binding:"max=50"`
}

// MCPToolBatchUpdateRequest 工具批量更新请求
type MCPToolBatchUpdateRequest struct {
	ToolIDs   []uint `json:"tool_ids" binding:"required"`
	IsEnabled *bool  `json:"is_enabled"`
	Category  string `json:"category" binding:"max=50"`
}

// MCPToolParameter 工具参数结构（用于解析Parameters字段）
type MCPToolParameter struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Required    bool        `json:"required"`
	Default     interface{} `json:"default,omitempty"`
	Enum        []string    `json:"enum,omitempty"`
}

// MCPToolSchema 工具完整模式（从MCP服务器获取的原始数据）
type MCPToolSchema struct {
	Name        string             `json:"name"`
	Description string             `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// TableName 指定表名
func (MCPServer) TableName() string {
	return "mcp_servers"
}

// TableName 指定表名
func (MCPTool) TableName() string {
	return "mcp_tools"
}

// GetParameters 解析工具参数
func (m *MCPTool) GetParameters() ([]MCPToolParameter, error) {
	if m.Parameters == "" {
		return []MCPToolParameter{}, nil
	}
	
	var params []MCPToolParameter
	if err := json.Unmarshal([]byte(m.Parameters), &params); err != nil {
		return nil, err
	}
	return params, nil
}

// SetParameters 设置工具参数
func (m *MCPTool) SetParameters(params []MCPToolParameter) error {
	data, err := json.Marshal(params)
	if err != nil {
		return err
	}
	m.Parameters = string(data)
	return nil
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