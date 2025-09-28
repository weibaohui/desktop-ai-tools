import axios from 'axios';
import type {
  MCPServer,
  MCPServerCreateRequest,
  MCPServerUpdateRequest,
  MCPServerListRequest,
  MCPServerListResponse,
  MCPServerStatusUpdateRequest,
  ApiResponse
} from '../types/mcpServer';

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

/**
 * MCP Server API 服务类
 */
export class MCPServerService {
  /**
   * 获取MCP服务器列表
   */
  static async getList(params: MCPServerListRequest = {}): Promise<MCPServerListResponse> {
    const response = await api.get<ApiResponse<MCPServerListResponse>>('/mcp-servers', {
      params,
    });
    
    if (!response.data.success) {
      throw new Error(response.data.error || '获取服务器列表失败');
    }
    
    return response.data.data!;
  }

  /**
   * 根据ID获取MCP服务器详情
   */
  static async getById(id: number): Promise<MCPServer> {
    const response = await api.get<ApiResponse<MCPServer>>(`/mcp-servers/${id}`);
    
    if (!response.data.success) {
      throw new Error(response.data.error || '获取服务器详情失败');
    }
    
    return response.data.data!;
  }

  /**
   * 创建MCP服务器
   */
  static async create(data: MCPServerCreateRequest): Promise<MCPServer> {
    const response = await api.post<ApiResponse<MCPServer>>('/mcp-servers', data);
    
    if (!response.data.success) {
      throw new Error(response.data.error || '创建服务器失败');
    }
    
    return response.data.data!;
  }

  /**
   * 更新MCP服务器
   */
  static async update(id: number, data: MCPServerUpdateRequest): Promise<MCPServer> {
    const response = await api.put<ApiResponse<MCPServer>>(`/mcp-servers/${id}`, data);
    
    if (!response.data.success) {
      throw new Error(response.data.error || '更新服务器失败');
    }
    
    return response.data.data!;
  }

  /**
   * 删除MCP服务器
   */
  static async delete(id: number): Promise<void> {
    const response = await api.delete<ApiResponse>(`/mcp-servers/${id}`);
    
    if (!response.data.success) {
      throw new Error(response.data.error || '删除服务器失败');
    }
  }

  /**
   * 更新服务器状态
   */
  static async updateStatus(id: number, status: string): Promise<void> {
    const response = await api.put<ApiResponse>(`/mcp-servers/${id}/status`, { status });
    
    if (!response.data.success) {
      throw new Error(response.data.error || '更新服务器状态失败');
    }
  }

  /**
   * 切换服务器启用状态
   */
  static async toggle(id: number): Promise<MCPServer> {
    const response = await api.put<ApiResponse<MCPServer>>(`/mcp-servers/${id}/toggle`);
    
    if (!response.data.success) {
      throw new Error(response.data.error || '切换服务器状态失败');
    }
    
    return response.data.data!;
  }

  /**
   * 获取所有标签
   */
  static async getTags(): Promise<string[]> {
    const response = await api.get<ApiResponse<string[]>>('/mcp-servers/tags');
    
    if (!response.data.success) {
      throw new Error(response.data.error || '获取标签失败');
    }
    
    return response.data.data || [];
  }

  /**
   * 测试服务器连接
   */
  static async testConnection(url: string, authConfig?: any): Promise<boolean> {
    try {
      const response = await api.post<ApiResponse<{ connected: boolean }>>('/mcp-servers/test-connection', {
        url,
        auth_config: authConfig,
      });
      
      return response.data.success && response.data.data?.connected === true;
    } catch (error) {
      console.error('连接测试失败:', error);
      return false;
    }
  }

  /**
   * 发现工具
   */
  static async discoverTools(id: number): Promise<{ success: boolean; message?: string; tools_count?: number }> {
    try {
      const response = await api.post<ApiResponse<{ tools_count: number }>>(`/mcp-servers/${id}/discover-tools`);
      
      return {
        success: response.data.success,
        message: response.data.message,
        tools_count: response.data.data?.tools_count,
      };
    } catch (error) {
      console.error('工具发现失败:', error);
      return {
        success: false,
        message: '工具发现失败',
      };
    }
  }

  /**
   * 刷新工具列表
   */
  static async refreshTools(id: number): Promise<{ success: boolean; message?: string; tools?: any[] }> {
    try {
      const response = await api.post<ApiResponse<{ tools: any[] }>>(`/mcp-tools/refresh/${id}`);
      
      return {
        success: response.data.success,
        message: response.data.message,
        tools: response.data.data?.tools,
      };
    } catch (error) {
      console.error('刷新工具失败:', error);
      return {
        success: false,
        message: '刷新工具失败',
      };
    }
  }
}

// 导出便捷的API对象
export const mcpServerApi = {
  getList: MCPServerService.getList,
  getById: MCPServerService.getById,
  create: MCPServerService.create,
  update: MCPServerService.update,
  delete: MCPServerService.delete,
  updateStatus: MCPServerService.updateStatus,
  toggle: MCPServerService.toggle,
  getTags: MCPServerService.getTags,
  testConnection: MCPServerService.testConnection,
  discoverTools: MCPServerService.discoverTools,
  refreshTools: MCPServerService.refreshTools,
};

export default MCPServerService;