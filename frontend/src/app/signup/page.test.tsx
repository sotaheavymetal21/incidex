import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import { server } from '@/test/mocks/server';
import SignupPage from './page';
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

describe('SignupPage', () => {
  beforeEach(() => {
    mockPush.mockClear();
    localStorage.clear();
  });

  describe('rendering', () => {
    it('renders signup form with all fields', () => {
      render(<SignupPage />);

      expect(screen.getByLabelText(/氏名/)).toBeInTheDocument();
      expect(screen.getByLabelText(/メールアドレス/)).toBeInTheDocument();
      expect(screen.getByLabelText(/社員番号/)).toBeInTheDocument();
      expect(screen.getByLabelText(/所属部署/)).toBeInTheDocument();
      expect(screen.getByLabelText(/パスワード/)).toBeInTheDocument();
      expect(screen.getByRole('button', { name: 'アカウント作成' })).toBeInTheDocument();
    });

    it('renders login link', () => {
      render(<SignupPage />);

      const loginLink = screen.getByText('すでにアカウントをお持ちの方はこちら →');
      expect(loginLink).toBeInTheDocument();
      expect(loginLink).toHaveAttribute('href', '/login');
    });

    it('renders logo', () => {
      render(<SignupPage />);

      const logo = screen.getByAltText('Incidex');
      expect(logo).toBeInTheDocument();
    });

    it('renders password requirements', () => {
      render(<SignupPage />);

      expect(screen.getByText('パスワード要件:')).toBeInTheDocument();
      expect(screen.getByText('8文字以上')).toBeInTheDocument();
      expect(screen.getByText('大文字（A-Z）を含む')).toBeInTheDocument();
      expect(screen.getByText('小文字（a-z）を含む')).toBeInTheDocument();
      expect(screen.getByText('数字（0-9）を含む')).toBeInTheDocument();
    });
  });

  describe('field validation', () => {
    it('shows name validation error when name is empty', async () => {
      render(<SignupPage />);

      const nameInput = screen.getByLabelText(/氏名/);
      fireEvent.focus(nameInput);
      fireEvent.blur(nameInput);

      await waitFor(() => {
        expect(screen.getByText(/名前は必須です/)).toBeInTheDocument();
      });
    });

    it('shows email validation error for invalid email', async () => {
      render(<SignupPage />);

      const emailInput = screen.getByLabelText(/メールアドレス/);
      fireEvent.change(emailInput, { target: { value: 'invalid-email' } });
      fireEvent.blur(emailInput);

      await waitFor(() => {
        expect(screen.getByText(/有効なメールアドレスを入力してください/)).toBeInTheDocument();
      });
    });

    it('shows employee number validation error for invalid characters', async () => {
      render(<SignupPage />);

      const employeeNumberInput = screen.getByLabelText(/社員番号/);
      fireEvent.change(employeeNumberInput, { target: { value: 'EMP@001' } });
      fireEvent.blur(employeeNumberInput);

      await waitFor(() => {
        expect(screen.getByText(/社員番号は英数字とハイフンのみ使用できます/)).toBeInTheDocument();
      });
    });

    it('validates department max length', () => {
      render(<SignupPage />);

      const departmentInput = screen.getByLabelText(/所属部署/);
      // 50文字以内のバリデーションを確認
      expect(departmentInput).toHaveAttribute('maxLength', '50');
    });

    it('shows password validation error for weak password', async () => {
      render(<SignupPage />);

      const passwordInput = screen.getByLabelText(/パスワード/);
      fireEvent.change(passwordInput, { target: { value: 'weak' } });
      fireEvent.blur(passwordInput);

      await waitFor(() => {
        expect(screen.getByText(/パスワードは8文字以上である必要があります/)).toBeInTheDocument();
      });
    });

    it('shows password validation error when missing uppercase', async () => {
      render(<SignupPage />);

      const passwordInput = screen.getByLabelText(/パスワード/);
      fireEvent.change(passwordInput, { target: { value: 'password123' } });
      fireEvent.blur(passwordInput);

      await waitFor(() => {
        expect(screen.getByText(/パスワードには大文字を含める必要があります/)).toBeInTheDocument();
      });
    });

    it('clears validation error when field is corrected', async () => {
      render(<SignupPage />);

      const nameInput = screen.getByLabelText(/氏名/);

      // 空のままblur
      fireEvent.focus(nameInput);
      fireEvent.blur(nameInput);

      await waitFor(() => {
        expect(screen.getByText(/名前は必須です/)).toBeInTheDocument();
      });

      // 有効な値を入力
      fireEvent.change(nameInput, { target: { value: '山田 太郎' } });

      await waitFor(() => {
        expect(screen.queryByText(/名前は必須です/)).not.toBeInTheDocument();
      });
    });
  });

  describe('form submission', () => {
    const fillValidForm = () => {
      fireEvent.change(screen.getByLabelText(/氏名/), { target: { value: '山田 太郎' } });
      fireEvent.change(screen.getByLabelText(/メールアドレス/), { target: { value: 'test@example.com' } });
      fireEvent.change(screen.getByLabelText(/社員番号/), { target: { value: 'EMP-001' } });
      fireEvent.change(screen.getByLabelText(/所属部署/), { target: { value: '開発部' } });
      fireEvent.change(screen.getByLabelText(/パスワード/), { target: { value: 'Password123' } });
    };

    it('submits form with valid data and redirects to login', async () => {
      server.use(
        http.post(`${API_URL}/auth/register`, () => {
          return HttpResponse.json({
            access_token: 'test-token',
            user: testUsers.viewer,
          });
        })
      );

      render(<SignupPage />);
      fillValidForm();
      fireEvent.click(screen.getByRole('button', { name: 'アカウント作成' }));

      await waitFor(() => {
        expect(mockPush).toHaveBeenCalledWith('/login');
      });
    });

    it('sends correct data to API', async () => {
      server.use(
        http.post(`${API_URL}/auth/register`, async ({ request }) => {
          const body = await request.json() as Record<string, string>;
          expect(body.name).toBe('山田 太郎');
          expect(body.email).toBe('test@example.com');
          expect(body.employee_number).toBe('EMP-001');
          expect(body.department).toBe('開発部');
          expect(body.password).toBe('Password123');
          return HttpResponse.json({
            access_token: 'test-token',
            user: testUsers.viewer,
          });
        })
      );

      render(<SignupPage />);
      fillValidForm();
      fireEvent.click(screen.getByRole('button', { name: 'アカウント作成' }));

      await waitFor(() => {
        expect(mockPush).toHaveBeenCalledWith('/login');
      });
    });

    it('shows error message when email already exists', async () => {
      server.use(
        http.post(`${API_URL}/auth/register`, () => {
          return HttpResponse.json(
            { error: 'Email already exists' },
            { status: 409 }
          );
        })
      );

      render(<SignupPage />);
      fillValidForm();
      fireEvent.click(screen.getByRole('button', { name: 'アカウント作成' }));

      await waitFor(() => {
        expect(screen.getByText('Email already exists')).toBeInTheDocument();
      });

      expect(mockPush).not.toHaveBeenCalled();
    });

    it('validates required fields on form submission', async () => {
      render(<SignupPage />);

      // 名前フィールドにフォーカスしてblur
      const nameInput = screen.getByLabelText(/氏名/);
      fireEvent.focus(nameInput);
      fireEvent.blur(nameInput);

      // バリデーションエラーが表示される
      await waitFor(() => {
        expect(screen.getByText(/名前は必須です/)).toBeInTheDocument();
      });
    });

    it('trims input values before submission', async () => {
      let receivedBody: Record<string, string> | null = null;

      server.use(
        http.post(`${API_URL}/auth/register`, async ({ request }) => {
          receivedBody = await request.json() as Record<string, string>;
          return HttpResponse.json({
            access_token: 'test-token',
            user: testUsers.viewer,
          });
        })
      );

      render(<SignupPage />);

      // 前後に空白を含む値を入力（emailとemployeeNumberは空白を含むとバリデーションエラーになるため、nameとdepartmentのみ）
      fireEvent.change(screen.getByLabelText(/氏名/), { target: { value: '  山田 太郎  ' } });
      fireEvent.change(screen.getByLabelText(/メールアドレス/), { target: { value: 'test@example.com' } });
      fireEvent.change(screen.getByLabelText(/社員番号/), { target: { value: 'EMP-001' } });
      fireEvent.change(screen.getByLabelText(/所属部署/), { target: { value: '  開発部  ' } });
      fireEvent.change(screen.getByLabelText(/パスワード/), { target: { value: 'Password123' } });
      fireEvent.click(screen.getByRole('button', { name: 'アカウント作成' }));

      await waitFor(() => {
        expect(mockPush).toHaveBeenCalledWith('/login');
      });

      // トリムされた値を確認（name と department のみ空白がトリムされる）
      expect(receivedBody).not.toBeNull();
      expect(receivedBody!.name).toBe('山田 太郎');
      expect(receivedBody!.email).toBe('test@example.com');
      expect(receivedBody!.employee_number).toBe('EMP-001');
      expect(receivedBody!.department).toBe('開発部');
    });
  });

  describe('accessibility', () => {
    it('has accessible form labels', () => {
      render(<SignupPage />);

      expect(screen.getByLabelText(/氏名/)).toHaveAttribute('type', 'text');
      expect(screen.getByLabelText(/メールアドレス/)).toHaveAttribute('type', 'email');
      expect(screen.getByLabelText(/社員番号/)).toHaveAttribute('type', 'text');
      expect(screen.getByLabelText(/所属部署/)).toHaveAttribute('type', 'text');
      expect(screen.getByLabelText(/パスワード/)).toHaveAttribute('type', 'password');
    });

    it('has required attributes on form fields', () => {
      render(<SignupPage />);

      expect(screen.getByLabelText(/氏名/)).toHaveAttribute('required');
      expect(screen.getByLabelText(/メールアドレス/)).toHaveAttribute('required');
      expect(screen.getByLabelText(/社員番号/)).toHaveAttribute('required');
      expect(screen.getByLabelText(/所属部署/)).toHaveAttribute('required');
      expect(screen.getByLabelText(/パスワード/)).toHaveAttribute('required');
    });

    it('has submit button with correct type', () => {
      render(<SignupPage />);

      expect(screen.getByRole('button', { name: 'アカウント作成' })).toHaveAttribute('type', 'submit');
    });
  });

  describe('input handling', () => {
    it('updates field values on change', () => {
      render(<SignupPage />);

      const nameInput = screen.getByLabelText(/氏名/);
      fireEvent.change(nameInput, { target: { value: '山田 太郎' } });
      expect(nameInput).toHaveValue('山田 太郎');

      const emailInput = screen.getByLabelText(/メールアドレス/);
      fireEvent.change(emailInput, { target: { value: 'test@example.com' } });
      expect(emailInput).toHaveValue('test@example.com');

      const employeeNumberInput = screen.getByLabelText(/社員番号/);
      fireEvent.change(employeeNumberInput, { target: { value: 'EMP-001' } });
      expect(employeeNumberInput).toHaveValue('EMP-001');

      const departmentInput = screen.getByLabelText(/所属部署/);
      fireEvent.change(departmentInput, { target: { value: '開発部' } });
      expect(departmentInput).toHaveValue('開発部');

      const passwordInput = screen.getByLabelText(/パスワード/);
      fireEvent.change(passwordInput, { target: { value: 'Password123' } });
      expect(passwordInput).toHaveValue('Password123');
    });

    it('has correct placeholder text', () => {
      render(<SignupPage />);

      expect(screen.getByLabelText(/氏名/)).toHaveAttribute('placeholder', '山田 太郎');
      expect(screen.getByLabelText(/メールアドレス/)).toHaveAttribute('placeholder', 'user@example.com');
      expect(screen.getByLabelText(/社員番号/)).toHaveAttribute('placeholder', 'EMP-001');
      expect(screen.getByLabelText(/所属部署/)).toHaveAttribute('placeholder', '開発部');
      expect(screen.getByLabelText(/パスワード/)).toHaveAttribute('placeholder', '••••••••');
    });
  });

  describe('error handling', () => {
    const fillValidForm = () => {
      fireEvent.change(screen.getByLabelText(/氏名/), { target: { value: '山田 太郎' } });
      fireEvent.change(screen.getByLabelText(/メールアドレス/), { target: { value: 'test@example.com' } });
      fireEvent.change(screen.getByLabelText(/社員番号/), { target: { value: 'EMP-001' } });
      fireEvent.change(screen.getByLabelText(/所属部署/), { target: { value: '開発部' } });
      fireEvent.change(screen.getByLabelText(/パスワード/), { target: { value: 'Password123' } });
    };

    it('clears general error when form is resubmitted', async () => {
      let callCount = 0;

      server.use(
        http.post(`${API_URL}/auth/register`, () => {
          callCount++;
          if (callCount === 1) {
            return HttpResponse.json(
              { error: 'Server error' },
              { status: 500 }
            );
          }
          return HttpResponse.json({
            access_token: 'test-token',
            user: testUsers.viewer,
          });
        })
      );

      render(<SignupPage />);

      // 最初の試行（失敗）
      fillValidForm();
      fireEvent.click(screen.getByRole('button', { name: 'アカウント作成' }));

      await waitFor(() => {
        expect(screen.getByText('Server error')).toBeInTheDocument();
      });

      // 再試行（成功）
      fireEvent.click(screen.getByRole('button', { name: 'アカウント作成' }));

      await waitFor(() => {
        expect(screen.queryByText('Server error')).not.toBeInTheDocument();
        expect(mockPush).toHaveBeenCalledWith('/login');
      });
    });

    it('handles network error', async () => {
      server.use(
        http.post(`${API_URL}/auth/register`, () => {
          return HttpResponse.json(
            { error: 'Network error' },
            { status: 500 }
          );
        })
      );

      render(<SignupPage />);
      fillValidForm();
      fireEvent.click(screen.getByRole('button', { name: 'アカウント作成' }));

      await waitFor(() => {
        expect(screen.getByText('Network error')).toBeInTheDocument();
      });

      expect(mockPush).not.toHaveBeenCalled();
    });
  });
});
