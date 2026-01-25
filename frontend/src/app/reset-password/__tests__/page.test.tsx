import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import ResetPasswordPage from '../page';

// Mock next/navigation
const mockSearchParams = vi.fn();
const mockRouterPush = vi.fn();
vi.mock('next/navigation', () => ({
  useRouter: () => ({
    push: mockRouterPush,
  }),
  useSearchParams: () => ({
    get: mockSearchParams,
  }),
}));

// Mock next/link
vi.mock('next/link', () => ({
  default: ({ children, href }: { children: React.ReactNode; href: string }) => (
    <a href={href}>{children}</a>
  ),
}));

// Mock authApi
const mockValidateResetToken = vi.fn();
const mockResetPassword = vi.fn();
vi.mock('@/lib/api', () => ({
  authApi: {
    validateResetToken: (token: string) => mockValidateResetToken(token),
    resetPassword: (token: string, password: string) => mockResetPassword(token, password),
  },
}));

describe('ResetPasswordPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('token validation', () => {
    it('shows loading state while validating token', async () => {
      mockSearchParams.mockReturnValue('valid-token');
      mockValidateResetToken.mockImplementation(
        () => new Promise((resolve) => setTimeout(() => resolve({ valid: true }), 100))
      );

      render(<ResetPasswordPage />);

      expect(screen.getByText('トークンを検証しています...')).toBeInTheDocument();
    });

    it('shows error when no token is provided', async () => {
      mockSearchParams.mockReturnValue(null);

      render(<ResetPasswordPage />);

      await waitFor(() => {
        expect(screen.getByText('無効なリンク')).toBeInTheDocument();
      });
    });

    it('shows error for invalid token', async () => {
      mockSearchParams.mockReturnValue('invalid-token');
      mockValidateResetToken.mockResolvedValue({ valid: false });

      render(<ResetPasswordPage />);

      await waitFor(() => {
        expect(screen.getByText('無効なリンク')).toBeInTheDocument();
        expect(screen.getByText('このリンクは無効か、有効期限が切れています')).toBeInTheDocument();
      });
    });

    it('shows password form for valid token', async () => {
      mockSearchParams.mockReturnValue('valid-token');
      mockValidateResetToken.mockResolvedValue({ valid: true });

      render(<ResetPasswordPage />);

      await waitFor(() => {
        expect(screen.getByText('新しいパスワードを設定')).toBeInTheDocument();
      });
    });
  });

  describe('password form', () => {
    beforeEach(async () => {
      mockSearchParams.mockReturnValue('valid-token');
      mockValidateResetToken.mockResolvedValue({ valid: true });
    });

    it('shows password requirements when typing', async () => {
      render(<ResetPasswordPage />);

      await waitFor(() => {
        expect(screen.getByText('新しいパスワードを設定')).toBeInTheDocument();
      });

      const passwordInput = screen.getByPlaceholderText('新しいパスワード');
      fireEvent.change(passwordInput, { target: { value: 'Test' } });

      expect(screen.getByText('8文字以上')).toBeInTheDocument();
      expect(screen.getByText('大文字を含む')).toBeInTheDocument();
      expect(screen.getByText('小文字を含む')).toBeInTheDocument();
      expect(screen.getByText('数字を含む')).toBeInTheDocument();
    });

    it('shows password mismatch error', async () => {
      render(<ResetPasswordPage />);

      await waitFor(() => {
        expect(screen.getByText('新しいパスワードを設定')).toBeInTheDocument();
      });

      const passwordInput = screen.getByPlaceholderText('新しいパスワード');
      const confirmInput = screen.getByPlaceholderText('パスワード（確認）');

      fireEvent.change(passwordInput, { target: { value: 'Password123' } });
      fireEvent.change(confirmInput, { target: { value: 'Password456' } });

      expect(screen.getByText('パスワードが一致しません')).toBeInTheDocument();
    });

    it('shows password match confirmation', async () => {
      render(<ResetPasswordPage />);

      await waitFor(() => {
        expect(screen.getByText('新しいパスワードを設定')).toBeInTheDocument();
      });

      const passwordInput = screen.getByPlaceholderText('新しいパスワード');
      const confirmInput = screen.getByPlaceholderText('パスワード（確認）');

      fireEvent.change(passwordInput, { target: { value: 'Password123' } });
      fireEvent.change(confirmInput, { target: { value: 'Password123' } });

      expect(screen.getByText('パスワードが一致しています')).toBeInTheDocument();
    });
  });

  describe('password submission', () => {
    beforeEach(async () => {
      mockSearchParams.mockReturnValue('valid-token');
      mockValidateResetToken.mockResolvedValue({ valid: true });
    });

    it('submits password and shows success', async () => {
      mockResetPassword.mockResolvedValue(undefined);

      render(<ResetPasswordPage />);

      await waitFor(() => {
        expect(screen.getByText('新しいパスワードを設定')).toBeInTheDocument();
      });

      const passwordInput = screen.getByPlaceholderText('新しいパスワード');
      const confirmInput = screen.getByPlaceholderText('パスワード（確認）');

      fireEvent.change(passwordInput, { target: { value: 'Password123' } });
      fireEvent.change(confirmInput, { target: { value: 'Password123' } });

      const submitButton = screen.getByText('パスワードをリセット');
      fireEvent.click(submitButton);

      await waitFor(() => {
        expect(screen.getByText('パスワードをリセットしました')).toBeInTheDocument();
      });

      expect(mockResetPassword).toHaveBeenCalledWith('valid-token', 'Password123');
    });

    it('shows error on submission failure', async () => {
      mockResetPassword.mockRejectedValue(new Error('リセットに失敗しました'));

      render(<ResetPasswordPage />);

      await waitFor(() => {
        expect(screen.getByText('新しいパスワードを設定')).toBeInTheDocument();
      });

      const passwordInput = screen.getByPlaceholderText('新しいパスワード');
      const confirmInput = screen.getByPlaceholderText('パスワード（確認）');

      fireEvent.change(passwordInput, { target: { value: 'Password123' } });
      fireEvent.change(confirmInput, { target: { value: 'Password123' } });

      const submitButton = screen.getByText('パスワードをリセット');
      fireEvent.click(submitButton);

      await waitFor(() => {
        expect(screen.getByText('リセットに失敗しました')).toBeInTheDocument();
      });
    });

    it('disables submit button when password requirements not met', async () => {
      render(<ResetPasswordPage />);

      await waitFor(() => {
        expect(screen.getByText('新しいパスワードを設定')).toBeInTheDocument();
      });

      const passwordInput = screen.getByPlaceholderText('新しいパスワード');
      const confirmInput = screen.getByPlaceholderText('パスワード（確認）');

      // Password too short, no numbers
      fireEvent.change(passwordInput, { target: { value: 'Pass' } });
      fireEvent.change(confirmInput, { target: { value: 'Pass' } });

      const submitButton = screen.getByText('パスワードをリセット');
      expect(submitButton).toBeDisabled();
    });
  });

  describe('invalid token page', () => {
    it('shows links to request new reset and login', async () => {
      mockSearchParams.mockReturnValue('invalid-token');
      mockValidateResetToken.mockResolvedValue({ valid: false });

      render(<ResetPasswordPage />);

      await waitFor(() => {
        expect(screen.getByText('無効なリンク')).toBeInTheDocument();
      });

      expect(screen.getByText('パスワードリセットをリクエスト')).toBeInTheDocument();
      expect(screen.getByText('ログインページに戻る')).toBeInTheDocument();
    });
  });
});
