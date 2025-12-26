'use client';

import { useState, useEffect } from 'react';
import { useAuth } from '../../context/AuthContext';
import { useRouter } from 'next/navigation';
import { auditLogApi } from '../../lib/api';
import { AuditLog, AuditLogFilters, AuditAction } from '../../types/auditLog';
import { usePermissions } from '../../hooks/usePermissions';

export default function AuditLogsPage() {
  const { token, user } = useAuth();
  const router = useRouter();
  const permissions = usePermissions();
  const [logs, setLogs] = useState<AuditLog[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Pagination state
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [total, setTotal] = useState(0);
  const limit = 50;

  // Filter state
  const [filters, setFilters] = useState<AuditLogFilters>({
    page: 1,
    limit: limit,
  });

  // Check if user is admin
  useEffect(() => {
    if (!permissions.isAdmin) {
      router.push('/dashboard');
    }
  }, [permissions.isAdmin, router]);

  useEffect(() => {
    if (token && permissions.isAdmin) {
      fetchLogs();
    }
  }, [token, filters, permissions.isAdmin]);

  const fetchLogs = async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await auditLogApi.getAll(token!, filters);
      setLogs(response.logs);
      setTotal(response.pagination.total);
      setTotalPages(response.pagination.total_pages);
      setCurrentPage(response.pagination.page);
    } catch (err) {
      setError(err instanceof Error ? err.message : '監査ログの取得に失敗しました');
    } finally {
      setLoading(false);
    }
  };

  const handlePageChange = (newPage: number) => {
    setFilters({ ...filters, page: newPage });
  };

  const handleFilterChange = (key: keyof AuditLogFilters, value: any) => {
    setFilters({ ...filters, [key]: value, page: 1 });
  };

  const clearFilters = () => {
    setFilters({ page: 1, limit: limit });
  };

  const getActionBadgeColor = (action: AuditAction) => {
    switch (action) {
      case 'create':
        return 'bg-green-100 text-green-800';
      case 'update':
        return 'bg-blue-100 text-blue-800';
      case 'delete':
        return 'bg-red-100 text-red-800';
      case 'read':
        return 'bg-gray-100 text-gray-800';
      case 'login':
        return 'bg-purple-100 text-purple-800';
      case 'logout':
        return 'bg-yellow-100 text-yellow-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  const getActionLabel = (action: AuditAction) => {
    switch (action) {
      case 'create':
        return '作成';
      case 'update':
        return '更新';
      case 'delete':
        return '削除';
      case 'read':
        return '閲覧';
      case 'login':
        return 'ログイン';
      case 'logout':
        return 'ログアウト';
      default:
        return action;
    }
  };

  const getStatusCodeColor = (code: number) => {
    if (code >= 200 && code < 300) return 'text-green-600';
    if (code >= 400 && code < 500) return 'text-yellow-600';
    if (code >= 500) return 'text-red-600';
    return 'text-gray-600';
  };

  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleString('ja-JP', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    });
  };

  if (!permissions.isAdmin) {
    return null;
  }

  return (
    <div className="p-6">
      <div className="mb-6">
        <h1 className="text-3xl font-bold text-gray-900">監査ログ</h1>
        <p className="mt-2 text-sm text-gray-600">
          システム内のすべてのアクションが自動的に記録され、セキュリティ分析やコンプライアンス遵守に活用できます。
        </p>
      </div>

      {/* Filters */}
      <div className="mb-6 bg-white p-4 rounded-lg shadow">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              アクション
            </label>
            <select
              className="w-full px-3 py-2 border border-gray-300 rounded-md"
              value={filters.action || ''}
              onChange={(e) => handleFilterChange('action', e.target.value || undefined)}
            >
              <option value="">すべて</option>
              <option value="create">作成</option>
              <option value="update">更新</option>
              <option value="delete">削除</option>
              <option value="read">閲覧</option>
              <option value="login">ログイン</option>
              <option value="logout">ログアウト</option>
            </select>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              リソースタイプ
            </label>
            <input
              type="text"
              className="w-full px-3 py-2 border border-gray-300 rounded-md"
              placeholder="例: incident, user"
              value={filters.resource_type || ''}
              onChange={(e) => handleFilterChange('resource_type', e.target.value || undefined)}
            />
          </div>

          <div className="flex items-end">
            <button
              onClick={clearFilters}
              className="px-4 py-2 text-sm text-gray-700 bg-gray-100 hover:bg-gray-200 rounded-md"
            >
              フィルタをクリア
            </button>
          </div>
        </div>
      </div>

      {/* Stats */}
      <div className="mb-4 text-sm text-gray-600">
        全 {total} 件中 {(currentPage - 1) * limit + 1} - {Math.min(currentPage * limit, total)} 件を表示
      </div>

      {/* Error Message */}
      {error && (
        <div className="mb-4 p-4 bg-red-50 border border-red-200 text-red-700 rounded-lg">
          {error}
        </div>
      )}

      {/* Loading */}
      {loading ? (
        <div className="flex justify-center items-center h-64">
          <div className="text-gray-900">読み込み中...</div>
        </div>
      ) : (
        <>
          {/* Logs Table */}
          <div className="bg-white shadow overflow-hidden rounded-lg">
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      日時
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      ユーザー
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      アクション
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      リソース
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      メソッド
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      パス
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      ステータス
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      IPアドレス
                    </th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {logs.length === 0 ? (
                    <tr>
                      <td colSpan={8} className="px-6 py-4 text-center text-gray-700">
                        監査ログがありません
                      </td>
                    </tr>
                  ) : (
                    logs.map((log) => (
                      <tr key={log.id} className="hover:bg-gray-50">
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                          {formatDate(log.created_at)}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap">
                          <div className="text-sm font-medium text-gray-900">
                            {log.user_name || '不明'}
                          </div>
                          <div className="text-sm text-gray-700">
                            {log.user_email || '-'}
                          </div>
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap">
                          <span
                            className={`px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${getActionBadgeColor(
                              log.action
                            )}`}
                          >
                            {getActionLabel(log.action)}
                          </span>
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                          {log.resource_type}
                          {log.resource_id && ` #${log.resource_id}`}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                          <span className="font-mono text-xs">{log.method}</span>
                        </td>
                        <td className="px-6 py-4 text-sm text-gray-900">
                          <span className="font-mono text-xs">{log.path}</span>
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm">
                          <span className={`font-semibold ${getStatusCodeColor(log.status_code)}`}>
                            {log.status_code}
                          </span>
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-700">
                          {log.ip_address}
                        </td>
                      </tr>
                    ))
                  )}
                </tbody>
              </table>
            </div>
          </div>

          {/* Pagination */}
          {totalPages > 1 && (
            <div className="mt-6 flex justify-center items-center space-x-2">
              <button
                onClick={() => handlePageChange(currentPage - 1)}
                disabled={currentPage === 1}
                className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                前へ
              </button>

              <span className="text-sm text-gray-700">
                ページ {currentPage} / {totalPages}
              </span>

              <button
                onClick={() => handlePageChange(currentPage + 1)}
                disabled={currentPage === totalPages}
                className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                次へ
              </button>
            </div>
          )}
        </>
      )}
    </div>
  );
}
