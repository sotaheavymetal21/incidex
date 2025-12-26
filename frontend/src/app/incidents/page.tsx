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

  const getSeverityStyle = (severity: Severity) => {
    switch (severity) {
      case 'critical': return { background: 'var(--critical-light)', color: 'var(--critical)', borderColor: 'var(--critical)' };
      case 'high': return { background: 'var(--high-light)', color: 'var(--high)', borderColor: 'var(--high)' };
      case 'medium': return { background: 'var(--medium-light)', color: 'var(--medium)', borderColor: 'var(--medium)' };
      case 'low': return { background: 'var(--low-light)', color: 'var(--low)', borderColor: 'var(--low)' };
      default: return { background: 'var(--gray-100)', color: 'var(--gray-700)', borderColor: 'var(--gray-300)' };
    }
  };

  const getStatusStyle = (status: Status) => {
    switch (status) {
      case 'open': return { background: 'var(--gray-100)', color: 'var(--gray-700)', borderColor: 'var(--gray-400)' };
      case 'investigating': return { background: 'var(--info-light)', color: 'var(--info)', borderColor: 'var(--info)' };
      case 'resolved': return { background: 'var(--success-light)', color: 'var(--success)', borderColor: 'var(--success)' };
      case 'closed': return { background: 'var(--secondary-light)', color: 'var(--secondary-dark)', borderColor: 'var(--secondary)' };
      default: return { background: 'var(--gray-100)', color: 'var(--gray-700)', borderColor: 'var(--gray-300)' };
    }
  };

  // 旧関数との互換性のため残す（使っている箇所があれば）
  const getSeverityColor = (severity: Severity) => {
    return '';  // style属性に置き換えるため、classNameは空文字列
  };

  const getStatusColor = (status: Status) => {
    return '';  // style属性に置き換えるため、classNameは空文字列
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
      <div className="min-h-screen flex items-center justify-center" style={{ background: 'var(--background)' }}>
        <div style={{ color: 'var(--secondary)' }}>Loading...</div>
      </div>
    );
  }

  return (
    <div className="min-h-screen py-8 px-4 sm:px-6 lg:px-8" style={{ background: 'var(--background)' }}>
      <div className="max-w-7xl mx-auto">
        {/* Header */}
        <div className="mb-8 flex justify-between items-center">
          <div>
            <h1 className="text-3xl font-bold" style={{ color: 'var(--foreground)' }}>インシデント一覧</h1>
            <p className="text-sm mt-1" style={{ color: 'var(--secondary)' }}>インシデントの管理と追跡</p>
          </div>
          <div className="flex space-x-3">
            <button
              onClick={handleExportCSV}
              className="px-4 py-2.5 text-white rounded-lg flex items-center shadow-md transition-all"
              style={{ background: 'var(--accent)' }}
              onMouseEnter={(e) => e.currentTarget.style.opacity = '0.9'}
              onMouseLeave={(e) => e.currentTarget.style.opacity = '1'}
            >
              <svg className="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
              </svg>
              CSVエクスポート
            </button>
            {permissions.canCreateIncidents && (
              <button
                onClick={() => router.push('/incidents/create')}
                className="px-4 py-2.5 text-white rounded-lg shadow-lg transition-all"
                style={{ background: 'var(--primary)' }}
                onMouseEnter={(e) => e.currentTarget.style.background = 'var(--primary-hover)'}
                onMouseLeave={(e) => e.currentTarget.style.background = 'var(--primary)'}
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
              <div className="rounded-xl shadow-lg p-5 sticky top-8 border" style={{
                background: 'var(--surface)',
                borderColor: 'var(--border)'
              }}>
                <div className="flex items-center justify-between mb-5">
                  <h2 className="text-lg font-semibold" style={{ color: 'var(--foreground)' }}>フィルター</h2>
                  <button
                    onClick={() => setShowSidebar(false)}
                    className="lg:hidden transition-colors"
                    style={{ color: 'var(--secondary)' }}
                    onMouseEnter={(e) => e.currentTarget.style.color = 'var(--foreground)'}
                    onMouseLeave={(e) => e.currentTarget.style.color = 'var(--secondary)'}
                  >
                    <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                    </svg>
                  </button>
                </div>

                {/* Filter Presets */}
                <div className="mb-6">
                  <h3 className="text-sm font-medium mb-3" style={{ color: 'var(--foreground)' }}>クイックフィルター</h3>
                  <div className="space-y-2">
                    <button
                      onClick={() => applyPreset('unresolved')}
                      className="w-full px-3 py-2 text-sm text-left rounded-lg transition-all"
                      style={{ background: 'var(--secondary-light)', color: 'var(--foreground)' }}
                      onMouseEnter={(e) => e.currentTarget.style.background = 'var(--primary-light)'}
                      onMouseLeave={(e) => e.currentTarget.style.background = 'var(--secondary-light)'}
                    >
                      未解決のインシデント
                    </button>
                    <button
                      onClick={() => applyPreset('critical')}
                      className="w-full px-3 py-2 text-sm text-left rounded-lg transition-all"
                      style={{ background: 'var(--secondary-light)', color: 'var(--foreground)' }}
                      onMouseEnter={(e) => e.currentTarget.style.background = 'var(--primary-light)'}
                      onMouseLeave={(e) => e.currentTarget.style.background = 'var(--secondary-light)'}
                    >
                      Critical のみ
                    </button>
                    <button
                      onClick={clearFilters}
                      className="w-full px-3 py-2 text-sm text-left rounded-lg border transition-all"
                      style={{ borderColor: 'var(--border)', color: 'var(--secondary)' }}
                      onMouseEnter={(e) => {
                        e.currentTarget.style.background = 'var(--error-light)';
                        e.currentTarget.style.borderColor = 'var(--error)';
                        e.currentTarget.style.color = 'var(--error)';
                      }}
                      onMouseLeave={(e) => {
                        e.currentTarget.style.background = 'transparent';
                        e.currentTarget.style.borderColor = 'var(--border)';
                        e.currentTarget.style.color = 'var(--secondary)';
                      }}
                    >
                      すべてクリア
                    </button>
                  </div>
                </div>

                {/* Search */}
                <div className="mb-6">
                  <label className="block text-sm font-medium mb-2" style={{ color: 'var(--foreground)' }}>
                    検索
                  </label>
                  <input
                    type="text"
                    value={search}
                    onChange={handleSearchChange}
                    placeholder="タイトル、説明を検索..."
                    className="w-full px-3 py-2.5 border-2 rounded-lg focus:outline-none focus:ring-2 transition-all"
                    style={{
                      background: 'var(--surface)',
                      borderColor: 'var(--border)',
                      color: 'var(--foreground)'
                    }}
                    onFocus={(e) => {
                      e.currentTarget.style.borderColor = 'var(--primary)';
                      e.currentTarget.style.boxShadow = '0 0 0 3px var(--primary-light)';
                    }}
                    onBlur={(e) => {
                      e.currentTarget.style.borderColor = 'var(--border)';
                      e.currentTarget.style.boxShadow = 'none';
                    }}
                  />
                </div>

                {/* Severity Filter */}
                <div className="mb-6">
                  <label className="block text-sm font-medium mb-2" style={{ color: 'var(--foreground)' }}>
                    深刻度
                  </label>
                  <div className="space-y-2">
                    {(Object.keys(SEVERITY_LABELS) as Severity[]).map((sev) => (
                      <label key={sev} className="flex items-center cursor-pointer">
                        <input
                          type="radio"
                          name="severity"
                          checked={severity === sev}
                          onChange={() => {
                            setSeverity(sev);
                            setPagination({ ...pagination, page: 1 });
                          }}
                          className="h-4 w-4 accent-[--primary] border-2"
                          style={{ borderColor: 'var(--border)' }}
                        />
                        <span className="ml-2 text-sm" style={{ color: 'var(--foreground)' }}>{SEVERITY_LABELS[sev]}</span>
                      </label>
                    ))}
                    <label className="flex items-center cursor-pointer">
                      <input
                        type="radio"
                        name="severity"
                        checked={severity === ''}
                        onChange={() => {
                          setSeverity('');
                          setPagination({ ...pagination, page: 1 });
                        }}
                        className="h-4 w-4 accent-[--primary] border-2"
                        style={{ borderColor: 'var(--border)' }}
                      />
                      <span className="ml-2 text-sm" style={{ color: 'var(--foreground)' }}>すべて</span>
                    </label>
                  </div>
                </div>

                {/* Status Filter */}
                <div className="mb-6">
                  <label className="block text-sm font-medium mb-2" style={{ color: 'var(--foreground)' }}>
                    ステータス
                  </label>
                  <div className="space-y-2">
                    {(Object.keys(STATUS_LABELS) as Status[]).map((st) => (
                      <label key={st} className="flex items-center cursor-pointer">
                        <input
                          type="radio"
                          name="status"
                          checked={status === st}
                          onChange={() => {
                            setStatus(st);
                            setPagination({ ...pagination, page: 1 });
                          }}
                          className="h-4 w-4 accent-[--primary] border-2"
                          style={{ borderColor: 'var(--border)' }}
                        />
                        <span className="ml-2 text-sm" style={{ color: 'var(--foreground)' }}>{STATUS_LABELS[st]}</span>
                      </label>
                    ))}
                    <label className="flex items-center cursor-pointer">
                      <input
                        type="radio"
                        name="status"
                        checked={status === ''}
                        onChange={() => {
                          setStatus('');
                          setPagination({ ...pagination, page: 1 });
                        }}
                        className="h-4 w-4 accent-[--primary] border-2"
                        style={{ borderColor: 'var(--border)' }}
                      />
                      <span className="ml-2 text-sm" style={{ color: 'var(--foreground)' }}>すべて</span>
                    </label>
                  </div>
                </div>

                {/* Tag Filter */}
                <div>
                  <label className="block text-sm font-medium mb-2" style={{ color: 'var(--foreground)' }}>
                    タグ
                  </label>
                  <div className="space-y-2">
                    {tags.map((tag) => (
                      <label key={tag.id} className="flex items-center cursor-pointer">
                        <input
                          type="checkbox"
                          checked={selectedTagIds.includes(tag.id)}
                          onChange={() => handleTagToggle(tag.id)}
                          className="h-4 w-4 rounded accent-[--primary] border-2"
                          style={{ borderColor: 'var(--border)' }}
                        />
                        <span className="ml-2 text-sm flex-1" style={{ color: 'var(--foreground)' }}>{tag.name}</span>
                        <span
                          className="w-3 h-3 rounded-full shadow-sm"
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
                className="mb-4 px-4 py-2 rounded-lg flex items-center border transition-all"
                style={{ background: 'var(--surface)', borderColor: 'var(--border)', color: 'var(--foreground)' }}
                onMouseEnter={(e) => e.currentTarget.style.background = 'var(--secondary-light)'}
                onMouseLeave={(e) => e.currentTarget.style.background = 'var(--surface)'}
              >
                <svg className="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z" />
                </svg>
                フィルターを表示
              </button>
            )}

            {/* Filter Chips */}
            {hasActiveFilters && (
              <div className="mb-4 p-4 rounded-xl shadow-lg border" style={{ background: 'var(--surface)', borderColor: 'var(--border)' }}>
                <div className="flex items-center justify-between mb-2">
                  <h3 className="text-sm font-medium" style={{ color: 'var(--foreground)' }}>適用中のフィルター:</h3>
                  <button
                    onClick={clearFilters}
                    className="text-sm transition-colors"
                    style={{ color: 'var(--primary)' }}
                    onMouseEnter={(e) => e.currentTarget.style.color = 'var(--primary-hover)'}
                    onMouseLeave={(e) => e.currentTarget.style.color = 'var(--primary)'}
                  >
                    すべてクリア
                  </button>
                </div>
                <div className="flex flex-wrap gap-2">
                  {search && (
                    <span className="inline-flex items-center px-3 py-1.5 rounded-full text-sm border-2" style={{ background: 'var(--info-light)', color: 'var(--info)', borderColor: 'var(--info)' }}>
                      検索: {search}
                      <button
                        onClick={() => clearFilter('search')}
                        className="ml-2 transition-opacity"
                        onMouseEnter={(e) => e.currentTarget.style.opacity = '0.7'}
                        onMouseLeave={(e) => e.currentTarget.style.opacity = '1'}
                      >
                        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                        </svg>
                      </button>
                    </span>
                  )}
                  {severity && (
                    <span className="inline-flex items-center px-3 py-1.5 rounded-full text-sm border-2" style={getSeverityStyle(severity)}>
                      深刻度: {SEVERITY_LABELS[severity]}
                      <button
                        onClick={() => clearFilter('severity')}
                        className="ml-2 transition-opacity"
                        onMouseEnter={(e) => e.currentTarget.style.opacity = '0.7'}
                        onMouseLeave={(e) => e.currentTarget.style.opacity = '1'}
                      >
                        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                        </svg>
                      </button>
                    </span>
                  )}
                  {status && (
                    <span className="inline-flex items-center px-3 py-1.5 rounded-full text-sm border-2" style={getStatusStyle(status)}>
                      ステータス: {STATUS_LABELS[status]}
                      <button
                        onClick={() => clearFilter('status')}
                        className="ml-2 transition-opacity"
                        onMouseEnter={(e) => e.currentTarget.style.opacity = '0.7'}
                        onMouseLeave={(e) => e.currentTarget.style.opacity = '1'}
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
                        className="inline-flex items-center px-3 py-1.5 rounded-full text-sm text-white shadow-sm"
                        style={{ backgroundColor: tag.color }}
                      >
                        {tag.name}
                        <button
                          onClick={() => clearFilter('tag', tagId)}
                          className="ml-2 text-white transition-opacity"
                          onMouseEnter={(e) => e.currentTarget.style.opacity = '0.7'}
                          onMouseLeave={(e) => e.currentTarget.style.opacity = '1'}
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
              <div className="px-4 py-3 rounded-xl mb-4 border-2" style={{ background: 'var(--error-light)', borderColor: 'var(--error)', color: 'var(--error)' }}>
                {error}
              </div>
            )}

            {/* Incidents Table */}
            {loading ? (
              <div className="text-center py-8" style={{ color: 'var(--secondary)' }}>Loading incidents...</div>
            ) : incidents.length === 0 ? (
              <div className="text-center py-12 px-4 rounded-xl shadow-lg border" style={{ background: 'var(--surface)', borderColor: 'var(--border)', color: 'var(--secondary)' }}>
                <svg className="mx-auto h-12 w-12 mb-4" style={{ color: 'var(--secondary)' }} fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                </svg>
                <p className="text-base font-medium" style={{ color: 'var(--foreground)' }}>インシデントが見つかりませんでした</p>
                <p className="text-sm mt-1">フィルターを変更するか、新しいインシデントを作成してください。</p>
              </div>
            ) : (
              <div className="rounded-xl shadow-lg overflow-hidden border" style={{ background: 'var(--surface)', borderColor: 'var(--border)' }}>
                <table className="min-w-full">
                  <thead style={{ background: 'var(--secondary-light)' }}>
                    <tr>
                      <th className="px-6 py-3.5 text-left text-xs font-semibold uppercase tracking-wider" style={{ color: 'var(--foreground)' }}>
                        Title
                      </th>
                      <th className="px-6 py-3.5 text-left text-xs font-semibold uppercase tracking-wider" style={{ color: 'var(--foreground)' }}>
                        Severity
                      </th>
                      <th className="px-6 py-3.5 text-left text-xs font-semibold uppercase tracking-wider" style={{ color: 'var(--foreground)' }}>
                        Status
                      </th>
                      <th className="px-6 py-3.5 text-left text-xs font-semibold uppercase tracking-wider" style={{ color: 'var(--foreground)' }}>
                        Detected At
                      </th>
                      <th className="px-6 py-3.5 text-left text-xs font-semibold uppercase tracking-wider" style={{ color: 'var(--foreground)' }}>
                        Assignee
                      </th>
                      <th className="px-6 py-3.5 text-left text-xs font-semibold uppercase tracking-wider" style={{ color: 'var(--foreground)' }}>
                        Tags
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    {incidents.map((incident, index) => (
                      <tr
                        key={incident.id}
                        onClick={() => router.push(`/incidents/${incident.id}`)}
                        className="cursor-pointer transition-all"
                        style={{
                          borderTop: index > 0 ? `1px solid var(--border)` : 'none'
                        }}
                        onMouseEnter={(e) => e.currentTarget.style.background = 'var(--secondary-light)'}
                        onMouseLeave={(e) => e.currentTarget.style.background = 'transparent'}
                      >
                        <td className="px-6 py-4">
                          <div className="text-sm font-medium" style={{ color: 'var(--foreground)' }}>
                            {incident.title}
                          </div>
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap">
                          <span
                            className="px-2.5 py-1 inline-flex text-xs leading-5 font-semibold rounded-full border-2"
                            style={getSeverityStyle(incident.severity)}
                          >
                            {incident.severity.toUpperCase()}
                          </span>
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap">
                          <span
                            className="px-2.5 py-1 inline-flex text-xs leading-5 font-semibold rounded-full border-2"
                            style={getStatusStyle(incident.status)}
                          >
                            {incident.status.charAt(0).toUpperCase() +
                              incident.status.slice(1)}
                          </span>
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm" style={{ color: 'var(--secondary)' }}>
                          {new Date(incident.detected_at).toLocaleString()}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm" style={{ color: 'var(--secondary)' }}>
                          {incident.assignee?.name || '-'}
                        </td>
                        <td className="px-6 py-4">
                          <div className="flex flex-wrap gap-1.5 max-w-xs">
                            {incident.tags?.map((tag) => (
                              <span
                                key={tag.id}
                                className="px-2.5 py-1 text-xs rounded-full text-white whitespace-nowrap shadow-sm"
                                style={{ backgroundColor: tag.color }}
                              >
                                {tag.name}
                              </span>
                            )) || '-'}
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>

                {/* Pagination */}
                <div className="px-4 py-3.5 flex items-center justify-between border-t sm:px-6" style={{ background: 'var(--secondary-light)', borderColor: 'var(--border)' }}>
                  <div className="flex-1 flex justify-between sm:hidden">
                    <button
                      onClick={() =>
                        setPagination({ ...pagination, page: Math.max(pagination.page - 1, 1) })
                      }
                      disabled={pagination.page === 1}
                      className="relative inline-flex items-center px-4 py-2 border-2 text-sm font-medium rounded-lg transition-all disabled:opacity-40"
                      style={{
                        background: 'var(--surface)',
                        borderColor: 'var(--border)',
                        color: 'var(--foreground)'
                      }}
                      onMouseEnter={(e) => !e.currentTarget.disabled && (e.currentTarget.style.borderColor = 'var(--primary)')}
                      onMouseLeave={(e) => e.currentTarget.style.borderColor = 'var(--border)'}
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
                      className="ml-3 relative inline-flex items-center px-4 py-2 border-2 text-sm font-medium rounded-lg transition-all disabled:opacity-40"
                      style={{
                        background: 'var(--surface)',
                        borderColor: 'var(--border)',
                        color: 'var(--foreground)'
                      }}
                      onMouseEnter={(e) => !e.currentTarget.disabled && (e.currentTarget.style.borderColor = 'var(--primary)')}
                      onMouseLeave={(e) => e.currentTarget.style.borderColor = 'var(--border)'}
                    >
                      Next
                    </button>
                  </div>
                  <div className="hidden sm:flex-1 sm:flex sm:items-center sm:justify-between">
                    <div>
                      <p className="text-sm" style={{ color: 'var(--foreground)' }}>
                        Showing page <span className="font-semibold">{pagination.page}</span> of{' '}
                        <span className="font-semibold">{pagination.total_pages}</span> (
                        <span className="font-semibold">{pagination.total}</span> total)
                      </p>
                    </div>
                    <div>
                      <nav className="relative z-0 inline-flex rounded-lg shadow-sm gap-2">
                        <button
                          onClick={() =>
                            setPagination({ ...pagination, page: Math.max(pagination.page - 1, 1) })
                          }
                          disabled={pagination.page === 1}
                          className="relative inline-flex items-center px-4 py-2 rounded-lg border-2 text-sm font-medium transition-all disabled:opacity-40"
                          style={{
                            background: 'var(--surface)',
                            borderColor: 'var(--border)',
                            color: 'var(--foreground)'
                          }}
                          onMouseEnter={(e) => !e.currentTarget.disabled && (e.currentTarget.style.borderColor = 'var(--primary)')}
                          onMouseLeave={(e) => e.currentTarget.style.borderColor = 'var(--border)'}
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
                          className="relative inline-flex items-center px-4 py-2 rounded-lg border-2 text-sm font-medium transition-all disabled:opacity-40"
                          style={{
                            background: 'var(--surface)',
                            borderColor: 'var(--border)',
                            color: 'var(--foreground)'
                          }}
                          onMouseEnter={(e) => !e.currentTarget.disabled && (e.currentTarget.style.borderColor = 'var(--primary)')}
                          onMouseLeave={(e) => e.currentTarget.style.borderColor = 'var(--border)'}
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
    <Suspense fallback={<div className="min-h-screen flex items-center justify-center" style={{ background: 'var(--background)' }}><div style={{ color: 'var(--secondary)' }}>Loading...</div></div>}>
      <IncidentsPageContent />
    </Suspense>
  );
}
