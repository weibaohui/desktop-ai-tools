// MCP Server 相关类型定义

export interface MCPServer {
  id: number;
  name: string;
  description: string;
  url: string;
  auth_type: 'none' | 'bearer' | 'basic' | 'api_key';
  auth_config: string;
  status: 'active' | 'inactive' | 'error';
  is_enabled: boolean;
  tags: string;
  created_at: string;
  updated_at: string;
  tools?: MCPTool[];
}

export interface MCPTool {
  id: number;
  server_id: number;
  name: string;
  description: string;
  category: string;
  parameters: string;
  is_enabled: boolean;
  created_at: string;
  updated_at: string;
}

export interface MCPServerCreateRequest {
  name: string;
  description?: string;
  url: string;
  auth_type: 'none' | 'bearer' | 'basic' | 'api_key';
  auth_config?: string;
  tags?: string;
}

export interface MCPServerUpdateRequest {
  name: string;
  description?: string;
  url: string;
  auth_type: 'none' | 'bearer' | 'basic' | 'api_key';
  auth_config?: string;
  is_enabled?: boolean;
  tags?: string;
}

export interface MCPServerListRequest {
  page?: number;
  size?: number;
  search?: string;
  status?: 'active' | 'inactive' | 'error' | '';
  enabled?: boolean;
  order_by?: 'created_at' | 'updated_at' | 'name';
  order_dir?: 'asc' | 'desc';
}

export interface MCPServerListResponse {
  total: number;
  page: number;
  size: number;
  servers: MCPServer[];
}

export interface ApiResponse<T = any> {
  success: boolean;
  data?: T;
  message?: string;
  error?: string;
}

export interface MCPServerStatusUpdateRequest {
  status: 'active' | 'inactive' | 'error';
}

// 认证配置类型
export interface AuthConfig {
  token?: string;
  username?: string;
  password?: string;
  api_key?: string;
  [key: string]: any;
}

// 表单状态
export interface MCPServerFormData {
  name: string;
  description: string;
  url: string;
  auth_type: 'none' | 'bearer' | 'basic' | 'api_key';
  auth_config: AuthConfig;
  tags: string[];
}

// 列表查询状态
export interface MCPServerListState {
  loading: boolean;
  data: MCPServer[];
  total: number;
  current: number;
  pageSize: number;
  filters: {
    search: string;
    status: string;
    enabled?: boolean;
  };
  sorter: {
    field: string;
    order: 'ascend' | 'descend';
  };
}