import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { authApi, userApi } from '../api';

interface User {
  id: number;
  username: string;
  nickname: string;
  email: string;
  website?: string;
  avatar: string;
  created_at: string;
}

interface AuthContextType {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  login: (username: string, password: string) => Promise<void>;
  register: (username: string, password: string, nickname?: string, email?: string) => Promise<void>;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};

interface AuthProviderProps {
  children: ReactNode;
}

export const AuthProvider: React.FC<AuthProviderProps> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null);
  const [token, setToken] = useState<string | null>(localStorage.getItem('token'));

  useEffect(() => {
    const fetchUser = async () => {
      if (!token) return;
      try {
        const response = await userApi.getCurrentUser();
        setUser(response.data);
      } catch (error) {
        console.error('获取用户信息失败:', error);
      }
    };

    fetchUser();
  }, [token]);

  useEffect(() => {
    if (token) {
      localStorage.setItem('token', token);
    }
  }, [token]);

  const login = async (username: string, password: string) => {
    const response = await authApi.login(username, password);
    const { token: newToken, user: userData } = response.data;
    setToken(newToken);
    if (userData) {
      setUser(userData);
    }
  };

  const register = async (username: string, password: string, nickname?: string, email?: string) => {
    await authApi.register(username, password, nickname, email);
    // 注册成功后自动登录
    await login(username, password);
  };

  const logout = () => {
    setUser(null);
    setToken(null);
    localStorage.removeItem('token');
  };

  const value: AuthContextType = {
    user,
    token,
    isAuthenticated: !!token,
    login,
    register,
    logout,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};
