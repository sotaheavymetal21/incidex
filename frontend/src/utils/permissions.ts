import { Role } from '@/types/user';

export type Permission =
  | 'view_incidents'
  | 'create_incidents'
  | 'edit_incidents'
  | 'delete_incidents'
  | 'view_tags'
  | 'manage_tags'
  | 'view_templates'
  | 'manage_templates'
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
    'view_templates',
    'manage_templates',
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
    'view_templates',
    'manage_templates',
    'view_postmortems',
    'manage_postmortems',
    'view_stats',
    'export_data',
  ],
  viewer: [
    'view_incidents',
    'view_tags',
    'view_templates',
    'view_postmortems',
    'view_stats',
  ],
};

export function hasPermission(userRole: Role | string, permission: Permission): boolean {
  const role = userRole as Role;
  return rolePermissions[role]?.includes(permission) ?? false;
}

export function hasAnyPermission(userRole: Role | string, permissions: Permission[]): boolean {
  return permissions.some(permission => hasPermission(userRole, permission));
}

export function hasAllPermissions(userRole: Role | string, permissions: Permission[]): boolean {
  return permissions.every(permission => hasPermission(userRole, permission));
}

export function canManageUsers(userRole: Role | string): boolean {
  return hasPermission(userRole, 'manage_users');
}

export function canEditContent(userRole: Role | string): boolean {
  return hasAnyPermission(userRole, ['create_incidents', 'edit_incidents', 'manage_tags']);
}

export function isAdmin(userRole: Role | string): boolean {
  return userRole === 'admin';
}

export function isEditorOrAdmin(userRole: Role | string): boolean {
  return userRole === 'admin' || userRole === 'editor';
}

export function isViewer(userRole: Role | string): boolean {
  return userRole === 'viewer';
}
