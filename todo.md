## 框架
### 前端
#### antd
**技术栈选择分析：**
- **UI组件库**: Ant Design (antd) - 提供丰富的企业级UI组件
- **前端框架**: React + TypeScript (基于Wails模板)
- **状态管理**: Redux Toolkit 或 Zustand (推荐Zustand，轻量级)
- **路由管理**: React Router v6
- **样式方案**: CSS Modules + Ant Design主题定制
- **图标库**: @ant-design/icons + 自定义SVG图标

**关键实现要点：**
- 响应式布局设计，支持窗口大小调整
- 主题切换功能（明暗模式）
- 国际化支持（中英文）
- 组件懒加载优化性能

### 后端
#### gin
**技术栈选择分析：**
- **Web框架**: Gin - 高性能HTTP框架
- **数据库**: SQLite (本地存储) + GORM (ORM)
- **配置管理**: Viper
- **日志系统**: logrus 或 zap

**架构设计：**
- RESTful API设计规范
- 分层架构：Controller -> Service -> Repository -> Model
- 依赖注入容器
- 统一错误处理和响应格式
- 数据验证和序列化

## MCP功能
### MCP Server 
####  MCP Server 功能

**1、新增MCP Server**
- **功能描述**: 支持添加外部MCP服务器连接
- **实现要点**:
  - 服务器配置表单：名称、描述、连接地址、认证信息
  - 连接测试功能，验证服务器可用性
  - 支持多种连接方式：HTTP/HTTPS、WebSocket、本地进程
  - 配置验证和错误提示
  - 添加后将tools 入库存储

**2、增删改查 MCP Server**
- **CRUD操作**:
  - 创建：POST /api/mcp-servers
  - 查询：GET /api/mcp-servers (支持分页、搜索、过滤)
  - 更新：PUT /api/mcp-servers/:id
  - 删除：DELETE /api/mcp-servers/:id (软删除)
- **前端界面**:
  - 服务器列表页面（Table组件 + 搜索过滤）
  - 新增/编辑表单（Modal + Form组件）
  - 批量操作功能
- **业务逻辑**:
  - 删除前检查是否有依赖的工具在使用
  - 更新配置时重新连接验证
  - 操作日志记录

**3、展示MCP Server Tools，并设置针对tools的开关**
- **工具发现机制**:
  - 连接到MCP服务器后自动获取可用工具列表
  - 定期同步工具状态和元数据
  - 缓存工具信息提高响应速度
- **工具管理界面**:
  - 树形结构展示：服务器 -> 工具分类 -> 具体工具
  - 每个工具的开关控制（Switch组件）
  - 工具详情展示：名称、描述、参数说明、使用示例
  - 批量启用/禁用功能


**4、对话时，调用打开的MCP Server中的Tools**
- **工具调用流程**:
  - AI对话时识别需要使用的工具
  - 根据工具配置动态路由到对应的MCP服务器
  - 执行工具调用并处理返回结果
  - 将结果整合到对话上下文中
- **技术实现**:
  - 工具调用代理服务
  - 异步任务队列处理长时间运行的工具
  - 错误处理和重试机制
  - 调用日志和性能监控
- **安全考虑**:
  - 工具权限控制
  - 参数验证和清理
  - 调用频率限制
  - 敏感操作确认机制



### 内置MCP Server
#### 文件MCP Server

**1、文件读取、写入、搜索文本内容**
- **文件读取功能**:
  - 支持多种文件格式：txt, md, json, xml, csv, log等
  - 大文件分块读取，避免内存溢出
  - 文件编码自动检测（UTF-8, GBK, ASCII等）
  - 二进制文件检测和处理
  - API接口：
    ```go
    // 读取文件内容
    GET /api/files/read?path={filePath}&encoding={encoding}&offset={offset}&limit={limit}
    
    // 获取文件信息
    GET /api/files/info?path={filePath}
    ```

- **文件写入功能**:
  - 支持新建文件和覆盖写入
  - 追加写入模式
  - 原子写入操作，确保数据完整性
  - 写入前备份机制
  - 权限检查和安全验证
  - API接口：
    ```go
    // 写入文件
    POST /api/files/write
    {
        "path": "string",
        "content": "string", 
        "mode": "overwrite|append|create",
        "encoding": "utf-8",
        "backup": true
    }
    ```

- **文本内容搜索**:
  - 全文搜索功能（支持正则表达式）
  - 多文件批量搜索
  - 搜索结果高亮显示
  - 搜索历史记录
  - 性能优化：索引缓存、并发搜索
  - API接口：
    ```go
    // 搜索文件内容
    POST /api/files/search
    {
        "pattern": "string",
        "paths": ["string"],
        "regex": true,
        "case_sensitive": false,
        "max_results": 100
    }
    ```

**2、文件夹读取、创建、删除、重命名、按文件夹名称搜索**
- **文件夹读取**:
  - 递归遍历目录结构
  - 文件/文件夹过滤（按类型、大小、时间等）
  - 树形结构数据返回
  - 权限检查和隐藏文件处理
  - 大目录分页加载
  - API接口：
    ```go
    // 读取目录内容
    GET /api/directories/list?path={dirPath}&recursive={bool}&filter={filter}&page={page}&size={size}
    
    // 获取目录树
    GET /api/directories/tree?path={dirPath}&depth={depth}
    ```

- **文件夹操作**:
  - 创建目录（支持递归创建）
  - 删除目录（安全删除，回收站机制）
  - 重命名/移动目录
  - 目录权限管理
  - 操作日志记录
  - API接口：
    ```go
    // 创建目录
    POST /api/directories/create
    {
        "path": "string",
        "recursive": true,
        "permissions": "755"
    }
    
    // 删除目录
    DELETE /api/directories/delete?path={dirPath}&force={bool}
    
    // 重命名目录
    PUT /api/directories/rename
    {
        "old_path": "string",
        "new_path": "string"
    }
    ```

- **目录搜索**:
  - 按名称模糊搜索
  - 按路径深度搜索
  - 按创建/修改时间搜索
  - 搜索结果排序和分页
  - API接口：
    ```go
    // 搜索目录
    POST /api/directories/search
    {
        "name_pattern": "string",
        "base_path": "string",
        "max_depth": 5,
        "sort_by": "name|size|time",
        "order": "asc|desc"
    }
    ```

**安全和性能考虑**:
- **安全措施**:
  - 路径遍历攻击防护
  - 文件访问权限控制
  - 敏感目录黑名单
  - 操作审计日志
- **性能优化**:
  - 文件操作缓存
  - 异步I/O处理
  - 大文件流式处理
  - 并发限制和资源管理

**数据模型**:
```go
type FileInfo struct {
    Path        string    `json:"path"`
    Name        string    `json:"name"`
    Size        int64     `json:"size"`
    IsDir       bool      `json:"is_dir"`
    ModTime     time.Time `json:"mod_time"`
    Permissions string    `json:"permissions"`
    MimeType    string    `json:"mime_type"`
}

type DirectoryTree struct {
    Path     string          `json:"path"`
    Name     string          `json:"name"`
    IsDir    bool            `json:"is_dir"`
    Children []*DirectoryTree `json:"children,omitempty"`
    FileInfo *FileInfo       `json:"file_info"`
}
```



## AI
### AI 模型管理

**1、AI 模型的增删改查**
- **模型配置管理**:
  - 支持多种AI服务提供商：OpenAI、Claude、本地模型、自定义API
  - 模型参数配置：API Key、Base URL、模型名称、温度、最大Token等
  - 模型分类管理：文本生成、代码生成、图像生成、嵌入模型等
  - 配置模板功能，快速添加常用模型
- **数据模型**:
  ```go
  type AIModel struct {
      ID          uint      `json:"id" gorm:"primaryKey"`
      Name        string    `json:"name" gorm:"not null"`
      Provider    string    `json:"provider"` // openai, claude, local, custom
      ModelType   string    `json:"model_type"` // text, code, image, embedding
      APIKey      string    `json:"api_key" gorm:"type:text"`
      BaseURL     string    `json:"base_url"`
      ModelName   string    `json:"model_name"`
      Config      string    `json:"config" gorm:"type:text"` // JSON配置
      IsEnabled   bool      `json:"is_enabled" gorm:"default:true"`
      IsDefault   bool      `json:"is_default" gorm:"default:false"`
      CreatedAt   time.Time `json:"created_at"`
      UpdatedAt   time.Time `json:"updated_at"`
  }
  ```
- **API接口设计**:
  ```go
  // 模型CRUD操作
  POST   /api/ai-models          // 创建模型
  GET    /api/ai-models          // 获取模型列表
  GET    /api/ai-models/:id      // 获取单个模型
  PUT    /api/ai-models/:id      // 更新模型
  DELETE /api/ai-models/:id      // 删除模型
  POST   /api/ai-models/:id/test // 测试模型连接
  ```

**2、AI 模型的启用、禁用**
- **状态管理功能**:
  - 批量启用/禁用模型
  - 默认模型设置（每种类型只能有一个默认模型）
  - 模型健康检查和状态监控
  - 使用统计和性能监控
- **前端界面**:
  - 模型列表页面（Table + Switch组件）
  - 模型状态指示器（在线/离线/错误）
  - 批量操作工具栏
  - 模型测试功能
- **业务逻辑**:
  - 禁用模型前检查是否有正在进行的对话
  - 默认模型切换时的平滑过渡
  - 模型故障时的自动降级机制

### AI 对话聊天

**1、对话界面要可选AI模型**
- **聊天界面设计**:
  - 现代化聊天UI：消息气泡、头像、时间戳
  - 模型选择器：下拉菜单显示可用模型
  - 实时模型切换功能
  - 对话历史管理和搜索
  - 消息导出功能（Markdown、PDF等）

- **核心功能实现**:
  - **多模型支持**:
    - 同一对话中可切换不同模型
    - 模型特性展示（支持的功能、限制等）
    - 模型响应时间和成本统计
  
  - **对话管理**:
    - 会话创建、删除、重命名
    - 对话分组和标签管理
    - 对话分享和协作功能
    - 自动保存和恢复

  - **消息处理**:
    - 流式响应显示（打字机效果）
    - 消息重新生成功能
    - 消息编辑和删除
    - 代码块语法高亮
    - 数学公式渲染（LaTeX）

- **高级功能**:
  - **工具集成**:
    - MCP工具自动调用
    - 文件上传和处理
    - 图片识别和生成
    - 代码执行环境
  
  - **对话增强**:
    - 上下文记忆管理
    - 角色扮演模式
    - 提示词模板库
    - 对话总结功能

- **数据模型**:
  ```go
  type Conversation struct {
      ID          uint      `json:"id" gorm:"primaryKey"`
      Title       string    `json:"title"`
      ModelID     uint      `json:"model_id"`
      UserID      string    `json:"user_id"` // 用户标识
      Tags        string    `json:"tags"`    // JSON数组
      IsArchived  bool      `json:"is_archived" gorm:"default:false"`
      CreatedAt   time.Time `json:"created_at"`
      UpdatedAt   time.Time `json:"updated_at"`
      Messages    []Message `json:"messages" gorm:"foreignKey:ConversationID"`
  }

  type Message struct {
      ID             uint      `json:"id" gorm:"primaryKey"`
      ConversationID uint      `json:"conversation_id"`
      Role           string    `json:"role"` // user, assistant, system
      Content        string    `json:"content" gorm:"type:text"`
      ModelID        uint      `json:"model_id"`
      TokenCount     int       `json:"token_count"`
      Cost           float64   `json:"cost"`
      CreatedAt      time.Time `json:"created_at"`
  }
  ```

- **API接口设计**:
  ```go
  // 对话管理
  POST   /api/conversations              // 创建对话
  GET    /api/conversations              // 获取对话列表
  GET    /api/conversations/:id          // 获取对话详情
  PUT    /api/conversations/:id          // 更新对话
  DELETE /api/conversations/:id          // 删除对话
  
  // 消息处理
  POST   /api/conversations/:id/messages // 发送消息
  GET    /api/conversations/:id/messages // 获取消息历史
  PUT    /api/messages/:id               // 编辑消息
  DELETE /api/messages/:id               // 删除消息
  POST   /api/messages/:id/regenerate    // 重新生成消息
  
  // 流式响应
  GET    /api/conversations/:id/stream   // WebSocket连接
  ```

**性能和用户体验优化**:
- **响应优化**:
  - 消息预加载和缓存
  - 图片懒加载
  - 虚拟滚动处理长对话
- **离线支持**:
  - 本地消息缓存
  - 离线模式下的基本功能
  - 网络恢复后的数据同步


## 具体功能
### 整理目录

**功能描述**: 智能化的目录整理工具，帮助用户自动或半自动地整理文件和文件夹

**核心功能**:
- **智能分类**:
  - 按文件类型自动分类（图片、文档、视频、音频等）
  - 按文件大小分类（大文件、小文件）
  - 按创建/修改时间分类（按年、月、日）
  - 按文件名模式分类（正则表达式匹配）
  - 自定义分类规则配置

- **重复文件检测**:
  - 基于文件内容的MD5/SHA256哈希比较
  - 基于文件名和大小的快速检测
  - 相似文件检测（图片、文档内容相似度）
  - 重复文件处理策略：删除、移动、重命名

- **目录结构优化**:
  - 空文件夹清理
  - 深层嵌套目录扁平化
  - 目录命名规范化
  - 文件路径长度优化

**实现方案**:
```go
type DirectoryOrganizer struct {
    Rules       []OrganizeRule `json:"rules"`
    TargetPath  string         `json:"target_path"`
    BackupPath  string         `json:"backup_path"`
    DryRun      bool           `json:"dry_run"`
}

type OrganizeRule struct {
    ID          uint   `json:"id"`
    Name        string `json:"name"`
    Type        string `json:"type"` // file_type, size, date, pattern
    Pattern     string `json:"pattern"`
    TargetDir   string `json:"target_dir"`
    Action      string `json:"action"` // move, copy, link
    Enabled     bool   `json:"enabled"`
}
```

**前端界面**:
- 拖拽式规则配置器
- 实时预览整理结果
- 进度条显示整理进度
- 操作日志和撤销功能

### 开发小工具

**1、代码格式化工具**
- **支持语言**: Go, JavaScript, TypeScript, Python, Java, C++等
- **功能特性**:
  - 多种代码风格配置
  - 批量格式化
  - 格式化前后对比
  - 自定义格式化规则
- **实现**: 集成各语言的格式化工具（gofmt, prettier, black等）

**2、JSON/XML处理工具**
- **JSON工具**:
  - 格式化和压缩
  - 语法验证和错误提示
  - JSON转换（转XML、YAML、CSV等）
  - JSONPath查询
- **XML工具**:
  - 格式化和验证
  - XPath查询
  - XML转JSON/YAML
  - DTD/XSD验证

**3、文本处理工具**
- **编码转换**: UTF-8, GBK, ASCII等编码互转
- **文本统计**: 字符数、行数、词频统计
- **文本替换**: 批量查找替换，支持正则表达式
- **文本比较**: 文件差异对比，高亮显示差异

**4、哈希计算工具**
- **支持算法**: MD5, SHA1, SHA256, SHA512, CRC32等
- **功能特性**:
  - 文件哈希计算
  - 文本哈希计算
  - 批量哈希计算
  - 哈希值验证

**5、Base64编解码工具**
- 文本和文件的Base64编解码
- URL安全的Base64编码
- 图片Base64编码预览
- 批量处理功能

**6、时间戳转换工具**
- Unix时间戳转换
- 多种时间格式支持
- 时区转换
- 批量时间处理

**7、正则表达式测试工具**
- 实时正则匹配测试
- 常用正则表达式库
- 匹配结果高亮显示
- 正则表达式解释和优化建议

**通用实现框架**:
```go
type Tool interface {
    Name() string
    Description() string
    Process(input interface{}) (interface{}, error)
    Validate(input interface{}) error
}

type ToolManager struct {
    tools map[string]Tool
}

func (tm *ToolManager) RegisterTool(tool Tool) {
    tm.tools[tool.Name()] = tool
}

func (tm *ToolManager) ExecuteTool(name string, input interface{}) (interface{}, error) {
    tool, exists := tm.tools[name]
    if !exists {
        return nil, fmt.Errorf("tool not found: %s", name)
    }
    
    if err := tool.Validate(input); err != nil {
        return nil, err
    }
    
    return tool.Process(input)
}
```

### 各种功能

**1、系统信息监控**
- **硬件信息**: CPU、内存、磁盘、网络使用情况
- **系统信息**: 操作系统版本、运行时间、进程列表
- **实时监控**: 资源使用率图表、历史数据记录
- **告警功能**: 资源使用率超阈值提醒

**2、网络工具集**
- **网络测试**:
  - Ping测试（延迟、丢包率）
  - 端口扫描和连通性测试
  - 网速测试
  - DNS查询工具
- **网络信息**:
  - 本机IP地址查询
  - 公网IP和地理位置
  - 网络接口信息
  - 路由表查看

**3、文件加密解密**
- **加密算法**: AES-256, RSA, ChaCha20等
- **功能特性**:
  - 文件和文本加密
  - 密码强度检测
  - 密钥管理
  - 批量加密处理
- **安全考虑**:
  - 密钥安全存储
  - 加密过程内存清理
  - 操作审计日志

**4、二维码生成器**
- **支持类型**: 文本、URL、WiFi配置、联系人信息等
- **自定义选项**: 尺寸、颜色、Logo嵌入、容错级别
- **批量生成**: 支持批量数据生成二维码
- **格式导出**: PNG, SVG, PDF等格式

**5、颜色工具**
- **颜色转换**: RGB, HEX, HSL, HSV等格式互转
- **调色板生成**: 基于主色生成配色方案
- **颜色提取**: 从图片中提取主要颜色
- **无障碍检测**: 颜色对比度检测

**6、单位转换器**
- **支持类型**: 长度、重量、温度、面积、体积、速度等
- **货币转换**: 实时汇率查询和转换
- **自定义单位**: 支持添加自定义转换规则
- **历史记录**: 转换历史记录和收藏功能

**技术架构设计**:
- **插件化架构**: 每个工具作为独立插件，支持动态加载
- **统一界面**: 使用Ant Design组件构建一致的用户界面
- **数据持久化**: 用户配置和历史记录本地存储
- **性能优化**: 工具懒加载、结果缓存、异步处理

**API设计示例**:
```go
// 工具管理
GET    /api/tools                    // 获取工具列表
POST   /api/tools/:name/execute      // 执行工具
GET    /api/tools/:name/config       // 获取工具配置
PUT    /api/tools/:name/config       // 更新工具配置

// 历史记录
GET    /api/tools/:name/history      // 获取使用历史
POST   /api/tools/:name/history      // 保存历史记录
DELETE /api/tools/:name/history/:id  // 删除历史记录
```

