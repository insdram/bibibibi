import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';

const THEME_KEY = 'bibibibi_theme';

type ThemeMode = 'light' | 'dark' | 'system';

interface ThemeContextType {
  themeMode: ThemeMode;
  darkMode: boolean;
  setThemeMode: (mode: ThemeMode) => void;
}

const ThemeContext = createContext<ThemeContextType | undefined>(undefined);

export const useTheme = () => {
  const context = useContext(ThemeContext);
  if (!context) {
    throw new Error('useTheme must be used within a ThemeProvider');
  }
  return context;
};

interface ThemeProviderProps {
  children: ReactNode;
}

export const ThemeProvider: React.FC<ThemeProviderProps> = ({ children }) => {
  const [themeMode, setThemeMode] = useState<ThemeMode>(() => {
    const saved = localStorage.getItem(THEME_KEY);
    return (saved as ThemeMode) || 'system';
  });

  const [darkMode, setDarkMode] = useState(() => {
    const saved = localStorage.getItem(THEME_KEY);
    if (saved === 'light') return false;
    if (saved === 'dark') return true;
    return window.matchMedia('(prefers-color-scheme: dark)').matches;
  });

  useEffect(() => {
    localStorage.setItem(THEME_KEY, themeMode);
  }, [themeMode]);

  useEffect(() => {
    const updateDarkMode = () => {
      if (themeMode === 'system') {
        setDarkMode(window.matchMedia('(prefers-color-scheme: dark)').matches);
      } else {
        setDarkMode(themeMode === 'dark');
      }
    };

    updateDarkMode();

    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
    mediaQuery.addEventListener('change', updateDarkMode);
    return () => mediaQuery.removeEventListener('change', updateDarkMode);
  }, [themeMode]);

  useEffect(() => {
    if (darkMode) {
      document.documentElement.classList.add('dark');
    } else {
      document.documentElement.classList.remove('dark');
    }
  }, [darkMode]);

  return (
    <ThemeContext.Provider value={{ themeMode, darkMode, setThemeMode }}>
      {children}
    </ThemeContext.Provider>
  );
};