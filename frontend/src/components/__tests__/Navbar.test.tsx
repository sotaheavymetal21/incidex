import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import Navbar from '../Navbar';

// Mock usePathname
const mockPathname = vi.fn();
vi.mock('next/navigation', () => ({
  usePathname: () => mockPathname(),
  useRouter: () => ({
    push: vi.fn(),
  }),
}));

// Mock useAuth
const mockUser = vi.fn();
const mockLogout = vi.fn();
vi.mock('@/context/AuthContext', () => ({
  useAuth: () => ({
    user: mockUser(),
    logout: mockLogout,
  }),
}));

// Mock usePermissions
const mockPermissions = vi.fn();
vi.mock('@/hooks/usePermissions', () => ({
  usePermissions: () => mockPermissions(),
}));

describe('Navbar', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockPathname.mockReturnValue('/dashboard');
    mockUser.mockReturnValue({
      id: 1,
      name: 'Test User',
      email: 'test@example.com',
      role: 'admin',
    });
    mockPermissions.mockReturnValue({
      canViewIncidents: true,
      canViewTags: true,
      canManageUsers: true,
      isAdmin: true,
    });
  });

  describe('visibility', () => {
    it('does not render on login page', () => {
      mockPathname.mockReturnValue('/login');

      const { container } = render(<Navbar />);

      expect(container.firstChild).toBeNull();
    });

    it('does not render on signup page', () => {
      mockPathname.mockReturnValue('/signup');

      const { container } = render(<Navbar />);

      expect(container.firstChild).toBeNull();
    });

    it('does not render when user is not logged in', () => {
      mockUser.mockReturnValue(null);

      const { container } = render(<Navbar />);

      expect(container.firstChild).toBeNull();
    });

    it('renders when user is logged in on dashboard', () => {
      render(<Navbar />);

      expect(screen.getByText('ダッシュボード')).toBeInTheDocument();
    });
  });

  describe('navigation items', () => {
    it('displays all navigation items for admin', () => {
      render(<Navbar />);

      expect(screen.getByText('ダッシュボード')).toBeInTheDocument();
      expect(screen.getByText('インシデント')).toBeInTheDocument();
      expect(screen.getByText('タグ管理')).toBeInTheDocument();
      expect(screen.getByText('ユーザー管理')).toBeInTheDocument();
      expect(screen.getByText('監査ログ')).toBeInTheDocument();
      expect(screen.getByText('レポート')).toBeInTheDocument();
    });

    it('hides user management for non-admin', () => {
      mockUser.mockReturnValue({
        id: 2,
        name: 'Editor User',
        email: 'editor@example.com',
        role: 'editor',
      });
      mockPermissions.mockReturnValue({
        canViewIncidents: true,
        canViewTags: true,
        canManageUsers: false,
        isAdmin: false,
      });

      render(<Navbar />);

      expect(screen.getByText('ダッシュボード')).toBeInTheDocument();
      expect(screen.getByText('インシデント')).toBeInTheDocument();
      expect(screen.getByText('タグ管理')).toBeInTheDocument();
      expect(screen.queryByText('ユーザー管理')).not.toBeInTheDocument();
      expect(screen.queryByText('監査ログ')).not.toBeInTheDocument();
    });

    it('hides incidents for viewer without permission', () => {
      mockPermissions.mockReturnValue({
        canViewIncidents: false,
        canViewTags: false,
        canManageUsers: false,
        isAdmin: false,
      });

      render(<Navbar />);

      expect(screen.getByText('ダッシュボード')).toBeInTheDocument();
      expect(screen.queryByText('インシデント')).not.toBeInTheDocument();
      expect(screen.queryByText('タグ管理')).not.toBeInTheDocument();
    });
  });

  describe('user display', () => {
    it('displays user name', () => {
      render(<Navbar />);

      expect(screen.getByText('Test User')).toBeInTheDocument();
    });

    it('displays user role', () => {
      render(<Navbar />);

      expect(screen.getByText('admin')).toBeInTheDocument();
    });

    it('displays user initial in avatar', () => {
      render(<Navbar />);

      expect(screen.getByText('T')).toBeInTheDocument();
    });
  });

  describe('logout functionality', () => {
    it('calls logout when logout button is clicked', () => {
      render(<Navbar />);

      const logoutButton = screen.getByText('ログアウト');
      fireEvent.click(logoutButton);

      expect(mockLogout).toHaveBeenCalled();
    });
  });

  describe('active state', () => {
    it('applies active styling to current page', () => {
      mockPathname.mockReturnValue('/dashboard');
      render(<Navbar />);

      const dashboardButton = screen.getByText('ダッシュボード').closest('button');
      // Component uses inline styles with gradient background for active state
      expect(dashboardButton).toHaveAttribute('style');
      expect(dashboardButton?.getAttribute('style')).toContain('linear-gradient');
    });

    it('applies active styling when on incidents subpage', () => {
      mockPathname.mockReturnValue('/incidents/123');
      render(<Navbar />);

      const incidentsButton = screen.getByText('インシデント').closest('button');
      // Component uses inline styles for active state
      expect(incidentsButton).toHaveAttribute('style');
      expect(incidentsButton?.getAttribute('style')).toContain('linear-gradient');
    });
  });

  describe('mobile menu', () => {
    it('renders mobile menu button', () => {
      render(<Navbar />);

      // Find the mobile menu toggle button (contains hamburger icon)
      const mobileMenuContainer = document.querySelector('.md\\:hidden');
      expect(mobileMenuContainer).toBeInTheDocument();
    });
  });

  describe('notification settings', () => {
    it('has notification settings button', () => {
      render(<Navbar />);

      // Check for notification settings link/button
      const settingsButton = document.querySelector('button[title="通知設定"]');
      expect(settingsButton).toBeInTheDocument();
    });
  });
});
