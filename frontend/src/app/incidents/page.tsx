'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from '@/context/AuthContext';
import { incidentApi, tagApi, exportApi } from '@/lib/api';
import { Incident, Severity, Status, PaginationResult } from '@/types/incident';
import { Tag } from '@/types/tag';
import SeverityGuide from '@/components/SeverityGuide';

export default function IncidentsPage() {
  const { token, loading: authLoading } = useAuth();
  const router = useRouter();
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
    setPagination({ ...pagination, page: 1 }); // Reset to first page
  };

  const handleTagToggle = (tagId: number) => {
    setSelectedTagIds((prev) =>
      prev.includes(tagId) ? prev.filter((id) => id !== tagId) : [...prev, tagId]
    );
    setPagination({ ...pagination, page: 1 }); // Reset to first page
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

      // Create download link
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
            <button
              onClick={() => router.push('/incidents/create')}
              className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
            >
              新規作成
            </button>
          </div>
        </div>

        {/* Severity Guide */}
        <SeverityGuide />

        {/* Filters */}
        <div className="bg-white p-4 rounded-lg shadow mb-6">
          <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
            {/* Search */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Search
              </label>
              <input
                type="text"
                value={search}
                onChange={handleSearchChange}
                placeholder="Search title or description..."
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 text-gray-900 placeholder-gray-500"
              />
            </div>

            {/* Severity Filter */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Severity
              </label>
              <select
                value={severity}
                onChange={(e) => {
                  setSeverity(e.target.value as Severity | '');
                  setPagination({ ...pagination, page: 1 });
                }}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 text-gray-900"
              >
                <option value="">All</option>
                <option value="critical">Critical</option>
                <option value="high">High</option>
                <option value="medium">Medium</option>
                <option value="low">Low</option>
              </select>
            </div>

            {/* Status Filter */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Status
              </label>
              <select
                value={status}
                onChange={(e) => {
                  setStatus(e.target.value as Status | '');
                  setPagination({ ...pagination, page: 1 });
                }}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 text-gray-900"
              >
                <option value="">All</option>
                <option value="open">Open</option>
                <option value="investigating">Investigating</option>
                <option value="resolved">Resolved</option>
                <option value="closed">Closed</option>
              </select>
            </div>

            {/* Tag Filter */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Tags
              </label>
              <div className="flex flex-wrap gap-2">
                {tags.map((tag) => (
                  <button
                    key={tag.id}
                    onClick={() => handleTagToggle(tag.id)}
                    className={`px-3 py-1 rounded-full text-sm border ${
                      selectedTagIds.includes(tag.id)
                        ? 'bg-blue-100 border-blue-500 text-blue-800'
                        : 'bg-white border-gray-300 text-gray-700'
                    }`}
                    style={
                      selectedTagIds.includes(tag.id)
                        ? { borderColor: tag.color }
                        : undefined
                    }
                  >
                    {tag.name}
                  </button>
                ))}
              </div>
            </div>
          </div>
        </div>

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
            No incidents found. Create your first incident!
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
                    className="hover:bg-gray-50 cursor-pointer"
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
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="flex flex-wrap gap-1">
                        {incident.tags.slice(0, 3).map((tag) => (
                          <span
                            key={tag.id}
                            className="px-2 py-1 text-xs rounded-full text-white"
                            style={{ backgroundColor: tag.color }}
                          >
                            {tag.name}
                          </span>
                        ))}
                        {incident.tags.length > 3 && (
                          <span className="px-2 py-1 text-xs rounded-full bg-gray-200 text-gray-700">
                            +{incident.tags.length - 3}
                          </span>
                        )}
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
  );
}
