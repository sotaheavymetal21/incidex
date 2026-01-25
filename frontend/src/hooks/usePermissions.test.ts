import { describe, it, expect, beforeEach, vi } from 'vitest';
import { renderHook } from '@testing-library/react';
import React from 'react';
import { usePermissions } from './usePermissions';
import { testUsers } from '@/test/fixtures';

// AuthContext をモック
const mockUseAuth = vi.fn();
vi.mock('@/context/AuthContext', () => ({
  useAuth: () => mockUseAuth(),
}));

describe('usePermissions', () => {
  beforeEach(() => {
    mockUseAuth.mockClear();
  });

  describe('when user is not logged in (null)', () => {
    beforeEach(() => {
      mockUseAuth.mockReturnValue({ user: null });
    });

    it('returns false for all permission checks', () => {
      const { result } = renderHook(() => usePermissions());

      // すべての権限チェックがfalseを返すことを確認
      expect(result.current.canViewIncidents).toBe(false);
      expect(result.current.canCreateIncidents).toBe(false);
      expect(result.current.canEditIncidents).toBe(false);
      expect(result.current.canDeleteIncidents).toBe(false);
      expect(result.current.canViewTags).toBe(false);
      expect(result.current.canManageTags).toBe(false);
      expect(result.current.canViewPostMortems).toBe(false);
      expect(result.current.canManagePostMortems).toBe(false);
      expect(result.current.canViewUsers).toBe(false);
      expect(result.current.canManageUsers).toBe(false);
      expect(result.current.canViewStats).toBe(false);
      expect(result.current.canExportData).toBe(false);
    });

    it('returns false for all role checks', () => {
      const { result } = renderHook(() => usePermissions());

      expect(result.current.isAdmin).toBe(false);
      expect(result.current.isEditorOrAdmin).toBe(false);
      expect(result.current.isViewer).toBe(false);
      expect(result.current.canEdit).toBe(false);
    });

    it('can() returns false for any permission', () => {
      const { result } = renderHook(() => usePermissions());

      expect(result.current.can('view_incidents')).toBe(false);
      expect(result.current.can('manage_users')).toBe(false);
    });

    it('canAny() returns false for any permissions', () => {
      const { result } = renderHook(() => usePermissions());

      expect(result.current.canAny(['view_incidents', 'create_incidents'])).toBe(false);
    });

    it('canAll() returns false for any permissions', () => {
      const { result } = renderHook(() => usePermissions());

      expect(result.current.canAll(['view_incidents', 'view_stats'])).toBe(false);
    });
  });

  describe('when user is admin', () => {
    beforeEach(() => {
      mockUseAuth.mockReturnValue({ user: testUsers.admin });
    });

    it('returns true for all incident permissions', () => {
      const { result } = renderHook(() => usePermissions());

      expect(result.current.canViewIncidents).toBe(true);
      expect(result.current.canCreateIncidents).toBe(true);
      expect(result.current.canEditIncidents).toBe(true);
      expect(result.current.canDeleteIncidents).toBe(true);
    });

    it('returns true for all tag permissions', () => {
      const { result } = renderHook(() => usePermissions());

      expect(result.current.canViewTags).toBe(true);
      expect(result.current.canManageTags).toBe(true);
    });

    it('returns true for all postmortem permissions', () => {
      const { result } = renderHook(() => usePermissions());

      expect(result.current.canViewPostMortems).toBe(true);
      expect(result.current.canManagePostMortems).toBe(true);
    });

    it('returns true for all user permissions', () => {
      const { result } = renderHook(() => usePermissions());

      expect(result.current.canViewUsers).toBe(true);
      expect(result.current.canManageUsers).toBe(true);
    });

    it('returns true for stats and export permissions', () => {
      const { result } = renderHook(() => usePermissions());

      expect(result.current.canViewStats).toBe(true);
      expect(result.current.canExportData).toBe(true);
    });

    it('returns correct role checks', () => {
      const { result } = renderHook(() => usePermissions());

      expect(result.current.isAdmin).toBe(true);
      expect(result.current.isEditorOrAdmin).toBe(true);
      expect(result.current.isViewer).toBe(false);
      expect(result.current.canEdit).toBe(true);
    });

    it('can() returns correct values', () => {
      const { result } = renderHook(() => usePermissions());

      expect(result.current.can('view_incidents')).toBe(true);
      expect(result.current.can('manage_users')).toBe(true);
      expect(result.current.can('export_data')).toBe(true);
    });

    it('canAny() returns true if any permission matches', () => {
      const { result } = renderHook(() => usePermissions());

      expect(result.current.canAny(['manage_users', 'delete_incidents'])).toBe(true);
    });

    it('canAll() returns true if all permissions are present', () => {
      const { result } = renderHook(() => usePermissions());

      expect(result.current.canAll(['view_incidents', 'create_incidents', 'manage_users'])).toBe(true);
    });
  });

  describe('when user is editor', () => {
    beforeEach(() => {
      mockUseAuth.mockReturnValue({ user: testUsers.editor });
    });

    it('returns true for all incident permissions', () => {
      const { result } = renderHook(() => usePermissions());

      expect(result.current.canViewIncidents).toBe(true);
      expect(result.current.canCreateIncidents).toBe(true);
      expect(result.current.canEditIncidents).toBe(true);
      expect(result.current.canDeleteIncidents).toBe(true);
    });

    it('returns true for all tag permissions', () => {
      const { result } = renderHook(() => usePermissions());

      expect(result.current.canViewTags).toBe(true);
      expect(result.current.canManageTags).toBe(true);
    });

    it('returns true for all postmortem permissions', () => {
      const { result } = renderHook(() => usePermissions());

      expect(result.current.canViewPostMortems).toBe(true);
      expect(result.current.canManagePostMortems).toBe(true);
    });

    it('returns false for user management', () => {
      const { result } = renderHook(() => usePermissions());

      expect(result.current.canViewUsers).toBe(false);
      expect(result.current.canManageUsers).toBe(false);
    });

    it('returns true for stats and export permissions', () => {
      const { result } = renderHook(() => usePermissions());

      expect(result.current.canViewStats).toBe(true);
      expect(result.current.canExportData).toBe(true);
    });

    it('returns correct role checks', () => {
      const { result } = renderHook(() => usePermissions());

      expect(result.current.isAdmin).toBe(false);
      expect(result.current.isEditorOrAdmin).toBe(true);
      expect(result.current.isViewer).toBe(false);
      expect(result.current.canEdit).toBe(true);
    });

    it('can() returns correct values for editor permissions', () => {
      const { result } = renderHook(() => usePermissions());

      expect(result.current.can('view_incidents')).toBe(true);
      expect(result.current.can('manage_users')).toBe(false);
      expect(result.current.can('export_data')).toBe(true);
    });

    it('canAll() returns false when admin-only permission is required', () => {
      const { result } = renderHook(() => usePermissions());

      // editor は manage_users を持っていない
      expect(result.current.canAll(['view_incidents', 'manage_users'])).toBe(false);
    });

    it('canAny() returns true for mixed permissions if any match', () => {
      const { result } = renderHook(() => usePermissions());

      // manage_users は持っていないが、view_incidents は持っている
      expect(result.current.canAny(['manage_users', 'view_incidents'])).toBe(true);
    });
  });

  describe('when user is viewer', () => {
    beforeEach(() => {
      mockUseAuth.mockReturnValue({ user: testUsers.viewer });
    });

    it('returns true only for view incidents permission', () => {
      const { result } = renderHook(() => usePermissions());

      expect(result.current.canViewIncidents).toBe(true);
      expect(result.current.canCreateIncidents).toBe(false);
      expect(result.current.canEditIncidents).toBe(false);
      expect(result.current.canDeleteIncidents).toBe(false);
    });

    it('returns true only for view tags permission', () => {
      const { result } = renderHook(() => usePermissions());

      expect(result.current.canViewTags).toBe(true);
      expect(result.current.canManageTags).toBe(false);
    });

    it('returns true only for view postmortems permission', () => {
      const { result } = renderHook(() => usePermissions());

      expect(result.current.canViewPostMortems).toBe(true);
      expect(result.current.canManagePostMortems).toBe(false);
    });

    it('returns false for all user permissions', () => {
      const { result } = renderHook(() => usePermissions());

      expect(result.current.canViewUsers).toBe(false);
      expect(result.current.canManageUsers).toBe(false);
    });

    it('returns true for view stats but false for export', () => {
      const { result } = renderHook(() => usePermissions());

      expect(result.current.canViewStats).toBe(true);
      expect(result.current.canExportData).toBe(false);
    });

    it('returns correct role checks', () => {
      const { result } = renderHook(() => usePermissions());

      expect(result.current.isAdmin).toBe(false);
      expect(result.current.isEditorOrAdmin).toBe(false);
      expect(result.current.isViewer).toBe(true);
      expect(result.current.canEdit).toBe(false);
    });

    it('can() returns correct values for viewer permissions', () => {
      const { result } = renderHook(() => usePermissions());

      expect(result.current.can('view_incidents')).toBe(true);
      expect(result.current.can('create_incidents')).toBe(false);
      expect(result.current.can('manage_users')).toBe(false);
    });

    it('canAny() returns true for view permissions only', () => {
      const { result } = renderHook(() => usePermissions());

      expect(result.current.canAny(['view_incidents', 'view_stats'])).toBe(true);
      expect(result.current.canAny(['create_incidents', 'delete_incidents'])).toBe(false);
    });

    it('canAll() returns true only for view permissions', () => {
      const { result } = renderHook(() => usePermissions());

      expect(result.current.canAll(['view_incidents', 'view_stats', 'view_tags'])).toBe(true);
      expect(result.current.canAll(['view_incidents', 'create_incidents'])).toBe(false);
    });
  });

  describe('edge cases', () => {
    it('handles empty permission array for canAny()', () => {
      mockUseAuth.mockReturnValue({ user: testUsers.admin });
      const { result } = renderHook(() => usePermissions());

      // 空配列の場合、some()は常にfalseを返す
      expect(result.current.canAny([])).toBe(false);
    });

    it('handles empty permission array for canAll()', () => {
      mockUseAuth.mockReturnValue({ user: testUsers.admin });
      const { result } = renderHook(() => usePermissions());

      // 空配列の場合、every()は常にtrueを返す
      expect(result.current.canAll([])).toBe(true);
    });

    it('handles single permission array for canAny()', () => {
      mockUseAuth.mockReturnValue({ user: testUsers.viewer });
      const { result } = renderHook(() => usePermissions());

      expect(result.current.canAny(['view_incidents'])).toBe(true);
      expect(result.current.canAny(['manage_users'])).toBe(false);
    });

    it('handles single permission array for canAll()', () => {
      mockUseAuth.mockReturnValue({ user: testUsers.viewer });
      const { result } = renderHook(() => usePermissions());

      expect(result.current.canAll(['view_incidents'])).toBe(true);
      expect(result.current.canAll(['manage_users'])).toBe(false);
    });
  });
});
