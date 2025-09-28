package services

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"desktop-ai-tools/models"
	"github.com/modelcontextprotocol/go-sdk/src/client"
	"github.com/modelcontextprotocol/go-sdk/src/protocol"
	"github.com/modelcontextprotocol/go-sdk/src/transport"
)

// MCPClient MCP客户端封装
type MCPClient struct {
	client *client.Client
	url    string
}

// NewMCPClient 创建新的MCP客户端
func NewMCPClient(serverURL string) (*MCPClient, error) {
	// 解析URL以确定传输类型
	parsedURL, err := url.Parse(serverURL)
	if err != nil {
		return nil, fmt.Errorf("无效的服务器URL: %v", err)
	}

	var transport transport.Transport

	// 根据URL协议选择传输方式
	switch parsedURL.Scheme {
	case "http", "https":
		// 对于HTTP/HTTPS，使用SSE传输
		transport, err = transport.NewSSETransport(serverURL)
		if err != nil {
			return nil, fmt.Errorf("创建SSE传输失败: %v", err)
		}
	default:
		return nil, fmt.Errorf("不支持的协议: %s", parsedURL.Scheme)
	}

	// 创建MCP客户端
	mcpClient := client.NewClient(transport)

	return &MCPClient{
		client: mcpClient,
		url:    serverURL,
	}, nil
}

// Connect 连接到MCP服务器
func (c *MCPClient) Connect(ctx context.Context) error {
	// 连接到服务器
	err := c.client.Connect(ctx)
	if err != nil {
		return fmt.Errorf("连接MCP服务器失败: %v", err)
	}

	// 初始化客户端
	initRequest := &protocol.InitializeRequest{
		ProtocolVersion: "2024-11-05",
		Capabilities: &protocol.ClientCapabilities{
			Roots: &protocol.RootsCapability{
				ListChanged: true,
			},
		},
		ClientInfo: &protocol.Implementation{
			Name:    "desktop-ai-tools",
			Version: "1.0.0",
		},
	}

	_, err = c.client.Initialize(ctx, initRequest)
	if err != nil {
		return fmt.Errorf("初始化MCP客户端失败: %v", err)
	}

	return nil
}

// ListTools 获取工具列表
func (c *MCPClient) ListTools(ctx context.Context) ([]models.MCPTool, error) {
	// 调用tools/list方法
	response, err := c.client.ListTools(ctx, &protocol.ListToolsRequest{})
	if err != nil {
		return nil, fmt.Errorf("获取工具列表失败: %v", err)
	}

	// 转换为内部模型
	var tools []models.MCPTool
	for _, tool := range response.Tools {
		mcpTool := models.MCPTool{
			Name:        tool.Name,
			Description: tool.Description,
			Category:    c.inferCategory(tool.Name),
		}

		// 解析参数
		if tool.InputSchema != nil {
			parameters, err := c.parseToolParameters(tool.InputSchema)
			if err != nil {
				log.Printf("解析工具 %s 参数失败: %v", tool.Name, err)
				continue
			}
			mcpTool.Parameters = parameters
		}

		tools = append(tools, mcpTool)
	}

	return tools, nil
}

// CallTool 调用工具
func (c *MCPClient) CallTool(ctx context.Context, name string, arguments map[string]interface{}) (*protocol.CallToolResult, error) {
	request := &protocol.CallToolRequest{
		Name:      name,
		Arguments: arguments,
	}

	response, err := c.client.CallTool(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("调用工具 %s 失败: %v", name, err)
	}

	return response, nil
}

// Close 关闭连接
func (c *MCPClient) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

// inferCategory 推断工具类别
func (c *MCPClient) inferCategory(toolName string) string {
	toolName = strings.ToLower(toolName)
	
	if strings.Contains(toolName, "k8s") || strings.Contains(toolName, "kubernetes") {
		return "kubernetes"
	}
	if strings.Contains(toolName, "pod") {
		return "kubernetes"
	}
	if strings.Contains(toolName, "deploy") {
		return "kubernetes"
	}
	if strings.Contains(toolName, "service") {
		return "kubernetes"
	}
	if strings.Contains(toolName, "node") {
		return "kubernetes"
	}
	if strings.Contains(toolName, "namespace") {
		return "kubernetes"
	}
	if strings.Contains(toolName, "log") {
		return "monitoring"
	}
	if strings.Contains(toolName, "metric") {
		return "monitoring"
	}
	if strings.Contains(toolName, "file") {
		return "file"
	}
	if strings.Contains(toolName, "search") {
		return "search"
	}
	
	return "general"
}

// parseToolParameters 解析工具参数
func (c *MCPClient) parseToolParameters(inputSchema interface{}) ([]models.MCPToolParameter, error) {
	var parameters []models.MCPToolParameter
	
	// 将interface{}转换为map
	schemaMap, ok := inputSchema.(map[string]interface{})
	if !ok {
		return parameters, nil
	}
	
	// 获取properties
	properties, ok := schemaMap["properties"].(map[string]interface{})
	if !ok {
		return parameters, nil
	}
	
	// 获取required字段
	requiredFields := make(map[string]bool)
	if required, ok := schemaMap["required"].([]interface{}); ok {
		for _, field := range required {
			if fieldName, ok := field.(string); ok {
				requiredFields[fieldName] = true
			}
		}
	}
	
	// 解析每个参数
	for paramName, paramDef := range properties {
		paramDefMap, ok := paramDef.(map[string]interface{})
		if !ok {
			continue
		}
		
		param := models.MCPToolParameter{
			Name:     paramName,
			Required: requiredFields[paramName],
		}
		
		// 获取类型
		if paramType, ok := paramDefMap["type"].(string); ok {
			param.Type = paramType
		}
		
		// 获取描述
		if description, ok := paramDefMap["description"].(string); ok {
			param.Description = description
		}
		
		// 获取默认值
		if defaultValue, ok := paramDefMap["default"]; ok {
			param.DefaultValue = fmt.Sprintf("%v", defaultValue)
		}
		
		parameters = append(parameters, param)
	}
	
	return parameters, nil
}