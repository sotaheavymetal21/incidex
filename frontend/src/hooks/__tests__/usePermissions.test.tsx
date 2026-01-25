import { describe, it, expect, beforeEach, vi } from 'vitest';
import { renderHook } from '@testing-library/react';
import { ReactNode } from 'react';
import { AuthProvider } from '@/context/AuthContext';
import { usePermissions } from '../usePermissions';
import { testUsers } from '@/test/fixtures';

// Next.js router をモック
vi.mock('next/navigation', () => ({
  useRouter: () => ({
    push: vi.fn(),
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

/**
 * AuthProvider ラッパー
 * テストで認証状態を設定するために使用します
 */
function createWrapper(
  options: { token?: string; user?: typeof testUsers.admin | null } = {}
) {
  return function Wrapper({ children }: { children: ReactNode }) {
    // localStorage を設定
    if (options.token && options.user) {
      localStorage.setItem('token', options.token);
      localStorage.setItem('user', JSON.stringify(options.user));
    }

    return <AuthProvider>{children}</AuthProvider>;
  };
}

describe('usePermissions', () => {
  beforeEach(() => {
    localStorage.clear();
  });

  describe('Hook initialization', () => {
    it('returns false for all checks when no user', async () => {
      const { result } = renderHook(() => usePermissions(), {
        wrapper: createWrapper(),
      });

      // 汎用的なチェック関数
      expect(result.current.can('view_incidents')).toBe(false);
      expect(result.current.canAny(['view_incidents', 'create_incidents'])).toBe(false);
      expect(result.current.canAll(['view_incidents', 'create_incidents'])).toBe(false);

      // 具体的な権限チェック
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

      // ロールベースのチェック
      expect(result.current.isAdmin).toBe(false);
      expect(result.current.isEditorOrAdmin).toBe(false);
      expect(result.current.isViewer).toBe(false);
      expect(result.current.canEdit).toBe(false);
    });

    it('returns correct values when user is present', async () => {
      const { result } = renderHook(() => usePermissions(), {
        wrapper: createWrapper({ token: 'test-token', user: testUsers.admin }),
      });

      // Admin ユーザーの場合、すべての権限が true
      expect(result.current.canViewIncidents).toBe(true);
      expect(result.current.canCreateIncidents).toBe(true);
      expect(result.current.isAdmin).toBe(true);
    });
  });

  describe('Permission checks with different roles', () => {
    describe('Admin user', () => {
      it('has all permissions', async () => {
        const { result } = renderHook(() => usePermissions(), {
          wrapper: createWrapper({ token: 'admin-token', user: testUsers.admin }),
        });

        // インシデント権限
        expect(result.current.canViewIncidents).toBe(true);
        expect(result.current.canCreateIncidents).toBe(true);
        expect(result.current.canEditIncidents).toBe(true);
        expect(result.current.canDeleteIncidents).toBe(true);

        // タグ権限
        expect(result.current.canViewTags).toBe(true);
        expect(result.current.canManageTags).toBe(true);

        // ポストモーテム権限
        expect(result.current.canViewPostMortems).toBe(true);
        expect(result.current.canManagePostMortems).toBe(true);

        // ユーザー権限
        expect(result.current.canViewUsers).toBe(true);
        expect(result.current.canManageUsers).toBe(true);

        // その他の権限
        expect(result.current.canViewStats).toBe(true);
        expect(result.current.canExportData).toBe(true);
      });

      it('role helpers return correct values', async () => {
        const { result } = renderHook(() => usePermissions(), {
          wrapper: createWrapper({ token: 'admin-token', user: testUsers.admin }),
        });

        expect(result.current.isAdmin).toBe(true);
        expect(result.current.isEditorOrAdmin).toBe(true);
        expect(result.current.isViewer).toBe(false);
        expect(result.current.canEdit).toBe(true);
      });
    });

    describe('Editor user', () => {
      it('has limited permissions', async () => {
        const { result } = renderHook(() => usePermissions(), {
          wrapper: createWrapper({ token: 'editor-token', user: testUsers.editor }),
        });

        // コンテンツ管理権限は持つ
        expect(result.current.canViewIncidents).toBe(true);
        expect(result.current.canCreateIncidents).toBe(true);
        expect(result.current.canEditIncidents).toBe(true);
        expect(result.current.canDeleteIncidents).toBe(true);
        expect(result.current.canManageTags).toBe(true);
        expect(result.current.canManagePostMortems).toBe(true);
        expect(result.current.canViewStats).toBe(true);
        expect(result.current.canExportData).toBe(true);

        // ユーザー管理権限は持たない
        expect(result.current.canViewUsers).toBe(false);
        expect(result.current.canManageUsers).toBe(false);
      });

      it('role helpers return correct values', async () => {
        const { result } = renderHook(() => usePermissions(), {
          wrapper: createWrapper({ token: 'editor-token', user: testUsers.editor }),
        });

        expect(result.current.isAdmin).toBe(false);
        expect(result.current.isEditorOrAdmin).toBe(true);
        expect(result.current.isViewer).toBe(false);
        expect(result.current.canEdit).toBe(true);
      });
    });

    describe('Viewer user', () => {
      it('has read-only permissions', async () => {
        const { result } = renderHook(() => usePermissions(), {
          wrapper: createWrapper({ token: 'viewer-token', user: testUsers.viewer }),
        });

        // 閲覧権限のみ
        expect(result.current.canViewIncidents).toBe(true);
        expect(result.current.canViewTags).toBe(true);
        expect(result.current.canViewPostMortems).toBe(true);
        expect(result.current.canViewStats).toBe(true);

        // 編集・削除権限は持たない
        expect(result.current.canCreateIncidents).toBe(false);
        expect(result.current.canEditIncidents).toBe(false);
        expect(result.current.canDeleteIncidents).toBe(false);
        expect(result.current.canManageTags).toBe(false);
        expect(result.current.canManagePostMortems).toBe(false);
        expect(result.current.canViewUsers).toBe(false);
        expect(result.current.canManageUsers).toBe(false);
        expect(result.current.canExportData).toBe(false);
      });

      it('role helpers return correct values', async () => {
        const { result } = renderHook(() => usePermissions(), {
          wrapper: createWrapper({ token: 'viewer-token', user: testUsers.viewer }),
        });

        expect(result.current.isAdmin).toBe(false);
        expect(result.current.isEditorOrAdmin).toBe(false);
        expect(result.current.isViewer).toBe(true);
        expect(result.current.canEdit).toBe(false);
      });
    });
  });

  describe('can function - hasPermission wrapper', () => {
    it('returns true for allowed permissions', async () => {
      const { result } = renderHook(() => usePermissions(), {
        wrapper: createWrapper({ token: 'admin-token', user: testUsers.admin }),
      });

      expect(result.current.can('view_incidents')).toBe(true);
      expect(result.current.can('manage_users')).toBe(true);
      expect(result.current.can('export_data')).toBe(true);
    });

    it('returns false for denied permissions', async () => {
      const { result } = renderHook(() => usePermissions(), {
        wrapper: createWrapper({ token: 'viewer-token', user: testUsers.viewer }),
      });

      expect(result.current.can('manage_users')).toBe(false);
      expect(result.current.can('create_incidents')).toBe(false);
      expect(result.current.can('manage_tags')).toBe(false);
    });

    it('returns false when no user is present', async () => {
      const { result } = renderHook(() => usePermissions(), {
        wrapper: createWrapper(),
      });

      expect(result.current.can('view_incidents')).toBe(false);
      expect(result.current.can('manage_users')).toBe(false);
    });
  });

  describe('canAny function - hasAnyPermission wrapper', () => {
    it('returns true if at least one permission is granted for admin', async () => {
      const { result } = renderHook(() => usePermissions(), {
        wrapper: createWrapper({ token: 'admin-token', user: testUsers.admin }),
      });

      expect(result.current.canAny(['view_incidents', 'manage_users'])).toBe(true);
      expect(result.current.canAny(['create_incidents', 'edit_incidents'])).toBe(true);
    });

    it('returns true if at least one permission is granted for viewer', async () => {
      const { result } = renderHook(() => usePermissions(), {
        wrapper: createWrapper({ token: 'viewer-token', user: testUsers.viewer }),
      });

      // viewer は view_incidents を持つが manage_users を持たない
      expect(result.current.canAny(['view_incidents', 'manage_users'])).toBe(true);
    });

    it('returns false if no permissions are granted', async () => {
      const { result } = renderHook(() => usePermissions(), {
        wrapper: createWrapper({ token: 'viewer-token', user: testUsers.viewer }),
      });

      expect(result.current.canAny(['manage_users', 'create_incidents'])).toBe(false);
    });

    it('returns false when no user is present', async () => {
      const { result } = renderHook(() => usePermissions(), {
        wrapper: createWrapper(),
      });

      expect(result.current.canAny(['view_incidents', 'manage_users'])).toBe(false);
    });

    it('returns false for empty array', async () => {
      const { result } = renderHook(() => usePermissions(), {
        wrapper: createWrapper({ token: 'admin-token', user: testUsers.admin }),
      });

      expect(result.current.canAny([])).toBe(false);
    });
  });

  describe('canAll function - hasAllPermissions wrapper', () => {
    it('returns true only if all permissions are granted for admin', async () => {
      const { result } = renderHook(() => usePermissions(), {
        wrapper: createWrapper({ token: 'admin-token', user: testUsers.admin }),
      });

      expect(result.current.canAll(['view_incidents', 'create_incidents'])).toBe(true);
      expect(result.current.canAll(['manage_users', 'manage_tags'])).toBe(true);
    });

    it('returns false if any permission is denied for viewer', async () => {
      const { result } = renderHook(() => usePermissions(), {
        wrapper: createWrapper({ token: 'viewer-token', user: testUsers.viewer }),
      });

      // viewer は view_incidents を持つが create_incidents を持たない
      expect(result.current.canAll(['view_incidents', 'create_incidents'])).toBe(false);
    });

    it('returns true for empty array', async () => {
      const { result } = renderHook(() => usePermissions(), {
        wrapper: createWrapper({ token: 'viewer-token', user: testUsers.viewer }),
      });

      expect(result.current.canAll([])).toBe(true);
    });

    it('returns false when no user is present', async () => {
      const { result } = renderHook(() => usePermissions(), {
        wrapper: createWrapper(),
      });

      expect(result.current.canAll(['view_incidents', 'create_incidents'])).toBe(false);
    });
  });

  describe('Role helper functions', () => {
    describe('isAdmin', () => {
      it('returns true for admin role', async () => {
        const { result } = renderHook(() => usePermissions(), {
          wrapper: createWrapper({ token: 'admin-token', user: testUsers.admin }),
        });

        expect(result.current.isAdmin).toBe(true);
      });

      it('returns false for editor role', async () => {
        const { result } = renderHook(() => usePermissions(), {
          wrapper: createWrapper({ token: 'editor-token', user: testUsers.editor }),
        });

        expect(result.current.isAdmin).toBe(false);
      });

      it('returns false for viewer role', async () => {
        const { result } = renderHook(() => usePermissions(), {
          wrapper: createWrapper({ token: 'viewer-token', user: testUsers.viewer }),
        });

        expect(result.current.isAdmin).toBe(false);
      });

      it('returns false when no user', async () => {
        const { result } = renderHook(() => usePermissions(), {
          wrapper: createWrapper(),
        });

        expect(result.current.isAdmin).toBe(false);
      });
    });

    describe('isEditorOrAdmin', () => {
      it('returns true for admin role', async () => {
        const { result } = renderHook(() => usePermissions(), {
          wrapper: createWrapper({ token: 'admin-token', user: testUsers.admin }),
        });

        expect(result.current.isEditorOrAdmin).toBe(true);
      });

      it('returns true for editor role', async () => {
        const { result } = renderHook(() => usePermissions(), {
          wrapper: createWrapper({ token: 'editor-token', user: testUsers.editor }),
        });

        expect(result.current.isEditorOrAdmin).toBe(true);
      });

      it('returns false for viewer role', async () => {
        const { result } = renderHook(() => usePermissions(), {
          wrapper: createWrapper({ token: 'viewer-token', user: testUsers.viewer }),
        });

        expect(result.current.isEditorOrAdmin).toBe(false);
      });

      it('returns false when no user', async () => {
        const { result } = renderHook(() => usePermissions(), {
          wrapper: createWrapper(),
        });

        expect(result.current.isEditorOrAdmin).toBe(false);
      });
    });

    describe('canManageUsers', () => {
      it('returns true for admin role', async () => {
        const { result } = renderHook(() => usePermissions(), {
          wrapper: createWrapper({ token: 'admin-token', user: testUsers.admin }),
        });

        expect(result.current.canManageUsers).toBe(true);
      });

      it('returns false for editor role', async () => {
        const { result } = renderHook(() => usePermissions(), {
          wrapper: createWrapper({ token: 'editor-token', user: testUsers.editor }),
        });

        expect(result.current.canManageUsers).toBe(false);
      });

      it('returns false for viewer role', async () => {
        const { result } = renderHook(() => usePermissions(), {
          wrapper: createWrapper({ token: 'viewer-token', user: testUsers.viewer }),
        });

        expect(result.current.canManageUsers).toBe(false);
      });

      it('returns false when no user', async () => {
        const { result } = renderHook(() => usePermissions(), {
          wrapper: createWrapper(),
        });

        expect(result.current.canManageUsers).toBe(false);
      });
    });

    describe('canEditContent (canEdit)', () => {
      it('returns true for admin role', async () => {
        const { result } = renderHook(() => usePermissions(), {
          wrapper: createWrapper({ token: 'admin-token', user: testUsers.admin }),
        });

        expect(result.current.canEdit).toBe(true);
      });

      it('returns true for editor role', async () => {
        const { result } = renderHook(() => usePermissions(), {
          wrapper: createWrapper({ token: 'editor-token', user: testUsers.editor }),
        });

        expect(result.current.canEdit).toBe(true);
      });

      it('returns false for viewer role', async () => {
        const { result } = renderHook(() => usePermissions(), {
          wrapper: createWrapper({ token: 'viewer-token', user: testUsers.viewer }),
        });

        expect(result.current.canEdit).toBe(false);
      });

      it('returns false when no user', async () => {
        const { result } = renderHook(() => usePermissions(), {
          wrapper: createWrapper(),
        });

        expect(result.current.canEdit).toBe(false);
      });
    });
  });

  describe('Edge cases', () => {
    it('handles undefined user', async () => {
      localStorage.clear();

      const { result } = renderHook(() => usePermissions(), {
        wrapper: createWrapper(),
      });

      expect(result.current.can('view_incidents')).toBe(false);
      expect(result.current.canAny(['view_incidents'])).toBe(false);
      expect(result.current.canAll(['view_incidents'])).toBe(false);
      expect(result.current.isAdmin).toBe(false);
      expect(result.current.canManageUsers).toBe(false);
    });

    it('handles null user', async () => {
      const { result } = renderHook(() => usePermissions(), {
        wrapper: createWrapper({ user: null }),
      });

      expect(result.current.can('view_incidents')).toBe(false);
      expect(result.current.canViewIncidents).toBe(false);
      expect(result.current.isAdmin).toBe(false);
      expect(result.current.canEdit).toBe(false);
    });

    it('handles user with empty role', async () => {
      const userWithEmptyRole = {
        ...testUsers.viewer,
        role: '' as any,
      };

      const { result } = renderHook(() => usePermissions(), {
        wrapper: createWrapper({ token: 'test-token', user: userWithEmptyRole }),
      });

      // 空のロールは権限なし
      expect(result.current.can('view_incidents')).toBe(false);
      expect(result.current.canViewIncidents).toBe(false);
      expect(result.current.isAdmin).toBe(false);
      expect(result.current.isViewer).toBe(false);
      expect(result.current.isEditorOrAdmin).toBe(false);
    });

    it('handles invalid role values', async () => {
      const userWithInvalidRole = {
        ...testUsers.viewer,
        role: 'invalid_role' as any,
      };

      const { result } = renderHook(() => usePermissions(), {
        wrapper: createWrapper({ token: 'test-token', user: userWithInvalidRole }),
      });

      // 無効なロールは権限なし
      expect(result.current.can('view_incidents')).toBe(false);
      expect(result.current.canViewIncidents).toBe(false);
      expect(result.current.canManageUsers).toBe(false);
      expect(result.current.isAdmin).toBe(false);
      expect(result.current.isEditorOrAdmin).toBe(false);
      expect(result.current.isViewer).toBe(false);
    });

    it('handles user object with missing role field', async () => {
      const userWithoutRole = {
        id: 999,
        name: 'No Role User',
        email: 'norole@example.com',
      } as any;

      const { result } = renderHook(() => usePermissions(), {
        wrapper: createWrapper({ token: 'test-token', user: userWithoutRole }),
      });

      // ロールが未定義の場合は権限なし
      expect(result.current.can('view_incidents')).toBe(false);
      expect(result.current.isAdmin).toBe(false);
    });

    it('handles rapid user changes', async () => {
      const { result, rerender } = renderHook(() => usePermissions(), {
        wrapper: createWrapper({ token: 'viewer-token', user: testUsers.viewer }),
      });

      expect(result.current.isViewer).toBe(true);
      expect(result.current.canManageUsers).toBe(false);

      // ユーザーを変更
      localStorage.clear();
      localStorage.setItem('token', 'admin-token');
      localStorage.setItem('user', JSON.stringify(testUsers.admin));

      rerender();

      // 再レンダリング後も古い値が返される（localStorage の変更は検知されない）
      // AuthContext が再初期化されないため
      expect(result.current.isViewer).toBe(true);
    });
  });

  describe('Specific permission checks', () => {
    it('checks incident permissions correctly for each role', async () => {
      // Admin
      const adminResult = renderHook(() => usePermissions(), {
        wrapper: createWrapper({ token: 'admin-token', user: testUsers.admin }),
      }).result;

      expect(adminResult.current.canViewIncidents).toBe(true);
      expect(adminResult.current.canCreateIncidents).toBe(true);
      expect(adminResult.current.canEditIncidents).toBe(true);
      expect(adminResult.current.canDeleteIncidents).toBe(true);

      // Editor
      const editorResult = renderHook(() => usePermissions(), {
        wrapper: createWrapper({ token: 'editor-token', user: testUsers.editor }),
      }).result;

      expect(editorResult.current.canViewIncidents).toBe(true);
      expect(editorResult.current.canCreateIncidents).toBe(true);
      expect(editorResult.current.canEditIncidents).toBe(true);
      expect(editorResult.current.canDeleteIncidents).toBe(true);

      // Viewer
      const viewerResult = renderHook(() => usePermissions(), {
        wrapper: createWrapper({ token: 'viewer-token', user: testUsers.viewer }),
      }).result;

      expect(viewerResult.current.canViewIncidents).toBe(true);
      expect(viewerResult.current.canCreateIncidents).toBe(false);
      expect(viewerResult.current.canEditIncidents).toBe(false);
      expect(viewerResult.current.canDeleteIncidents).toBe(false);
    });

    it('checks tag permissions correctly for each role', async () => {
      // Admin
      const adminResult = renderHook(() => usePermissions(), {
        wrapper: createWrapper({ token: 'admin-token', user: testUsers.admin }),
      }).result;

      expect(adminResult.current.canViewTags).toBe(true);
      expect(adminResult.current.canManageTags).toBe(true);

      // Editor
      const editorResult = renderHook(() => usePermissions(), {
        wrapper: createWrapper({ token: 'editor-token', user: testUsers.editor }),
      }).result;

      expect(editorResult.current.canViewTags).toBe(true);
      expect(editorResult.current.canManageTags).toBe(true);

      // Viewer
      const viewerResult = renderHook(() => usePermissions(), {
        wrapper: createWrapper({ token: 'viewer-token', user: testUsers.viewer }),
      }).result;

      expect(viewerResult.current.canViewTags).toBe(true);
      expect(viewerResult.current.canManageTags).toBe(false);
    });

    it('checks postmortem permissions correctly for each role', async () => {
      // Admin
      const adminResult = renderHook(() => usePermissions(), {
        wrapper: createWrapper({ token: 'admin-token', user: testUsers.admin }),
      }).result;

      expect(adminResult.current.canViewPostMortems).toBe(true);
      expect(adminResult.current.canManagePostMortems).toBe(true);

      // Editor
      const editorResult = renderHook(() => usePermissions(), {
        wrapper: createWrapper({ token: 'editor-token', user: testUsers.editor }),
      }).result;

      expect(editorResult.current.canViewPostMortems).toBe(true);
      expect(editorResult.current.canManagePostMortems).toBe(true);

      // Viewer
      const viewerResult = renderHook(() => usePermissions(), {
        wrapper: createWrapper({ token: 'viewer-token', user: testUsers.viewer }),
      }).result;

      expect(viewerResult.current.canViewPostMortems).toBe(true);
      expect(viewerResult.current.canManagePostMortems).toBe(false);
    });

    it('checks user management permissions correctly for each role', async () => {
      // Admin
      const adminResult = renderHook(() => usePermissions(), {
        wrapper: createWrapper({ token: 'admin-token', user: testUsers.admin }),
      }).result;

      expect(adminResult.current.canViewUsers).toBe(true);
      expect(adminResult.current.canManageUsers).toBe(true);

      // Editor
      const editorResult = renderHook(() => usePermissions(), {
        wrapper: createWrapper({ token: 'editor-token', user: testUsers.editor }),
      }).result;

      expect(editorResult.current.canViewUsers).toBe(false);
      expect(editorResult.current.canManageUsers).toBe(false);

      // Viewer
      const viewerResult = renderHook(() => usePermissions(), {
        wrapper: createWrapper({ token: 'viewer-token', user: testUsers.viewer }),
      }).result;

      expect(viewerResult.current.canViewUsers).toBe(false);
      expect(viewerResult.current.canManageUsers).toBe(false);
    });

    it('checks stats and export permissions correctly for each role', async () => {
      // Admin
      const adminResult = renderHook(() => usePermissions(), {
        wrapper: createWrapper({ token: 'admin-token', user: testUsers.admin }),
      }).result;

      expect(adminResult.current.canViewStats).toBe(true);
      expect(adminResult.current.canExportData).toBe(true);

      // Editor
      const editorResult = renderHook(() => usePermissions(), {
        wrapper: createWrapper({ token: 'editor-token', user: testUsers.editor }),
      }).result;

      expect(editorResult.current.canViewStats).toBe(true);
      expect(editorResult.current.canExportData).toBe(true);

      // Viewer
      const viewerResult = renderHook(() => usePermissions(), {
        wrapper: createWrapper({ token: 'viewer-token', user: testUsers.viewer }),
      }).result;

      expect(viewerResult.current.canViewStats).toBe(true);
      expect(viewerResult.current.canExportData).toBe(false);
    });
  });

  describe('isViewer helper', () => {
    it('returns true for viewer role', async () => {
      const { result } = renderHook(() => usePermissions(), {
        wrapper: createWrapper({ token: 'viewer-token', user: testUsers.viewer }),
      });

      expect(result.current.isViewer).toBe(true);
    });

    it('returns false for admin role', async () => {
      const { result } = renderHook(() => usePermissions(), {
        wrapper: createWrapper({ token: 'admin-token', user: testUsers.admin }),
      });

      expect(result.current.isViewer).toBe(false);
    });

    it('returns false for editor role', async () => {
      const { result } = renderHook(() => usePermissions(), {
        wrapper: createWrapper({ token: 'editor-token', user: testUsers.editor }),
      });

      expect(result.current.isViewer).toBe(false);
    });

    it('returns false when no user', async () => {
      const { result } = renderHook(() => usePermissions(), {
        wrapper: createWrapper(),
      });

      expect(result.current.isViewer).toBe(false);
    });
  });
});
