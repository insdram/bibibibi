import React, { useState, useEffect } from 'react';
import { Layout, Card, Avatar, Button, Space, Tag, Dropdown, message, Pagination, Spin, List, Tabs, Form, Input, Divider, Switch, Modal, Menu, Segmented, Select, Table, Typography, Skeleton } from 'antd';
import { PlusOutlined, MessageOutlined, MoreOutlined, DeleteOutlined, PushpinOutlined, LockOutlined, HomeOutlined, UserOutlined, UserOutlined as ProfileIcon, MailOutlined, LockOutlined as PasswordIcon, SmileOutlined, LikeOutlined, LikeFilled, SettingOutlined, MoonOutlined, SunOutlined, ApiOutlined, DeleteOutlined as ClearOutlined, GlobalOutlined } from '@ant-design/icons';
import type { MenuProps } from 'antd';
import axios from 'axios';
import { useNavigate, useLocation } from 'react-router-dom';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { bibiApi, userApi, systemApi, tokenApi, feedApi } from '../api';
import { useAuth } from '../stores/AuthContext';
import { useTheme } from '../stores/ThemeContext';
import CommentSection from '../components/CommentSection';
import BibiEditor from '../components/BibiEditor';

const { Sider, Content } = Layout;
const { TabPane } = Tabs;

const API_ADDRESS = import.meta.env.VITE_API_URL || '/api/v1';

interface Bibi {
  id: string | number;
  content: string;
  visibility: string;
  pinned: boolean;
  like_count: number;
  liked: boolean;
  created_at: string;
  updated_at?: string;
  creator: {
    id: number;
    username: string;
    nickname: string;
    avatar: string;
  };
  tags: Array<{ id: number; name: string }>;
  comments: Array<any>;
  is_remote?: boolean;
  source_id?: number;
  source_url?: string;
}

interface UserInfo {
  id: number;
  username: string;
  nickname: string;
  email: string;
  website?: string;
  is_admin?: boolean;
  avatar: string;
}

const Home: React.FC = () => {
  const { user, logout } = useAuth();
  const { darkMode, themeMode, setThemeMode } = useTheme();
  const navigate = useNavigate();
  const location = useLocation();
  const [bibis, setBibis] = useState<Bibi[]>([]);
  const [loading, setLoading] = useState(true);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [showEditor, setShowEditor] = useState(false);
  const [expandedComments, setExpandedComments] = useState<string | number | null>(null);
  const [likedAnimating, setLikedAnimating] = useState<string | number | null>(null);
  const [activeTab, setActiveTab] = useState('home');
  const [homeSubTab, setHomeSubTab] = useState<'square' | 'mine'>('mine');
  const [userInfo, setUserInfo] = useState<UserInfo | null>(null);
  const [profileLoading, setProfileLoading] = useState(false);
  const [updateLoading, setUpdateLoading] = useState(false);
  const [registrationEnabled, setRegistrationEnabled] = useState(true);
  const [gravatarSource, setGravatarSource] = useState('https://weavatar.com/avatar/');
  const [settingsLoading, setSettingsLoading] = useState(false);
  const [previewImage, setPreviewImage] = useState<string | null>(null);
  const [tokens, setTokens] = useState<any[]>([]);
  const [tokensLoading, setTokensLoading] = useState(false);
  const [createTokenModalVisible, setCreateTokenModalVisible] = useState(false);
  const [createTokenLoading, setCreateTokenLoading] = useState(false);
  const [users, setUsers] = useState<any[]>([]);
  const [usersLoading, setUsersLoading] = useState(false);
  const [feedSources, setFeedSources] = useState<any[]>([]);
  const [feedSourcesLoading, setFeedSourcesLoading] = useState(false);
  const [createFeedSourceModalVisible, setCreateFeedSourceModalVisible] = useState(false);
  const [createFeedSourceLoading, setCreateFeedSourceLoading] = useState(false);
  const [syncLoading, setSyncLoading] = useState(false);
  const [form] = Form.useForm();

  const fetchBibis = async () => {
    try {
      setLoading(true);
      const params: any = { page, page_size: 20 };
      if (homeSubTab === 'mine' && user) {
        params.creator_id = user.id;
      }
      const response = await bibiApi.getBibis(params);
      console.log('笔记广场数据:', response.data);
      const bibisWithLiked = (response.data.bibis || []).map((bibi: Bibi) => {
        // 用 uuid 或本地 id 作为 localStorage 的 key
        const storageId = bibi.is_remote ? String(bibi.id) : bibi.id;
        const likedKey = `bibibibi_liked_${storageId}`;
        const likeCountKey = `bibibibi_likecount_${storageId}`;
        const isLiked = localStorage.getItem(likedKey) === 'true';
        const storedLikeCount = parseInt(localStorage.getItem(likeCountKey) || '0', 10);
        // 如果是远程笔记且有本地存储的点赞数，使用本地的
        const likeCount = bibi.is_remote && storedLikeCount > 0 ? storedLikeCount : bibi.like_count;
        return {
          ...bibi,
          liked: bibi.liked || isLiked,
          like_count: likeCount,
        };
      });
      setBibis(bibisWithLiked);
      setTotal(response.data.total || 0);
    } catch (error) {
      console.error('获取笔记失败:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (activeTab === 'home') {
      fetchBibis();
    }
  }, [page, homeSubTab, user, activeTab]);

  useEffect(() => {
    if (user) {
      setHomeSubTab('mine');
    }
  }, [user]);

  const fetchUserInfo = async () => {
    try {
      const response = await userApi.getCurrentUser();
      setUserInfo(response.data);
      form.setFieldsValue({
        username: response.data.username,
        nickname: response.data.nickname,
        email: response.data.email,
        website: response.data.website || '',
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

  useEffect(() => {
    if (activeTab === 'settings' && user?.is_admin) {
      fetchSettings();
      fetchTokens();
      fetchFeedSources();
    }
  }, [activeTab, user?.is_admin]);

  useEffect(() => {
    if (activeTab === 'members') {
      fetchUsers();
    }
  }, [activeTab]);

  useEffect(() => {
    const handleImageClick = (e: MouseEvent) => {
      const target = e.target as HTMLElement;
      if (target.tagName === 'IMG') {
        setPreviewImage((target as HTMLImageElement).src);
      }
    };

    document.addEventListener('click', handleImageClick);
    return () => document.removeEventListener('click', handleImageClick);
  }, []);

  const fetchSettings = async () => {
    try {
      const response = await systemApi.getSettings();
      setRegistrationEnabled(response.data.registration_enabled);
      setGravatarSource(response.data.gravatar_source || 'https://www.gravatar.com/avatar/');
    } catch (error) {
      console.error('获取设置失败:', error);
    }
  };

  const fetchUsers = async () => {
    setUsersLoading(true);
    try {
      const response = await userApi.getUsers();
      setUsers(response.data || []);
    } catch (error) {
      console.error('获取用户列表失败:', error);
    } finally {
      setUsersLoading(false);
    }
  };

  const fetchFeedSources = async () => {
    setFeedSourcesLoading(true);
    try {
      const response = await feedApi.getFeedSources();
      setFeedSources(response.data || []);
    } catch (error) {
      console.error('获取广场数据源失败:', error);
    } finally {
      setFeedSourcesLoading(false);
    }
  };

  const handleCreateFeedSource = async (values: { name: string; url: string }) => {
    setCreateFeedSourceLoading(true);
    try {
      await feedApi.createFeedSource(values);
      message.success('数据源创建成功');
      setCreateFeedSourceModalVisible(false);
      form.resetFields();
      fetchFeedSources();
    } catch (error: any) {
      message.error(error.response?.data?.error || '创建失败');
    } finally {
      setCreateFeedSourceLoading(false);
    }
  };

  const handleDeleteFeedSource = async (id: number) => {
    try {
      await feedApi.deleteFeedSource(id);
      message.success('数据源已删除');
      fetchFeedSources();
    } catch (error: any) {
      message.error(error.response?.data?.error || '删除失败');
    }
  };

  const handleSyncFeedSources = async () => {
    setSyncLoading(true);
    try {
      await feedApi.syncFeedSources();
      message.success('同步任务已启动');
    } catch (error: any) {
      message.error(error.response?.data?.error || '同步失败');
    } finally {
      setSyncLoading(false);
    }
  };

  const handleUpdateRegistration = async (enabled: boolean) => {
    setSettingsLoading(true);
    try {
      await systemApi.updateSettings({ registration_enabled: enabled, gravatar_source: gravatarSource });
      setRegistrationEnabled(enabled);
      message.success('设置已更新');
    } catch (error: any) {
      message.error(error.response?.data?.error || '更新失败');
    } finally {
      setSettingsLoading(false);
    }
  };

  const handleUpdateGravatarSource = async (source: string) => {
    setSettingsLoading(true);
    try {
      await systemApi.updateSettings({ registration_enabled: registrationEnabled, gravatar_source: source });
      setGravatarSource(source);
      message.success('设置已更新');
    } catch (error: any) {
      message.error(error.response?.data?.error || '更新失败');
    } finally {
      setSettingsLoading(false);
    }
  };

  const fetchTokens = async () => {
    if (!user) return;
    setTokensLoading(true);
    try {
      const response = await tokenApi.getTokens();
      setTokens(response.data.tokens || []);
    } catch (error) {
      console.error('获取 Token 列表失败:', error);
    } finally {
      setTokensLoading(false);
    }
  };

  const handleCreateToken = async (values: { description: string; expires_in_hours: number }) => {
    setCreateTokenLoading(true);
    try {
      const data = {
        description: values.description,
        expires_in_hours: values.expires_in_hours || undefined,
      };
      await tokenApi.createToken(data);
      message.success('Token 创建成功');
      setCreateTokenModalVisible(false);
      form.resetFields();
      fetchTokens();
    } catch (error: any) {
      message.error(error.response?.data?.error || '创建失败');
    } finally {
      setCreateTokenLoading(false);
    }
  };

  const handleDeleteToken = async (id: number) => {
    try {
      await tokenApi.deleteToken(id);
      message.success('Token 已删除');
      fetchTokens();
    } catch (error: any) {
      message.error(error.response?.data?.error || '删除失败');
    }
  };

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

  const handleToggleLike = async (id: string | number, isRemote?: boolean, sourceUrl?: string) => {
    const LIKE_COOLDOWN = 10000;
    const storageId = String(id);
    const lastLikeKey = `bibibibi_like_${storageId}`;
    const likedKey = `bibibibi_liked_${storageId}`;
    const likeCountKey = `bibibibi_likecount_${storageId}`;
    const lastLikeTime = localStorage.getItem(lastLikeKey);
    const isAlreadyLiked = localStorage.getItem(likedKey) === 'true';
    
    if (isAlreadyLiked) {
      message.warning('已点赞，无需重复点赞');
      return;
    }
    
    if (lastLikeTime) {
      const elapsed = Date.now() - parseInt(lastLikeTime, 10);
      if (elapsed < LIKE_COOLDOWN) {
        message.warning('已点赞，无需重复点赞');
        return;
      }
    }
    
    setLikedAnimating(id);
    
    try {
      if (isRemote && sourceUrl) {
        const baseUrl = sourceUrl.replace(/\/$/, '');
        await axios.post(`${baseUrl}/api/v1/bibis/${id}/like`);
      } else {
        await bibiApi.toggleLike(id);
      }
      localStorage.setItem(lastLikeKey, Date.now().toString());
      localStorage.setItem(likedKey, 'true');
      setBibis((prev) =>
        prev.map((b) => {
          if (String(b.id) === storageId) {
            const newCount = b.like_count + 1;
            localStorage.setItem(likeCountKey, newCount.toString());
            return { ...b, like_count: newCount, liked: true };
          }
          return b;
        })
      );
    } catch (error) {
      console.error('点赞失败:', error);
    } finally {
      setTimeout(() => setLikedAnimating(null), 400);
    }
  };

  const handleUpdateProfile = async (values: { username: string; nickname: string; email: string; website?: string; password?: string }) => {
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
        <Avatar src={bibi.creator.avatar} size={40} className="avatar-hover">
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
    </div>
  );

  const renderCardFooter = (bibi: Bibi) => (
    <div className="flex items-center justify-between border-t border-[#f0f0f0] dark:border-[#303030] pt-3 mt-3">
      <Space size={8} wrap>
        {bibi.tags && bibi.tags.length > 0 && bibi.tags.map((tag) => (
          <Tag key={tag.id} color="blue" className="tag-hover m-0">#{tag.name}</Tag>
        ))}
      </Space>
      <Space size={16}>
        <Button
          type="text"
          icon={bibi.liked ? <LikeFilled style={{ color: '#ff4d4f' }} /> : <LikeOutlined />}
          onClick={() => handleToggleLike(bibi.id, bibi.is_remote, bibi.source_url)}
          className={`${bibi.liked ? 'text-red-500 dark:text-red-400' : 'text-gray-500 dark:text-gray-400'} ${likedAnimating === bibi.id ? 'heart-beat' : ''}`}
        >
          {bibi.like_count || 0}
        </Button>
        <Button
          type="text"
          icon={<MessageOutlined />}
          onClick={() => setExpandedComments(expandedComments === bibi.id ? null : bibi.id)}
          className="text-gray-500 dark:text-gray-400"
        >
          {bibi.comment_count || 0}
        </Button>
      </Space>
    </div>
  );

  const renderEmpty = () => (
    <div className="text-center py-16">
      <div className="text-gray-400 dark:text-gray-500">
        {!user && homeSubTab === 'mine' ? '登录后查看我的笔记' : '还没有动态'}
      </div>
      {!user && homeSubTab === 'mine' && (
        <Button type="primary" className="mt-4" onClick={() => navigate('/login')}>
          登录
        </Button>
      )}
    </div>
  );

  const renderNotesList = () => (
    <>
      {showEditor && (
        <Card
          className="mb-6 zoom-in"
          title="发布动态"
          extra={<Button onClick={() => setShowEditor(false)}>收起</Button>}
        >
          <BibiEditor
            onSubmit={handleCreateBibi}
            onCancel={() => setShowEditor(false)}
          />
        </Card>
      )}

      <Card className="fade-in">
        {loading ? (
          <>
            {[1, 2, 3].map((i) => (
              <Card key={i} className="mb-4 skeleton-pulse" bodyStyle={{ padding: '16px' }}>
                <Skeleton active avatar paragraph={{ rows: 2 }} />
              </Card>
            ))}
          </>
        ) : bibis.length === 0 ? (
          renderEmpty()
        ) : (
          <List
            dataSource={bibis}
            renderItem={(bibi) => (
              <Card
                className="mb-4 card-hover stagger-item"
                bodyStyle={{ padding: '16px' }}
                styles={{ body: { padding: '16px' } }}
              >
                {renderCardHeader(bibi, user?.id === bibi.creator.id)}
                {renderCardContent(bibi)}
                {renderCardFooter(bibi)}
                  {expandedComments === bibi.id && (
                    <div className="mt-4 pt-4 border-t border-[#f0f0f0] dark:border-[#303030] comment-appear">
                      <CommentSection
                        bibiId={bibi.id}
                        comments={bibi.comments || []}
                        onBibiUpdate={fetchBibis}
                        isOwner={user?.id === bibi.creator.id}
                        isRemote={bibi.is_remote}
                        remoteSourceUrl={bibi.source_url}
                      />
                    </div>
                  )}
              </Card>
            )}
          />
        )}
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

          <Form.Item
            name="website"
            label="网址"
          >
            <Input prefix={<GlobalOutlined />} placeholder="个人网站或博客地址" />
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

  const renderMembers = () => (
    <Card title="成员列表">
      <Table
        dataSource={users}
        rowKey="id"
        loading={usersLoading}
        pagination={false}
        columns={[
          {
            title: 'ID',
            dataIndex: 'id',
            key: 'id',
            width: 80,
          },
          {
            title: '用户名',
            dataIndex: 'username',
            key: 'username',
          },
          {
            title: '昵称',
            dataIndex: 'nickname',
            key: 'nickname',
          },
          {
            title: '邮箱',
            dataIndex: 'email',
            key: 'email',
          },
          {
            title: '管理员',
            dataIndex: 'is_admin',
            key: 'is_admin',
            render: (isAdmin: boolean) => isAdmin ? '是' : '否',
            filters: [
              { text: '管理员', value: true },
              { text: '普通用户', value: false },
            ],
            onFilter: (value, record) => record.is_admin === value,
          },
        ]}
      />
    </Card>
  );

  const renderSettings = () => (
    <Card title="系统设置">
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <div className="font-medium dark:text-white">主题模式</div>
            <div className="text-sm text-gray-500 dark:text-gray-400">跟随系统/白天/黑夜模式</div>
          </div>
          <Segmented
            value={themeMode}
            onChange={(value) => setThemeMode(value as any)}
            options={[
              { label: <SunOutlined />, value: 'light' },
              { label: <MoonOutlined />, value: 'dark' },
              { label: '跟随系统', value: 'system' },
            ]}
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

          {user && (
            <div className="mt-4">
              <div className="flex items-center justify-between mb-2">
                <div className="font-medium dark:text-white">API Token</div>
                <Button
                  type="primary"
                  size="small"
                  icon={<PlusOutlined />}
                  onClick={() => setCreateTokenModalVisible(true)}
                >
                  创建
                </Button>
              </div>
              <Spin spinning={tokensLoading}>
                <Table
                  dataSource={tokens}
                  rowKey="id"
                  pagination={false}
                  size="small"
                  columns={[
                    {
                      title: '描述',
                      dataIndex: 'description',
                      key: 'description',
                      render: (text) => text || '无描述',
                    },
                    {
                      title: 'Token',
                      dataIndex: 'token',
                      key: 'token',
                      ellipsis: true,
                      render: (text) => <Typography.Text type="secondary" className="font-mono text-xs">{text}</Typography.Text>,
                    },
                    {
                      title: '创建时间',
                      dataIndex: 'created_at',
                      key: 'created_at',
                      width: 180,
                      render: (text) => new Date(text).toLocaleString('zh-CN'),
                    },
                    {
                      title: '过期时间',
                      dataIndex: 'expires_at',
                      key: 'expires_at',
                      width: 180,
                      render: (text) => text ? new Date(text).toLocaleString('zh-CN') : '不过期',
                    },
                    {
                      title: '操作',
                      key: 'action',
                      width: 80,
                      render: (_, record) => (
                        <Button
                          type="text"
                          danger
                          size="small"
                          icon={<DeleteOutlined />}
                          onClick={() => handleDeleteToken(record.id)}
                        />
                      ),
                    },
                  ]}
                />
              </Spin>
            </div>
          )}
        </div>

        <Modal
          title="创建 API Token"
          open={createTokenModalVisible}
          onCancel={() => {
            setCreateTokenModalVisible(false);
            form.resetFields();
          }}
          footer={null}
        >
          <Form
            form={form}
            layout="vertical"
            onFinish={handleCreateToken}
          >
            <Form.Item
              name="description"
              label="描述"
              rules={[{ required: true, message: '请输入描述' }]}
            >
              <Input placeholder="Token 用途描述" />
            </Form.Item>
            <Form.Item
              name="expires_in_hours"
              label="过期时间"
              initialValue={0}
            >
              <Select>
                <Select.Option value={0}>不过期</Select.Option>
                <Select.Option value={24}>24 小时</Select.Option>
              </Select>
            </Form.Item>
            <Form.Item className="mb-0">
              <Space>
                <Button type="primary" htmlType="submit" loading={createTokenLoading}>
                  创建
                </Button>
                <Button onClick={() => {
                  setCreateTokenModalVisible(false);
                  form.resetFields();
                }}>
                  取消
                </Button>
              </Space>
            </Form.Item>
          </Form>
        </Modal>

        <Modal
          title="添加广场数据源"
          open={createFeedSourceModalVisible}
          onCancel={() => {
            setCreateFeedSourceModalVisible(false);
            form.resetFields();
          }}
          footer={null}
        >
          <Form
            form={form}
            layout="vertical"
            onFinish={handleCreateFeedSource}
          >
            <Form.Item
              name="name"
              label="名称"
              rules={[{ required: true, message: '请输入名称' }]}
            >
              <Input placeholder="数据源名称，如：隔壁的bibibibi" />
            </Form.Item>
            <Form.Item
              name="url"
              label="地址"
              rules={[{ required: true, message: '请输入实例地址' }]}
              extra="填写其他 bibibibi 实例的基础 URL，如：https://bibi.example.com"
            >
              <Input placeholder="https://bibi.example.com" />
            </Form.Item>
            <Form.Item className="mb-0">
              <Space>
                <Button type="primary" htmlType="submit" loading={createFeedSourceLoading}>
                  添加
                </Button>
                <Button onClick={() => {
                  setCreateFeedSourceModalVisible(false);
                  form.resetFields();
                }}>
                  取消
                </Button>
              </Space>
            </Form.Item>
          </Form>
        </Modal>

        {user?.is_admin && (
          <>
            <Divider />

            <div className="flex items-center justify-between">
              <div>
                <div className="font-medium dark:text-white">开放注册</div>
                <div className="text-sm text-gray-500 dark:text-gray-400">允许新用户注册账号</div>
              </div>
              <Switch
                checked={registrationEnabled}
                loading={settingsLoading}
                onChange={handleUpdateRegistration}
              />
            </div>

            <Divider />

            <div>
              <div className="flex items-center justify-between mb-2">
                <div>
                  <div className="font-medium dark:text-white">Gravatar 头像源</div>
                  <div className="text-sm text-gray-500 dark:text-gray-400">用于生成用户和评论的头像</div>
                </div>
              </div>
              <Space.Compact style={{ width: '100%' }}>
                <Input
                  value={gravatarSource}
                  onChange={(e) => setGravatarSource(e.target.value)}
                  placeholder="https://www.gravatar.com/avatar/"
                  style={{ flex: 1 }}
                />
                <Button type="primary" loading={settingsLoading} onClick={() => handleUpdateGravatarSource(gravatarSource)}>
                  保存
                </Button>
              </Space.Compact>
              <div className="text-xs text-gray-400 dark:text-gray-500 mt-1">
                常用镜像：https://weavatar.com/avatar/ 或 https://cdn.v2ex.com/gravatar/
              </div>
            </div>

            <Divider />

            <div>
              <div className="flex items-center justify-between mb-2">
                <div>
                  <div className="font-medium dark:text-white">广场数据源</div>
                  <div className="text-sm text-gray-500 dark:text-gray-400">从其他 bibibibi 实例聚合公开笔记</div>
                </div>
                <Space>
                  <Button size="small" onClick={handleSyncFeedSources} loading={syncLoading}>
                    同步
                  </Button>
                  <Button
                    type="primary"
                    size="small"
                    icon={<PlusOutlined />}
                    onClick={() => setCreateFeedSourceModalVisible(true)}
                  >
                    添加
                  </Button>
                </Space>
              </div>
              <Spin spinning={feedSourcesLoading}>
                <Table
                  dataSource={feedSources}
                  rowKey="id"
                  pagination={false}
                  size="small"
                  columns={[
                    {
                      title: '名称',
                      dataIndex: 'name',
                      key: 'name',
                    },
                    {
                      title: '地址',
                      dataIndex: 'url',
                      key: 'url',
                      ellipsis: true,
                    },
                    {
                      title: '最后同步',
                      dataIndex: 'last_fetch_at',
                      key: 'last_fetch_at',
                      width: 180,
                      render: (text) => text ? new Date(text).toLocaleString('zh-CN') : '从未同步',
                    },
                    {
                      title: '操作',
                      key: 'action',
                      width: 80,
                      render: (_, record) => (
                        <Button
                          type="text"
                          danger
                          size="small"
                          icon={<DeleteOutlined />}
                          onClick={() => handleDeleteFeedSource(record.id)}
                        />
                      ),
                    },
                  ]}
                  locale={{ emptyText: '暂无数据源' }}
                />
              </Spin>
            </div>
          </>
        )}

        {user && (
          <>
            <Divider />

            <div>
              <div className="font-medium mb-2 dark:text-white">清空数据</div>
              <div className="text-sm text-gray-500 dark:text-gray-400 mb-3">清空所有本地缓存数据，包括登录信息和评论记录</div>
              <Button danger icon={<ClearOutlined />} onClick={handleClearData}>
                清空所有数据
              </Button>
            </div>

            <Divider />
          </>
        )}

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
    ...(user ? [{
      key: 'members',
      icon: <UserOutlined />,
      label: '成员列表',
      onClick: () => setActiveTab('members'),
    }] : []),
    {
      key: 'settings',
      icon: <SettingOutlined />,
      label: '系统设置',
      onClick: () => setActiveTab('settings'),
    },
    ...(user ? [{
      key: 'profile',
      icon: <ProfileIcon />,
      label: '个人中心',
      onClick: () => setActiveTab('profile'),
    }] : []),
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
        {user && (
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
        )}

        <Button
          type="primary"
          icon={user ? <PlusOutlined /> : <UserOutlined />}
          size="large"
          style={{ margin: user ? '0 16px 24px' : '24px 16px', width: 'calc(100% - 32px)' }}
          onClick={() => {
            if (user) {
              setShowEditor(true);
              setActiveTab('home');
            } else {
              navigate('/login');
            }
          }}
        >
          {user ? '发布' : '登录'}
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

        {user && (
          <div style={{ position: 'absolute', bottom: 24, left: 0, right: 0, padding: '16px 24px', borderTop: '1px solid rgba(255,255,255,0.14)' }}>
            <Button type="text" block onClick={handleLogout} style={{ textAlign: 'left', color: 'rgba(255,255,255,0.65)' }}>
              退出登录
            </Button>
          </div>
        )}
      </Sider>

      <Layout style={{ marginLeft: 240 }}>
        <Content style={{ padding: '24px 24px 0', minHeight: '100vh' }}>
          {activeTab === 'home' && (
            <div className="fade-in">
              <Tabs
                activeKey={homeSubTab}
                onChange={(key) => setHomeSubTab(key as any)}
                size="small"
                className="mb-4"
              >
                <TabPane tab="我的笔记" key="mine" />
                {user && <TabPane tab="笔记广场" key="square" />}
              </Tabs>
              {renderNotesList()}
            </div>
          )}
          {activeTab === 'profile' && <div className="fade-in">{renderProfile()}</div>}
          {activeTab === 'members' && <div className="fade-in">{renderMembers()}</div>}
          {activeTab === 'settings' && <div className="fade-in">{renderSettings()}</div>}
        </Content>
      </Layout>

      {previewImage && (
        <div className="image-modal-overlay modal-fade" onClick={() => setPreviewImage(null)}>
          <img src={previewImage} alt="Preview" />
        </div>
      )}
    </Layout>
  );
};

export default Home;