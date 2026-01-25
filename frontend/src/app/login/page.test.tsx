import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import { server } from '@/test/mocks/server';
import LoginPage from './page';
import { testUsers } from '@/test/fixtures';

const API_URL = 'http://localhost:8080/api';

// Next.js router をモック
const mockPush = vi.fn();
vi.mock('next/navigation', () => ({
  useRouter: () => ({
    push: mockPush,
    replace: vi.fn(),
    prefetch: vi.fn(),
  }),
}));

// AuthContext をモック
const mockLogin = vi.fn();
vi.mock('@/context/AuthContext', () => ({
  useAuth: () => ({
    login: mockLogin,
    user: null,
    token: null,
    loading: false,
    logout: vi.fn(),
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

describe('LoginPage', () => {
  beforeEach(() => {
    mockPush.mockClear();
    mockLogin.mockClear();
    localStorage.clear();
  });

  describe('rendering', () => {
    it('renders login form with all fields', () => {
      render(<LoginPage />);

      expect(screen.getByLabelText('メールアドレス')).toBeInTheDocument();
      expect(screen.getByLabelText('パスワード')).toBeInTheDocument();
      expect(screen.getByRole('button', { name: 'ログイン' })).toBeInTheDocument();
    });

    it('renders forgot password link', () => {
      render(<LoginPage />);

      const forgotPasswordLink = screen.getByText('パスワードを忘れた場合');
      expect(forgotPasswordLink).toBeInTheDocument();
      expect(forgotPasswordLink).toHaveAttribute('href', '/forgot-password');
    });

    it('renders signup link', () => {
      render(<LoginPage />);

      const signupLink = screen.getByText('アカウントをお持ちでない方はこちら →');
      expect(signupLink).toBeInTheDocument();
      expect(signupLink).toHaveAttribute('href', '/signup');
    });

    it('renders logo', () => {
      render(<LoginPage />);

      const logo = screen.getByAltText('Incidex');
      expect(logo).toBeInTheDocument();
    });
  });

  describe('field validation', () => {
    it('shows email validation error for invalid email', async () => {
      render(<LoginPage />);

      const emailInput = screen.getByLabelText('メールアドレス');
      fireEvent.change(emailInput, { target: { value: 'invalid-email' } });
      fireEvent.blur(emailInput);

      await waitFor(() => {
        expect(screen.getByText(/有効なメールアドレスを入力してください/)).toBeInTheDocument();
      });
    });

    it('shows email validation error when email is empty', async () => {
      render(<LoginPage />);

      const emailInput = screen.getByLabelText('メールアドレス');
      fireEvent.focus(emailInput);
      fireEvent.blur(emailInput);

      await waitFor(() => {
        expect(screen.getByText(/メールアドレスは必須です/)).toBeInTheDocument();
      });
    });

    it('shows password validation error when password is empty', async () => {
      render(<LoginPage />);

      const passwordInput = screen.getByLabelText('パスワード');
      fireEvent.focus(passwordInput);
      fireEvent.blur(passwordInput);

      await waitFor(() => {
        expect(screen.getByText('パスワードは必須です')).toBeInTheDocument();
      });
    });

    it('clears validation error when field is corrected', async () => {
      render(<LoginPage />);

      const emailInput = screen.getByLabelText('メールアドレス');

      // 無効なメールを入力
      fireEvent.change(emailInput, { target: { value: 'invalid' } });
      fireEvent.blur(emailInput);

      await waitFor(() => {
        expect(screen.getByText(/有効なメールアドレスを入力してください/)).toBeInTheDocument();
      });

      // 有効なメールに修正
      fireEvent.change(emailInput, { target: { value: 'valid@example.com' } });

      await waitFor(() => {
        expect(screen.queryByText(/有効なメールアドレスを入力してください/)).not.toBeInTheDocument();
      });
    });
  });

  describe('form submission', () => {
    it('submits form with valid credentials', async () => {
      server.use(
        http.post(`${API_URL}/auth/login`, () => {
          return HttpResponse.json({
            access_token: 'test-token',
            user: testUsers.viewer,
          });
        })
      );

      render(<LoginPage />);

      fireEvent.change(screen.getByLabelText('メールアドレス'), { target: { value: 'test@example.com' } });
      fireEvent.change(screen.getByLabelText('パスワード'), { target: { value: 'password123' } });
      fireEvent.click(screen.getByRole('button', { name: 'ログイン' }));

      await waitFor(() => {
        expect(mockLogin).toHaveBeenCalledWith('test-token', testUsers.viewer);
        expect(mockPush).toHaveBeenCalledWith('/');
      });
    });

    it('shows error message for invalid credentials', async () => {
      server.use(
        http.post(`${API_URL}/auth/login`, () => {
          return HttpResponse.json(
            { error: 'Invalid credentials' },
            { status: 401 }
          );
        })
      );

      render(<LoginPage />);

      fireEvent.change(screen.getByLabelText('メールアドレス'), { target: { value: 'wrong@example.com' } });
      fireEvent.change(screen.getByLabelText('パスワード'), { target: { value: 'wrongpassword' } });
      fireEvent.click(screen.getByRole('button', { name: 'ログイン' }));

      await waitFor(() => {
        expect(screen.getByText('Invalid credentials')).toBeInTheDocument();
      });
    });

    it('does not submit form with invalid email', async () => {
      render(<LoginPage />);

      const emailInput = screen.getByLabelText('メールアドレス');
      const passwordInput = screen.getByLabelText('パスワード');

      fireEvent.change(emailInput, { target: { value: 'invalid-email' } });
      fireEvent.blur(emailInput); // バリデーションをトリガー
      fireEvent.change(passwordInput, { target: { value: 'password123' } });

      // バリデーションエラーが表示されることを確認
      await waitFor(() => {
        expect(screen.getByText(/有効なメールアドレスを入力してください/)).toBeInTheDocument();
      });

      // フォーム送信
      fireEvent.click(screen.getByRole('button', { name: 'ログイン' }));

      // ログインが呼ばれないことを確認
      expect(mockLogin).not.toHaveBeenCalled();
    });

    it('does not submit form with empty password', async () => {
      render(<LoginPage />);

      const emailInput = screen.getByLabelText('メールアドレス');
      const passwordInput = screen.getByLabelText('パスワード');

      fireEvent.change(emailInput, { target: { value: 'test@example.com' } });
      // パスワードを空のままにしてblur
      fireEvent.focus(passwordInput);
      fireEvent.blur(passwordInput);

      // バリデーションエラーが表示されることを確認
      await waitFor(() => {
        expect(screen.getByText('パスワードは必須です')).toBeInTheDocument();
      });

      // フォーム送信
      fireEvent.click(screen.getByRole('button', { name: 'ログイン' }));

      // ログインが呼ばれないことを確認
      expect(mockLogin).not.toHaveBeenCalled();
    });

    it('trims email before submission', async () => {
      server.use(
        http.post(`${API_URL}/auth/login`, async ({ request }) => {
          const body = await request.json() as { email: string };
          // トリムされたメールアドレスを確認
          expect(body.email).toBe('test@example.com');
          return HttpResponse.json({
            access_token: 'test-token',
            user: testUsers.viewer,
          });
        })
      );

      render(<LoginPage />);

      fireEvent.change(screen.getByLabelText('メールアドレス'), { target: { value: '  test@example.com  ' } });
      fireEvent.change(screen.getByLabelText('パスワード'), { target: { value: 'password123' } });
      fireEvent.click(screen.getByRole('button', { name: 'ログイン' }));

      await waitFor(() => {
        expect(mockLogin).toHaveBeenCalled();
      });
    });
  });

  describe('accessibility', () => {
    it('has accessible form labels', () => {
      render(<LoginPage />);

      expect(screen.getByLabelText('メールアドレス')).toHaveAttribute('type', 'email');
      expect(screen.getByLabelText('パスワード')).toHaveAttribute('type', 'password');
    });

    it('has required attributes on form fields', () => {
      render(<LoginPage />);

      expect(screen.getByLabelText('メールアドレス')).toHaveAttribute('required');
      expect(screen.getByLabelText('パスワード')).toHaveAttribute('required');
    });

    it('has submit button with correct type', () => {
      render(<LoginPage />);

      expect(screen.getByRole('button', { name: 'ログイン' })).toHaveAttribute('type', 'submit');
    });
  });

  describe('input handling', () => {
    it('updates email value on change', () => {
      render(<LoginPage />);

      const emailInput = screen.getByLabelText('メールアドレス');
      fireEvent.change(emailInput, { target: { value: 'test@example.com' } });

      expect(emailInput).toHaveValue('test@example.com');
    });

    it('updates password value on change', () => {
      render(<LoginPage />);

      const passwordInput = screen.getByLabelText('パスワード');
      fireEvent.change(passwordInput, { target: { value: 'mypassword' } });

      expect(passwordInput).toHaveValue('mypassword');
    });

    it('has correct placeholder text', () => {
      render(<LoginPage />);

      expect(screen.getByLabelText('メールアドレス')).toHaveAttribute('placeholder', 'user@example.com');
      expect(screen.getByLabelText('パスワード')).toHaveAttribute('placeholder', '••••••••');
    });
  });

  describe('error handling', () => {
    it('clears general error when form is resubmitted', async () => {
      let callCount = 0;

      server.use(
        http.post(`${API_URL}/auth/login`, () => {
          callCount++;
          if (callCount === 1) {
            return HttpResponse.json(
              { error: 'Invalid credentials' },
              { status: 401 }
            );
          }
          return HttpResponse.json({
            access_token: 'test-token',
            user: testUsers.viewer,
          });
        })
      );

      render(<LoginPage />);

      // 最初のログイン試行（失敗）
      fireEvent.change(screen.getByLabelText('メールアドレス'), { target: { value: 'test@example.com' } });
      fireEvent.change(screen.getByLabelText('パスワード'), { target: { value: 'wrongpassword' } });
      fireEvent.click(screen.getByRole('button', { name: 'ログイン' }));

      await waitFor(() => {
        expect(screen.getByText('Invalid credentials')).toBeInTheDocument();
      });

      // パスワードを変更して再試行
      fireEvent.change(screen.getByLabelText('パスワード'), { target: { value: 'correctpassword' } });
      fireEvent.click(screen.getByRole('button', { name: 'ログイン' }));

      await waitFor(() => {
        expect(screen.queryByText('Invalid credentials')).not.toBeInTheDocument();
        expect(mockLogin).toHaveBeenCalled();
      });
    });

    it('handles network error', async () => {
      server.use(
        http.post(`${API_URL}/auth/login`, () => {
          return HttpResponse.json(
            { error: 'Network error' },
            { status: 500 }
          );
        })
      );

      render(<LoginPage />);

      fireEvent.change(screen.getByLabelText('メールアドレス'), { target: { value: 'test@example.com' } });
      fireEvent.change(screen.getByLabelText('パスワード'), { target: { value: 'password123' } });
      fireEvent.click(screen.getByRole('button', { name: 'ログイン' }));

      await waitFor(() => {
        expect(screen.getByText('Network error')).toBeInTheDocument();
      });

      expect(mockLogin).not.toHaveBeenCalled();
      expect(mockPush).not.toHaveBeenCalled();
    });
  });
});
