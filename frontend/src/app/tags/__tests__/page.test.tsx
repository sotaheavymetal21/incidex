import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import TagsPage from '../page';

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

// Mock usePermissions
const mockPermissions = vi.fn();
vi.mock('@/hooks/usePermissions', () => ({
  usePermissions: () => mockPermissions(),
}));

// Mock tagApi
const mockGetAll = vi.fn();
const mockCreate = vi.fn();
const mockUpdate = vi.fn();
const mockDelete = vi.fn();
vi.mock('@/lib/api', () => ({
  tagApi: {
    getAll: (token: string) => mockGetAll(token),
    create: (token: string, data: any) => mockCreate(token, data),
    update: (token: string, id: number, data: any) => mockUpdate(token, id, data),
    delete: (token: string, id: number) => mockDelete(token, id),
  },
}));

// Mock window.confirm
const mockConfirm = vi.fn();
global.confirm = mockConfirm;

describe('TagsPage', () => {
  const mockTags = [
    { id: 1, name: 'Bug', color: '#ef4444' },
    { id: 2, name: 'Feature', color: '#22c55e' },
  ];

  beforeEach(() => {
    vi.clearAllMocks();
    mockToken.mockReturnValue('test-token');
    mockAuthLoading.mockReturnValue(false);
    mockPermissions.mockReturnValue({
      canManageTags: true,
    });
    mockGetAll.mockResolvedValue(mockTags);
  });

  describe('authentication', () => {
    it('redirects to login when not authenticated', () => {
      mockToken.mockReturnValue(null);

      render(<TagsPage />);

      expect(mockRouterPush).toHaveBeenCalledWith('/login');
    });

    it('shows loading state during auth check', () => {
      mockAuthLoading.mockReturnValue(true);

      render(<TagsPage />);

      expect(screen.getByText('Loading...')).toBeInTheDocument();
    });
  });

  describe('tag list', () => {
    it('displays page title', async () => {
      render(<TagsPage />);

      await waitFor(() => {
        expect(screen.getByText('タグ管理')).toBeInTheDocument();
      });
    });

    it('displays tags from API', async () => {
      render(<TagsPage />);

      await waitFor(() => {
        expect(screen.getByText('Bug')).toBeInTheDocument();
        expect(screen.getByText('Feature')).toBeInTheDocument();
      });
    });

    it('displays tag colors', async () => {
      render(<TagsPage />);

      await waitFor(() => {
        expect(screen.getByText('#ef4444')).toBeInTheDocument();
        expect(screen.getByText('#22c55e')).toBeInTheDocument();
      });
    });

    it('shows error message on fetch failure', async () => {
      mockGetAll.mockRejectedValue(new Error('API Error'));

      render(<TagsPage />);

      await waitFor(() => {
        expect(screen.getByText('API Error')).toBeInTheDocument();
      });
    });
  });

  describe('create tag', () => {
    it('shows create button for users with permission', async () => {
      render(<TagsPage />);

      await waitFor(() => {
        expect(screen.getByText('タグを作成')).toBeInTheDocument();
      });
    });

    it('hides create button for users without permission', async () => {
      mockPermissions.mockReturnValue({
        canManageTags: false,
      });

      render(<TagsPage />);

      await waitFor(() => {
        expect(screen.queryByText('タグを作成')).not.toBeInTheDocument();
      });
    });

    it('opens create modal when button is clicked', async () => {
      render(<TagsPage />);

      await waitFor(() => {
        expect(screen.getByText('タグを作成')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('タグを作成'));

      expect(screen.getByRole('heading', { name: 'タグを作成' })).toBeInTheDocument();
    });

    it('creates tag on form submission', async () => {
      mockCreate.mockResolvedValue({ id: 3, name: 'New Tag', color: '#3b82f6' });

      render(<TagsPage />);

      await waitFor(() => {
        expect(screen.getByText('タグを作成')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('タグを作成'));

      const nameInput = screen.getByRole('textbox');
      fireEvent.change(nameInput, { target: { value: 'New Tag' } });

      fireEvent.click(screen.getByText('保存'));

      await waitFor(() => {
        expect(mockCreate).toHaveBeenCalledWith('test-token', {
          name: 'New Tag',
          color: '#10b981',
        });
      });
    });
  });

  describe('edit tag', () => {
    it('shows edit button for users with permission', async () => {
      render(<TagsPage />);

      await waitFor(() => {
        expect(screen.getAllByText('Edit')[0]).toBeInTheDocument();
      });
    });

    it('opens edit modal with tag data', async () => {
      render(<TagsPage />);

      await waitFor(() => {
        expect(screen.getAllByText('Edit')[0]).toBeInTheDocument();
      });

      fireEvent.click(screen.getAllByText('Edit')[0]);

      expect(screen.getByRole('heading', { name: 'タグを編集' })).toBeInTheDocument();
      expect(screen.getByRole('textbox')).toHaveValue('Bug');
    });

    it('updates tag on form submission', async () => {
      mockUpdate.mockResolvedValue({ id: 1, name: 'Updated Bug', color: '#ef4444' });

      render(<TagsPage />);

      await waitFor(() => {
        expect(screen.getAllByText('Edit')[0]).toBeInTheDocument();
      });

      fireEvent.click(screen.getAllByText('Edit')[0]);

      const nameInput = screen.getByRole('textbox');
      fireEvent.change(nameInput, { target: { value: 'Updated Bug' } });

      fireEvent.click(screen.getByText('保存'));

      await waitFor(() => {
        expect(mockUpdate).toHaveBeenCalledWith('test-token', 1, {
          name: 'Updated Bug',
          color: '#ef4444',
        });
      });
    });
  });

  describe('delete tag', () => {
    it('shows delete button for users with permission', async () => {
      render(<TagsPage />);

      await waitFor(() => {
        expect(screen.getAllByText('Delete')[0]).toBeInTheDocument();
      });
    });

    it('deletes tag after confirmation', async () => {
      mockConfirm.mockReturnValue(true);
      mockDelete.mockResolvedValue(undefined);

      render(<TagsPage />);

      await waitFor(() => {
        expect(screen.getAllByText('Delete')[0]).toBeInTheDocument();
      });

      fireEvent.click(screen.getAllByText('Delete')[0]);

      expect(mockConfirm).toHaveBeenCalled();
      await waitFor(() => {
        expect(mockDelete).toHaveBeenCalledWith('test-token', 1);
      });
    });

    it('does not delete when confirmation is cancelled', async () => {
      mockConfirm.mockReturnValue(false);

      render(<TagsPage />);

      await waitFor(() => {
        expect(screen.getAllByText('Delete')[0]).toBeInTheDocument();
      });

      fireEvent.click(screen.getAllByText('Delete')[0]);

      expect(mockDelete).not.toHaveBeenCalled();
    });
  });

  describe('modal behavior', () => {
    it('closes modal when cancel is clicked', async () => {
      render(<TagsPage />);

      await waitFor(() => {
        expect(screen.getByText('タグを作成')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('タグを作成'));
      expect(screen.getByRole('heading', { name: 'タグを作成' })).toBeInTheDocument();

      fireEvent.click(screen.getByText('キャンセル'));

      expect(screen.queryByRole('heading', { name: 'タグを作成' })).not.toBeInTheDocument();
    });
  });

  describe('viewer permissions', () => {
    it('shows read-only message for viewers', async () => {
      mockPermissions.mockReturnValue({
        canManageTags: false,
      });

      render(<TagsPage />);

      await waitFor(() => {
        expect(screen.getAllByText('閲覧のみ')[0]).toBeInTheDocument();
      });
    });
  });
});
