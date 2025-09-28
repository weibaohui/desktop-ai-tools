import React, { useState } from 'react';
import { 
  Button, 
  Input, 
  Card, 
  Space, 
  Typography, 
  message as antdMessage, 
  Row, 
  Col,
  Divider,
  Alert
} from 'antd';
import { SendOutlined, ApiOutlined } from '@ant-design/icons';
import axios from 'axios';
import { Greet } from "../wailsjs/go/main/App";
import './App.css';

const { Title, Text } = Typography;

/**
 * 主应用组件
 * 实现了基于Ant Design + Gin的Hello World功能
 */
function App() {
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState('');
  const [response, setResponse] = useState('');
  const [wailsName, setWailsName] = useState('');
  const [wailsResult, setWailsResult] = useState("请在下方输入您的姓名 👇");

  /**
   * 调用Gin后端API的Hello接口
   */
  const handleSendToBackend = async () => {
    if (!message.trim()) {
      antdMessage.warning('请输入消息内容');
      return;
    }

    setLoading(true);
    try {
      const response = await axios.post('http://localhost:8080/api/hello', {
        message: message
      });

      if (response.data.success) {
        setResponse(response.data.response);
        antdMessage.success('消息发送成功！');
      } else {
        antdMessage.error('发送失败');
      }
    } catch (error) {
      console.error('Error calling backend:', error);
      antdMessage.error('连接后端失败，请确保服务器正在运行');
    } finally {
      setLoading(false);
    }
  };

  /**
   * 调用原有的Wails方法
   */
  const handleWailsGreet = () => {
    if (!wailsName.trim()) {
      antdMessage.warning('请输入姓名');
      return;
    }
    Greet(wailsName).then(setWailsResult);
  };

  return (
    <div style={{ padding: '24px', minHeight: '100vh', backgroundColor: '#f5f5f5' }}>
      <Row justify="center">
        <Col xs={24} sm={20} md={16} lg={12} xl={10}>
          <Space direction="vertical" size="large" style={{ width: '100%' }}>
            
            {/* 标题 */}
            <Card>
              <div style={{ textAlign: 'center' }}>
                <Title level={2}>
                  <ApiOutlined style={{ color: '#1890ff', marginRight: '8px' }} />
                  Ant Design + Gin 框架演示
                </Title>
                <Text type="secondary">
                  基于Wails的桌面应用，集成了Ant Design前端和Gin后端
                </Text>
              </div>
            </Card>

            {/* Gin API 演示 */}
            <Card 
              title={
                <span>
                  <SendOutlined style={{ marginRight: '8px' }} />
                  Gin 后端 API 演示
                </span>
              }
              bordered={false}
            >
              <Space direction="vertical" size="middle" style={{ width: '100%' }}>
                <div>
                  <Text strong>发送消息到后端：</Text>
                  <Input.TextArea
                    value={message}
                    onChange={(e) => setMessage(e.target.value)}
                    placeholder="请输入要发送给后端的消息..."
                    rows={3}
                    style={{ marginTop: '8px' }}
                  />
                </div>
                
                <Button 
                  type="primary" 
                  icon={<SendOutlined />}
                  loading={loading}
                  onClick={handleSendToBackend}
                  size="large"
                  style={{ width: '100%' }}
                >
                  发送到 Gin 后端
                </Button>

                {response && (
                  <Alert
                    message="后端响应"
                    description={response}
                    type="success"
                    showIcon
                    style={{ marginTop: '16px' }}
                  />
                )}
              </Space>
            </Card>

            <Divider>或者</Divider>

            {/* 原有Wails方法演示 */}
            <Card 
              title="原有 Wails 方法演示"
              bordered={false}
            >
              <Space direction="vertical" size="middle" style={{ width: '100%' }}>
                <Alert
                  message={wailsResult}
                  type="info"
                  showIcon
                />
                
                <Input
                  value={wailsName}
                  onChange={(e) => setWailsName(e.target.value)}
                  placeholder="请输入您的姓名"
                  size="large"
                />
                
                <Button 
                  type="default" 
                  onClick={handleWailsGreet}
                  size="large"
                  style={{ width: '100%' }}
                >
                  Wails 问候
                </Button>
              </Space>
            </Card>

          </Space>
        </Col>
      </Row>
    </div>
  );
}

export default App;
