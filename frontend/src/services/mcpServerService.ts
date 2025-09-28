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
      // 这里可以实现实际的连接测试逻辑
      // 暂时返回模拟结果
      await new Promise(resolve => setTimeout(resolve, 1000));
      return Math.random() > 0.3; // 70% 成功率
    } catch (error) {
      console.error('连接测试失败:', error);
      return false;
    }
  }
}

export default MCPServerService;