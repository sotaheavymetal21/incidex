'use client';

import { useEffect, useState, Suspense } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { useAuth } from '@/context/AuthContext';
import { usePermissions } from '@/hooks/usePermissions';
import { incidentApi, tagApi, exportApi } from '@/lib/api';
import { Incident, Severity, Status, PaginationResult } from '@/types/incident';
import { Tag } from '@/types/tag';
import SeverityGuide from '@/components/SeverityGuide';

const SEVERITY_LABELS: Record<Severity, string> = {
  critical: 'Critical',
  high: 'High',
  medium: 'Medium',
  low: 'Low',
};

const STATUS_LABELS: Record<Status, string> = {
  open: 'Open',
  investigating: 'Investigating',
  resolved: 'Resolved',
  closed: 'Closed',
};

function IncidentsPageContent() {
  const { token, loading: authLoading, user } = useAuth();
  const permissions = usePermissions();
  const router = useRouter();
  const searchParams = useSearchParams();
  const [incidents, setIncidents] = useState<Incident[]>([]);
  const [tags, setTags] = useState<Tag[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [pagination, setPagination] = useState<PaginationResult>({
    page: 1,
    limit: 20,
    total: 0,
    total_pages: 0,
  });

  // Filters
  const [search, setSearch] = useState('');
  const [severity, setSeverity] = useState<Severity | ''>('');
  const [status, setStatus] = useState<Status | ''>('');
  const [selectedTagIds, setSelectedTagIds] = useState<number[]>([]);
  const [showSidebar, setShowSidebar] = useState(true);

  // Load filters from URL params
  useEffect(() => {
    const paramSeverity = searchParams.get('severity') as Severity | null;
    const paramStatus = searchParams.get('status') as Status | null;
    const paramTags = searchParams.get('tags');
    const paramSearch = searchParams.get('search');

    if (paramSeverity) setSeverity(paramSeverity);
    if (paramStatus) setStatus(paramStatus);
    if (paramTags) setSelectedTagIds(paramTags.split(',').map(Number));
    if (paramSearch) setSearch(paramSearch);
  }, [searchParams]);

  useEffect(() => {
    if (!authLoading && !token) {
      router.push('/login');
    }
  }, [token, authLoading, router]);

  useEffect(() => {
    if (token) {
      fetchTags();
    }
  }, [token]);

  useEffect(() => {
    if (token) {
      fetchIncidents();
    }
  }, [token, pagination.page, search, severity, status, selectedTagIds]);

  const fetchTags = async () => {
    try {
      const fetchedTags = await tagApi.getAll(token!);
      setTags(fetchedTags);
    } catch (err) {
      console.error('Failed to fetch tags:', err);
    }
  };

  const fetchIncidents = async () => {
    setLoading(true);
    setError('');
    try {
      const response = await incidentApi.getAll(token!, {
        page: pagination.page,
        limit: pagination.limit,
        search: search || undefined,
        severity: severity || undefined,
        status: status || undefined,
        tag_ids: selectedTagIds.length > 0 ? selectedTagIds.join(',') : undefined,
      });
      setIncidents(response.incidents);
      setPagination(response.pagination);
    } catch (err: any) {
      setError(err.message || 'Failed to fetch incidents');
    } finally {
      setLoading(false);
    }
  };

  const handleSearchChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setSearch(e.target.value);
    setPagination({ ...pagination, page: 1 });
  };

  const handleTagToggle = (tagId: number) => {
    setSelectedTagIds((prev) =>
      prev.includes(tagId) ? prev.filter((id) => id !== tagId) : [...prev, tagId]
    );
    setPagination({ ...pagination, page: 1 });
  };

  const clearFilters = () => {
    setSearch('');
    setSeverity('');
    setStatus('');
    setSelectedTagIds([]);
    setPagination({ ...pagination, page: 1 });
  };

  const clearFilter = (filterType: 'search' | 'severity' | 'status' | 'tag', value?: number) => {
    switch (filterType) {
      case 'search':
        setSearch('');
        break;
      case 'severity':
        setSeverity('');
        break;
      case 'status':
        setStatus('');
        break;
      case 'tag':
        if (value !== undefined) {
          setSelectedTagIds((prev) => prev.filter((id) => id !== value));
        }
        break;
    }
    setPagination({ ...pagination, page: 1 });
  };

  const applyPreset = (preset: 'unresolved' | 'my-assigned' | 'critical') => {
    clearFilters();
    switch (preset) {
      case 'unresolved':
        setStatus('open');
        break;
      case 'my-assigned':
        // Note: This requires API support for assignee filter
        // For now, we'll just use the user state
        break;
      case 'critical':
        setSeverity('critical');
        break;
    }
  };

  const getSeverityColor = (severity: Severity) => {
    switch (severity) {
      case 'critical': return 'bg-red-100 text-red-800 border-red-300';
      case 'high': return 'bg-orange-100 text-orange-800 border-orange-300';
      case 'medium': return 'bg-yellow-100 text-yellow-800 border-yellow-300';
      case 'low': return 'bg-green-100 text-green-800 border-green-300';
      default: return 'bg-gray-100 text-gray-800 border-gray-300';
    }
  };

  const getStatusColor = (status: Status) => {
    switch (status) {
      case 'open': return 'bg-gray-100 text-gray-800 border-gray-300';
      case 'investigating': return 'bg-blue-100 text-blue-800 border-blue-300';
      case 'resolved': return 'bg-green-100 text-green-800 border-green-300';
      case 'closed': return 'bg-purple-100 text-purple-800 border-purple-300';
      default: return 'bg-gray-100 text-gray-800 border-gray-300';
    }
  };

  const handleExportCSV = async () => {
    try {
      const blob = await exportApi.exportIncidentsCSV(token!, {
        search: search || undefined,
        severity: severity || undefined,
        status: status || undefined,
        tag_ids: selectedTagIds.length > 0 ? selectedTagIds.join(',') : undefined,
      });

      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = `incidents_${new Date().toISOString().split('T')[0]}.csv`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      window.URL.revokeObjectURL(url);
    } catch (err) {
      console.error('Export failed:', err);
      setError('Failed to export incidents');
    }
  };

  const hasActiveFilters = search || severity || status || selectedTagIds.length > 0;

  if (authLoading || !token) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-gray-600">Loading...</div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50 py-8 px-4 sm:px-6 lg:px-8">
      <div className="max-w-7xl mx-auto">
        {/* Header */}
        <div className="mb-6 flex justify-between items-center">
          <h1 className="text-3xl font-bold text-gray-900">インシデント一覧</h1>
          <div className="flex space-x-3">
            <button
              onClick={handleExportCSV}
              className="px-4 py-2 bg-green-600 text-white rounded-md hover:bg-green-700 flex items-center"
            >
              <svg className="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
              </svg>
              CSVエクスポート
            </button>
            {permissions.canCreateIncidents && (
              <button
                onClick={() => router.push('/incidents/create')}
                className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
              >
                新規作成
              </button>
            )}
          </div>
        </div>

        {/* Severity Guide */}
        <SeverityGuide />

        <div className="flex gap-6">
          {/* Sidebar Filter Panel */}
          {showSidebar && (
            <div className="w-64 flex-shrink-0">
              <div className="bg-white rounded-lg shadow p-4 sticky top-8">
                <div className="flex items-center justify-between mb-4">
                  <h2 className="text-lg font-semibold text-gray-900">フィルター</h2>
                  <button
                    onClick={() => setShowSidebar(false)}
                    className="text-gray-400 hover:text-gray-600 lg:hidden"
                  >
                    <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                    </svg>
                  </button>
                </div>

                {/* Filter Presets */}
                <div className="mb-6">
                  <h3 className="text-sm font-medium text-gray-700 mb-2">クイックフィルター</h3>
                  <div className="space-y-2">
                    <button
                      onClick={() => applyPreset('unresolved')}
                      className="w-full px-3 py-2 text-sm text-left bg-gray-50 hover:bg-gray-100 rounded-md text-gray-700"
                    >
                      未解決のインシデント
                    </button>
                    <button
                      onClick={() => applyPreset('critical')}
                      className="w-full px-3 py-2 text-sm text-left bg-gray-50 hover:bg-gray-100 rounded-md text-gray-700"
                    >
                      Critical のみ
                    </button>
                    <button
                      onClick={clearFilters}
                      className="w-full px-3 py-2 text-sm text-left bg-gray-50 hover:bg-gray-100 rounded-md text-gray-700"
                    >
                      すべてクリア
                    </button>
                  </div>
                </div>

                {/* Search */}
                <div className="mb-6">
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    検索
                  </label>
                  <input
                    type="text"
                    value={search}
                    onChange={handleSearchChange}
                    placeholder="タイトル、説明を検索..."
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 text-gray-900 placeholder-gray-500"
                  />
                </div>

                {/* Severity Filter */}
                <div className="mb-6">
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    深刻度
                  </label>
                  <div className="space-y-2">
                    {(Object.keys(SEVERITY_LABELS) as Severity[]).map((sev) => (
                      <label key={sev} className="flex items-center">
                        <input
                          type="radio"
                          name="severity"
                          checked={severity === sev}
                          onChange={() => {
                            setSeverity(sev);
                            setPagination({ ...pagination, page: 1 });
                          }}
                          className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300"
                        />
                        <span className="ml-2 text-sm text-gray-700">{SEVERITY_LABELS[sev]}</span>
                      </label>
                    ))}
                    <label className="flex items-center">
                      <input
                        type="radio"
                        name="severity"
                        checked={severity === ''}
                        onChange={() => {
                          setSeverity('');
                          setPagination({ ...pagination, page: 1 });
                        }}
                        className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300"
                      />
                      <span className="ml-2 text-sm text-gray-700">すべて</span>
                    </label>
                  </div>
                </div>

                {/* Status Filter */}
                <div className="mb-6">
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    ステータス
                  </label>
                  <div className="space-y-2">
                    {(Object.keys(STATUS_LABELS) as Status[]).map((st) => (
                      <label key={st} className="flex items-center">
                        <input
                          type="radio"
                          name="status"
                          checked={status === st}
                          onChange={() => {
                            setStatus(st);
                            setPagination({ ...pagination, page: 1 });
                          }}
                          className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300"
                        />
                        <span className="ml-2 text-sm text-gray-700">{STATUS_LABELS[st]}</span>
                      </label>
                    ))}
                    <label className="flex items-center">
                      <input
                        type="radio"
                        name="status"
                        checked={status === ''}
                        onChange={() => {
                          setStatus('');
                          setPagination({ ...pagination, page: 1 });
                        }}
                        className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300"
                      />
                      <span className="ml-2 text-sm text-gray-700">すべて</span>
                    </label>
                  </div>
                </div>

                {/* Tag Filter */}
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    タグ
                  </label>
                  <div className="space-y-2">
                    {tags.map((tag) => (
                      <label key={tag.id} className="flex items-center">
                        <input
                          type="checkbox"
                          checked={selectedTagIds.includes(tag.id)}
                          onChange={() => handleTagToggle(tag.id)}
                          className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                        />
                        <span className="ml-2 text-sm text-gray-700">{tag.name}</span>
                        <span
                          className="ml-auto w-3 h-3 rounded-full"
                          style={{ backgroundColor: tag.color }}
                        ></span>
                      </label>
                    ))}
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* Main Content */}
          <div className="flex-1 min-w-0">
            {/* Toggle Sidebar Button (Mobile) */}
            {!showSidebar && (
              <button
                onClick={() => setShowSidebar(true)}
                className="mb-4 px-4 py-2 bg-white border border-gray-300 rounded-md text-gray-700 hover:bg-gray-50 flex items-center"
              >
                <svg className="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z" />
                </svg>
                フィルターを表示
              </button>
            )}

            {/* Filter Chips */}
            {hasActiveFilters && (
              <div className="mb-4 bg-white p-4 rounded-lg shadow">
                <div className="flex items-center justify-between mb-2">
                  <h3 className="text-sm font-medium text-gray-700">適用中のフィルター:</h3>
                  <button
                    onClick={clearFilters}
                    className="text-sm text-blue-600 hover:text-blue-800"
                  >
                    すべてクリア
                  </button>
                </div>
                <div className="flex flex-wrap gap-2">
                  {search && (
                    <span className="inline-flex items-center px-3 py-1 rounded-full text-sm bg-blue-100 text-blue-800">
                      検索: {search}
                      <button
                        onClick={() => clearFilter('search')}
                        className="ml-2 text-blue-600 hover:text-blue-900"
                      >
                        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                        </svg>
                      </button>
                    </span>
                  )}
                  {severity && (
                    <span className="inline-flex items-center px-3 py-1 rounded-full text-sm bg-orange-100 text-orange-800">
                      深刻度: {SEVERITY_LABELS[severity]}
                      <button
                        onClick={() => clearFilter('severity')}
                        className="ml-2 text-orange-600 hover:text-orange-900"
                      >
                        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                        </svg>
                      </button>
                    </span>
                  )}
                  {status && (
                    <span className="inline-flex items-center px-3 py-1 rounded-full text-sm bg-green-100 text-green-800">
                      ステータス: {STATUS_LABELS[status]}
                      <button
                        onClick={() => clearFilter('status')}
                        className="ml-2 text-green-600 hover:text-green-900"
                      >
                        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                        </svg>
                      </button>
                    </span>
                  )}
                  {selectedTagIds.map((tagId) => {
                    const tag = tags.find((t) => t.id === tagId);
                    return tag ? (
                      <span
                        key={tagId}
                        className="inline-flex items-center px-3 py-1 rounded-full text-sm text-white"
                        style={{ backgroundColor: tag.color }}
                      >
                        {tag.name}
                        <button
                          onClick={() => clearFilter('tag', tagId)}
                          className="ml-2 text-white hover:text-gray-200"
                        >
                          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                          </svg>
                        </button>
                      </span>
                    ) : null;
                  })}
                </div>
              </div>
            )}

            {/* Error Message */}
            {error && (
              <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-4">
                {error}
              </div>
            )}

            {/* Incidents Table */}
            {loading ? (
              <div className="text-center py-8 text-gray-600">Loading incidents...</div>
            ) : incidents.length === 0 ? (
              <div className="text-center py-8 text-gray-600">
                インシデントが見つかりませんでした。フィルターを変更するか、新しいインシデントを作成してください。
              </div>
            ) : (
              <div className="bg-white rounded-lg shadow overflow-hidden">
                <table className="min-w-full divide-y divide-gray-200">
                  <thead className="bg-gray-50">
                    <tr>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Title
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Severity
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Status
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Detected At
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Assignee
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Tags
                      </th>
                    </tr>
                  </thead>
                  <tbody className="bg-white divide-y divide-gray-200">
                    {incidents.map((incident) => (
                      <tr
                        key={incident.id}
                        onClick={() => router.push(`/incidents/${incident.id}`)}
                        className="hover:bg-gray-50 cursor-pointer transition-colors"
                      >
                        <td className="px-6 py-4 whitespace-nowrap">
                          <div className="text-sm font-medium text-gray-900">
                            {incident.title}
                          </div>
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap">
                          <span
                            className={`px-2 py-1 inline-flex text-xs leading-5 font-semibold rounded-full border ${getSeverityColor(
                              incident.severity
                            )}`}
                          >
                            {incident.severity.toUpperCase()}
                          </span>
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap">
                          <span
                            className={`px-2 py-1 inline-flex text-xs leading-5 font-semibold rounded-full border ${getStatusColor(
                              incident.status
                            )}`}
                          >
                            {incident.status.charAt(0).toUpperCase() +
                              incident.status.slice(1)}
                          </span>
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                          {new Date(incident.detected_at).toLocaleString()}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                          {incident.assignee?.name || '-'}
                        </td>
                        <td className="px-6 py-4">
                          <div className="flex flex-wrap gap-1 max-w-xs">
                            {incident.tags.map((tag) => (
                              <span
                                key={tag.id}
                                className="px-2 py-1 text-xs rounded-full text-white whitespace-nowrap"
                                style={{ backgroundColor: tag.color }}
                              >
                                {tag.name}
                              </span>
                            ))}
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>

                {/* Pagination */}
                <div className="bg-gray-50 px-4 py-3 flex items-center justify-between border-t border-gray-200 sm:px-6">
                  <div className="flex-1 flex justify-between sm:hidden">
                    <button
                      onClick={() =>
                        setPagination({ ...pagination, page: Math.max(pagination.page - 1, 1) })
                      }
                      disabled={pagination.page === 1}
                      className="relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 disabled:opacity-50"
                    >
                      Previous
                    </button>
                    <button
                      onClick={() =>
                        setPagination({
                          ...pagination,
                          page: Math.min(pagination.page + 1, pagination.total_pages),
                        })
                      }
                      disabled={pagination.page === pagination.total_pages}
                      className="ml-3 relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 disabled:opacity-50"
                    >
                      Next
                    </button>
                  </div>
                  <div className="hidden sm:flex-1 sm:flex sm:items-center sm:justify-between">
                    <div>
                      <p className="text-sm text-gray-700">
                        Showing page <span className="font-medium">{pagination.page}</span> of{' '}
                        <span className="font-medium">{pagination.total_pages}</span> (
                        <span className="font-medium">{pagination.total}</span> total)
                      </p>
                    </div>
                    <div>
                      <nav className="relative z-0 inline-flex rounded-md shadow-sm -space-x-px">
                        <button
                          onClick={() =>
                            setPagination({ ...pagination, page: Math.max(pagination.page - 1, 1) })
                          }
                          disabled={pagination.page === 1}
                          className="relative inline-flex items-center px-2 py-2 rounded-l-md border border-gray-300 bg-white text-sm font-medium text-gray-500 hover:bg-gray-50 disabled:opacity-50"
                        >
                          Previous
                        </button>
                        <button
                          onClick={() =>
                            setPagination({
                              ...pagination,
                              page: Math.min(pagination.page + 1, pagination.total_pages),
                            })
                          }
                          disabled={pagination.page === pagination.total_pages}
                          className="relative inline-flex items-center px-2 py-2 rounded-r-md border border-gray-300 bg-white text-sm font-medium text-gray-500 hover:bg-gray-50 disabled:opacity-50"
                        >
                          Next
                        </button>
                      </nav>
                    </div>
                  </div>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}

export default function IncidentsPage() {
  return (
    <Suspense fallback={<div className="min-h-screen flex items-center justify-center"><div className="text-gray-600">Loading...</div></div>}>
      <IncidentsPageContent />
    </Suspense>
  );
}
