import { Role } from '@/types/user';

export type Permission =
  | 'view_incidents'
  | 'create_incidents'
  | 'edit_incidents'
  | 'delete_incidents'
  | 'view_tags'
  | 'manage_tags'
  | 'view_postmortems'
  | 'manage_postmortems'
  | 'view_users'
  | 'manage_users'
  | 'view_stats'
  | 'export_data';

const rolePermissions: Record<Role, Permission[]> = {
  admin: [
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
  ],
  editor: [
    'view_incidents',
    'create_incidents',
    'edit_incidents',
    'delete_incidents',
    'view_tags',
    'manage_tags',
    'view_postmortems',
    'manage_postmortems',
    'view_stats',
    'export_data',
  ],
  viewer: [
    'view_incidents',
    'view_tags',
    'view_postmortems',
    'view_stats',
  ],
};

/**
 * ユーザーが指定された権限を持っているかチェックします
 */
export function hasPermission(userRole: Role | string, permission: Permission): boolean {
  const role = userRole as Role;
  return rolePermissions[role]?.includes(permission) ?? false;
}

/**
 * ユーザーが指定された権限のいずれかを持っているかチェックします
 */
export function hasAnyPermission(userRole: Role | string, permissions: Permission[]): boolean {
  return permissions.some(permission => hasPermission(userRole, permission));
}

/**
 * ユーザーが指定されたすべての権限を持っているかチェックします
 */
export function hasAllPermissions(userRole: Role | string, permissions: Permission[]): boolean {
  return permissions.every(permission => hasPermission(userRole, permission));
}

/**
 * ユーザーがユーザー管理権限を持っているかチェックします
 */
export function canManageUsers(userRole: Role | string): boolean {
  return hasPermission(userRole, 'manage_users');
}

/**
 * ユーザーがコンテンツ編集権限を持っているかチェックします
 */
export function canEditContent(userRole: Role | string): boolean {
  return hasAnyPermission(userRole, ['create_incidents', 'edit_incidents', 'manage_tags']);
}

/**
 * ユーザーが管理者かどうかチェックします
 */
export function isAdmin(userRole: Role | string): boolean {
  return userRole === 'admin';
}

/**
 * ユーザーが編集者または管理者かどうかチェックします
 */
export function isEditorOrAdmin(userRole: Role | string): boolean {
  return userRole === 'admin' || userRole === 'editor';
}

/**
 * ユーザーが閲覧者かどうかチェックします
 */
export function isViewer(userRole: Role | string): boolean {
  return userRole === 'viewer';
}
