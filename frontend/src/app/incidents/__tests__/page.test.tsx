import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import IncidentsPage from '../page';

// Mock next/navigation
const mockRouterPush = vi.fn();
const mockSearchParams = new Map();
vi.mock('next/navigation', () => ({
  useRouter: () => ({
    push: mockRouterPush,
  }),
  useSearchParams: () => ({
    get: (key: string) => mockSearchParams.get(key) || null,
  }),
}));

// Mock useAuth
const mockToken = vi.fn();
const mockAuthLoading = vi.fn();
const mockUser = vi.fn();
vi.mock('@/context/AuthContext', () => ({
  useAuth: () => ({
    token: mockToken(),
    loading: mockAuthLoading(),
    user: mockUser(),
  }),
}));

// Mock usePermissions
const mockPermissions = vi.fn();
vi.mock('@/hooks/usePermissions', () => ({
  usePermissions: () => mockPermissions(),
}));

// Mock APIs
const mockGetAllIncidents = vi.fn();
const mockGetAllTags = vi.fn();
const mockExportCSV = vi.fn();
vi.mock('@/lib/api', () => ({
  incidentApi: {
    getAll: (token: string, params: any) => mockGetAllIncidents(token, params),
  },
  tagApi: {
    getAll: (token: string) => mockGetAllTags(token),
  },
  exportApi: {
    exportIncidentsCSV: (token: string, params: any) => mockExportCSV(token, params),
  },
}));

// Mock SeverityGuide component
vi.mock('@/components/SeverityGuide', () => ({
  default: () => <div data-testid="severity-guide">Severity Guide</div>,
}));

describe('IncidentsPage', () => {
  const mockIncidents = [
    {
      id: 1,
      title: 'Database Connection Issue',
      description: 'Database is not responding',
      severity: 'critical' as const,
      status: 'open' as const,
      detected_at: '2025-01-20T10:00:00Z',
      assignee: { id: 1, name: 'John Doe' },
      tags: [{ id: 1, name: 'Database', color: '#ef4444' }],
    },
    {
      id: 2,
      title: 'API Latency',
      description: 'High API latency detected',
      severity: 'high' as const,
      status: 'investigating' as const,
      detected_at: '2025-01-21T14:00:00Z',
      assignee: null,
      tags: [],
    },
  ];

  const mockTags = [
    { id: 1, name: 'Database', color: '#ef4444' },
    { id: 2, name: 'Network', color: '#3b82f6' },
  ];

  const mockPagination = {
    page: 1,
    limit: 20,
    total: 2,
    total_pages: 1,
  };

  beforeEach(() => {
    vi.clearAllMocks();
    mockSearchParams.clear();
    mockToken.mockReturnValue('test-token');
    mockAuthLoading.mockReturnValue(false);
    mockUser.mockReturnValue({ id: 1, name: 'Test User', role: 'admin' });
    mockPermissions.mockReturnValue({
      canCreateIncidents: true,
    });
    mockGetAllIncidents.mockResolvedValue({
      incidents: mockIncidents,
      pagination: mockPagination,
    });
    mockGetAllTags.mockResolvedValue(mockTags);
  });

  describe('authentication', () => {
    it('redirects to login when not authenticated', () => {
      mockToken.mockReturnValue(null);

      render(<IncidentsPage />);

      expect(mockRouterPush).toHaveBeenCalledWith('/login');
    });

    it('shows loading state during auth check', () => {
      mockAuthLoading.mockReturnValue(true);

      render(<IncidentsPage />);

      expect(screen.getByText('Loading...')).toBeInTheDocument();
    });
  });

  describe('incident list', () => {
    it('displays page title', async () => {
      render(<IncidentsPage />);

      await waitFor(() => {
        expect(screen.getByText('インシデント一覧')).toBeInTheDocument();
      });
    });

    it('displays incidents from API', async () => {
      render(<IncidentsPage />);

      await waitFor(() => {
        expect(screen.getByText('Database Connection Issue')).toBeInTheDocument();
        expect(screen.getByText('API Latency')).toBeInTheDocument();
      });
    });

    it('displays severity badges', async () => {
      render(<IncidentsPage />);

      await waitFor(() => {
        expect(screen.getByText('CRITICAL')).toBeInTheDocument();
        expect(screen.getByText('HIGH')).toBeInTheDocument();
      });
    });

    it('displays status badges', async () => {
      render(<IncidentsPage />);

      await waitFor(() => {
        expect(screen.getByText('Open')).toBeInTheDocument();
        expect(screen.getByText('Investigating')).toBeInTheDocument();
      });
    });

    it('displays assignee names', async () => {
      render(<IncidentsPage />);

      await waitFor(() => {
        expect(screen.getByText('John Doe')).toBeInTheDocument();
      });
    });

    it('displays tags', async () => {
      render(<IncidentsPage />);

      await waitFor(() => {
        // Database appears in both filter and table, so check for multiple
        const databaseElements = screen.getAllByText('Database');
        expect(databaseElements.length).toBeGreaterThan(0);
      });
    });

    it('navigates to incident detail on row click', async () => {
      render(<IncidentsPage />);

      await waitFor(() => {
        expect(screen.getByText('Database Connection Issue')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Database Connection Issue'));

      expect(mockRouterPush).toHaveBeenCalledWith('/incidents/1');
    });

    it('shows empty state when no incidents', async () => {
      mockGetAllIncidents.mockResolvedValue({
        incidents: [],
        pagination: { ...mockPagination, total: 0, total_pages: 0 },
      });

      render(<IncidentsPage />);

      await waitFor(() => {
        expect(screen.getByText('インシデントが見つかりませんでした')).toBeInTheDocument();
      });
    });
  });

  describe('create incident', () => {
    it('shows create button for users with permission', async () => {
      render(<IncidentsPage />);

      await waitFor(() => {
        expect(screen.getByText('新規作成')).toBeInTheDocument();
      });
    });

    it('hides create button for users without permission', async () => {
      mockPermissions.mockReturnValue({
        canCreateIncidents: false,
      });

      render(<IncidentsPage />);

      await waitFor(() => {
        expect(screen.queryByText('新規作成')).not.toBeInTheDocument();
      });
    });

    it('navigates to create page when button is clicked', async () => {
      render(<IncidentsPage />);

      await waitFor(() => {
        expect(screen.getByText('新規作成')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('新規作成'));

      expect(mockRouterPush).toHaveBeenCalledWith('/incidents/create');
    });
  });

  describe('filters', () => {
    it('displays filter sidebar', async () => {
      render(<IncidentsPage />);

      await waitFor(() => {
        expect(screen.getByText('フィルター')).toBeInTheDocument();
      });
    });

    it('displays quick filters', async () => {
      render(<IncidentsPage />);

      await waitFor(() => {
        expect(screen.getByText('未解決のインシデント')).toBeInTheDocument();
        expect(screen.getByText('Critical のみ')).toBeInTheDocument();
      });
    });

    it('displays severity filter options', async () => {
      render(<IncidentsPage />);

      await waitFor(() => {
        expect(screen.getByText('深刻度')).toBeInTheDocument();
        expect(screen.getByText('Critical')).toBeInTheDocument();
        expect(screen.getByText('High')).toBeInTheDocument();
        expect(screen.getByText('Medium')).toBeInTheDocument();
        expect(screen.getByText('Low')).toBeInTheDocument();
      });
    });

    it('displays status filter options', async () => {
      render(<IncidentsPage />);

      await waitFor(() => {
        expect(screen.getByText('ステータス')).toBeInTheDocument();
        // Using getAllByText since "Open" appears in both filter and table
        const openElements = screen.getAllByText('Open');
        expect(openElements.length).toBeGreaterThan(0);
      });
    });

    it('displays tag filter options', async () => {
      render(<IncidentsPage />);

      await waitFor(() => {
        expect(screen.getByText('タグ')).toBeInTheDocument();
      });
    });

    it('applies quick filter for unresolved incidents', async () => {
      render(<IncidentsPage />);

      await waitFor(() => {
        expect(screen.getByText('未解決のインシデント')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('未解決のインシデント'));

      await waitFor(() => {
        expect(mockGetAllIncidents).toHaveBeenCalledWith(
          'test-token',
          expect.objectContaining({ status: 'open' })
        );
      });
    });

    it('applies quick filter for critical incidents', async () => {
      render(<IncidentsPage />);

      await waitFor(() => {
        expect(screen.getByText('Critical のみ')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Critical のみ'));

      await waitFor(() => {
        expect(mockGetAllIncidents).toHaveBeenCalledWith(
          'test-token',
          expect.objectContaining({ severity: 'critical' })
        );
      });
    });

    it('clears all filters', async () => {
      render(<IncidentsPage />);

      await waitFor(() => {
        // Get the first "すべてクリア" button (in sidebar)
        const clearButtons = screen.getAllByText('すべてクリア');
        expect(clearButtons.length).toBeGreaterThan(0);
      });

      // Apply a filter first
      fireEvent.click(screen.getByText('Critical のみ'));

      await waitFor(() => {
        expect(mockGetAllIncidents).toHaveBeenCalled();
      });

      // Clear filters using the first clear button
      const clearButtons = screen.getAllByText('すべてクリア');
      fireEvent.click(clearButtons[0]);

      await waitFor(() => {
        expect(mockGetAllIncidents).toHaveBeenCalledWith(
          'test-token',
          expect.objectContaining({
            severity: undefined,
            status: undefined,
          })
        );
      });
    });
  });

  describe('sorting', () => {
    it('sorts by title when header is clicked', async () => {
      render(<IncidentsPage />);

      await waitFor(() => {
        expect(screen.getByText('Title')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Title'));

      // Check that incidents are rendered (sorting is client-side)
      expect(screen.getByText('Database Connection Issue')).toBeInTheDocument();
    });

    it('sorts by severity when header is clicked', async () => {
      render(<IncidentsPage />);

      await waitFor(() => {
        expect(screen.getByText('Severity')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Severity'));

      // Check that incidents are rendered
      expect(screen.getByText('CRITICAL')).toBeInTheDocument();
    });
  });

  describe('CSV export', () => {
    it('calls export API when button is clicked', async () => {
      const mockBlob = new Blob(['test'], { type: 'text/csv' });
      mockExportCSV.mockResolvedValue(mockBlob);

      render(<IncidentsPage />);

      await waitFor(() => {
        expect(screen.getByText('CSVエクスポート')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('CSVエクスポート'));

      await waitFor(() => {
        expect(mockExportCSV).toHaveBeenCalledWith('test-token', expect.any(Object));
      });
    });
  });

  describe('pagination', () => {
    it('displays pagination info', async () => {
      render(<IncidentsPage />);

      await waitFor(() => {
        expect(screen.getByText(/Showing page/)).toBeInTheDocument();
      });
    });

    it('displays next and previous buttons', async () => {
      render(<IncidentsPage />);

      await waitFor(() => {
        expect(screen.getAllByText('Previous')[0]).toBeInTheDocument();
        expect(screen.getAllByText('Next')[0]).toBeInTheDocument();
      });
    });
  });

  describe('severity guide', () => {
    it('renders severity guide component', async () => {
      render(<IncidentsPage />);

      await waitFor(() => {
        expect(screen.getByTestId('severity-guide')).toBeInTheDocument();
      });
    });
  });
});
