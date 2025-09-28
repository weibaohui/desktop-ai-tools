import axios from 'axios';

// 创建axios实例
const api = axios.create({
  baseURL: 'http://localhost:8080/api',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// 响应拦截器
api.interceptors.response.use(
  (response) => response,
  (error) => {
    console.error('API请求错误:', error);
    return Promise.reject(error);
  }
);

export interface MCPToolListRequest {
  server_id?: number;
  category?: string;
  search?: string;
  is_enabled?: boolean;
}

export interface MCPToolUpdateRequest {
  is_enabled?: boolean;
  category?: string;
}

export interface MCPToolBatchUpdateRequest {
  tool_ids: number[];
  is_enabled?: boolean;
  category?: string;
}

export interface MCPTool {
  id: number;
  server_id: number;
  name: string;
  description: string;
  category: string;
  parameters: any;
  is_enabled: boolean;
  created_at: string;
  updated_at: string;
}

export interface MCPToolListResponse {
  tools: MCPTool[];
  total: number;
}

export interface ApiResponse<T = any> {
  success: boolean;
  data?: T;
  message?: string;
  error?: string;
}

/**
 * MCP工具API服务类
 */
export class MCPToolService {
  /**
   * 获取工具列表
   */
  static async getList(params: MCPToolListRequest = {}): Promise<MCPToolListResponse> {
    const response = await api.get<ApiResponse<MCPToolListResponse>>('/mcp-tools', {
      params,
    });
    
    if (!response.data.success) {
      throw new Error(response.data.error || '获取工具列表失败');
    }
    
    return response.data.data!;
  }

  /**
   * 更新工具
   */
  static async update(id: number, data: MCPToolUpdateRequest): Promise<void> {
    const response = await api.put<ApiResponse>(`/mcp-tools/${id}`, data);
    
    if (!response.data.success) {
      throw new Error(response.data.error || '更新工具失败');
    }
  }

  /**
   * 批量更新工具
   */
  static async batchUpdate(data: MCPToolBatchUpdateRequest): Promise<void> {
    const response = await api.put<ApiResponse>('/mcp-tools/batch', data);
    
    if (!response.data.success) {
      throw new Error(response.data.error || '批量更新工具失败');
    }
  }

  /**
   * 获取工具分类
   */
  static async getCategories(serverId?: number): Promise<string[]> {
    const params = serverId ? { server_id: serverId } : {};
    const response = await api.get<ApiResponse<string[]>>('/mcp-tools/categories', {
      params,
    });
    
    if (!response.data.success) {
      throw new Error(response.data.error || '获取工具分类失败');
    }
    
    return response.data.data || [];
  }
}

// 导出便捷的API对象
export const mcpToolApi = {
  getList: MCPToolService.getList,
  update: MCPToolService.update,
  batchUpdate: MCPToolService.batchUpdate,
  getCategories: MCPToolService.getCategories,
};

export default MCPToolService;