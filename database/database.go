package database

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"desktop-ai-tools/models"
)

var DB *gorm.DB

// InitDatabase 初始化数据库连接
func InitDatabase() error {
	// 获取用户主目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("获取用户主目录失败: %v", err)
	}

	// 创建应用数据目录
	appDataDir := filepath.Join(homeDir, ".desktop-ai-tools")
	if err := os.MkdirAll(appDataDir, 0755); err != nil {
		return fmt.Errorf("创建应用数据目录失败: %v", err)
	}

	// 数据库文件路径
	dbPath := filepath.Join(appDataDir, "app.db")

	// 配置GORM日志
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	// 连接数据库
	db, err := gorm.Open(sqlite.Open(dbPath), gormConfig)
	if err != nil {
		return fmt.Errorf("连接数据库失败: %v", err)
	}

	DB = db

	// 自动迁移数据库表
	if err := autoMigrate(); err != nil {
		return fmt.Errorf("数据库迁移失败: %v", err)
	}

	log.Printf("数据库初始化成功，数据库文件: %s", dbPath)
	return nil
}

// autoMigrate 自动迁移数据库表
func autoMigrate() error {
	return DB.AutoMigrate(
		&models.MCPServer{},
		&models.MCPTool{},
	)
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return DB
}

// CloseDatabase 关闭数据库连接
func CloseDatabase() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

// SeedData 初始化种子数据
func SeedData() error {
	// 检查是否已有数据
	var count int64
	if err := DB.Model(&models.MCPServer{}).Count(&count).Error; err != nil {
		return err
	}

	// 如果已有数据，则不插入种子数据
	if count > 0 {
		return nil
	}

	// 插入示例MCP服务器数据
	sampleServers := []models.MCPServer{
		{
			Name:        "示例MCP服务器",
			Description: "这是一个示例MCP服务器，用于演示功能",
			URL:         "https://api.example.com/mcp",
			AuthType:    "bearer",
			AuthConfig:  `{"token": "your-api-token"}`,
			Status:      "active",
			IsEnabled:   true,
			Tags:        "示例,测试,API",
		},
		{
			Name:        "本地开发服务器",
			Description: "本地开发环境的MCP服务器",
			URL:         "http://localhost:3000/mcp",
			AuthType:    "none",
			AuthConfig:  "",
			Status:      "inactive",
			IsEnabled:   false,
			Tags:        "本地,开发",
		},
	}

	for _, server := range sampleServers {
		if err := DB.Create(&server).Error; err != nil {
			log.Printf("插入种子数据失败: %v", err)
			return err
		}
	}

	log.Println("种子数据插入成功")
	return nil
}