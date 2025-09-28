import React, { useState, useEffect } from 'react';
import {
  Card,
  Tree,
  Switch,
  Button,
  Space,
  Typography,
  Spin,
  message,
  Modal,
  Descriptions,
  Tag,
  Tooltip,
  Input,
  Select,
  Checkbox,
  Row,
  Col,
  Divider,
  Alert,
} from 'antd';
import {
  ToolOutlined,
  ReloadOutlined,
  SettingOutlined,
  InfoCircleOutlined,
  SearchOutlined,
  FilterOutlined,
  CheckSquareOutlined,
  MinusSquareOutlined,
} from '@ant-design/icons';
import type { DataNode } from 'antd/es/tree';
import { mcpServerApi } from '../services/mcpServerService';
import type { MCPServer } from '../types/mcpServer';

const { Title, Text, Paragraph } = Typography;
const { Search } = Input;
const { Option } = Select;

// 定义MCPTool接口
interface MCPTool {
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

// 定义API服务
const mcpToolApi = {
  async getList(params: { server_id?: number; category?: string; search?: string; is_enabled?: boolean } = {}) {
    // 这里暂时返回模拟数据，后续会连接真实API
    return {
      tools: [] as MCPTool[],
      total: 0,
    };
  },
  async update(id: number, data: { is_enabled?: boolean; category?: string }) {
    // 暂时返回成功
    return Promise.resolve();
  },
  async batchUpdate(data: { tool_ids: number[]; is_enabled?: boolean; category?: string }) {
    // 暂时返回成功
    return Promise.resolve();
  },
  async getCategories(serverId?: number) {
    // 返回默认分类
    return ['工具', '数据处理', '文件操作', '网络请求', '其他'];
  },
};

interface ToolTreeNode extends DataNode {
  type: 'server' | 'category' | 'tool';
  data?: MCPServer | MCPTool;
  server_id?: number;
  category?: string;
}

/**
 * MCP工具管理页面组件
 * 提供工具发现、管理、开关控制等功能
 */
const MCPTools: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [servers, setServers] = useState<MCPServer[]>([]);
  const [tools, setTools] = useState<MCPTool[]>([]);
  const [treeData, setTreeData] = useState<ToolTreeNode[]>([]);
  const [selectedTool, setSelectedTool] = useState<MCPTool | null>(null);
  const [detailModalVisible, setDetailModalVisible] = useState(false);
  const [searchText, setSearchText] = useState('');
  const [selectedCategory, setSelectedCategory] = useState<string>('');
  const [selectedServer, setSelectedServer] = useState<number | undefined>();
  const [categories, setCategories] = useState<string[]>([]);
  const [checkedKeys, setCheckedKeys] = useState<React.Key[]>([]);
  const [expandedKeys, setExpandedKeys] = useState<React.Key[]>([]);

  /**
   * 加载MCP服务器列表
   */
  const loadServers = async () => {
    try {
      setLoading(true);
      const response = await mcpServerApi.getList({
        status: 'active',
        enabled: true,
      });
      
      setServers(response.servers || []);
    } catch (error) {
      message.error('加载服务器列表失败');
    } finally {
      setLoading(false);
    }
  };

  /**
   * 加载工具列表
   */
  const loadTools = async () => {
    try {
      const response = await mcpToolApi.getList({
        server_id: selectedServer,
        category: selectedCategory,
        search: searchText,
      });
      setTools(response.tools || []);
    } catch (error) {
      console.error('加载工具列表失败:', error);
      message.error('加载工具列表失败');
    }
  };

  /**
   * 加载工具分类
   */
  const loadCategories = async () => {
    try {
      const response = await mcpToolApi.getCategories(selectedServer);
      setCategories(response || []);
    } catch (error) {
      console.error('加载工具分类失败:', error);
    }
  };

  /**
   * 发现工具
   */
  const discoverTools = async (serverId: number) => {
    setLoading(true);
    try {
      const response = await mcpServerApi.discoverTools(serverId);
      if (response.success) {
        message.success(`成功发现 ${response.tools_count} 个工具`);
        await loadTools();
        await loadCategories();
      } else {
        message.error(response.message || '工具发现失败');
      }
    } catch (error) {
      console.error('工具发现失败:', error);
      message.error('工具发现失败');
    } finally {
      setLoading(false);
    }
  };

  /**
   * 构建树形数据结构
   */
  const buildTreeData = (): ToolTreeNode[] => {
    const serverNodes: ToolTreeNode[] = [];

    servers.forEach(server => {
      const serverTools = tools.filter(tool => tool.server_id === server.id);
      const serverCategories = [...new Set(serverTools.map(tool => tool.category))];

      const categoryNodes: ToolTreeNode[] = serverCategories.map(category => {
        const categoryTools = serverTools.filter(tool => tool.category === category);
        const toolNodes: ToolTreeNode[] = categoryTools.map(tool => ({
          title: (
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
              <span>
                <ToolOutlined style={{ marginRight: 8 }} />
                {tool.name}
              </span>
              <Space>
                <Switch
                  size="small"
                  checked={tool.is_enabled}
                  onChange={(checked) => handleToolToggle(tool.id, checked)}
                />
                <Button
                  type="text"
                  size="small"
                  icon={<InfoCircleOutlined />}
                  onClick={() => showToolDetail(tool)}
                />
              </Space>
            </div>
          ),
          key: `tool-${tool.id}`,
          type: 'tool',
          data: tool,
          server_id: server.id,
          category,
          isLeaf: true,
        }));

        return {
          title: (
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
              <span>
                <FilterOutlined style={{ marginRight: 8 }} />
                {category} ({categoryTools.length})
              </span>
              <Space>
                <Button
                  type="text"
                  size="small"
                  onClick={() => handleCategoryToggle(server.id, category, true)}
                >
                  全开
                </Button>
                <Button
                  type="text"
                  size="small"
                  onClick={() => handleCategoryToggle(server.id, category, false)}
                >
                  全关
                </Button>
              </Space>
            </div>
          ),
          key: `category-${server.id}-${category}`,
          type: 'category',
          server_id: server.id,
          category,
          children: toolNodes,
        };
      });

      serverNodes.push({
        title: (
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
            <span>
              <SettingOutlined style={{ marginRight: 8 }} />
              {server.name} ({serverTools.length} 工具)
            </span>
            <Space>
              <Button
                type="text"
                size="small"
                icon={<ReloadOutlined />}
                onClick={() => discoverTools(server.id)}
                loading={loading}
              >
                发现工具
              </Button>
            </Space>
          </div>
        ),
        key: `server-${server.id}`,
        type: 'server',
        data: server,
        children: categoryNodes,
      });
    });

    return serverNodes;
  };

  /**
   * 处理工具开关切换
   */
  const handleToolToggle = async (toolId: number, enabled: boolean) => {
    try {
      await mcpToolApi.update(toolId, { is_enabled: enabled });
      setTools(prev => prev.map(tool => 
        tool.id === toolId ? { ...tool, is_enabled: enabled } : tool
      ));
      message.success(`工具已${enabled ? '启用' : '禁用'}`);
    } catch (error) {
      console.error('更新工具状态失败:', error);
      message.error('更新工具状态失败');
    }
  };

  /**
   * 处理分类批量开关
   */
  const handleCategoryToggle = async (serverId: number, category: string, enabled: boolean) => {
    const categoryTools = tools.filter(
      tool => tool.server_id === serverId && tool.category === category
    );
    const toolIds = categoryTools.map(tool => tool.id);

    try {
      await mcpToolApi.batchUpdate({
        tool_ids: toolIds,
        is_enabled: enabled,
      });
      setTools(prev => prev.map(tool => 
        toolIds.includes(tool.id) ? { ...tool, is_enabled: enabled } : tool
      ));
      message.success(`${category} 分类下的 ${toolIds.length} 个工具已${enabled ? '启用' : '禁用'}`);
    } catch (error) {
      console.error('批量更新工具状态失败:', error);
      message.error('批量更新工具状态失败');
    }
  };

  /**
   * 显示工具详情
   */
  const showToolDetail = (tool: MCPTool) => {
    setSelectedTool(tool);
    setDetailModalVisible(true);
  };

  /**
   * 渲染工具参数
   */
  const renderToolParameters = (parameters: any) => {
    if (!parameters || typeof parameters !== 'object') {
      return <Text type="secondary">无参数</Text>;
    }

    return (
      <div>
        {Object.entries(parameters).map(([key, value]: [string, any]) => (
          <Tag key={key} style={{ marginBottom: 4 }}>
            {key}: {typeof value === 'object' ? JSON.stringify(value) : String(value)}
          </Tag>
        ))}
      </div>
    );
  };

  useEffect(() => {
    loadServers();
  }, []);

  useEffect(() => {
    if (servers.length > 0) {
      loadTools();
      loadCategories();
    }
  }, [servers, selectedServer, selectedCategory, searchText]);

  useEffect(() => {
    setTreeData(buildTreeData());
  }, [servers, tools]);

  return (
    <div style={{ padding: 24 }}>
      <div style={{ marginBottom: 24 }}>
        <Title level={2}>
          <ToolOutlined style={{ marginRight: 12 }} />
          MCP 工具管理
        </Title>
        <Paragraph type="secondary">
          管理 MCP 服务器的工具，支持工具发现、开关控制和批量操作
        </Paragraph>
      </div>

      {/* 筛选控件 */}
      <Card style={{ marginBottom: 24 }}>
        <Row gutter={16}>
          <Col span={8}>
            <Search
              placeholder="搜索工具名称或描述"
              value={searchText}
              onChange={(e) => setSearchText(e.target.value)}
              onSearch={loadTools}
              enterButton={<SearchOutlined />}
            />
          </Col>
          <Col span={6}>
            <Select
              placeholder="选择服务器"
              value={selectedServer}
              onChange={setSelectedServer}
              allowClear
              style={{ width: '100%' }}
            >
              {servers.map(server => (
                <Option key={server.id} value={server.id}>
                  {server.name}
                </Option>
              ))}
            </Select>
          </Col>
          <Col span={6}>
            <Select
              placeholder="选择分类"
              value={selectedCategory}
              onChange={setSelectedCategory}
              allowClear
              style={{ width: '100%' }}
            >
              {categories.map(category => (
                <Option key={category} value={category}>
                  {category}
                </Option>
              ))}
            </Select>
          </Col>
          <Col span={4}>
            <Button
              type="primary"
              icon={<ReloadOutlined />}
              onClick={loadTools}
              loading={loading}
            >
              刷新
            </Button>
          </Col>
        </Row>
      </Card>

      {/* 工具树 */}
      <Card>
        {servers.length === 0 ? (
          <Alert
            message="暂无可用的 MCP 服务器"
            description="请先添加并连接 MCP 服务器，然后返回此页面管理工具。"
            type="info"
            showIcon
          />
        ) : (
          <Spin spinning={loading}>
            <Tree
              treeData={treeData}
              expandedKeys={expandedKeys}
              onExpand={setExpandedKeys}
              showLine
              showIcon={false}
              height={600}
              style={{ overflow: 'auto' }}
            />
          </Spin>
        )}
      </Card>

      {/* 工具详情模态框 */}
      <Modal
        title={
          <Space>
            <ToolOutlined />
            工具详情
          </Space>
        }
        open={detailModalVisible}
        onCancel={() => setDetailModalVisible(false)}
        footer={[
          <Button key="close" onClick={() => setDetailModalVisible(false)}>
            关闭
          </Button>,
        ]}
        width={800}
      >
        {selectedTool && (
          <Descriptions column={1} bordered>
            <Descriptions.Item label="工具名称">
              {selectedTool.name}
            </Descriptions.Item>
            <Descriptions.Item label="描述">
              {selectedTool.description || '暂无描述'}
            </Descriptions.Item>
            <Descriptions.Item label="分类">
              <Tag color="blue">{selectedTool.category}</Tag>
            </Descriptions.Item>
            <Descriptions.Item label="状态">
              <Tag color={selectedTool.is_enabled ? 'green' : 'red'}>
                {selectedTool.is_enabled ? '已启用' : '已禁用'}
              </Tag>
            </Descriptions.Item>
            <Descriptions.Item label="参数">
              {renderToolParameters(selectedTool.parameters)}
            </Descriptions.Item>
            <Descriptions.Item label="创建时间">
              {new Date(selectedTool.created_at).toLocaleString()}
            </Descriptions.Item>
            <Descriptions.Item label="更新时间">
              {new Date(selectedTool.updated_at).toLocaleString()}
            </Descriptions.Item>
          </Descriptions>
        )}
      </Modal>
    </div>
  );
};

export default MCPTools;