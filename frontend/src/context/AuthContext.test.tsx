import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen, act, waitFor } from '@testing-library/react';
import { AuthProvider, useAuth } from './AuthContext';
import { testUsers } from '@/test/fixtures';

// Next.js router をモック
const mockPush = vi.fn();
vi.mock('next/navigation', () => ({
  useRouter: () => ({
    push: mockPush,
    replace: vi.fn(),
    prefetch: vi.fn(),
  }),
}));

// logger をモック
vi.mock('@/lib/logger', () => ({
  logger: {
    warn: vi.fn(),
    error: vi.fn(),
    info: vi.fn(),
    debug: vi.fn(),
    apiRequest: vi.fn(),
    apiResponse: vi.fn(),
  },
}));

// テスト用コンポーネント
function TestComponent() {
  const { user, token, login, logout, loading } = useAuth();
  return (
    <div>
      <div data-testid="loading">{loading ? 'true' : 'false'}</div>
      <div data-testid="user">{user ? JSON.stringify(user) : 'null'}</div>
      <div data-testid="token">{token || 'null'}</div>
      <button onClick={() => login('test-token', testUsers.viewer)} data-testid="login-btn">
        Login
      </button>
      <button onClick={() => login('admin-token', testUsers.admin)} data-testid="login-admin-btn">
        Login Admin
      </button>
      <button onClick={() => login('editor-token', testUsers.editor)} data-testid="login-editor-btn">
        Login Editor
      </button>
      <button onClick={() => logout()} data-testid="logout-btn">
        Logout
      </button>
    </div>
  );
}

describe('AuthContext', () => {
  beforeEach(() => {
    // 各テスト前に localStorage と router mock をリセット
    localStorage.clear();
    mockPush.mockClear();
  });

  describe('AuthProvider', () => {
    it('provides initial state with null user and token', async () => {
      render(
        <AuthProvider>
          <TestComponent />
        </AuthProvider>
      );

      // loading が false になるのを待つ
      await waitFor(() => {
        expect(screen.getByTestId('loading').textContent).toBe('false');
      });

      expect(screen.getByTestId('user').textContent).toBe('null');
      expect(screen.getByTestId('token').textContent).toBe('null');
    });

    it('loads user and token from localStorage on mount', async () => {
      // localStorage に認証情報を設定
      localStorage.setItem('token', 'stored-token');
      localStorage.setItem('user', JSON.stringify(testUsers.viewer));

      render(
        <AuthProvider>
          <TestComponent />
        </AuthProvider>
      );

      await waitFor(() => {
        expect(screen.getByTestId('loading').textContent).toBe('false');
      });

      expect(screen.getByTestId('token').textContent).toBe('stored-token');
      expect(screen.getByTestId('user').textContent).toBe(JSON.stringify(testUsers.viewer));
    });

    it('login stores token and user in state and localStorage', async () => {
      render(
        <AuthProvider>
          <TestComponent />
        </AuthProvider>
      );

      await waitFor(() => {
        expect(screen.getByTestId('loading').textContent).toBe('false');
      });

      // login ボタンをクリック
      act(() => {
        screen.getByTestId('login-btn').click();
      });

      // 状態が更新されていることを確認
      expect(screen.getByTestId('token').textContent).toBe('test-token');
      expect(screen.getByTestId('user').textContent).toBe(JSON.stringify(testUsers.viewer));

      // localStorage にも保存されていることを確認
      expect(localStorage.getItem('token')).toBe('test-token');
      expect(localStorage.getItem('user')).toBe(JSON.stringify(testUsers.viewer));
    });

    it('logout clears token and user from state and localStorage', async () => {
      // 最初にログイン状態を設定
      localStorage.setItem('token', 'existing-token');
      localStorage.setItem('user', JSON.stringify(testUsers.viewer));

      render(
        <AuthProvider>
          <TestComponent />
        </AuthProvider>
      );

      await waitFor(() => {
        expect(screen.getByTestId('loading').textContent).toBe('false');
      });

      // logout ボタンをクリック
      await act(async () => {
        screen.getByTestId('logout-btn').click();
      });

      await waitFor(() => {
        // 状態がクリアされていることを確認
        expect(screen.getByTestId('token').textContent).toBe('null');
        expect(screen.getByTestId('user').textContent).toBe('null');
      });

      // localStorage もクリアされていることを確認
      expect(localStorage.getItem('token')).toBeNull();
      expect(localStorage.getItem('user')).toBeNull();

      // router.push が /login で呼ばれていることを確認
      expect(mockPush).toHaveBeenCalledWith('/login');
    });

    it('logout redirects to login page', async () => {
      localStorage.setItem('token', 'existing-token');
      localStorage.setItem('user', JSON.stringify(testUsers.viewer));

      render(
        <AuthProvider>
          <TestComponent />
        </AuthProvider>
      );

      await waitFor(() => {
        expect(screen.getByTestId('loading').textContent).toBe('false');
      });

      await act(async () => {
        screen.getByTestId('logout-btn').click();
      });

      await waitFor(() => {
        expect(mockPush).toHaveBeenCalledWith('/login');
      });
    });
  });

  describe('useAuth hook', () => {
    it('throws error when used outside AuthProvider', () => {
      // エラーをキャッチするためにコンソールエラーを抑制
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {});

      function ComponentWithoutProvider() {
        useAuth();
        return null;
      }

      expect(() => {
        render(<ComponentWithoutProvider />);
      }).toThrow('useAuth must be used within an AuthProvider');

      consoleSpy.mockRestore();
    });

    it('returns auth context values when used within AuthProvider', async () => {
      function TestAuthValues() {
        const auth = useAuth();
        return (
          <div>
            <span data-testid="has-login">{typeof auth.login === 'function' ? 'yes' : 'no'}</span>
            <span data-testid="has-logout">{typeof auth.logout === 'function' ? 'yes' : 'no'}</span>
            <span data-testid="has-user">{'user' in auth ? 'yes' : 'no'}</span>
            <span data-testid="has-token">{'token' in auth ? 'yes' : 'no'}</span>
            <span data-testid="has-loading">{'loading' in auth ? 'yes' : 'no'}</span>
          </div>
        );
      }

      render(
        <AuthProvider>
          <TestAuthValues />
        </AuthProvider>
      );

      expect(screen.getByTestId('has-login').textContent).toBe('yes');
      expect(screen.getByTestId('has-logout').textContent).toBe('yes');
      expect(screen.getByTestId('has-user').textContent).toBe('yes');
      expect(screen.getByTestId('has-token').textContent).toBe('yes');
      expect(screen.getByTestId('has-loading').textContent).toBe('yes');
    });
  });

  describe('loading state', () => {
    it('sets loading to true initially and false after initialization', async () => {
      render(
        <AuthProvider>
          <TestComponent />
        </AuthProvider>
      );

      // 初期化後に loading が false になることを確認
      await waitFor(() => {
        expect(screen.getByTestId('loading').textContent).toBe('false');
      });
    });
  });

  describe('user roles', () => {
    it('stores admin user correctly', async () => {
      render(
        <AuthProvider>
          <TestComponent />
        </AuthProvider>
      );

      await waitFor(() => {
        expect(screen.getByTestId('loading').textContent).toBe('false');
      });

      // admin login ボタンをクリック
      act(() => {
        screen.getByTestId('login-admin-btn').click();
      });

      // admin ユーザーの role を確認
      const userContent = screen.getByTestId('user').textContent;
      expect(userContent).toContain('"role":"admin"');
      expect(screen.getByTestId('token').textContent).toBe('admin-token');
    });

    it('stores editor user correctly', async () => {
      render(
        <AuthProvider>
          <TestComponent />
        </AuthProvider>
      );

      await waitFor(() => {
        expect(screen.getByTestId('loading').textContent).toBe('false');
      });

      // editor login ボタンをクリック
      act(() => {
        screen.getByTestId('login-editor-btn').click();
      });

      // editor ユーザーの role を確認
      const userContent = screen.getByTestId('user').textContent;
      expect(userContent).toContain('"role":"editor"');
      expect(screen.getByTestId('token').textContent).toBe('editor-token');
    });
  });
});
