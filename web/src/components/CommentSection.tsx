import React, { useState, useEffect } from 'react';
import { Input, Button, Avatar, List, Space, message, Popconfirm } from 'antd';
import { SendOutlined, CommentOutlined, DeleteOutlined } from '@ant-design/icons';
import { commentApi } from '../api';
import { useAuth } from '../stores/AuthContext';

const { TextArea } = Input;

const COMMENT_INFO_KEY = 'bibibibi_comment_info';

interface Comment {
  id: number;
  parent_id: number;
  name: string;
  email: string;
  website: string;
  content: string;
  avatar: string;
  created_at: string;
}

interface CommentInfo {
  name: string;
  email: string;
  website: string;
}

interface CommentSectionProps {
  bibiId: number;
  comments: Comment[];
  onUpdate: () => void;
  isOwner?: boolean;
}

const CommentSection: React.FC<CommentSectionProps> = ({ bibiId, comments, onUpdate, isOwner }) => {
  const { user } = useAuth();
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [website, setWebsite] = useState('');
  const [content, setContent] = useState('');
  const [loading, setLoading] = useState(false);
  const [replyTo, setReplyTo] = useState<Comment | null>(null);
  const [replyName, setReplyName] = useState('');
  const [replyEmail, setReplyEmail] = useState('');
  const [replyWebsite, setReplyWebsite] = useState('');
  const [replyContent, setReplyContent] = useState('');

  useEffect(() => {
    if (user) {
      setName(user.nickname || user.username);
      setEmail(user.email || '');
      setWebsite(user.website || '');
    } else {
      const savedInfo = localStorage.getItem(COMMENT_INFO_KEY);
      if (savedInfo) {
        try {
          const info: CommentInfo = JSON.parse(savedInfo);
          setName(info.name || '');
          setEmail(info.email || '');
          setWebsite(info.website || '');
        } catch (e) {
          console.error('Failed to parse saved comment info', e);
        }
      }
    }
  }, [user]);

  const saveCommentInfo = (info: CommentInfo) => {
    localStorage.setItem(COMMENT_INFO_KEY, JSON.stringify(info));
  };

  const handleSubmit = async () => {
    if (!name.trim() || !email.trim() || !content.trim()) {
      message.warning('请填写必填项');
      return;
    }

    setLoading(true);
    try {
      await commentApi.createComment(bibiId, {
        name: name.trim(),
        email: email.trim(),
        website: website.trim() || undefined,
        content: content.trim(),
      });
      saveCommentInfo({ name: name.trim(), email: email.trim(), website: website.trim() });
      setContent('');
      message.success('评论成功');
      onUpdate();
    } catch (error) {
      console.error('发表评论失败:', error);
      message.error('评论失败');
    } finally {
      setLoading(false);
    }
  };

  const handleReply = async () => {
    if (!replyName.trim() || !replyEmail.trim() || !replyContent.trim()) {
      message.warning('请填写必填项');
      return;
    }

    setLoading(true);
    try {
      await commentApi.createComment(bibiId, {
        name: replyName.trim(),
        email: replyEmail.trim(),
        website: replyWebsite.trim() || undefined,
        content: replyContent.trim(),
        parent_id: replyTo?.id,
      });
      saveCommentInfo({ name: replyName.trim(), email: replyEmail.trim(), website: replyWebsite.trim() });
      setReplyName('');
      setReplyEmail('');
      setReplyWebsite('');
      setReplyContent('');
      setReplyTo(null);
      message.success('回复成功');
      onUpdate();
    } catch (error) {
      console.error('回复失败:', error);
      message.error('回复失败');
    } finally {
      setLoading(false);
    }
  };

  const handleCancelReply = () => {
    setReplyTo(null);
    setReplyName('');
    setReplyEmail('');
    setReplyWebsite('');
    setReplyContent('');
  };

  const handleDeleteComment = async (commentId: number) => {
    try {
      await commentApi.deleteComment(commentId);
      message.success('删除成功');
      onUpdate();
    } catch (error) {
      console.error('删除评论失败:', error);
      message.error('删除失败');
    }
  };

  const handleSetReplyInfo = () => {
    if (name.trim()) setReplyName(name.trim());
    if (email.trim()) setReplyEmail(email.trim());
    if (website.trim()) setReplyWebsite(website.trim());
  };

  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleString('zh-CN', { month: 'numeric', day: 'numeric', hour: '2-digit', minute: '2-digit' });
  };

  const renderCommentItem = (comment: Comment) => (
    <List.Item
      key={comment.id}
      className="border-b border-gray-100 dark:border-gray-700 py-3"
      actions={[
        <Button
          key="reply"
          type="text"
          size="small"
          icon={<CommentOutlined />}
          onClick={() => {
            setReplyTo(comment);
            handleSetReplyInfo();
          }}
          className="text-gray-400 dark:text-gray-500"
        >
          回复
        </Button>,
        isOwner && (
          <Popconfirm
            key="delete"
            title="删除评论"
            description="确定要删除这条评论吗？"
            onConfirm={() => handleDeleteComment(comment.id)}
            okText="删除"
            cancelText="取消"
            okButtonProps={{ danger: true }}
          >
            <Button
              type="text"
              size="small"
              icon={<DeleteOutlined />}
              className="text-gray-400 dark:text-gray-500"
            >
              删除
            </Button>
          </Popconfirm>
        ),
      ]}
    >
      <Space align="start" size={12}>
        <Avatar
          src={comment.avatar}
        >
          {comment.name?.charAt(0).toUpperCase()}
        </Avatar>
        <div className="flex-1">
          <div className="flex items-center gap-2 mb-1">
            <span className="font-medium text-gray-800 dark:text-white">{comment.name}</span>
            {comment.website && (
              <a
                href={comment.website}
                target="_blank"
                rel="noopener noreferrer"
                className="text-xs text-blue-500 dark:text-blue-400 hover:text-blue-600"
              >
                {comment.website}
              </a>
            )}
            <span className="text-xs text-gray-400 dark:text-gray-500">
              {formatDate(comment.created_at)}
            </span>
          </div>
          {comment.parent_id > 0 && (
            <div className="text-xs text-blue-500 dark:text-blue-400 mb-1">
              回复 @{comments.find(c => c.id === comment.parent_id)?.name || '已删除'}
            </div>
          )}
          <p className="text-gray-600 dark:text-gray-300 text-sm m-0 whitespace-pre-wrap">{comment.content}</p>
        </div>
      </Space>
    </List.Item>
  );

  return (
    <div>
      <div className="mb-4">
        <h4 className="text-sm font-medium text-gray-700 dark:text-gray-200 mb-3">评论 ({comments.length})</h4>
        
        {comments.length > 0 ? (
          <List
            dataSource={comments}
            renderItem={renderCommentItem}
            locale={{ emptyText: '暂无评论' }}
          />
        ) : (
          <div className="text-center text-gray-400 dark:text-gray-500 py-4 text-sm">暂无评论</div>
        )}
      </div>

      {replyTo && (
        <div className="border-t border-[#f0f0f0] dark:border-[#303030] pt-4 mb-4">
          <div className="flex items-center gap-2 mb-2">
            <span className="text-sm text-blue-500 dark:text-blue-400">回复 @{replyTo.name}</span>
            <Button type="text" size="small" onClick={handleCancelReply}>
              取消
            </Button>
          </div>
          <Space direction="vertical" size={12} className="w-full">
            <div className="flex gap-3">
              <Input
                placeholder="昵称 *"
                value={replyName}
                onChange={(e) => setReplyName(e.target.value)}
                maxLength={50}
                style={{ flex: 1 }}
              />
              <Input
                placeholder="邮箱 *"
                type="email"
                value={replyEmail}
                onChange={(e) => setReplyEmail(e.target.value)}
                maxLength={100}
                style={{ flex: 1 }}
              />
              <Input
                placeholder="网址"
                value={replyWebsite}
                onChange={(e) => setReplyWebsite(e.target.value)}
                maxLength={200}
                style={{ flex: 1 }}
              />
            </div>
            <TextArea
              placeholder={`回复 @${replyTo.name}...`}
              value={replyContent}
              onChange={(e) => setReplyContent(e.target.value)}
              autoSize={{ minRows: 2, maxRows: 4 }}
              maxLength={1000}
            />
            <div className="flex justify-end">
              <Button 
                type="primary" 
                icon={<SendOutlined />} 
                loading={loading}
                onClick={handleReply}
                disabled={!replyName.trim() || !replyEmail.trim() || !replyContent.trim()}
              >
                发送回复
              </Button>
            </div>
          </Space>
        </div>
      )}

      <div className="border-t border-[#f0f0f0] dark:border-[#303030] pt-4">
        <h4 className="text-sm font-medium text-gray-700 dark:text-gray-200 mb-3">发表评论</h4>
        <Space direction="vertical" size={12} className="w-full">
          <div className="flex gap-3">
            <Input
              placeholder="昵称 *"
              value={name}
              onChange={(e) => setName(e.target.value)}
              maxLength={50}
              style={{ flex: 1 }}
            />
            <Input
              placeholder="邮箱 *"
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              maxLength={100}
              style={{ flex: 1 }}
            />
            <Input
              placeholder="网址"
              value={website}
              onChange={(e) => setWebsite(e.target.value)}
              maxLength={200}
              style={{ flex: 1 }}
            />
          </div>
          <TextArea
            placeholder="写下你的评论... *"
            value={content}
            onChange={(e) => setContent(e.target.value)}
            autoSize={{ minRows: 2, maxRows: 4 }}
            maxLength={1000}
          />
          <div className="flex justify-end">
            <Button 
              type="primary" 
              icon={<SendOutlined />} 
              loading={loading}
              onClick={handleSubmit}
              disabled={!name.trim() || !email.trim() || !content.trim()}
            >
              发表
            </Button>
          </div>
        </Space>
      </div>
    </div>
  );
};

export default CommentSection;