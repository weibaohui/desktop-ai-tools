import React, { useState, useEffect } from 'react';
import {
  Form,
  Input,
  Select,
  Button,
  Space,
  Card,
  Row,
  Col,
  Switch,
  message,
  Divider,
  Alert,
  Tooltip,
  Tag,
} from 'antd';
import {
  SaveOutlined,
  CloseOutlined,
  ApiOutlined,
  InfoCircleOutlined,
  TagsOutlined,
  PlusOutlined,
} from '@ant-design/icons';
import type { MCPServer, MCPServerFormData, AuthConfig } from '../types/mcpServer';
import { MCPServerService } from '../services/mcpServerService';

const { Option } = Select;
const { TextArea } = Input;

interface MCPServerFormProps {
  server?: MCPServer | null;
  onSuccess: () => void;
  onCancel: () => void;
}

/**
 * MCP Server表单组件
 */
const MCPServerForm: React.FC<MCPServerFormProps> = ({
  server,
  onSuccess,
  onCancel,
}) => {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [testLoading, setTestLoading] = useState(false);
  const [authConfig, setAuthConfig] = useState<AuthConfig>({});
  const [tags, setTags] = useState<string[]>([]);
  const [inputTag, setInputTag] = useState('');

  /**
   * 初始化表单数据
   */
  useEffect(() => {
    if (server) {
      // 解析认证配置
      let parsedAuthConfig: AuthConfig = {};
      try {
        if (server.auth_config) {
          parsedAuthConfig = JSON.parse(server.auth_config);
        }
      } catch (error) {
        console.error('解析认证配置失败:', error);
      }

      // 解析标签
      const parsedTags = server.tags ? server.tags.split(',').map(tag => tag.trim()).filter(Boolean) : [];

      // 设置表单值
      form.setFieldsValue({
        name: server.name,
        description: server.description,
        url: server.url,
        auth_type: server.auth_type,
        is_enabled: server.is_enabled,
      });

      setAuthConfig(parsedAuthConfig);
      setTags(parsedTags);
    } else {
      // 新增时的默认值
      form.setFieldsValue({
        auth_type: 'none',
        is_enabled: true,
      });
      setAuthConfig({});
      setTags([]);
    }
  }, [server, form]);

  /**
   * 处理认证类型变化
   */
  const handleAuthTypeChange = (value: string) => {
    setAuthConfig({});
    form.setFieldsValue({ auth_config: {} });
  };

  /**
   * 处理认证配置变化
   */
  const handleAuthConfigChange = (field: string, value: string) => {
    const newConfig = { ...authConfig, [field]: value };
    setAuthConfig(newConfig);
  };

  /**
   * 添加标签
   */
  const handleAddTag = () => {
    if (inputTag && !tags.includes(inputTag)) {
      setTags([...tags, inputTag]);
      setInputTag('');
    }
  };

  /**
   * 删除标签
   */
  const handleRemoveTag = (tagToRemove: string) => {
    setTags(tags.filter(tag => tag !== tagToRemove));
  };

  /**
   * 测试连接
   */
  const handleTestConnection = async () => {
    try {
      const url = form.getFieldValue('url');
      if (!url) {
        message.warning('请先输入服务器URL');
        return;
      }

      setTestLoading(true);
      const success = await MCPServerService.testConnection(url, authConfig);
      
      if (success) {
        message.success('连接测试成功');
      } else {
        message.error('连接测试失败');
      }
    } catch (error) {
      message.error('连接测试失败: ' + (error as Error).message);
    } finally {
      setTestLoading(false);
    }
  };

  /**
   * 提交表单
   */
  const handleSubmit = async (values: any) => {
    try {
      setLoading(true);

      const formData = {
        name: values.name,
        description: values.description || '',
        url: values.url,
        auth_type: values.auth_type,
        auth_config: JSON.stringify(authConfig),
        tags: tags.join(','),
      };

      if (server) {
        // 更新
        await MCPServerService.update(server.id, {
          ...formData,
          is_enabled: values.is_enabled,
        });
        message.success('更新成功');
      } else {
        // 创建
        await MCPServerService.create(formData);
        message.success('创建成功');
      }

      onSuccess();
    } catch (error) {
      message.error((server ? '更新' : '创建') + '失败: ' + (error as Error).message);
    } finally {
      setLoading(false);
    }
  };

  /**
   * 渲染认证配置表单
   */
  const renderAuthConfig = () => {
    const authType = form.getFieldValue('auth_type');

    switch (authType) {
      case 'bearer':
        return (
          <Form.Item label="Bearer Token">
            <Input.Password
              placeholder="请输入Bearer Token"
              value={authConfig.token || ''}
              onChange={(e) => handleAuthConfigChange('token', e.target.value)}
            />
          </Form.Item>
        );

      case 'basic':
        return (
          <>
            <Form.Item label="用户名">
              <Input
                placeholder="请输入用户名"
                value={authConfig.username || ''}
                onChange={(e) => handleAuthConfigChange('username', e.target.value)}
              />
            </Form.Item>
            <Form.Item label="密码">
              <Input.Password
                placeholder="请输入密码"
                value={authConfig.password || ''}
                onChange={(e) => handleAuthConfigChange('password', e.target.value)}
              />
            </Form.Item>
          </>
        );

      case 'api_key':
        return (
          <Form.Item label="API Key">
            <Input.Password
              placeholder="请输入API Key"
              value={authConfig.api_key || ''}
              onChange={(e) => handleAuthConfigChange('api_key', e.target.value)}
            />
          </Form.Item>
        );

      default:
        return null;
    }
  };

  return (
    <Form
      form={form}
      layout="vertical"
      onFinish={handleSubmit}
      autoComplete="off"
    >
      <Row gutter={16}>
        <Col span={12}>
          <Form.Item
            label="服务器名称"
            name="name"
            rules={[
              { required: true, message: '请输入服务器名称' },
              { min: 1, max: 100, message: '名称长度应在1-100字符之间' },
            ]}
          >
            <Input placeholder="请输入服务器名称" />
          </Form.Item>
        </Col>
        <Col span={12}>
          <Form.Item
            label="服务器URL"
            name="url"
            rules={[
              { required: true, message: '请输入服务器URL' },
              { type: 'url', message: '请输入有效的URL' },
            ]}
          >
            <Input
              placeholder="https://api.example.com/mcp"
              addonAfter={
                <Tooltip title="测试连接">
                  <Button
                    type="text"
                    icon={<ApiOutlined />}
                    loading={testLoading}
                    onClick={handleTestConnection}
                  />
                </Tooltip>
              }
            />
          </Form.Item>
        </Col>
      </Row>

      <Form.Item
        label="描述"
        name="description"
        rules={[{ max: 500, message: '描述长度不能超过500字符' }]}
      >
        <TextArea
          rows={3}
          placeholder="请输入服务器描述（可选）"
          showCount
          maxLength={500}
        />
      </Form.Item>

      <Divider orientation="left">认证配置</Divider>

      <Row gutter={16}>
        <Col span={12}>
          <Form.Item
            label="认证方式"
            name="auth_type"
            rules={[{ required: true, message: '请选择认证方式' }]}
          >
            <Select onChange={handleAuthTypeChange}>
              <Option value="none">无认证</Option>
              <Option value="bearer">Bearer Token</Option>
              <Option value="basic">Basic Auth</Option>
              <Option value="api_key">API Key</Option>
            </Select>
          </Form.Item>
        </Col>
        {server && (
          <Col span={12}>
            <Form.Item
              label="启用状态"
              name="is_enabled"
              valuePropName="checked"
            >
              <Switch checkedChildren="启用" unCheckedChildren="禁用" />
            </Form.Item>
          </Col>
        )}
      </Row>

      {renderAuthConfig()}

      <Divider orientation="left">标签管理</Divider>

      <Form.Item label="标签">
        <Space direction="vertical" style={{ width: '100%' }}>
          <Space wrap>
            {tags.map((tag, index) => (
              <Tag
                key={index}
                closable
                onClose={() => handleRemoveTag(tag)}
                icon={<TagsOutlined />}
              >
                {tag}
              </Tag>
            ))}
          </Space>
          <Space.Compact style={{ width: '100%' }}>
            <Input
              placeholder="输入标签名称"
              value={inputTag}
              onChange={(e) => setInputTag(e.target.value)}
              onPressEnter={handleAddTag}
            />
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={handleAddTag}
              disabled={!inputTag || tags.includes(inputTag)}
            >
              添加
            </Button>
          </Space.Compact>
        </Space>
      </Form.Item>

      <Alert
        message="提示"
        description="创建后可以在列表中管理服务器状态，包括启用/禁用、状态更新等操作。"
        type="info"
        icon={<InfoCircleOutlined />}
        style={{ marginBottom: 16 }}
      />

      <Form.Item>
        <Space>
          <Button
            type="primary"
            htmlType="submit"
            loading={loading}
            icon={<SaveOutlined />}
          >
            {server ? '更新' : '创建'}
          </Button>
          <Button
            onClick={onCancel}
            icon={<CloseOutlined />}
          >
            取消
          </Button>
        </Space>
      </Form.Item>
    </Form>
  );
};

export default MCPServerForm;