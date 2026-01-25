import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import { server } from '@/test/mocks/server';
import DashboardPage from '../page';
import { testUsers, testIncidents } from '@/test/fixtures';
import { DashboardStats, TagStats } from '@/types/stats';

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
const mockUseAuth = vi.fn();
vi.mock('@/context/AuthContext', () => ({
  useAuth: () => mockUseAuth(),
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

// recharts をモック（テスト環境でのレンダリングエラーを防ぐ）
vi.mock('recharts', () => ({
  PieChart: ({ children }: any) => <div data-testid="pie-chart">{children}</div>,
  Pie: () => <div data-testid="pie" />,
  Cell: () => <div data-testid="cell" />,
  LineChart: ({ children }: any) => <div data-testid="line-chart">{children}</div>,
  Line: () => <div data-testid="line" />,
  BarChart: ({ children }: any) => <div data-testid="bar-chart">{children}</div>,
  Bar: () => <div data-testid="bar" />,
  XAxis: () => <div data-testid="x-axis" />,
  YAxis: () => <div data-testid="y-axis" />,
  CartesianGrid: () => <div data-testid="cartesian-grid" />,
  Tooltip: () => <div data-testid="tooltip" />,
  Legend: () => <div data-testid="legend" />,
  ResponsiveContainer: ({ children }: any) => <div data-testid="responsive-container">{children}</div>,
}));

describe('DashboardPage', () => {
  const mockDashboardStats: DashboardStats = {
    total_incidents: 42,
    by_severity: {
      critical: 5,
      high: 12,
      medium: 18,
      low: 7,
    },
    by_status: {
      open: 10,
      investigating: 8,
      resolved: 20,
      closed: 4,
    },
    recent_incidents: [
      testIncidents.critical,
      testIncidents.high,
      testIncidents.resolved,
    ],
    trend_data: [
      { date: '2025-01-20', count: 5 },
      { date: '2025-01-21', count: 8 },
      { date: '2025-01-22', count: 6 },
      { date: '2025-01-23', count: 10 },
      { date: '2025-01-24', count: 7 },
    ],
  };

  const mockTagStats: TagStats[] = [
    {
      tag_id: 1,
      tag_name: 'Database',
      tag_color: '#ef4444',
      count: 15,
      percentage: 35.7,
    },
    {
      tag_id: 2,
      tag_name: 'API',
      tag_color: '#3b82f6',
      count: 12,
      percentage: 28.6,
    },
    {
      tag_id: 3,
      tag_name: 'Security',
      tag_color: '#f59e0b',
      count: 8,
      percentage: 19.0,
    },
  ];

  beforeEach(() => {
    mockPush.mockClear();
    mockUseAuth.mockReturnValue({
      user: testUsers.viewer,
      token: 'test-token',
      loading: false,
      login: vi.fn(),
      logout: vi.fn(),
    });

    // デフォルトのAPI応答をセットアップ
    server.use(
      http.get(`${API_URL}/stats/dashboard`, () => {
        return HttpResponse.json(mockDashboardStats);
      }),
      http.get(`${API_URL}/stats/tags`, () => {
        return HttpResponse.json({ tag_stats: mockTagStats });
      })
    );
  });

  describe('rendering', () => {
    it('shows loading state initially', async () => {
      render(<DashboardPage />);

      expect(screen.getByText('読み込み中...')).toBeInTheDocument();

      await waitFor(() => {
        expect(screen.queryByText('読み込み中...')).not.toBeInTheDocument();
      });
    });

    it('renders dashboard stats after loading', async () => {
      render(<DashboardPage />);

      await waitFor(() => {
        expect(screen.getByText('総インシデント数')).toBeInTheDocument();
      });

      expect(screen.getByText('42')).toBeInTheDocument();
    });

    it('shows error message on API failure', async () => {
      server.use(
        http.get(`${API_URL}/stats/dashboard`, () => {
          return HttpResponse.json(
            { error: 'Failed to fetch stats' },
            { status: 500 }
          );
        })
      );

      render(<DashboardPage />);

      await waitFor(() => {
        expect(screen.getByText(/Failed to fetch stats|データの取得に失敗しました/)).toBeInTheDocument();
      });
    });

    it('redirects to login when not authenticated', async () => {
      mockUseAuth.mockReturnValue({
        user: null,
        token: null,
        loading: false,
        login: vi.fn(),
        logout: vi.fn(),
      });

      render(<DashboardPage />);

      await waitFor(() => {
        expect(mockPush).toHaveBeenCalledWith('/login');
      });
    });

    it('waits for auth loading before redirecting', () => {
      mockUseAuth.mockReturnValue({
        user: null,
        token: null,
        loading: true,
        login: vi.fn(),
        logout: vi.fn(),
      });

      render(<DashboardPage />);

      expect(mockPush).not.toHaveBeenCalled();
    });
  });

  describe('stats display', () => {
    it('shows total incidents count', async () => {
      render(<DashboardPage />);

      await waitFor(() => {
        expect(screen.getByText('総インシデント数')).toBeInTheDocument();
      });

      expect(screen.getByText('42')).toBeInTheDocument();
    });

    it('shows critical incidents count', async () => {
      render(<DashboardPage />);

      await waitFor(() => {
        expect(screen.getByText('Critical')).toBeInTheDocument();
      });

      expect(screen.getByText('5')).toBeInTheDocument();
    });

    it('shows open incidents count', async () => {
      render(<DashboardPage />);

      await waitFor(() => {
        expect(screen.getAllByText('Open（未対応）').length).toBeGreaterThan(0);
      });

      expect(screen.getByText('10')).toBeInTheDocument();
    });

    it('shows resolved incidents count', async () => {
      render(<DashboardPage />);

      await waitFor(() => {
        expect(screen.getAllByText('Resolved（解決済み）').length).toBeGreaterThan(0);
      });

      expect(screen.getByText('20')).toBeInTheDocument();
    });

    it('handles zero values correctly', async () => {
      server.use(
        http.get(`${API_URL}/stats/dashboard`, () => {
          return HttpResponse.json({
            ...mockDashboardStats,
            by_severity: {
              critical: 0,
              high: 0,
              medium: 0,
              low: 0,
            },
          });
        })
      );

      render(<DashboardPage />);

      await waitFor(() => {
        expect(screen.getByText('Critical')).toBeInTheDocument();
      });

      // critical の 0 を確認（複数ある場合は最初のもの）
      const zeroElements = screen.getAllByText('0');
      expect(zeroElements.length).toBeGreaterThan(0);
    });
  });

  describe('charts and visualizations', () => {
    it('renders severity distribution pie chart', async () => {
      render(<DashboardPage />);

      await waitFor(() => {
        expect(screen.getByText('重要度別分布')).toBeInTheDocument();
      });

      expect(screen.getAllByTestId('pie-chart').length).toBeGreaterThan(0);
    });

    it('renders status distribution pie chart', async () => {
      render(<DashboardPage />);

      await waitFor(() => {
        expect(screen.getByText('ステータス別分布')).toBeInTheDocument();
      });

      expect(screen.getAllByTestId('pie-chart').length).toBeGreaterThan(0);
    });

    it('renders trend chart', async () => {
      render(<DashboardPage />);

      await waitFor(() => {
        expect(screen.getByText('インシデント発生トレンド')).toBeInTheDocument();
      });

      expect(screen.getAllByTestId('line-chart').length).toBeGreaterThan(0);
    });

    it('allows switching graph types', async () => {
      render(<DashboardPage />);

      await waitFor(() => {
        expect(screen.getByText('分布グラフ')).toBeInTheDocument();
      });

      // 時系列グラフに切り替え
      fireEvent.click(screen.getByText('時系列グラフ'));

      await waitFor(() => {
        expect(screen.getByText('インシデント発生推移')).toBeInTheDocument();
      });

      // 棒グラフに切り替え
      fireEvent.click(screen.getByText('棒グラフ'));

      await waitFor(() => {
        expect(screen.getByText('重要度別件数')).toBeInTheDocument();
      });
    });

    it('allows switching trend periods', async () => {
      render(<DashboardPage />);

      await waitFor(() => {
        expect(screen.getByText('インシデント発生トレンド')).toBeInTheDocument();
      });

      // 期間セレクターボタンが存在することを確認
      const dailyButtons = screen.getAllByText('日次');
      const weeklyButtons = screen.getAllByText('週次');
      const monthlyButtons = screen.getAllByText('月次');

      expect(dailyButtons.length).toBeGreaterThan(0);
      expect(weeklyButtons.length).toBeGreaterThan(0);
      expect(monthlyButtons.length).toBeGreaterThan(0);

      // 週次に切り替え
      fireEvent.click(weeklyButtons[0]);

      // 月次に切り替え
      fireEvent.click(monthlyButtons[0]);

      // 切り替えが完了したことを確認（月次ボタンが存在し続ける）
      await waitFor(() => {
        expect(screen.getAllByText('月次').length).toBeGreaterThan(0);
      });
    });
  });

  describe('recent incidents section', () => {
    it('shows recent incidents list', async () => {
      render(<DashboardPage />);

      await waitFor(() => {
        expect(screen.getByText('最近のインシデント')).toBeInTheDocument();
      });

      expect(screen.getByText('Critical System Outage')).toBeInTheDocument();
      expect(screen.getByText('API Performance Degradation')).toBeInTheDocument();
      expect(screen.getByText('Login Issue Resolved')).toBeInTheDocument();
    });

    it('handles empty incidents list', async () => {
      server.use(
        http.get(`${API_URL}/stats/dashboard`, () => {
          return HttpResponse.json({
            ...mockDashboardStats,
            recent_incidents: [],
          });
        })
      );

      render(<DashboardPage />);

      await waitFor(() => {
        expect(screen.getByText('最近のインシデント')).toBeInTheDocument();
      });

      expect(screen.getByText('インシデントがありません')).toBeInTheDocument();
    });

    it('links to incident detail pages', async () => {
      render(<DashboardPage />);

      await waitFor(() => {
        expect(screen.getByText('Critical System Outage')).toBeInTheDocument();
      });

      const incidentRow = screen.getByText('Critical System Outage').closest('tr');
      fireEvent.click(incidentRow!);

      expect(mockPush).toHaveBeenCalledWith('/incidents/1');
    });

    it('displays incident severity badges', async () => {
      render(<DashboardPage />);

      await waitFor(() => {
        expect(screen.getByText('Critical System Outage')).toBeInTheDocument();
      });

      expect(screen.getByText('Critical（致命的）')).toBeInTheDocument();
      expect(screen.getByText('High（高）')).toBeInTheDocument();
      expect(screen.getByText('Medium（中）')).toBeInTheDocument();
    });

    it('displays incident status badges', async () => {
      render(<DashboardPage />);

      await waitFor(() => {
        expect(screen.getByText('Critical System Outage')).toBeInTheDocument();
      });

      expect(screen.getByText('Investigating（調査中）')).toBeInTheDocument();
      expect(screen.getAllByText('Open（未対応）').length).toBeGreaterThan(0);
      expect(screen.getAllByText('Resolved（解決済み）').length).toBeGreaterThan(0);
    });

    it('formats detection timestamps correctly', async () => {
      render(<DashboardPage />);

      await waitFor(() => {
        expect(screen.getByText('Critical System Outage')).toBeInTheDocument();
      });

      // 日本語ロケールでフォーマットされた日時が表示されることを確認
      const timestamps = screen.getAllByText(/2025/);
      expect(timestamps.length).toBeGreaterThan(0);
    });
  });

  describe('tag statistics section', () => {
    it('shows tag statistics when available', async () => {
      render(<DashboardPage />);

      await waitFor(() => {
        expect(screen.getByText('タグ別インシデント統計')).toBeInTheDocument();
      });

      expect(screen.getByText('Database')).toBeInTheDocument();
      expect(screen.getByText('API')).toBeInTheDocument();
      expect(screen.getByText('Security')).toBeInTheDocument();
    });

    it('displays tag counts', async () => {
      render(<DashboardPage />);

      await waitFor(() => {
        expect(screen.getByText('タグ別インシデント統計')).toBeInTheDocument();
      });

      expect(screen.getByText('15')).toBeInTheDocument();
      expect(screen.getByText('12')).toBeInTheDocument();
      expect(screen.getByText('8')).toBeInTheDocument();
    });

    it('displays tag percentages', async () => {
      render(<DashboardPage />);

      await waitFor(() => {
        expect(screen.getByText('タグ別インシデント統計')).toBeInTheDocument();
      });

      expect(screen.getByText('35.7%')).toBeInTheDocument();
      expect(screen.getByText('28.6%')).toBeInTheDocument();
      expect(screen.getByText('19.0%')).toBeInTheDocument();
    });

    it('hides tag section when no tags', async () => {
      server.use(
        http.get(`${API_URL}/stats/tags`, () => {
          return HttpResponse.json({ tag_stats: [] });
        })
      );

      render(<DashboardPage />);

      await waitFor(() => {
        expect(screen.getByText('総インシデント数')).toBeInTheDocument();
      });

      expect(screen.queryByText('タグ別インシデント統計')).not.toBeInTheDocument();
    });
  });

  describe('navigation and interaction', () => {
    it('navigates to incidents page when clicking total incidents card', async () => {
      render(<DashboardPage />);

      await waitFor(() => {
        expect(screen.getByText('総インシデント数')).toBeInTheDocument();
      });

      const totalCard = screen.getByText('総インシデント数').closest('div[class*="cursor-pointer"]');
      fireEvent.click(totalCard!);

      expect(mockPush).toHaveBeenCalledWith('/incidents');
    });

    it('navigates to filtered critical incidents', async () => {
      render(<DashboardPage />);

      await waitFor(() => {
        expect(screen.getByText('Critical')).toBeInTheDocument();
      });

      const criticalCard = screen.getByText('Critical').closest('div[class*="cursor-pointer"]');
      fireEvent.click(criticalCard!);

      expect(mockPush).toHaveBeenCalledWith('/incidents?severity=critical');
    });

    it('navigates to filtered open incidents', async () => {
      render(<DashboardPage />);

      await waitFor(() => {
        expect(screen.getAllByText('Open（未対応）').length).toBeGreaterThan(0);
      });

      const openElements = screen.getAllByText('Open（未対応）');
      const openCard = openElements[0].closest('div[class*="cursor-pointer"]');
      fireEvent.click(openCard!);

      expect(mockPush).toHaveBeenCalledWith('/incidents?status=open');
    });

    it('navigates to filtered resolved incidents', async () => {
      render(<DashboardPage />);

      await waitFor(() => {
        expect(screen.getAllByText('Resolved（解決済み）').length).toBeGreaterThan(0);
      });

      const resolvedElements = screen.getAllByText('Resolved（解決済み）');
      const resolvedCard = resolvedElements[0].closest('div[class*="cursor-pointer"]');
      fireEvent.click(resolvedCard!);

      expect(mockPush).toHaveBeenCalledWith('/incidents?status=resolved');
    });
  });

  describe('data refresh on period change', () => {
    it('refetches data when period changes', async () => {
      let requestCount = 0;
      server.use(
        http.get(`${API_URL}/stats/dashboard`, () => {
          requestCount++;
          return HttpResponse.json(mockDashboardStats);
        })
      );

      render(<DashboardPage />);

      await waitFor(() => {
        expect(screen.getByText('日次')).toBeInTheDocument();
      });

      const initialCount = requestCount;

      // 週次に切り替え
      const weeklyButtons = screen.getAllByText('週次');
      fireEvent.click(weeklyButtons[0]);

      await waitFor(() => {
        expect(requestCount).toBeGreaterThan(initialCount);
      });
    });
  });

  describe('error handling', () => {
    it('handles network error gracefully', async () => {
      server.use(
        http.get(`${API_URL}/stats/dashboard`, () => {
          return HttpResponse.error();
        })
      );

      render(<DashboardPage />);

      await waitFor(() => {
        expect(screen.getByText(/データの取得に失敗しました|バックエンドサーバーに接続できませんでした/)).toBeInTheDocument();
      });
    });

    it('handles tag stats API failure independently', async () => {
      server.use(
        http.get(`${API_URL}/stats/tags`, () => {
          return HttpResponse.json(
            { error: 'Failed to fetch tag stats' },
            { status: 500 }
          );
        })
      );

      render(<DashboardPage />);

      await waitFor(() => {
        expect(screen.getByText('Failed to fetch tag stats')).toBeInTheDocument();
      });
    });

    it('shows general error for unexpected API response', async () => {
      server.use(
        http.get(`${API_URL}/stats/dashboard`, () => {
          return HttpResponse.json(null);
        })
      );

      render(<DashboardPage />);

      await waitFor(() => {
        // null を受け取った場合でもエラーが表示されるか、適切に処理される
        const errorOrLoading = screen.queryByText('読み込み中...') || screen.queryByText(/データの取得に失敗しました/);
        expect(errorOrLoading).toBeTruthy();
      });
    });
  });

  describe('permission-based rendering', () => {
    it('shows dashboard for admin users', async () => {
      mockUseAuth.mockReturnValue({
        user: testUsers.admin,
        token: 'admin-token',
        loading: false,
        login: vi.fn(),
        logout: vi.fn(),
      });

      render(<DashboardPage />);

      await waitFor(() => {
        expect(screen.getByText('総インシデント数')).toBeInTheDocument();
      });
    });

    it('shows dashboard for editor users', async () => {
      mockUseAuth.mockReturnValue({
        user: testUsers.editor,
        token: 'editor-token',
        loading: false,
        login: vi.fn(),
        logout: vi.fn(),
      });

      render(<DashboardPage />);

      await waitFor(() => {
        expect(screen.getByText('総インシデント数')).toBeInTheDocument();
      });
    });

    it('shows dashboard for viewer users', async () => {
      mockUseAuth.mockReturnValue({
        user: testUsers.viewer,
        token: 'viewer-token',
        loading: false,
        login: vi.fn(),
        logout: vi.fn(),
      });

      render(<DashboardPage />);

      await waitFor(() => {
        expect(screen.getByText('総インシデント数')).toBeInTheDocument();
      });
    });
  });
});
