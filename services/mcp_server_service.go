package services

import (
	"fmt"
	"strings"

	"gorm.io/gorm"

	"desktop-ai-tools/database"
	"desktop-ai-tools/models"
)

// MCPServerService MCP服务器服务
type MCPServerService struct {
	db *gorm.DB
}

// NewMCPServerService 创建新的MCP服务器服务实例
func NewMCPServerService() *MCPServerService {
	return &MCPServerService{
		db: database.GetDB(),
	}
}

// GetList 获取MCP服务器列表
func (s *MCPServerService) GetList(req *models.MCPServerListRequest) (*models.MCPServerListResponse, error) {
	var servers []models.MCPServer
	var total int64

	// 构建查询
	query := s.db.Model(&models.MCPServer{})

	// 搜索条件
	if req.Search != "" {
		searchTerm := "%" + req.Search + "%"
		query = query.Where("name LIKE ? OR description LIKE ? OR tags LIKE ?", 
			searchTerm, searchTerm, searchTerm)
	}

	// 状态过滤
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	// 启用状态过滤
	if req.Enabled != nil {
		query = query.Where("is_enabled = ?", *req.Enabled)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("获取总数失败: %v", err)
	}

	// 排序
	orderClause := fmt.Sprintf("%s %s", req.OrderBy, req.OrderDir)
	query = query.Order(orderClause)

	// 分页
	offset := (req.Page - 1) * req.Size
	if err := query.Offset(offset).Limit(req.Size).Find(&servers).Error; err != nil {
		return nil, fmt.Errorf("查询服务器列表失败: %v", err)
	}

	return &models.MCPServerListResponse{
		Total:   total,
		Page:    req.Page,
		Size:    req.Size,
		Servers: servers,
	}, nil
}

// GetByID 根据ID获取MCP服务器
func (s *MCPServerService) GetByID(id uint) (*models.MCPServer, error) {
	var server models.MCPServer
	if err := s.db.Preload("Tools").First(&server, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("服务器不存在")
		}
		return nil, fmt.Errorf("查询服务器失败: %v", err)
	}
	return &server, nil
}

// Create 创建MCP服务器
func (s *MCPServerService) Create(req *models.MCPServerCreateRequest) (*models.MCPServer, error) {
	// 检查名称是否重复
	var count int64
	if err := s.db.Model(&models.MCPServer{}).Where("name = ?", req.Name).Count(&count).Error; err != nil {
		return nil, fmt.Errorf("检查名称重复失败: %v", err)
	}
	if count > 0 {
		return nil, fmt.Errorf("服务器名称已存在")
	}

	// 创建服务器
	server := &models.MCPServer{
		Name:        req.Name,
		Description: req.Description,
		URL:         req.URL,
		AuthType:    req.AuthType,
		AuthConfig:  req.AuthConfig,
		Status:      "inactive", // 默认为非活跃状态
		IsEnabled:   true,       // 默认启用
		Tags:        req.Tags,
	}

	if err := s.db.Create(server).Error; err != nil {
		return nil, fmt.Errorf("创建服务器失败: %v", err)
	}

	return server, nil
}

// Update 更新MCP服务器
func (s *MCPServerService) Update(id uint, req *models.MCPServerUpdateRequest) (*models.MCPServer, error) {
	// 检查服务器是否存在
	var server models.MCPServer
	if err := s.db.First(&server, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("服务器不存在")
		}
		return nil, fmt.Errorf("查询服务器失败: %v", err)
	}

	// 检查名称是否重复（排除当前记录）
	var count int64
	if err := s.db.Model(&models.MCPServer{}).Where("name = ? AND id != ?", req.Name, id).Count(&count).Error; err != nil {
		return nil, fmt.Errorf("检查名称重复失败: %v", err)
	}
	if count > 0 {
		return nil, fmt.Errorf("服务器名称已存在")
	}

	// 更新字段
	updates := map[string]interface{}{
		"name":        req.Name,
		"description": req.Description,
		"url":         req.URL,
		"auth_type":   req.AuthType,
		"auth_config": req.AuthConfig,
		"tags":        req.Tags,
	}

	if req.IsEnabled != nil {
		updates["is_enabled"] = *req.IsEnabled
	}

	if err := s.db.Model(&server).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("更新服务器失败: %v", err)
	}

	// 重新查询更新后的数据
	if err := s.db.First(&server, id).Error; err != nil {
		return nil, fmt.Errorf("查询更新后的服务器失败: %v", err)
	}

	return &server, nil
}

// Delete 删除MCP服务器
func (s *MCPServerService) Delete(id uint) error {
	// 检查服务器是否存在
	var server models.MCPServer
	if err := s.db.First(&server, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("服务器不存在")
		}
		return fmt.Errorf("查询服务器失败: %v", err)
	}

	// 软删除服务器（同时删除关联的工具）
	if err := s.db.Select("Tools").Delete(&server).Error; err != nil {
		return fmt.Errorf("删除服务器失败: %v", err)
	}

	return nil
}

// UpdateStatus 更新服务器状态
func (s *MCPServerService) UpdateStatus(id uint, status string) error {
	// 验证状态值
	validStatuses := []string{"active", "inactive", "error"}
	isValid := false
	for _, validStatus := range validStatuses {
		if status == validStatus {
			isValid = true
			break
		}
	}
	if !isValid {
		return fmt.Errorf("无效的状态值: %s", status)
	}

	// 更新状态
	result := s.db.Model(&models.MCPServer{}).Where("id = ?", id).Update("status", status)
	if result.Error != nil {
		return fmt.Errorf("更新状态失败: %v", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("服务器不存在")
	}

	return nil
}

// ToggleEnabled 切换服务器启用状态
func (s *MCPServerService) ToggleEnabled(id uint) (*models.MCPServer, error) {
	var server models.MCPServer
	if err := s.db.First(&server, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("服务器不存在")
		}
		return nil, fmt.Errorf("查询服务器失败: %v", err)
	}

	// 切换启用状态
	newEnabled := !server.IsEnabled
	if err := s.db.Model(&server).Update("is_enabled", newEnabled).Error; err != nil {
		return nil, fmt.Errorf("更新启用状态失败: %v", err)
	}

	server.IsEnabled = newEnabled
	return &server, nil
}

// GetTags 获取所有标签
func (s *MCPServerService) GetTags() ([]string, error) {
	var servers []models.MCPServer
	if err := s.db.Select("tags").Where("tags != ''").Find(&servers).Error; err != nil {
		return nil, fmt.Errorf("查询标签失败: %v", err)
	}

	tagSet := make(map[string]bool)
	for _, server := range servers {
		if server.Tags != "" {
			tags := strings.Split(server.Tags, ",")
			for _, tag := range tags {
				tag = strings.TrimSpace(tag)
				if tag != "" {
					tagSet[tag] = true
				}
			}
		}
	}

	var result []string
	for tag := range tagSet {
		result = append(result, tag)
	}

	return result, nil
}