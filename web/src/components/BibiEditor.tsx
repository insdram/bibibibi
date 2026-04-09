import React, { useState, useEffect } from 'react';
import { Input, Button, Space, Tag, Select, message } from 'antd';
import { PlusOutlined, CloseOutlined } from '@ant-design/icons';
import { tagApi } from '../api';
import { useAuth } from '../stores/AuthContext';

const { TextArea } = Input;

interface Tag {
  id: number;
  name: string;
}

interface BibiEditorProps {
  onSubmit: (content: string, visibility: string, tagIds: number[]) => void;
  onCancel: () => void;
}

const BibiEditor: React.FC<BibiEditorProps> = ({ onSubmit, onCancel }) => {
  const { user } = useAuth();
  const [content, setContent] = useState('');
  const [visibility, setVisibility] = useState<string>('PUBLIC');
  const [selectedTagIds, setSelectedTagIds] = useState<number[]>([]);
  const [tags, setTags] = useState<Tag[]>([]);
  const [newTagName, setNewTagName] = useState('');
  const [showTagInput, setShowTagInput] = useState(false);
  const [creating, setCreating] = useState(false);

  useEffect(() => {
    fetchTags();
  }, []);

  const fetchTags = async () => {
    try {
      const response = await tagApi.getTags({ creator_id: user?.id });
      setTags(response.data || []);
    } catch (error) {
      console.error('获取标签失败:', error);
    }
  };

  const handleSubmit = () => {
    if (!content.trim()) {
      message.warning('请输入内容');
      return;
    }
    onSubmit(content, visibility, selectedTagIds);
  };

  const handleCreateTag = async () => {
    if (!newTagName.trim()) return;
    setCreating(true);
    try {
      const response = await tagApi.createTag(newTagName.trim());
      setTags([...tags, response.data]);
      setSelectedTagIds([...selectedTagIds, response.data.id]);
      setNewTagName('');
      setShowTagInput(false);
      message.success('标签创建成功');
    } catch (error) {
      console.error('创建标签失败:', error);
      message.error('创建标签失败');
    } finally {
      setCreating(false);
    }
  };

  const toggleTag = (tagId: number) => {
    setSelectedTagIds((prev) =>
      prev.includes(tagId)
        ? prev.filter((id) => id !== tagId)
        : [...prev, tagId]
    );
  };

  const handleCloseTag = async (e: React.MouseEvent, tagId: number) => {
    e.stopPropagation();
    try {
      await tagApi.deleteTag(tagId);
      setTags((prev) => prev.filter((tag) => tag.id !== tagId));
      setSelectedTagIds((prev) => prev.filter((id) => id !== tagId));
      message.success('标签已删除');
    } catch (error) {
      console.error('删除标签失败:', error);
      message.error('删除标签失败');
    }
  };

  const tagOptions = tags.map((tag) => ({
    label: tag.name,
    value: tag.id,
  }));

  return (
    <div>
      <TextArea
        value={content}
        onChange={(e) => setContent(e.target.value)}
        placeholder="分享你的想法..."
        autoSize={{ minRows: 4, maxRows: 8 }}
        maxLength={2000}
        showCount
        style={{ marginBottom: 16 }}
      />

      <div className="mb-4">
        <Space size={8} wrap>
          <span className="text-gray-500 dark:text-gray-400 text-sm">可见性:</span>
          <Button.Group>
            <Button
              type={visibility === 'PUBLIC' ? 'primary' : 'default'}
              onClick={() => setVisibility('PUBLIC')}
              size="small"
            >
              公开
            </Button>
            <Button
              type={visibility === 'PRIVATE' ? 'primary' : 'default'}
              onClick={() => setVisibility('PRIVATE')}
              size="small"
            >
              私密
            </Button>
          </Button.Group>
        </Space>
      </div>

      <div className="mb-4">
        <Space size={8} wrap style={{ marginBottom: 8 }}>
          <span className="text-gray-500 dark:text-gray-400 text-sm">标签:</span>
          {tags.map((tag) => (
            <Tag
              key={tag.id}
              closable={selectedTagIds.includes(tag.id)}
              onClose={(e) => handleCloseTag(e, tag.id)}
              color={selectedTagIds.includes(tag.id) ? 'blue' : 'default'}
              style={{ cursor: 'pointer' }}
              onClick={() => toggleTag(tag.id)}
            >
              {tag.name}
            </Tag>
          ))}
          {showTagInput ? (
            <Space size={4}>
              <Input
                size="small"
                value={newTagName}
                onChange={(e) => setNewTagName(e.target.value)}
                placeholder="标签名"
                style={{ width: 100 }}
                autoFocus
                onPressEnter={handleCreateTag}
              />
              <Button size="small" type="primary" loading={creating} onClick={handleCreateTag}>
                创建
              </Button>
              <Button size="small" icon={<CloseOutlined />} onClick={() => setShowTagInput(false)} />
            </Space>
          ) : (
            <Tag
              className="cursor-pointer dark:!bg-gray-700 dark:!border-gray-600 dark:!text-gray-300"
              style={{ background: '#f0f0f0', borderStyle: 'dashed' }}
              onClick={() => setShowTagInput(true)}
            >
              <PlusOutlined /> 新标签
            </Tag>
          )}
        </Space>
      </div>

      <div className="flex justify-end">
        <Space>
          <Button onClick={onCancel}>取消</Button>
          <Button type="primary" onClick={handleSubmit} disabled={!content.trim()}>
            发布
          </Button>
        </Space>
      </div>
    </div>
  );
};

export default BibiEditor;