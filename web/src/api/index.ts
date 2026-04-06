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
  getBibis: (params?: { page?: number; page_size?: number; visibility?: string }) =>
    client.get('/bibis', { params }),
  
  getBibi: (id: number) =>
    client.get(`/bibis/${id}`),
  
  createBibi: (data: { content: string; visibility?: string; tag_ids?: number[] }) =>
    client.post('/bibis', data),
  
  updateBibi: (id: number, data: { content: string; visibility?: string; tag_ids?: number[] }) =>
    client.put(`/bibis/${id}`, data),
  
  deleteBibi: (id: number) =>
    client.delete(`/bibis/${id}`),
  
  togglePin: (id: number) =>
    client.post(`/bibis/${id}/pin`),
  
  toggleLike: (id: number) =>
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
  getComments: (bibiId: number, params?: { page?: number; page_size?: number }) =>
    client.get(`/bibis/${bibiId}/comments`, { params }),
  
  createComment: (bibiId: number, data: { name: string; email: string; website?: string; content: string; parent_id?: number }) =>
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
  
  updateCurrentUser: (data: { username?: string; nickname?: string; email?: string; password?: string }) =>
    client.put('/user/me', data),
};
