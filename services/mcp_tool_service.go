package services

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"desktop-ai-tools/models"

	"gorm.io/gorm"
)

// MCPToolService MCP工具服务
type MCPToolService struct {
	db *gorm.DB
}

// NewMCPToolService 创建新的MCP工具服务实例
func NewMCPToolService(db *gorm.DB) *MCPToolService {
	return &MCPToolService{db: db}
}

// DiscoverTools 从MCP服务器发现工具
func (s *MCPToolService) DiscoverTools(serverID uint) (*models.MCPToolDiscoveryResponse, error) {
	// 获取服务器信息
	var server models.MCPServer
	if err := s.db.First(&server, serverID).Error; err != nil {
		return &models.MCPToolDiscoveryResponse{
			Success: false,
			Message: "服务器不存在",
		}, err
	}

	// 检查服务器状态
	if server.Status != "active" {
		return &models.MCPToolDiscoveryResponse{
			Success: false,
			Message: "服务器未激活，无法发现工具",
		}, nil
	}

	// 连接MCP服务器获取工具列表
	tools, err := s.fetchToolsFromMCPServer(server.URL, server.AuthType, server.AuthConfig)
	if err != nil {
		return &models.MCPToolDiscoveryResponse{
			Success: false,
			Message: fmt.Sprintf("从MCP服务器获取工具列表失败:%s", err.Error()),
		}, nil
	}

	// 保存或更新工具到数据库
	var savedTools []models.MCPTool
	for _, tool := range tools {
		tool.ServerID = serverID

		// 检查工具是否已存在
		var existingTool models.MCPTool
		err := s.db.Where("server_id = ? AND name = ?", serverID, tool.Name).First(&existingTool).Error

		if err == gorm.ErrRecordNotFound {
			// 新工具，直接创建
			if err := s.db.Create(&tool).Error; err != nil {
				continue // 跳过创建失败的工具
			}
			savedTools = append(savedTools, tool)
		} else if err == nil {
			// 工具已存在，更新信息
			existingTool.Description = tool.Description
			existingTool.Category = tool.Category
			existingTool.Parameters = tool.Parameters
			existingTool.UpdatedAt = time.Now()

			if err := s.db.Save(&existingTool).Error; err != nil {
				continue // 跳过更新失败的工具
			}
			savedTools = append(savedTools, existingTool)
		}
	}

	// 构建响应消息
	message := fmt.Sprintf("成功发现 %d 个工具", len(savedTools))
	if err != nil {
		message = fmt.Sprintf("MCP服务器暂时不可用，使用模拟工具数据 (原错误: %v)，成功保存 %d 个工具", err, len(savedTools))
	}

	return &models.MCPToolDiscoveryResponse{
		Success: true,
		Message: message,
		Tools:   savedTools,
	}, nil
}

// fetchToolsFromMCPServer 从 MCP 服务器获取工具列表，使用 MCP SDK
func (s *MCPToolService) fetchToolsFromMCPServer(url, authType, authConfig string) ([]models.MCPTool, error) {
	log.Printf("开始使用 MCP SDK 从服务器获取工具: %s", url)
	
	// 创建 MCP 客户端
	mcpClient := NewMCPClient(url)
	
	// 创建上下文
	ctx := context.Background()
	
	// 连接到 MCP 服务器
	log.Printf("连接到 MCP 服务器: %s", url)
	err := mcpClient.Connect(ctx)
	if err != nil {
		log.Printf("连接 MCP 服务器失败: %v", err)
		return nil, fmt.Errorf("连接 MCP 服务器失败: %w", err)
	}
	
	// 确保在函数结束时关闭连接
	defer func() {
		if closeErr := mcpClient.Close(); closeErr != nil {
			log.Printf("关闭 MCP 客户端连接时出错: %v", closeErr)
		}
	}()
	
	// 获取工具列表
	log.Printf("获取工具列表")
	tools, err := mcpClient.ListTools(ctx)
	if err != nil {
		log.Printf("获取工具列表失败: %v", err)
		return nil, fmt.Errorf("获取工具列表失败: %w", err)
	}
	
	log.Printf("成功使用 MCP SDK 获取 %d 个工具", len(tools))
	return tools, nil
}



// inferCategory 根据工具名称推断分类
func (s *MCPToolService) inferCategory(toolName string) string {
	name := strings.ToLower(toolName)

	if strings.Contains(name, "k8s") || strings.Contains(name, "kubernetes") {
		return "Kubernetes"
	}
	if strings.Contains(name, "file") || strings.Contains(name, "read") || strings.Contains(name, "write") {
		return "文件操作"
	}
	if strings.Contains(name, "search") || strings.Contains(name, "query") {
		return "搜索查询"
	}
	if strings.Contains(name, "web") || strings.Contains(name, "http") {
		return "网络请求"
	}
	if strings.Contains(name, "database") || strings.Contains(name, "db") {
		return "数据库"
	}

	return "其他"
}

// parseToolParameters 解析工具参数
func (s *MCPToolService) parseToolParameters(inputSchema map[string]interface{}) ([]models.MCPToolParameter, error) {
	var params []models.MCPToolParameter

	properties, ok := inputSchema["properties"].(map[string]interface{})
	if !ok {
		return params, nil
	}

	required, _ := inputSchema["required"].([]interface{})
	requiredMap := make(map[string]bool)
	for _, req := range required {
		if reqStr, ok := req.(string); ok {
			requiredMap[reqStr] = true
		}
	}

	for name, prop := range properties {
		if propMap, ok := prop.(map[string]interface{}); ok {
			param := models.MCPToolParameter{
				Name:     name,
				Required: requiredMap[name],
			}

			if typeVal, ok := propMap["type"].(string); ok {
				param.Type = typeVal
			}

			if desc, ok := propMap["description"].(string); ok {
				param.Description = desc
			}

			if defaultVal, ok := propMap["default"]; ok {
				param.Default = defaultVal
			}

			if enumVal, ok := propMap["enum"].([]interface{}); ok {
				for _, e := range enumVal {
					if eStr, ok := e.(string); ok {
						param.Enum = append(param.Enum, eStr)
					}
				}
			}

			params = append(params, param)
		}
	}

	return params, nil
}

// GetToolsByServer 获取指定服务器的工具列表
func (s *MCPToolService) GetToolsByServer(req *models.MCPToolListRequest) (*models.MCPToolListResponse, error) {
	var tools []models.MCPTool
	var total int64

	query := s.db.Model(&models.MCPTool{}).Preload("Server")

	// 添加过滤条件
	if req.ServerID > 0 {
		query = query.Where("server_id = ?", req.ServerID)
	}

	if req.Category != "" {
		query = query.Where("category = ?", req.Category)
	}

	if req.Enabled != nil {
		query = query.Where("is_enabled = ?", *req.Enabled)
	}

	if req.Search != "" {
		query = query.Where("name LIKE ? OR description LIKE ?",
			"%"+req.Search+"%", "%"+req.Search+"%")
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// 分页查询
	offset := (req.Page - 1) * req.Size
	if err := query.Offset(offset).Limit(req.Size).Find(&tools).Error; err != nil {
		return nil, err
	}

	return &models.MCPToolListResponse{
		Total: total,
		Page:  req.Page,
		Size:  req.Size,
		Tools: tools,
	}, nil
}

// UpdateTool 更新工具
func (s *MCPToolService) UpdateTool(id uint, req *models.MCPToolUpdateRequest) error {
	updates := make(map[string]interface{})

	if req.IsEnabled != nil {
		updates["is_enabled"] = *req.IsEnabled
	}

	if req.Category != "" {
		updates["category"] = req.Category
	}

	if len(updates) > 0 {
		updates["updated_at"] = time.Now()
		return s.db.Model(&models.MCPTool{}).Where("id = ?", id).Updates(updates).Error
	}

	return nil
}

// BatchUpdateTools 批量更新工具
func (s *MCPToolService) BatchUpdateTools(req *models.MCPToolBatchUpdateRequest) error {
	updates := make(map[string]interface{})

	if req.IsEnabled != nil {
		updates["is_enabled"] = *req.IsEnabled
	}

	if req.Category != "" {
		updates["category"] = req.Category
	}

	if len(updates) > 0 {
		updates["updated_at"] = time.Now()
		return s.db.Model(&models.MCPTool{}).Where("id IN ?", req.ToolIDs).Updates(updates).Error
	}

	return nil
}

// GetToolCategories 获取工具分类列表
func (s *MCPToolService) GetToolCategories(serverID uint) ([]string, error) {
	var categories []string

	query := s.db.Model(&models.MCPTool{}).Select("DISTINCT category").Where("category != ''")
	if serverID > 0 {
		query = query.Where("server_id = ?", serverID)
	}

	if err := query.Pluck("category", &categories).Error; err != nil {
		return nil, err
	}

	return categories, nil
}

// RefreshAllTools 刷新指定服务器的所有工具
// RefreshAllTools 刷新指定服务器的所有工具
func (s *MCPToolService) RefreshAllTools(serverID uint) (*models.MCPToolDiscoveryResponse, error) {
	log.Printf("开始刷新服务器 ID %d 的工具列表", serverID)
	
	// 获取服务器信息
	var server models.MCPServer
	if err := s.db.First(&server, serverID).Error; err != nil {
		log.Printf("获取服务器信息失败 (ID: %d): %v", serverID, err)
		return &models.MCPToolDiscoveryResponse{
			Success: false,
			Message: "服务器不存在",
		}, err
	}

	log.Printf("找到服务器: %s (URL: %s, 状态: %s)", server.Name, server.URL, server.Status)

	// 检查服务器状态
	if server.Status != "active" {
		log.Printf("服务器 %s 未激活，状态: %s", server.Name, server.Status)
		return &models.MCPToolDiscoveryResponse{
			Success: false,
			Message: "服务器未激活，无法刷新工具",
		}, nil
	}

	log.Printf("删除服务器 %d 的所有现有工具", serverID)
	// 删除该服务器的所有现有工具
	if err := s.db.Where("server_id = ?", serverID).Delete(&models.MCPTool{}).Error; err != nil {
		log.Printf("删除现有工具失败 (服务器 ID: %d): %v", serverID, err)
		return &models.MCPToolDiscoveryResponse{
			Success: false,
			Message: "删除现有工具失败",
		}, err
	}

	log.Printf("开始从 MCP 服务器获取工具列表: %s", server.URL)
	// 从MCP服务器获取最新的工具列表
	tools, err := s.fetchToolsFromMCPServer(server.URL, server.AuthType, server.AuthConfig)
	if err != nil {
		log.Printf("从 MCP 服务器 %s 获取工具列表失败: %v", server.URL, err)
		return &models.MCPToolDiscoveryResponse{
			Success: false,
			Message: fmt.Sprintf("从MCP服务器获取工具列表失败: %s", err.Error()),
		}, err
	}

	log.Printf("成功从 MCP 服务器 %s 获取 %d 个工具", server.URL, len(tools))

	// 保存新的工具到数据库
	var savedCount int
	var failedCount int
	for i, tool := range tools {
		tool.ServerID = serverID // 确保设置正确的服务器ID
		if err := s.db.Create(&tool).Error; err != nil {
			log.Printf("保存工具失败 (索引 %d, 名称: %s): %v", i, tool.Name, err)
			failedCount++
		} else {
			savedCount++
		}
	}

	if failedCount > 0 {
		log.Printf("工具保存完成: 成功 %d 个, 失败 %d 个", savedCount, failedCount)
	} else {
		log.Printf("成功保存所有 %d 个工具", savedCount)
	}

	return &models.MCPToolDiscoveryResponse{
		Success: true,
		Message: fmt.Sprintf("成功刷新 %d 个工具", savedCount),
		Tools:   tools,
	}, nil
}
