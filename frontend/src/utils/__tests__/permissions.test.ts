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
  describe('admin role', () => {
    it('has all permissions', () => {
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

  describe('editor role', () => {
    it('has content management permissions', () => {
      expect(hasPermission('editor', 'view_incidents')).toBe(true);
      expect(hasPermission('editor', 'create_incidents')).toBe(true);
      expect(hasPermission('editor', 'edit_incidents')).toBe(true);
      expect(hasPermission('editor', 'delete_incidents')).toBe(true);
      expect(hasPermission('editor', 'manage_tags')).toBe(true);
      expect(hasPermission('editor', 'manage_postmortems')).toBe(true);
    });

    it('does not have user management permissions', () => {
      expect(hasPermission('editor', 'view_users')).toBe(false);
      expect(hasPermission('editor', 'manage_users')).toBe(false);
    });
  });

  describe('viewer role', () => {
    it('has view-only permissions', () => {
      expect(hasPermission('viewer', 'view_incidents')).toBe(true);
      expect(hasPermission('viewer', 'view_tags')).toBe(true);
      expect(hasPermission('viewer', 'view_postmortems')).toBe(true);
      expect(hasPermission('viewer', 'view_stats')).toBe(true);
    });

    it('does not have edit permissions', () => {
      expect(hasPermission('viewer', 'create_incidents')).toBe(false);
      expect(hasPermission('viewer', 'edit_incidents')).toBe(false);
      expect(hasPermission('viewer', 'delete_incidents')).toBe(false);
      expect(hasPermission('viewer', 'manage_tags')).toBe(false);
      expect(hasPermission('viewer', 'manage_users')).toBe(false);
    });
  });

  describe('invalid role', () => {
    it('returns false for unknown role', () => {
      expect(hasPermission('unknown' as any, 'view_incidents')).toBe(false);
    });
  });
});

describe('hasAnyPermission', () => {
  it('returns true if user has at least one permission', () => {
    expect(hasAnyPermission('viewer', ['view_incidents', 'manage_users'])).toBe(true);
  });

  it('returns false if user has none of the permissions', () => {
    expect(hasAnyPermission('viewer', ['manage_users', 'create_incidents'])).toBe(false);
  });

  it('returns true for admin with any permission', () => {
    expect(hasAnyPermission('admin', ['manage_users', 'manage_tags'])).toBe(true);
  });
});

describe('hasAllPermissions', () => {
  it('returns true if user has all permissions', () => {
    expect(hasAllPermissions('admin', ['view_incidents', 'create_incidents'])).toBe(true);
  });

  it('returns false if user is missing any permission', () => {
    expect(hasAllPermissions('viewer', ['view_incidents', 'create_incidents'])).toBe(false);
  });

  it('returns true for empty permissions array', () => {
    expect(hasAllPermissions('viewer', [])).toBe(true);
  });
});

describe('canManageUsers', () => {
  it('returns true for admin', () => {
    expect(canManageUsers('admin')).toBe(true);
  });

  it('returns false for editor', () => {
    expect(canManageUsers('editor')).toBe(false);
  });

  it('returns false for viewer', () => {
    expect(canManageUsers('viewer')).toBe(false);
  });
});

describe('canEditContent', () => {
  it('returns true for admin', () => {
    expect(canEditContent('admin')).toBe(true);
  });

  it('returns true for editor', () => {
    expect(canEditContent('editor')).toBe(true);
  });

  it('returns false for viewer', () => {
    expect(canEditContent('viewer')).toBe(false);
  });
});

describe('isAdmin', () => {
  it('returns true for admin role', () => {
    expect(isAdmin('admin')).toBe(true);
  });

  it('returns false for editor role', () => {
    expect(isAdmin('editor')).toBe(false);
  });

  it('returns false for viewer role', () => {
    expect(isAdmin('viewer')).toBe(false);
  });
});

describe('isEditorOrAdmin', () => {
  it('returns true for admin role', () => {
    expect(isEditorOrAdmin('admin')).toBe(true);
  });

  it('returns true for editor role', () => {
    expect(isEditorOrAdmin('editor')).toBe(true);
  });

  it('returns false for viewer role', () => {
    expect(isEditorOrAdmin('viewer')).toBe(false);
  });
});

describe('isViewer', () => {
  it('returns true for viewer role', () => {
    expect(isViewer('viewer')).toBe(true);
  });

  it('returns false for admin role', () => {
    expect(isViewer('admin')).toBe(false);
  });

  it('returns false for editor role', () => {
    expect(isViewer('editor')).toBe(false);
  });
});
