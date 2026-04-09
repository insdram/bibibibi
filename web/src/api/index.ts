import client from './client';

// 认证相关
export const authApi = {
  login: (username: string, password: string) =>
    client.post('/auth/login', { username, password }),
  
  register: (username: string, password: string, nickname?: string, email?: string) =>
    client.post('/auth/register', { username, password, nickname, email }),
};

// Bibi 相关
export const bibiApi = {
  getBibis: (params?: { page?: number; page_size?: number; visibility?: string; creator_id?: number }) =>
    client.get('/bibis', { params }),
  
  getBibi: (id: string) =>
    client.get(`/bibis/${id}`),
  
  createBibi: (data: { content: string; visibility?: string; tag_ids?: number[] }) =>
    client.post('/bibis', data),
  
  updateBibi: (id: string, data: { content: string; visibility?: string; tag_ids?: number[] }) =>
    client.put(`/bibis/${id}`, data),
  
  deleteBibi: (id: string) =>
    client.delete(`/bibis/${id}`),
  
  togglePin: (id: string) =>
    client.post(`/bibis/${id}/pin`),
  
  toggleLike: (id: string) =>
    client.post(`/bibis/${id}/like`),
  
  searchBibis: (keyword: string, params?: { page?: number; page_size?: number }) =>
    client.get('/bibis/search', { params: { keyword, ...params } }),
};

// 标签相关
export const tagApi = {
  getTags: (params?: { creator_id?: number }) =>
    client.get('/tags', { params }),
  
  createTag: (name: string) =>
    client.post('/tags', { name }),
  
  updateTag: (id: number, name: string) =>
    client.put(`/tags/${id}`, { name }),
  
  deleteTag: (id: number) =>
    client.delete(`/tags/${id}`),
};

// 评论相关
export const commentApi = {
  getComments: (bibiId: string, params?: { page?: number; page_size?: number }) =>
    client.get(`/bibis/${bibiId}/comments`, { params }),
  
  createComment: (bibiId: string, data: { name: string; email: string; website?: string; content: string; parent_id?: number }) =>
    client.post(`/bibis/${bibiId}/comments`, data),
  
  updateComment: (id: number, data: { name: string; email: string; website?: string; content: string }) =>
    client.put(`/comments/${id}`, data),
  
  deleteComment: (id: number) =>
    client.delete(`/comments/${id}`),
};

// 用户相关
export const userApi = {
  getCurrentUser: () =>
    client.get('/user/me'),
  
  getUsers: () =>
    client.get('/user/list'),
  
  updateCurrentUser: (data: { username?: string; nickname?: string; email?: string; website?: string; password?: string }) =>
    client.put('/user/me', data),
};

// Token 管理相关
export const tokenApi = {
  getTokens: () =>
    client.get('/tokens'),
  
  createToken: (data: { description?: string; expires_in_hours?: number }) =>
    client.post('/tokens', data),
  
  deleteToken: (id: number) =>
    client.delete(`/tokens/${id}`),
};

// 系统设置相关
export const systemApi = {
  getSettings: () =>
    client.get('/settings'),
  
  getPublicSettings: () =>
    client.get('/public/settings'),
  
  updateSettings: (data: { registration_enabled?: boolean; gravatar_source?: string }) =>
    client.put('/settings', data),
};

// 广场数据源相关
export const feedApi = {
  getFeedSources: () =>
    client.get('/feeds'),
  
  createFeedSource: (data: { name: string; url: string }) =>
    client.post('/feeds', data),
  
  deleteFeedSource: (id: number) =>
    client.delete(`/feeds/${id}`),
  
  syncFeedSources: () =>
    client.post('/feeds/sync'),
};
