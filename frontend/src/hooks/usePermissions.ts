import { useAuth } from '@/context/AuthContext';
import {
  hasPermission,
  hasAnyPermission,
  hasAllPermissions,
  canManageUsers,
  canEditContent,
  isAdmin,
  isEditorOrAdmin,
  isViewer,
  Permission,
} from '@/utils/permissions';

export function usePermissions() {
  const { user } = useAuth();

  const checkPermission = (permission: Permission): boolean => {
    if (!user) return false;
    return hasPermission(user.role, permission);
  };

  const checkAnyPermission = (permissions: Permission[]): boolean => {
    if (!user) return false;
    return hasAnyPermission(user.role, permissions);
  };

  const checkAllPermissions = (permissions: Permission[]): boolean => {
    if (!user) return false;
    return hasAllPermissions(user.role, permissions);
  };

  return {
    // 汎用的なチェック関数
    can: checkPermission,
    canAny: checkAnyPermission,
    canAll: checkAllPermissions,

    // 具体的な権限チェック
    canViewIncidents: checkPermission('view_incidents'),
    canCreateIncidents: checkPermission('create_incidents'),
    canEditIncidents: checkPermission('edit_incidents'),
    canDeleteIncidents: checkPermission('delete_incidents'),

    canViewTags: checkPermission('view_tags'),
    canManageTags: checkPermission('manage_tags'),

    canViewTemplates: checkPermission('view_templates'),
    canManageTemplates: checkPermission('manage_templates'),

    canViewPostMortems: checkPermission('view_postmortems'),
    canManagePostMortems: checkPermission('manage_postmortems'),

    canViewUsers: checkPermission('view_users'),
    canManageUsers: user ? canManageUsers(user.role) : false,

    canViewStats: checkPermission('view_stats'),
    canExportData: checkPermission('export_data'),

    // ロールベースのチェック
    isAdmin: user ? isAdmin(user.role) : false,
    isEditorOrAdmin: user ? isEditorOrAdmin(user.role) : false,
    isViewer: user ? isViewer(user.role) : false,
    canEdit: user ? canEditContent(user.role) : false,
  };
}
