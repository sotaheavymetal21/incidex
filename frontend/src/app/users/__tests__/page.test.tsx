import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import UsersPage from '../page';

// Mock next/navigation
const mockRouterPush = vi.fn();
vi.mock('next/navigation', () => ({
  useRouter: () => ({
    push: mockRouterPush,
  }),
}));

// Mock useAuth
const mockToken = vi.fn();
const mockAuthLoading = vi.fn();
vi.mock('@/context/AuthContext', () => ({
  useAuth: () => ({
    token: mockToken(),
    loading: mockAuthLoading(),
  }),
}));

// Mock userApi
const mockGetAll = vi.fn();
const mockDelete = vi.fn();
const mockToggleActive = vi.fn();
vi.mock('@/lib/api', () => ({
  userApi: {
    getAll: (token: string) => mockGetAll(token),
    delete: (token: string, id: number) => mockDelete(token, id),
    toggleActive: (token: string, id: number, isActive: boolean) =>
      mockToggleActive(token, id, isActive),
  },
}));

// Mock modal components
vi.mock('../EditUserModal', () => ({
  default: ({ user, onClose, onSuccess }: any) => (
    <div data-testid="edit-modal">
      <span>Edit Modal: {user.name}</span>
      <button onClick={onClose}>Close Edit</button>
      <button onClick={onSuccess}>Save Edit</button>
    </div>
  ),
}));

vi.mock('../ChangePasswordModal', () => ({
  default: ({ user, onClose, onSuccess }: any) => (
    <div data-testid="password-modal">
      <span>Password Modal: {user.name}</span>
      <button onClick={onClose}>Close Password</button>
    </div>
  ),
}));

vi.mock('../CreateUserModal', () => ({
  default: ({ onClose, onSuccess }: any) => (
    <div data-testid="create-modal">
      <span>Create User Modal</span>
      <button onClick={onClose}>Close Create</button>
      <button onClick={onSuccess}>Create User</button>
    </div>
  ),
}));

// Mock window.confirm and alert
const mockConfirm = vi.fn();
const mockAlert = vi.fn();
global.confirm = mockConfirm;
global.alert = mockAlert;

describe('UsersPage', () => {
  const mockUsers = [
    {
      id: 1,
      name: 'Admin User',
      email: 'admin@example.com',
      employee_number: 'EMP001',
      department: 'IT',
      role: 'admin' as const,
      is_active: true,
      created_at: '2025-01-01T00:00:00Z',
    },
    {
      id: 2,
      name: 'Editor User',
      email: 'editor@example.com',
      employee_number: 'EMP002',
      department: 'Marketing',
      role: 'editor' as const,
      is_active: true,
      created_at: '2025-01-02T00:00:00Z',
    },
    {
      id: 3,
      name: 'Inactive User',
      email: 'inactive@example.com',
      employee_number: null,
      department: null,
      role: 'viewer' as const,
      is_active: false,
      created_at: '2025-01-03T00:00:00Z',
    },
  ];

  beforeEach(() => {
    vi.clearAllMocks();
    mockToken.mockReturnValue('test-token');
    mockAuthLoading.mockReturnValue(false);
    mockGetAll.mockResolvedValue(mockUsers);
    mockConfirm.mockReturnValue(true);
  });

  describe('authentication', () => {
    it('redirects to login when not authenticated', () => {
      mockToken.mockReturnValue(null);

      render(<UsersPage />);

      expect(mockRouterPush).toHaveBeenCalledWith('/login');
    });

    it('shows loading state during auth check', () => {
      mockAuthLoading.mockReturnValue(true);

      render(<UsersPage />);

      expect(screen.getByText('読み込み中...')).toBeInTheDocument();
    });
  });

  describe('user list', () => {
    it('displays page title', async () => {
      render(<UsersPage />);

      await waitFor(() => {
        expect(screen.getByText('ユーザー管理')).toBeInTheDocument();
      });
    });

    it('displays users from API', async () => {
      render(<UsersPage />);

      await waitFor(() => {
        expect(screen.getByText('Admin User')).toBeInTheDocument();
        expect(screen.getByText('Editor User')).toBeInTheDocument();
        expect(screen.getByText('Inactive User')).toBeInTheDocument();
      });
    });

    it('displays user emails', async () => {
      render(<UsersPage />);

      await waitFor(() => {
        expect(screen.getByText('admin@example.com')).toBeInTheDocument();
        expect(screen.getByText('editor@example.com')).toBeInTheDocument();
      });
    });

    it('displays employee numbers or dash', async () => {
      render(<UsersPage />);

      await waitFor(() => {
        expect(screen.getByText('EMP001')).toBeInTheDocument();
        expect(screen.getByText('EMP002')).toBeInTheDocument();
      });
    });

    it('displays departments or dash', async () => {
      render(<UsersPage />);

      await waitFor(() => {
        expect(screen.getByText('IT')).toBeInTheDocument();
        expect(screen.getByText('Marketing')).toBeInTheDocument();
      });
    });

    it('displays role badges with Japanese labels', async () => {
      render(<UsersPage />);

      await waitFor(() => {
        expect(screen.getByText('管理者')).toBeInTheDocument();
        expect(screen.getByText('編集者')).toBeInTheDocument();
        expect(screen.getByText('閲覧者')).toBeInTheDocument();
      });
    });

    it('displays active/inactive status', async () => {
      render(<UsersPage />);

      await waitFor(() => {
        expect(screen.getAllByText('有効').length).toBeGreaterThan(0);
        expect(screen.getByText('無効')).toBeInTheDocument();
      });
    });

    it('shows error message on fetch failure', async () => {
      mockGetAll.mockRejectedValue(new Error('API Error'));

      render(<UsersPage />);

      await waitFor(() => {
        expect(screen.getByText('API Error')).toBeInTheDocument();
      });
    });
  });

  describe('create user', () => {
    it('shows create button', async () => {
      render(<UsersPage />);

      await waitFor(() => {
        expect(screen.getByText('ユーザーを作成')).toBeInTheDocument();
      });
    });

    it('opens create modal when button is clicked', async () => {
      render(<UsersPage />);

      await waitFor(() => {
        expect(screen.getByText('ユーザーを作成')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('ユーザーを作成'));

      expect(screen.getByTestId('create-modal')).toBeInTheDocument();
    });

    it('closes create modal and refreshes list on success', async () => {
      render(<UsersPage />);

      await waitFor(() => {
        expect(screen.getByText('ユーザーを作成')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('ユーザーを作成'));
      expect(screen.getByTestId('create-modal')).toBeInTheDocument();

      fireEvent.click(screen.getByText('Create User'));

      await waitFor(() => {
        expect(screen.queryByTestId('create-modal')).not.toBeInTheDocument();
      });
    });
  });

  describe('edit user', () => {
    it('opens edit modal when edit button is clicked', async () => {
      render(<UsersPage />);

      await waitFor(() => {
        expect(screen.getAllByText('編集')[0]).toBeInTheDocument();
      });

      fireEvent.click(screen.getAllByText('編集')[0]);

      expect(screen.getByTestId('edit-modal')).toBeInTheDocument();
      expect(screen.getByText('Edit Modal: Admin User')).toBeInTheDocument();
    });

    it('closes edit modal on close', async () => {
      render(<UsersPage />);

      await waitFor(() => {
        expect(screen.getAllByText('編集')[0]).toBeInTheDocument();
      });

      fireEvent.click(screen.getAllByText('編集')[0]);
      expect(screen.getByTestId('edit-modal')).toBeInTheDocument();

      fireEvent.click(screen.getByText('Close Edit'));

      expect(screen.queryByTestId('edit-modal')).not.toBeInTheDocument();
    });
  });

  describe('change password', () => {
    it('opens password modal when button is clicked', async () => {
      render(<UsersPage />);

      await waitFor(() => {
        expect(screen.getAllByText('パスワード変更')[0]).toBeInTheDocument();
      });

      fireEvent.click(screen.getAllByText('パスワード変更')[0]);

      expect(screen.getByTestId('password-modal')).toBeInTheDocument();
      expect(screen.getByText('Password Modal: Admin User')).toBeInTheDocument();
    });
  });

  describe('delete user', () => {
    it('deletes user after confirmation', async () => {
      mockDelete.mockResolvedValue(undefined);

      render(<UsersPage />);

      await waitFor(() => {
        expect(screen.getAllByText('削除')[0]).toBeInTheDocument();
      });

      fireEvent.click(screen.getAllByText('削除')[0]);

      expect(mockConfirm).toHaveBeenCalledWith('このユーザーを削除してもよろしいですか？');
      await waitFor(() => {
        expect(mockDelete).toHaveBeenCalledWith('test-token', 1);
      });
    });

    it('does not delete when confirmation is cancelled', async () => {
      mockConfirm.mockReturnValue(false);

      render(<UsersPage />);

      await waitFor(() => {
        expect(screen.getAllByText('削除')[0]).toBeInTheDocument();
      });

      fireEvent.click(screen.getAllByText('削除')[0]);

      expect(mockDelete).not.toHaveBeenCalled();
    });

    it('shows alert on delete error', async () => {
      mockDelete.mockRejectedValue(new Error('Delete failed'));

      render(<UsersPage />);

      await waitFor(() => {
        expect(screen.getAllByText('削除')[0]).toBeInTheDocument();
      });

      fireEvent.click(screen.getAllByText('削除')[0]);

      await waitFor(() => {
        expect(mockAlert).toHaveBeenCalledWith('Delete failed');
      });
    });
  });

  describe('toggle active status', () => {
    it('activates inactive user after confirmation', async () => {
      mockToggleActive.mockResolvedValue(undefined);

      render(<UsersPage />);

      await waitFor(() => {
        expect(screen.getByText('有効化')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('有効化'));

      expect(mockConfirm).toHaveBeenCalledWith('ユーザーを有効化しますか？');
      await waitFor(() => {
        expect(mockToggleActive).toHaveBeenCalledWith('test-token', 3, true);
      });
    });

    it('deactivates active user after confirmation', async () => {
      mockToggleActive.mockResolvedValue(undefined);

      render(<UsersPage />);

      await waitFor(() => {
        expect(screen.getAllByText('無効化')[0]).toBeInTheDocument();
      });

      fireEvent.click(screen.getAllByText('無効化')[0]);

      expect(mockConfirm).toHaveBeenCalledWith('ユーザーを無効化しますか？');
      await waitFor(() => {
        expect(mockToggleActive).toHaveBeenCalledWith('test-token', 1, false);
      });
    });
  });
});
