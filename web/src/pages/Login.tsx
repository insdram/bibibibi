import React, { useState, useEffect } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { useAuth } from '../stores/AuthContext';
import { Form, Input, Button, Card, message } from 'antd';
import { UserOutlined, LockOutlined } from '@ant-design/icons';
import { systemApi } from '../api';

const Login: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [registrationEnabled, setRegistrationEnabled] = useState(true);
  const { login } = useAuth();
  const navigate = useNavigate();

  useEffect(() => {
    const fetchSettings = async () => {
      try {
        const response = await systemApi.getPublicSettings();
        setRegistrationEnabled(response.data.registration_enabled);
      } catch (error) {
        console.error('获取设置失败:', error);
      }
    };
    fetchSettings();
  }, []);

  const onFinish = async (values: { username: string; password: string }) => {
    setLoading(true);
    try {
      await login(values.username, values.password);
      message.success('登录成功');
      navigate('/');
    } catch (err: any) {
      message.error(err.response?.data?.error || '登录失败');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{ 
      minHeight: '100vh', 
      display: 'flex', 
      alignItems: 'center', 
      justifyContent: 'center',
    }}>
      <Card style={{ width: 400 }}>
        <div style={{ textAlign: 'center', marginBottom: 24 }}>
          <h1 style={{ fontSize: 24, fontWeight: 'bold', margin: 0 }}>哔哔哔哔</h1>
        </div>

        <Form
          name="login"
          onFinish={onFinish}
          autoComplete="off"
          layout="vertical"
        >
          <Form.Item
            name="username"
            rules={[{ required: true, message: '请输入用户名' }]}
          >
            <Input prefix={<UserOutlined />} placeholder="用户名" size="large" />
          </Form.Item>

          <Form.Item
            name="password"
            rules={[{ required: true, message: '请输入密码' }]}
          >
            <Input.Password prefix={<LockOutlined />} placeholder="密码" size="large" />
          </Form.Item>

          <Form.Item>
            <Button type="primary" htmlType="submit" loading={loading} block size="large">
              登录
            </Button>
          </Form.Item>
        </Form>

        {registrationEnabled && (
          <div style={{ textAlign: 'center' }}>
            没有账号？<Link to="/register">立即注册</Link>
          </div>
        )}
      </Card>
    </div>
  );
};

export default Login;