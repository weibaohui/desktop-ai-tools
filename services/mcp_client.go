package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"desktop-ai-tools/models"
)

// MCPClient MCP客户端结构体
type MCPClient struct {
	client *client.Client
	url    string
}

// NewMCPClient 创建新的MCP客户端
func NewMCPClient(url string) *MCPClient {
	return &MCPClient{
		url: url,
	}
}

// Connect 连接到MCP服务器
func (c *MCPClient) Connect(ctx context.Context) error {
	// 创建SSE客户端
	mcpClient, err := client.NewSSEMCPClient(c.url)
	if err != nil {
		return fmt.Errorf("创建MCP客户端失败: %w", err)
	}
	
	c.client = mcpClient
	
	// 启动客户端
	err = c.client.Start(ctx)
	if err != nil {
		return fmt.Errorf("启动MCP客户端失败: %w", err)
	}
	
	// 初始化连接
	initRequest := mcp.InitializeRequest{
		Params: mcp.InitializeParams{
			ProtocolVersion: "2024-11-05",
			Capabilities: mcp.ClientCapabilities{
				Roots: &struct {
					ListChanged bool `json:"listChanged,omitempty"`
				}{
					ListChanged: true,
				},
			},
			ClientInfo: mcp.Implementation{
				Name:    "desktop-ai-tools",
				Version: "1.0.0",
			},
		},
	}
	
	_, err = c.client.Initialize(ctx, initRequest)
	if err != nil {
		return fmt.Errorf("初始化MCP客户端失败: %w", err)
	}
	
	return nil
}

// ListTools 获取可用工具列表
func (c *MCPClient) ListTools(ctx context.Context) ([]models.MCPTool, error) {
	if c.client == nil {
		return nil, fmt.Errorf("客户端未连接")
	}
	
	// 获取工具列表
	toolsResponse, err := c.client.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return nil, fmt.Errorf("获取工具列表失败: %w", err)
	}
	
	var tools []models.MCPTool
	for _, tool := range toolsResponse.Tools {
		// 解析参数
		var parameters []models.MCPToolParameter
		
		// 检查InputSchema是否有内容
		if tool.InputSchema.Type != "" || len(tool.InputSchema.Properties) > 0 {
			// 处理必需字段
			required := make(map[string]bool)
			for _, req := range tool.InputSchema.Required {
				required[req] = true
			}
			
			// 遍历属性
			for name, prop := range tool.InputSchema.Properties {
				if propMap, ok := prop.(map[string]interface{}); ok {
					param := models.MCPToolParameter{
						Name:        name,
						Type:        getStringValue(propMap, "type"),
						Description: getStringValue(propMap, "description"),
						Required:    required[name],
					}
					
					if defaultVal, exists := propMap["default"]; exists {
						param.Default = defaultVal
					}
					
					if enum, ok := propMap["enum"].([]interface{}); ok {
						for _, e := range enum {
							if enumStr, ok := e.(string); ok {
								param.Enum = append(param.Enum, enumStr)
							}
						}
					}
					
					parameters = append(parameters, param)
				}
			}
		}
		
		// 序列化参数
		parametersJSON, err := json.Marshal(parameters)
		if err != nil {
			log.Printf("序列化参数失败: %v", err)
			parametersJSON = []byte("[]")
		}
		
		mcpTool := models.MCPTool{
			Name:        tool.Name,
			Description: tool.Description,
			Parameters:  string(parametersJSON),
			IsEnabled:   true, // 默认启用
		}
		
		tools = append(tools, mcpTool)
	}
	
	return tools, nil
}

// CallTool 调用工具
func (c *MCPClient) CallTool(ctx context.Context, name string, arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	if c.client == nil {
		return nil, fmt.Errorf("客户端未连接")
	}
	
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      name,
			Arguments: arguments,
		},
	}
	
	result, err := c.client.CallTool(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("调用工具失败: %w", err)
	}
	
	return result, nil
}

// Close 关闭连接
func (c *MCPClient) Close() error {
	// mark3labs/mcp-go的客户端没有Close方法，这里暂时不做处理
	if c.client != nil {
		c.client = nil
	}
	return nil
}

// getStringValue 从map中获取字符串值的辅助函数
func getStringValue(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}
