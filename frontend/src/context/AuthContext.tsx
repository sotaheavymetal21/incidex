'use client';

import React, { createContext, useContext, useEffect, useState, ReactNode } from 'react';

import { useRouter } from 'next/navigation';
import { authApi } from '@/lib/api';
import { logger } from '@/lib/logger';

type User = {
  id: number;
  name: string;
  email: string;
  role: string;
};

type AuthContextType = {
  user: User | null;
  token: string | null;
  login: (token: string, user: User) => void;
  logout: () => void;
  loading: boolean;
};

const AuthContext = createContext<AuthContextType | undefined>(undefined);

/**
 * 認証プロバイダーコンポーネント
 * 認証状態を管理し、子コンポーネントに認証コンテキストを提供します
 */
export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [token, setToken] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const router = useRouter();

  useEffect(() => {
    const storedToken = localStorage.getItem('token');
    const storedUser = localStorage.getItem('user');

    if (storedToken && storedUser) {
      setToken(storedToken);
      setUser(JSON.parse(storedUser));
    }
    setLoading(false);
  }, []);

  /**
   * ログイン処理
   * token とユーザー情報を保存します
   */
  const login = (newToken: string, newUser: User) => {
    setToken(newToken);
    setUser(newUser);
    localStorage.setItem('token', newToken);
    localStorage.setItem('user', JSON.stringify(newUser));
  };

  /**
   * ログアウト処理
   * リフレッシュ token を無効化するためログアウト API を呼び出します
   */
  const logout = async () => {
    try {
      // リフレッシュ token を無効化するためログアウト API を呼び出します
      await authApi.logout();
    } catch (error) {
      // error は無視してローカルでログアウトを続行
      logger.warn('Logout API call failed, proceeding with local logout', { error });
    }

    setToken(null);
    setUser(null);
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    router.push('/login');
  };

  return (
    <AuthContext.Provider value={{ user, token, login, logout, loading }}>
      {children}
    </AuthContext.Provider>
  );
}

/**
 * 認証コンテキストを使用するためのカスタムフック
 * AuthProvider の外で使用するとエラーをスローします
 */
export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}
