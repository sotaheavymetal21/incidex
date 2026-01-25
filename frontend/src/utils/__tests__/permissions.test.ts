import { describe, it, expect } from 'vitest';
import {
  hasPermission,
  hasAnyPermission,
  hasAllPermissions,
  canManageUsers,
  canEditContent,
  isAdmin,
  isEditorOrAdmin,
  isViewer,
} from '../permissions';
import type { Permission } from '../permissions';

describe('hasPermission', () => {
  describe('admin ロール', () => {
    it('すべての権限を持っています', () => {
      const adminPermissions: Permission[] = [
        'view_incidents',
        'create_incidents',
        'edit_incidents',
        'delete_incidents',
        'view_tags',
        'manage_tags',
        'view_postmortems',
        'manage_postmortems',
        'view_users',
        'manage_users',
        'view_stats',
        'export_data',
      ];

      adminPermissions.forEach(permission => {
        expect(hasPermission('admin', permission)).toBe(true);
      });
    });
  });

  describe('editor ロール', () => {
    it('コンテンツ管理権限を持っています', () => {
      expect(hasPermission('editor', 'view_incidents')).toBe(true);
      expect(hasPermission('editor', 'create_incidents')).toBe(true);
      expect(hasPermission('editor', 'edit_incidents')).toBe(true);
      expect(hasPermission('editor', 'delete_incidents')).toBe(true);
      expect(hasPermission('editor', 'manage_tags')).toBe(true);
      expect(hasPermission('editor', 'manage_postmortems')).toBe(true);
    });

    it('ユーザー管理権限を持っていません', () => {
      expect(hasPermission('editor', 'view_users')).toBe(false);
      expect(hasPermission('editor', 'manage_users')).toBe(false);
    });
  });

  describe('viewer ロール', () => {
    it('閲覧のみの権限を持っています', () => {
      expect(hasPermission('viewer', 'view_incidents')).toBe(true);
      expect(hasPermission('viewer', 'view_tags')).toBe(true);
      expect(hasPermission('viewer', 'view_postmortems')).toBe(true);
      expect(hasPermission('viewer', 'view_stats')).toBe(true);
    });

    it('編集権限を持っていません', () => {
      expect(hasPermission('viewer', 'create_incidents')).toBe(false);
      expect(hasPermission('viewer', 'edit_incidents')).toBe(false);
      expect(hasPermission('viewer', 'delete_incidents')).toBe(false);
      expect(hasPermission('viewer', 'manage_tags')).toBe(false);
      expect(hasPermission('viewer', 'manage_users')).toBe(false);
    });
  });

  describe('無効なロール', () => {
    it('不明なロールの場合は false を返します', () => {
      expect(hasPermission('unknown' as any, 'view_incidents')).toBe(false);
    });
  });
});

describe('hasAnyPermission', () => {
  it('少なくとも1つの権限を持っている場合は true を返します', () => {
    expect(hasAnyPermission('viewer', ['view_incidents', 'manage_users'])).toBe(true);
  });

  it('いずれの権限も持っていない場合は false を返します', () => {
    expect(hasAnyPermission('viewer', ['manage_users', 'create_incidents'])).toBe(false);
  });

  it('admin は任意の権限で true を返します', () => {
    expect(hasAnyPermission('admin', ['manage_users', 'manage_tags'])).toBe(true);
  });
});

describe('hasAllPermissions', () => {
  it('すべての権限を持っている場合は true を返します', () => {
    expect(hasAllPermissions('admin', ['view_incidents', 'create_incidents'])).toBe(true);
  });

  it('いずれかの権限が欠けている場合は false を返します', () => {
    expect(hasAllPermissions('viewer', ['view_incidents', 'create_incidents'])).toBe(false);
  });

  it('空の権限配列の場合は true を返します', () => {
    expect(hasAllPermissions('viewer', [])).toBe(true);
  });
});

describe('canManageUsers', () => {
  it('admin の場合は true を返します', () => {
    expect(canManageUsers('admin')).toBe(true);
  });

  it('editor の場合は false を返します', () => {
    expect(canManageUsers('editor')).toBe(false);
  });

  it('viewer の場合は false を返します', () => {
    expect(canManageUsers('viewer')).toBe(false);
  });
});

describe('canEditContent', () => {
  it('admin の場合は true を返します', () => {
    expect(canEditContent('admin')).toBe(true);
  });

  it('editor の場合は true を返します', () => {
    expect(canEditContent('editor')).toBe(true);
  });

  it('viewer の場合は false を返します', () => {
    expect(canEditContent('viewer')).toBe(false);
  });
});

describe('isAdmin', () => {
  it('admin ロールの場合は true を返します', () => {
    expect(isAdmin('admin')).toBe(true);
  });

  it('editor ロールの場合は false を返します', () => {
    expect(isAdmin('editor')).toBe(false);
  });

  it('viewer ロールの場合は false を返します', () => {
    expect(isAdmin('viewer')).toBe(false);
  });
});

describe('isEditorOrAdmin', () => {
  it('admin ロールの場合は true を返します', () => {
    expect(isEditorOrAdmin('admin')).toBe(true);
  });

  it('editor ロールの場合は true を返します', () => {
    expect(isEditorOrAdmin('editor')).toBe(true);
  });

  it('viewer ロールの場合は false を返します', () => {
    expect(isEditorOrAdmin('viewer')).toBe(false);
  });
});

describe('isViewer', () => {
  it('viewer ロールの場合は true を返します', () => {
    expect(isViewer('viewer')).toBe(true);
  });

  it('admin ロールの場合は false を返します', () => {
    expect(isViewer('admin')).toBe(false);
  });

  it('editor ロールの場合は false を返します', () => {
    expect(isViewer('editor')).toBe(false);
  });
});
