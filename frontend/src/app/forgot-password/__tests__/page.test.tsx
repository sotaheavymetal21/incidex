import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import ForgotPasswordPage from '../page';

// Mock next/link
vi.mock('next/link', () => ({
  default: ({ children, href }: { children: React.ReactNode; href: string }) => (
    <a href={href}>{children}</a>
  ),
}));

// Mock authApi
const mockRequestPasswordReset = vi.fn();
vi.mock('@/lib/api', () => ({
  authApi: {
    requestPasswordReset: (email: string) => mockRequestPasswordReset(email),
  },
}));

describe('ForgotPasswordPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('initial rendering', () => {
    it('renders the page title', () => {
      render(<ForgotPasswordPage />);

      expect(screen.getByText('パスワードリセット')).toBeInTheDocument();
    });

    it('renders email input', () => {
      render(<ForgotPasswordPage />);

      expect(screen.getByPlaceholderText('メールアドレス')).toBeInTheDocument();
    });

    it('renders submit button', () => {
      render(<ForgotPasswordPage />);

      expect(screen.getByText('パスワードリセットリンクを送信')).toBeInTheDocument();
    });

    it('renders login link', () => {
      render(<ForgotPasswordPage />);

      expect(screen.getByText('ログインページに戻る')).toBeInTheDocument();
    });

    it('renders signup link', () => {
      render(<ForgotPasswordPage />);

      expect(screen.getByText('新規登録')).toBeInTheDocument();
    });
  });

  describe('form submission', () => {
    it('submits email and shows success message', async () => {
      mockRequestPasswordReset.mockResolvedValue(undefined);

      render(<ForgotPasswordPage />);

      const emailInput = screen.getByPlaceholderText('メールアドレス');
      fireEvent.change(emailInput, { target: { value: 'test@example.com' } });

      const submitButton = screen.getByText('パスワードリセットリンクを送信');
      fireEvent.click(submitButton);

      await waitFor(() => {
        expect(screen.getByText('メールを送信しました')).toBeInTheDocument();
      });

      expect(mockRequestPasswordReset).toHaveBeenCalledWith('test@example.com');
    });

    it('shows loading state during submission', async () => {
      mockRequestPasswordReset.mockImplementation(
        () => new Promise((resolve) => setTimeout(resolve, 100))
      );

      render(<ForgotPasswordPage />);

      const emailInput = screen.getByPlaceholderText('メールアドレス');
      fireEvent.change(emailInput, { target: { value: 'test@example.com' } });

      const submitButton = screen.getByText('パスワードリセットリンクを送信');
      fireEvent.click(submitButton);

      expect(screen.getByText('リクエスト中...')).toBeInTheDocument();
    });

    it('shows error message on failure', async () => {
      mockRequestPasswordReset.mockRejectedValue(new Error('サーバーエラー'));

      render(<ForgotPasswordPage />);

      const emailInput = screen.getByPlaceholderText('メールアドレス');
      fireEvent.change(emailInput, { target: { value: 'test@example.com' } });

      const submitButton = screen.getByText('パスワードリセットリンクを送信');
      fireEvent.click(submitButton);

      await waitFor(() => {
        expect(screen.getByText('サーバーエラー')).toBeInTheDocument();
      });
    });

    it('shows generic error message when error is not an Error instance', async () => {
      mockRequestPasswordReset.mockRejectedValue('unknown error');

      render(<ForgotPasswordPage />);

      const emailInput = screen.getByPlaceholderText('メールアドレス');
      fireEvent.change(emailInput, { target: { value: 'test@example.com' } });

      const submitButton = screen.getByText('パスワードリセットリンクを送信');
      fireEvent.click(submitButton);

      await waitFor(() => {
        expect(screen.getByText('パスワードリセットのリクエストに失敗しました')).toBeInTheDocument();
      });
    });
  });

  describe('success state', () => {
    it('displays success information after email is sent', async () => {
      mockRequestPasswordReset.mockResolvedValue(undefined);

      render(<ForgotPasswordPage />);

      const emailInput = screen.getByPlaceholderText('メールアドレス');
      fireEvent.change(emailInput, { target: { value: 'test@example.com' } });

      const submitButton = screen.getByText('パスワードリセットリンクを送信');
      fireEvent.click(submitButton);

      await waitFor(() => {
        expect(screen.getByText('メールを送信しました')).toBeInTheDocument();
        expect(
          screen.getByText('パスワードリセットの手順を記載したメールを送信しました。')
        ).toBeInTheDocument();
        expect(screen.getByText('リンクの有効期限は1時間です。')).toBeInTheDocument();
      });
    });

    it('shows login page link in success state', async () => {
      mockRequestPasswordReset.mockResolvedValue(undefined);

      render(<ForgotPasswordPage />);

      const emailInput = screen.getByPlaceholderText('メールアドレス');
      fireEvent.change(emailInput, { target: { value: 'test@example.com' } });

      fireEvent.click(screen.getByText('パスワードリセットリンクを送信'));

      await waitFor(() => {
        const loginLink = screen.getByText('ログインページに戻る');
        expect(loginLink.closest('a')).toHaveAttribute('href', '/login');
      });
    });
  });
});
