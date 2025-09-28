package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"desktop-ai-tools/models"
	"desktop-ai-tools/utils"

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

// fetchToolsFromMCPServer 从MCP服务器获取工具列表
func (s *MCPToolService) fetchToolsFromMCPServer(url, authType, authConfig string) ([]models.MCPTool, error) {
	// 构建请求
	reqBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/list",
		"params":  map[string]interface{}{},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("构建请求失败: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// 添加认证信息
	if err := s.addAuthentication(req, authType, authConfig); err != nil {
		return nil, fmt.Errorf("添加认证失败: %v", err)
	}

	// 发送请求
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("服务器返回错误状态: %d", resp.StatusCode)
	}

	// 解析响应
	var mcpResp struct {
		JSONRPC string `json:"jsonrpc"`
		ID      int    `json:"id"`
		Result  struct {
			Tools []models.MCPToolSchema `json:"tools"`
		} `json:"result"`
		Error *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&mcpResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	// 打印MCP服务器响应结果用于调试
	utils.PrintJSON(mcpResp, "MCP服务器响应")

	if mcpResp.Error != nil {
		return nil, fmt.Errorf("MCP服务器错误: %s", mcpResp.Error.Message)
	}

	// 转换为内部工具格式
	var tools []models.MCPTool
	for _, toolSchema := range mcpResp.Result.Tools {
		tool := models.MCPTool{
			Name:        toolSchema.Name,
			Description: toolSchema.Description,
			Category:    s.inferCategory(toolSchema.Name),
			IsEnabled:   true,
		}

		// 解析参数
		if params, err := s.parseToolParameters(toolSchema.InputSchema); err == nil {
			if paramData, err := json.Marshal(params); err == nil {
				tool.Parameters = string(paramData)
			}
		}

		tools = append(tools, tool)
	}

	return tools, nil
}

// addAuthentication 添加认证信息到请求
func (s *MCPToolService) addAuthentication(req *http.Request, authType, authConfig string) error {
	switch authType {
	case "bearer":
		var config struct {
			Token string `json:"token"`
		}
		if err := json.Unmarshal([]byte(authConfig), &config); err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+config.Token)
	case "basic":
		var config struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := json.Unmarshal([]byte(authConfig), &config); err != nil {
			return err
		}
		req.SetBasicAuth(config.Username, config.Password)
	case "api_key":
		var config struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		}
		if err := json.Unmarshal([]byte(authConfig), &config); err != nil {
			return err
		}
		req.Header.Set(config.Key, config.Value)
	}
	return nil
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
func (s *MCPToolService) RefreshAllTools(serverID uint) (*models.MCPToolDiscoveryResponse, error) {
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
			Message: "服务器未激活，无法刷新工具",
		}, nil
	}
	fmt.Printf("删除该服务器的所有现有工具 %d 的工具列表\n", uint(serverID))
	// 删除该服务器的所有现有工具
	if err := s.db.Where("server_id = ?", serverID).Delete(&models.MCPTool{}).Error; err != nil {
		return &models.MCPToolDiscoveryResponse{
			Success: false,
			Message: "删除现有工具失败",
		}, err
	}
	fmt.Printf("从MCP服务器 %s 获取最新的工具列表fetchToolsFromMCPServer\n", server.URL)
	// 从MCP服务器获取最新的工具列表
	tools, err := s.fetchToolsFromMCPServer(server.URL, server.AuthType, server.AuthConfig)
	if err != nil {
		fmt.Printf("从MCP服务器 %s 获取工具列表失败: %s\n", server.URL, err.Error())
		// 如果获取失败，直接返回错误，不再使用模拟数据
		return &models.MCPToolDiscoveryResponse{
			Success: false,
			Message: fmt.Sprintf("从MCP服务器获取工具列表失败: %s", err.Error()),
		}, err
	}
	fmt.Printf("成功从MCP服务器 %s 获取 %d 个工具\n", server.URL, len(tools))
	// 保存新的工具到数据库
	var savedCount int
	for _, tool := range tools {
		if err := s.db.Create(&tool).Error; err == nil {
			savedCount++
		}
	}

	return &models.MCPToolDiscoveryResponse{
		Success: true,
		Message: fmt.Sprintf("成功刷新 %d 个工具", savedCount),
		Tools:   tools,
	}, nil
}
