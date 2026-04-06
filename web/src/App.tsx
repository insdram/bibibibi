import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider, useAuth } from './stores/AuthContext';
import { ThemeProvider, useTheme } from './stores/ThemeContext';
import { ConfigProvider, theme } from 'antd';
import Home from './pages/Home';
import Login from './pages/Login';
import Register from './pages/Register';

const ProtectedRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { isAuthenticated } = useAuth();
  
  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }
  
  return <>{children}</>;
};

const ThemedApp: React.FC = () => {
  const { darkMode } = useTheme();
  
  return (
    <ConfigProvider
      theme={{
        algorithm: darkMode ? theme.darkAlgorithm : theme.defaultAlgorithm,
        token: {
          colorPrimary: '#1890ff',
          colorBgContainer: darkMode ? '#1f1f1f' : '#ffffff',
          colorBgElevated: darkMode ? '#2d2d2d' : '#ffffff',
          colorBgLayout: darkMode ? '#141414' : '#f0f2f5',
          colorText: darkMode ? '#ffffff' : '#000000e6',
          colorTextSecondary: darkMode ? '#a0a0a0' : '#00000073',
          colorBorder: darkMode ? '#303030' : '#d9d9d9',
          borderRadius: 6,
        },
        components: {
          Layout: {
            siderBg: darkMode ? '#141414' : '#001529',
            headerBg: darkMode ? '#1f1f1f' : '#ffffff',
            bodyBg: darkMode ? '#141414' : '#f0f2f5',
          },
          Menu: {
            darkItemBg: '#141414',
            darkSubMenuItemBg: '#141414',
            darkItemSelectedBg: '#1890ff',
            itemBg: darkMode ? '#141414' : '#001529',
            itemSelectedBg: darkMode ? '#1890ff33' : '#1890ff1a',
            itemHoverBg: darkMode ? '#ffffff0a' : '#ffffff14',
            darkItemHoverBg: '#ffffff0a',
            darkItemColor: '#ffffffa6',
            darkItemSelectedColor: '#ffffff',
            itemColor: '#ffffffa6',
            itemSelectedColor: '#ffffff',
            itemHoverColor: '#ffffff',
            subMenuItemBg: darkMode ? '#141414' : '#001529',
            darkSubMenuItemSelectedBg: '#1890ff1a',
            subMenuItemSelectedBg: darkMode ? '#1890ff33' : '#1890ff1a',
          },
          Card: {
            colorBgContainer: darkMode ? '#1f1f1f' : '#ffffff',
          },
          Input: {
            colorBgContainer: darkMode ? '#262626' : '#ffffff',
          },
          Tabs: {
            colorBgContainer: 'transparent',
            inkBarColor: '#1890ff',
            itemActiveColor: '#1890ff',
            itemSelectedColor: '#1890ff',
            itemHoverColor: '#1890ff',
          },
        },
      }}
    >
      <Router>
        <AuthProvider>
          <AppRoutes />
        </AuthProvider>
      </Router>
    </ConfigProvider>
  );
};

const AppRoutes: React.FC = () => {
  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route path="/register" element={<Register />} />
      <Route
        path="/"
        element={
          <ProtectedRoute>
            <Home />
          </ProtectedRoute>
        }
      />
    </Routes>
  );
};

const App: React.FC = () => {
  return (
    <ThemeProvider>
      <ThemedApp />
    </ThemeProvider>
  );
};

export default App;