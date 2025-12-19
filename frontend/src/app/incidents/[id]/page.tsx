'use client';

import { useEffect, useState } from 'react';
import { useRouter, useParams } from 'next/navigation';
import { useAuth } from '@/context/AuthContext';
import { incidentApi } from '@/lib/api';
import { Incident, Severity, Status } from '@/types/incident';

export default function IncidentDetailPage() {
  const { token, user, loading: authLoading } = useAuth();
  const router = useRouter();
  const params = useParams();
  const id = params.id as string;
  const [incident, setIncident] = useState<Incident | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [deleting, setDeleting] = useState(false);

  useEffect(() => {
    if (!authLoading && !token) {
      router.push('/login');
    }
  }, [token, authLoading, router]);

  useEffect(() => {
    if (token && id) {
      fetchIncident();
    }
  }, [token, id]);

  const fetchIncident = async () => {
    setLoading(true);
    setError('');
    try {
      const data = await incidentApi.getById(token!, parseInt(id));
      setIncident(data);
    } catch (err: any) {
      setError(err.message || 'Failed to fetch incident');
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async () => {
    if (!confirm('Are you sure you want to delete this incident?')) {
      return;
    }

    setDeleting(true);
    try {
      await incidentApi.delete(token!, parseInt(id));
      router.push('/incidents');
    } catch (err: any) {
      alert(err.message || 'Failed to delete incident');
    } finally {
      setDeleting(false);
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

  const canEdit = () => {
    if (!user || !incident) return false;
    return user.role === 'admin' || (user.role === 'editor' && incident.creator_id === user.id);
  };

  const canDelete = () => {
    return user?.role === 'admin';
  };

  if (authLoading || !token) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-gray-600">Loading...</div>
      </div>
    );
  }

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-gray-600">Loading incident...</div>
      </div>
    );
  }

  if (error || !incident) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-red-600">{error || 'Incident not found'}</div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50 py-8 px-4 sm:px-6 lg:px-8">
      <div className="max-w-4xl mx-auto">
        {/* Header */}
        <div className="mb-6">
          <button
            onClick={() => router.push('/incidents')}
            className="text-blue-600 hover:text-blue-800 mb-4 inline-flex items-center"
          >
            ← Back to List
          </button>

          <div className="flex justify-between items-start">
            <div>
              <h1 className="text-3xl font-bold text-gray-900 mb-2">{incident.title}</h1>
              <div className="flex gap-2">
                <span
                  className={`px-3 py-1 inline-flex text-sm font-semibold rounded-full border ${getSeverityColor(
                    incident.severity
                  )}`}
                >
                  {incident.severity.toUpperCase()}
                </span>
                <span
                  className={`px-3 py-1 inline-flex text-sm font-semibold rounded-full border ${getStatusColor(
                    incident.status
                  )}`}
                >
                  {incident.status.charAt(0).toUpperCase() + incident.status.slice(1)}
                </span>
              </div>
            </div>

            <div className="flex gap-2">
              {canEdit() && (
                <button
                  onClick={() => router.push(`/incidents/${id}/edit`)}
                  className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
                >
                  Edit
                </button>
              )}
              {canDelete() && (
                <button
                  onClick={handleDelete}
                  disabled={deleting}
                  className="px-4 py-2 bg-red-600 text-white rounded-md hover:bg-red-700 disabled:opacity-50"
                >
                  {deleting ? 'Deleting...' : 'Delete'}
                </button>
              )}
            </div>
          </div>
        </div>

        {/* Metadata Section */}
        <div className="bg-white rounded-lg shadow p-6 mb-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Metadata</h2>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <p className="text-sm font-medium text-gray-500">Detected At</p>
              <p className="mt-1 text-sm text-gray-900">
                {new Date(incident.detected_at).toLocaleString()}
              </p>
            </div>
            <div>
              <p className="text-sm font-medium text-gray-500">Resolved At</p>
              <p className="mt-1 text-sm text-gray-900">
                {incident.resolved_at
                  ? new Date(incident.resolved_at).toLocaleString()
                  : 'Not resolved yet'}
              </p>
            </div>
            <div>
              <p className="text-sm font-medium text-gray-500">Impact Scope</p>
              <p className="mt-1 text-sm text-gray-900">{incident.impact_scope || '-'}</p>
            </div>
            <div>
              <p className="text-sm font-medium text-gray-500">Assignee</p>
              <p className="mt-1 text-sm text-gray-900">
                {incident.assignee
                  ? `${incident.assignee.name} (${incident.assignee.email})`
                  : 'Unassigned'}
              </p>
            </div>
            <div>
              <p className="text-sm font-medium text-gray-500">Creator</p>
              <p className="mt-1 text-sm text-gray-900">
                {incident.creator.name} ({incident.creator.email})
              </p>
            </div>
            <div>
              <p className="text-sm font-medium text-gray-500">Created At</p>
              <p className="mt-1 text-sm text-gray-900">
                {new Date(incident.created_at).toLocaleString()}
              </p>
            </div>
          </div>
        </div>

        {/* Description Section */}
        <div className="bg-white rounded-lg shadow p-6 mb-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Description</h2>
          <p className="text-sm text-gray-700 whitespace-pre-wrap">{incident.description}</p>
        </div>

        {/* Summary Section */}
        <div className="bg-white rounded-lg shadow p-6 mb-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">AI Summary</h2>
          {incident.summary ? (
            <p className="text-sm text-gray-700">{incident.summary}</p>
          ) : (
            <p className="text-sm text-gray-500 italic">
              AI summary will be available in Phase 1C
            </p>
          )}
        </div>

        {/* Tags Section */}
        <div className="bg-white rounded-lg shadow p-6 mb-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Tags</h2>
          {incident.tags.length > 0 ? (
            <div className="flex flex-wrap gap-2">
              {incident.tags.map((tag) => (
                <span
                  key={tag.id}
                  className="px-3 py-1 rounded-full text-white text-sm"
                  style={{ backgroundColor: tag.color }}
                >
                  {tag.name}
                </span>
              ))}
            </div>
          ) : (
            <p className="text-sm text-gray-500">No tags assigned</p>
          )}
        </div>

        {/* Timeline Section - Placeholder */}
        <div className="bg-white rounded-lg shadow p-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Timeline</h2>
          <p className="text-sm text-gray-500 italic">
            Timeline feature will be available in Phase 1B
          </p>
        </div>
      </div>
    </div>
  );
}
