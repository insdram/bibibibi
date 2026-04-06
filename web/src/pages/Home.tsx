import React, { useState, useEffect } from 'react';
import { Layout, Card, Avatar, Button, Space, Tag, Dropdown, message, Pagination, Spin, List, Tabs, Form, Input, Divider, Switch, Modal, Menu } from 'antd';
import { PlusOutlined, MessageOutlined, MoreOutlined, DeleteOutlined, PushpinOutlined, LockOutlined, HomeOutlined, UserOutlined, UserOutlined as ProfileIcon, MailOutlined, LockOutlined as PasswordIcon, SmileOutlined, LikeOutlined, LikeFilled, SettingOutlined, MoonOutlined, SunOutlined, ApiOutlined, DeleteOutlined as ClearOutlined } from '@ant-design/icons';
import type { MenuProps } from 'antd';
import { useNavigate, useLocation } from 'react-router-dom';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { bibiApi, userApi } from '../api';
import { useAuth } from '../stores/AuthContext';
import { useTheme } from '../stores/ThemeContext';
import CommentSection from '../components/CommentSection';
import BibiEditor from '../components/BibiEditor';

const { Sider, Content } = Layout;
const { TabPane } = Tabs;

const API_ADDRESS = import.meta.env.VITE_API_URL || '/api/v1';

interface Bibi {
  id: number;
  content: string;
  visibility: string;
  pinned: boolean;
  like_count: number;
  liked: boolean;
  created_at: string;
  updated_at: string;
  creator: {
    id: number;
    username: string;
    nickname: string;
    avatar: string;
  };
  tags: Array<{ id: number; name: string }>;
  comments: Array<any>;
}

interface UserInfo {
  id: number;
  username: string;
  nickname: string;
  email: string;
  avatar: string;
}

const Home: React.FC = () => {
  const { user, logout } = useAuth();
  const { darkMode, toggleDarkMode } = useTheme();
  const navigate = useNavigate();
  const location = useLocation();
  const [bibis, setBibis] = useState<Bibi[]>([]);
  const [loading, setLoading] = useState(true);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [showEditor, setShowEditor] = useState(false);
  const [expandedComments, setExpandedComments] = useState<number | null>(null);
  const [activeTab, setActiveTab] = useState('home');
  const [userInfo, setUserInfo] = useState<UserInfo | null>(null);
  const [profileLoading, setProfileLoading] = useState(false);
  const [updateLoading, setUpdateLoading] = useState(false);
  const [form] = Form.useForm();

  const fetchBibis = async () => {
    try {
      setLoading(true);
      const response = await bibiApi.getBibis({ page, page_size: 20 });
      const bibisWithLiked = (response.data.bibis || []).map((bibi: Bibi) => ({
        ...bibi,
        liked: bibi.liked || false,
      }));
      setBibis(bibisWithLiked);
      setTotal(response.data.total || 0);
    } catch (error) {
      console.error('获取笔记失败:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchBibis();
  }, [page]);

  const fetchUserInfo = async () => {
    try {
      const response = await userApi.getCurrentUser();
      setUserInfo(response.data);
      form.setFieldsValue({
        username: response.data.username,
        nickname: response.data.nickname,
        email: response.data.email,
      });
    } catch (error) {
      console.error('获取用户信息失败:', error);
    }
  };

  useEffect(() => {
    if (activeTab === 'profile') {
      fetchUserInfo();
    }
  }, [activeTab]);

  const handleCreateBibi = async (content: string, visibility: string, tagIds: number[]) => {
    try {
      await bibiApi.createBibi({ content, visibility, tag_ids: tagIds });
      setShowEditor(false);
      message.success('发布成功');
      fetchBibis();
    } catch (error) {
      console.error('创建笔记失败:', error);
      message.error('发布失败');
    }
  };

  const handleDeleteBibi = async (id: number) => {
    try {
      await bibiApi.deleteBibi(id);
      message.success('删除成功');
      fetchBibis();
    } catch (error) {
      console.error('删除笔记失败:', error);
      message.error('删除失败');
    }
  };

  const handleTogglePin = async (id: number) => {
    try {
      await bibiApi.togglePin(id);
      message.success('操作成功');
      fetchBibis();
    } catch (error) {
      console.error('切换置顶状态失败:', error);
      message.error('操作失败');
    }
  };

  const handleToggleLike = async (id: number) => {
    try {
      const response = await bibiApi.toggleLike(id);
      const { liked } = response.data;
      // 乐观更新
      setBibis((prev) =>
        prev.map((bibi) =>
          bibi.id === id
            ? { ...bibi, like_count: liked ? bibi.like_count + 1 : bibi.like_count - 1, liked }
            : bibi
        )
      );
    } catch (error) {
      console.error('切换点赞状态失败:', error);
    }
  };

  const handleUpdateProfile = async (values: { username: string; nickname: string; email: string; password?: string }) => {
    setUpdateLoading(true);
    try {
      const response = await userApi.updateCurrentUser(values);
      setUserInfo(response.data);
      message.success('更新成功');
    } catch (error: any) {
      message.error(error.response?.data?.error || '更新失败');
    } finally {
      setUpdateLoading(false);
    }
  };

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  const handleClearData = () => {
    Modal.confirm({
      title: '确认清空',
      content: '确定要清空所有用户数据吗？此操作不可恢复！',
      okText: '确认清空',
      okType: 'danger',
      cancelText: '取消',
      onOk: () => {
        localStorage.clear();
        sessionStorage.clear();
        message.success('数据已清空');
        handleLogout();
      },
    });
  };

  const handleThemeChange = (checked: boolean) => {
    toggleDarkMode();
  };

  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleString('zh-CN', { month: 'numeric', day: 'numeric', hour: '2-digit', minute: '2-digit' });
  };

  const renderItemActions = (bibi: Bibi, isOwner: boolean) => {
    const items: MenuProps['items'] = [];

    if (isOwner) {
      items.push({
        key: 'pin',
        icon: <PushpinOutlined />,
        label: bibi.pinned ? '取消置顶' : '置顶',
        onClick: () => handleTogglePin(bibi.id),
      });
      items.push({
        key: 'delete',
        icon: <DeleteOutlined />,
        label: '删除',
        danger: true,
        onClick: () => handleDeleteBibi(bibi.id),
      });
    }

    return items;
  };

  const renderCardHeader = (bibi: Bibi, isOwner: boolean) => (
    <div className="flex items-start justify-between">
      <Space size={12}>
        <Avatar src={bibi.creator.avatar} size={40}>
          {(bibi.creator.nickname || bibi.creator.username).charAt(0).toUpperCase()}
        </Avatar>
        <div>
          <div className="font-medium dark:text-white">
            {bibi.creator.nickname || bibi.creator.username}
          </div>
          <div className="text-xs dark:text-gray-400">
            {formatDate(bibi.created_at)}
            {bibi.visibility === 'PRIVATE' && (
              <span className="ml-2 dark:text-gray-400">
                <LockOutlined /> 仅自己可见
              </span>
            )}
            {bibi.pinned && (
              <Tag color="gold" className="ml-2">置顶</Tag>
            )}
          </div>
        </div>
      </Space>
      {isOwner && (
        <Dropdown menu={{ items: renderItemActions(bibi, isOwner) }} placement="bottomRight">
          <Button type="text" icon={<MoreOutlined />} />
        </Dropdown>
      )}
    </div>
  );

  const renderCardContent = (bibi: Bibi) => (
    <div className="markdown-body">
      <ReactMarkdown remarkPlugins={[remarkGfm]}>{bibi.content}</ReactMarkdown>
      {bibi.tags && bibi.tags.length > 0 && (
        <div className="flex flex-wrap gap-1 mb-3 mt-3">
          {bibi.tags.map((tag) => (
            <Tag key={tag.id} color="blue">#{tag.name}</Tag>
          ))}
        </div>
      )}
    </div>
  );

  const renderCardFooter = (bibi: Bibi) => (
    <div className="flex items-center justify-between border-t border-[#f0f0f0] dark:border-[#303030] pt-3 mt-3">
      <Space size={16}>
        <Button
          type="text"
          icon={bibi.liked ? <LikeFilled style={{ color: '#ff4d4f' }} /> : <LikeOutlined />}
          onClick={() => handleToggleLike(bibi.id)}
          className={bibi.liked ? 'text-red-500 dark:text-red-400' : 'text-gray-500 dark:text-gray-400'}
        >
          {bibi.like_count || 0}
        </Button>
        <Button
          type="text"
          icon={<MessageOutlined />}
          onClick={() => setExpandedComments(expandedComments === bibi.id ? null : bibi.id)}
          className="text-gray-500 dark:text-gray-400"
        >
          {bibi.comments?.length || 0}
        </Button>
      </Space>
    </div>
  );

  const renderEmpty = () => (
    <div className="text-center py-16">
      <div className="text-gray-400 dark:text-gray-500 mb-4">还没有动态</div>
      <Button type="primary" icon={<PlusOutlined />} onClick={() => setShowEditor(true)}>
        发布第一条
      </Button>
    </div>
  );

  const renderNotesList = () => (
    <>
      {showEditor && (
        <Card
          className="mb-6"
          title="发布动态"
          extra={<Button onClick={() => setShowEditor(false)}>收起</Button>}
        >
          <BibiEditor
            onSubmit={handleCreateBibi}
            onCancel={() => setShowEditor(false)}
          />
        </Card>
      )}

      <Card>
        <Spin spinning={loading}>
          {bibis.length === 0 && !loading ? (
            renderEmpty()
          ) : (
            <List
              dataSource={bibis}
              renderItem={(bibi) => (
                <Card
                  className="mb-4"
                  bodyStyle={{ padding: '16px' }}
                  styles={{ body: { padding: '16px' } }}
                >
                  {renderCardHeader(bibi, user?.id === bibi.creator.id)}
                  {renderCardContent(bibi)}
                  {renderCardFooter(bibi)}
                  {expandedComments === bibi.id && (
                    <div className="mt-4 pt-4 border-t border-gray-200 dark:border-gray-700">
                      <CommentSection
                        bibiId={bibi.id}
                        comments={bibi.comments || []}
                        onUpdate={fetchBibis}
                      />
                    </div>
                  )}
                </Card>
              )}
            />
          )}
        </Spin>
      </Card>

      {total > 20 && (
        <div className="flex justify-center mt-6">
          <Pagination
            current={page}
            total={total}
            pageSize={20}
            onChange={setPage}
            showSizeChanger={false}
            showTotal={(total) => `共 ${total} 条`}
          />
        </div>
      )}
    </>
  );

  const renderProfile = () => (
    <Card title="个人信息">
      <Spin spinning={profileLoading}>
        <Form
          form={form}
          layout="vertical"
          onFinish={handleUpdateProfile}
        >
          <Form.Item
            name="username"
            label="用户名"
            rules={[{ required: true, message: '请输入用户名' }]}
          >
            <Input prefix={<UserOutlined />} placeholder="用户名" />
          </Form.Item>

          <Form.Item
            name="nickname"
            label="昵称"
          >
            <Input prefix={<SmileOutlined />} placeholder="昵称" />
          </Form.Item>

          <Form.Item
            name="email"
            label="邮箱"
            rules={[
              { required: true, message: '请输入邮箱' },
              { type: 'email', message: '请输入有效的邮箱地址' }
            ]}
          >
            <Input prefix={<MailOutlined />} placeholder="邮箱" />
          </Form.Item>

          <Divider>修改密码（不修改请留空）</Divider>

          <Form.Item
            name="password"
            label="新密码"
          >
            <Input.Password prefix={<PasswordIcon />} placeholder="请输入新密码" />
          </Form.Item>

          <Form.Item
            name="confirmPassword"
            label="确认密码"
            dependencies={['password']}
            rules={[
              ({ getFieldValue }) => ({
                validator(_, value) {
                  if (!value || getFieldValue('password') === value) {
                    return Promise.resolve();
                  }
                  return Promise.reject(new Error('两次输入的密码不一致'));
                },
              }),
            ]}
          >
            <Input.Password prefix={<PasswordIcon />} placeholder="请确认新密码" />
          </Form.Item>

          <Form.Item>
            <Button type="primary" htmlType="submit" loading={updateLoading} block size="large">
              保存修改
            </Button>
          </Form.Item>
        </Form>
      </Spin>
    </Card>
  );

  const renderSettings = () => (
    <Card title="系统设置">
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <div className="font-medium dark:text-white">主题模式</div>
            <div className="text-sm text-gray-500 dark:text-gray-400">切换白天/黑夜模式</div>
          </div>
          <Switch
            checked={darkMode}
            onChange={handleThemeChange}
            checkedChildren={<MoonOutlined />}
            unCheckedChildren={<SunOutlined />}
          />
        </div>

        <Divider />

        <div>
          <div className="flex items-center justify-between mb-2">
            <div>
              <div className="font-medium dark:text-white">API 地址</div>
              <div className="text-sm text-gray-500 dark:text-gray-400">后端服务接口地址</div>
            </div>
          </div>
          <Input value={API_ADDRESS} disabled addonBefore={<ApiOutlined />} />
        </div>

        <Divider />

        <div>
          <div className="font-medium mb-2 dark:text-white">清空数据</div>
          <div className="text-sm text-gray-500 dark:text-gray-400 mb-3">清空所有本地缓存数据，包括登录信息和评论记录</div>
          <Button danger icon={<ClearOutlined />} onClick={handleClearData}>
            清空所有数据
          </Button>
        </div>

        <Divider />

        <div>
          <div className="font-medium mb-2 dark:text-white">关于</div>
          <div className="text-sm text-gray-500 dark:text-gray-400">
            <p>bibibibi v1.0.0</p>
            <p>一个简洁的开源笔记应用</p>
          </div>
        </div>
      </div>
    </Card>
  );

  const menuItems = [
    {
      key: 'home',
      icon: <HomeOutlined />,
      label: '首页',
      onClick: () => setActiveTab('home'),
    },
    {
      key: 'profile',
      icon: <ProfileIcon />,
      label: '个人中心',
      onClick: () => setActiveTab('profile'),
    },
    {
      key: 'settings',
      icon: <SettingOutlined />,
      label: '系统设置',
      onClick: () => setActiveTab('settings'),
    },
  ];

  return (
    <Layout hasSider>
      <Sider
        width={240}
        style={{
          height: '100vh',
          position: 'fixed',
          left: 0,
          top: 0,
          overflow: 'auto',
        }}
      >
        <div style={{ padding: '24px 16px', textAlign: 'center' }}>
          <Avatar
            size={64}
            src={user?.avatar}
            style={{ marginBottom: 12 }}
          >
            {!user?.avatar && (user?.nickname?.charAt(0).toUpperCase() || user?.username?.charAt(0).toUpperCase() || 'U')}
          </Avatar>
            <div style={{ fontWeight: 500, marginBottom: 4, color: 'rgba(255,255,255,0.85)' }}>{user?.nickname || user?.username}</div>
          <div style={{ fontSize: 12, color: 'rgba(255,255,255,0.45)' }}>@{user?.username}</div>
        </div>

        <Button
          type="primary"
          icon={<PlusOutlined />}
          size="large"
          style={{ margin: '0 16px 24px', width: 'calc(100% - 32px)' }}
          onClick={() => { setShowEditor(true); setActiveTab('home'); }}
        >
          发布
        </Button>

        <Menu
          mode="inline"
          selectedKeys={[activeTab]}
          onClick={({ key }) => setActiveTab(key)}
          style={{ border: 'none' }}
          items={menuItems.map((item) => ({
            key: item.key,
            icon: item.icon,
            label: item.label,
          }))}
        />

        <div style={{ position: 'absolute', bottom: 24, left: 0, right: 0, padding: '16px 24px', borderTop: '1px solid rgba(255,255,255,0.14)' }}>
          <Button type="text" block onClick={handleLogout} style={{ textAlign: 'left', color: 'rgba(255,255,255,0.65)' }}>
            退出登录
          </Button>
        </div>
      </Sider>

      <Layout style={{ marginLeft: 240 }}>
        <Content style={{ padding: '24px 24px 0', minHeight: '100vh' }}>
          <Tabs activeKey={activeTab} onChange={setActiveTab}>
            <TabPane tab="笔记列表" key="home">
              {activeTab === 'home' && renderNotesList()}
            </TabPane>
            <TabPane tab="个人中心" key="profile">
              {activeTab === 'profile' && renderProfile()}
            </TabPane>
            <TabPane tab="系统设置" key="settings">
              {activeTab === 'settings' && renderSettings()}
            </TabPane>
          </Tabs>
        </Content>
      </Layout>
    </Layout>
  );
};

export default Home;