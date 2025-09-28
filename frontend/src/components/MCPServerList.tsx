import React, { useState, useEffect } from 'react';
import {
  Table,
  Button,
  Space,
  Input,
  Select,
  Tag,
  Switch,
  Modal,
  message,
  Tooltip,
  Card,
  Row,
  Col,
  Popconfirm,
  Badge,
  Typography,
} from 'antd';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  ReloadOutlined,
  SearchOutlined,
  SettingOutlined,
  LinkOutlined,
  TagsOutlined,
} from '@ant-design/icons';
import type { ColumnsType, TableProps } from 'antd/es/table';
import type {
  MCPServer,
  MCPServerListRequest,
  MCPServerListResponse,
  MCPServerListState
} from '../types/mcpServer';
import { MCPServerService } from '../services/mcpServerService';
import MCPServerForm from './MCPServerForm';

const { Search } = Input;
const { Option } = Select;
const { Text } = Typography;

/**
 * MCP Server列表组件
 */
const MCPServerList: React.FC = () => {
  // 状态管理
  const [state, setState] = useState<MCPServerListState>({
    loading: false,
    data: [],
    total: 0,
    current: 1,
    pageSize: 10,
    filters: {
      search: '',
      status: '',
    },
    sorter: {
      field: 'created_at',
      order: 'descend',
    },
  });

  // 表单相关状态
  const [formVisible, setFormVisible] = useState(false);
  const [editingServer, setEditingServer] = useState<MCPServer | null>(null);

  /**
   * 加载服务器列表
   */
  const loadServers = async () => {
    setState(prev => ({ ...prev, loading: true }));

    try {
      const params: MCPServerListRequest = {
        page: state.current,
        size: state.pageSize,
        search: state.filters.search || undefined,
        status: (state.filters.status === '' ? undefined : state.filters.status) as 'active' | 'inactive' | 'error' | undefined,
        enabled: state.filters.enabled,
        order_by: state.sorter.field as 'created_at' | 'updated_at' | 'name',
        order_dir: state.sorter.order === 'ascend' ? 'asc' : 'desc',
      };

      const response = await MCPServerService.getList(params);

      setState(prev => ({
        ...prev,
        data: response.servers,
        total: response.total,
        loading: false,
      }));
    } catch (error) {
      message.error('加载服务器列表失败: ' + (error as Error).message);
      setState(prev => ({ ...prev, loading: false }));
    }
  };

  /**
   * 处理表格变化
   */
  const handleTableChange: TableProps<MCPServer>['onChange'] = (pagination, filters, sorter) => {
    const newState = { ...state };

    if (pagination) {
      newState.current = pagination.current || 1;
      newState.pageSize = pagination.pageSize || 10;
    }

    if (Array.isArray(sorter)) {
      // 多列排序，取第一个
      const firstSorter = sorter[0];
      if (firstSorter) {
        newState.sorter.field = firstSorter.field as string;
        newState.sorter.order = firstSorter.order || 'descend';
      }
    } else if (sorter) {
      newState.sorter.field = sorter.field as string;
      newState.sorter.order = sorter.order || 'descend';
    }

    setState(newState);
  };

  /**
   * 处理搜索
   */
  const handleSearch = (value: string) => {
    setState(prev => ({
      ...prev,
      filters: { ...prev.filters, search: value },
      current: 1,
    }));
  };

  /**
   * 处理状态过滤
   */
  const handleStatusFilter = (value: string) => {
    setState(prev => ({
      ...prev,
      filters: { ...prev.filters, status: value },
      current: 1,
    }));
  };

  /**
   * 处理启用状态过滤
   */
  const handleEnabledFilter = (value: boolean | undefined) => {
    setState(prev => ({
      ...prev,
      filters: { ...prev.filters, enabled: value },
      current: 1,
    }));
  };

  /**
   * 切换服务器启用状态
   */
  const handleToggleEnabled = async (server: MCPServer) => {
    try {
      await MCPServerService.toggle(server.id);
      message.success('状态切换成功');
      loadServers();
    } catch (error) {
      message.error('状态切换失败: ' + (error as Error).message);
    }
  };

  /**
   * 删除服务器
   */
  const handleDelete = async (id: number) => {
    try {
      await MCPServerService.delete(id);
      message.success('删除成功');
      loadServers();
    } catch (error) {
      message.error('删除失败: ' + (error as Error).message);
    }
  };

  /**
   * 打开新增表单
   */
  const handleAdd = () => {
    setEditingServer(null);
    setFormVisible(true);
  };

  /**
   * 打开编辑表单
   */
  const handleEdit = (server: MCPServer) => {
    setEditingServer(server);
    setFormVisible(true);
  };

  /**
   * 表单提交成功回调
   */
  const handleFormSuccess = () => {
    setFormVisible(false);
    setEditingServer(null);
    loadServers();
  };

  /**
   * 获取状态标签
   */
  const getStatusTag = (status: string) => {
    const statusConfig = {
      active: { color: 'green', text: '活跃' },
      inactive: { color: 'default', text: '非活跃' },
      error: { color: 'red', text: '错误' },
    };

    const config = statusConfig[status as keyof typeof statusConfig] || statusConfig.inactive;
    return <Tag color={config.color}>{config.text}</Tag>;
  };

  /**
   * 获取认证类型标签
   */
  const getAuthTypeTag = (authType: string) => {
    const authConfig = {
      none: { color: 'default', text: '无认证' },
      bearer: { color: 'blue', text: 'Bearer' },
      basic: { color: 'orange', text: 'Basic' },
      api_key: { color: 'purple', text: 'API Key' },
    };

    const config = authConfig[authType as keyof typeof authConfig] || authConfig.none;
    return <Tag color={config.color}>{config.text}</Tag>;
  };

  /**
   * 渲染标签
   */
  const renderTags = (tags: string) => {
    if (!tags) return null;

    const tagList = tags.split(',').map(tag => tag.trim()).filter(Boolean);
    return (
      <Space size={4} wrap>
        {tagList.map((tag, index) => (
          <Tag key={index} icon={<TagsOutlined />}>
            {tag}
          </Tag>
        ))}
      </Space>
    );
  };

  // 表格列定义
  const columns: ColumnsType<MCPServer> = [
    {
      title: '服务器名称',
      dataIndex: 'name',
      key: 'name',
      sorter: true,
      render: (text, record) => (
        <Space direction="vertical" size={0}>
          <Text strong>{text}</Text>
          {record.description && (
            <Text type="secondary" style={{ fontSize: '12px' }}>
              {record.description}
            </Text>
          )}
        </Space>
      ),
    },
    {
      title: 'URL',
      dataIndex: 'url',
      key: 'url',
      render: (url) => (
        <Tooltip title={url}>
          <Text copyable={{ text: url }} style={{ maxWidth: 200 }}>
            <LinkOutlined /> {url.length > 30 ? url.substring(0, 30) + '...' : url}
          </Text>
        </Tooltip>
      ),
    },
    {
      title: '认证方式',
      dataIndex: 'auth_type',
      key: 'auth_type',
      render: (authType) => getAuthTypeTag(authType),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status) => getStatusTag(status),
    },
    {
      title: '启用状态',
      dataIndex: 'is_enabled',
      key: 'is_enabled',
      render: (enabled, record) => (
        <Switch
          checked={enabled}
          onChange={() => handleToggleEnabled(record)}
          size="small"
        />
      ),
    },
    {
      title: '标签',
      dataIndex: 'tags',
      key: 'tags',
      render: renderTags,
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      sorter: true,
      render: (time) => new Date(time).toLocaleString(),
    },
    {
      title: '操作',
      key: 'action',
      render: (_, record) => (
        <Space size="small">
          <Tooltip title="编辑">
            <Button
              type="text"
              icon={<EditOutlined />}
              onClick={() => handleEdit(record)}
            />
          </Tooltip>
          <Popconfirm
            title="确定要删除这个服务器吗？"
            onConfirm={() => handleDelete(record.id)}
            okText="确定"
            cancelText="取消"
          >
            <Tooltip title="删除">
              <Button
                type="text"
                danger
                icon={<DeleteOutlined />}
              />
            </Tooltip>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  // 组件挂载时加载数据
  useEffect(() => {
    loadServers();
  }, [state.current, state.pageSize, state.filters, state.sorter]);

  return (
    <Card>
      {/* 头部操作区 */}
      <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
        <Col flex="auto">
          <Space size="middle">
            <Search
              placeholder="搜索服务器名称、描述或标签"
              allowClear
              style={{ width: 300 }}
              onSearch={handleSearch}
            />
            <Select
              placeholder="状态筛选"
              allowClear
              style={{ width: 120 }}
              onChange={handleStatusFilter}
            >
              <Option value="active">活跃</Option>
              <Option value="inactive">非活跃</Option>
              <Option value="error">错误</Option>
            </Select>
            <Select
              placeholder="启用状态"
              allowClear
              style={{ width: 120 }}
              onChange={handleEnabledFilter}
            >
              <Option value={true}>已启用</Option>
              <Option value={false}>已禁用</Option>
            </Select>
          </Space>
        </Col>
        <Col>
          <Space>
            <Button
              icon={<ReloadOutlined />}
              onClick={loadServers}
              loading={state.loading}
            >
              刷新
            </Button>
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={handleAdd}
            >
              新增服务器
            </Button>
          </Space>
        </Col>
      </Row>

      {/* 统计信息 */}
      <Row gutter={16} style={{ marginBottom: 16 }}>
        <Col span={6}>
          <Badge count={state.total} showZero color="#1890ff">
            <div style={{ padding: '8px 12px', background: '#f0f2f5', borderRadius: 4 }}>
              总计
            </div>
          </Badge>
        </Col>
        <Col span={6}>
          <Badge count={state.data.filter(s => s.status === 'active').length} showZero color="#52c41a">
            <div style={{ padding: '8px 12px', background: '#f0f2f5', borderRadius: 4 }}>
              活跃
            </div>
          </Badge>
        </Col>
        <Col span={6}>
          <Badge count={state.data.filter(s => s.is_enabled).length} showZero color="#722ed1">
            <div style={{ padding: '8px 12px', background: '#f0f2f5', borderRadius: 4 }}>
              已启用
            </div>
          </Badge>
        </Col>
      </Row>

      {/* 表格 */}
      <Table<MCPServer>
        columns={columns}
        dataSource={state.data}
        rowKey="id"
        loading={state.loading}
        pagination={{
          current: state.current,
          pageSize: state.pageSize,
          total: state.total,
          showSizeChanger: true,
          showQuickJumper: true,
          showTotal: (total, range) => `第 ${range[0]}-${range[1]} 条，共 ${total} 条`,
        }}
        onChange={handleTableChange}
        scroll={{ x: 1200 }}
      />

      {/* 表单弹窗 */}
      <Modal
        title={editingServer ? '编辑MCP服务器' : '新增MCP服务器'}
        open={formVisible}
        onCancel={() => setFormVisible(false)}
        footer={null}
        width={800}
        destroyOnClose
      >
        <MCPServerForm
          server={editingServer}
          onSuccess={handleFormSuccess}
          onCancel={() => setFormVisible(false)}
        />
      </Modal>
    </Card>
  );
};

export default MCPServerList;