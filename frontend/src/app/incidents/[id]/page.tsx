'use client';

import { useEffect, useState } from 'react';
import { useRouter, useParams } from 'next/navigation';
import { useAuth } from '@/context/AuthContext';
import { incidentApi, activityApi, attachmentApi, userApi } from '@/lib/api';
import { Incident, Severity, Status } from '@/types/incident';
import { IncidentActivity } from '@/types/activity';
import { Attachment } from '@/types/attachment';
import { User } from '@/types/user';
import Timeline from '@/components/Timeline';
import { usePermissions } from '@/hooks/usePermissions';

export default function IncidentDetailPage() {
  const { token, user, loading: authLoading } = useAuth();
  const permissions = usePermissions();
  const router = useRouter();
  const params = useParams();
  const id = params.id as string;

  const [incident, setIncident] = useState<Incident | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [deleting, setDeleting] = useState(false);
  const [activities, setActivities] = useState<IncidentActivity[]>([]);
  const [loadingActivities, setLoadingActivities] = useState(true);
  const [newComment, setNewComment] = useState('');
  const [submittingComment, setSubmittingComment] = useState(false);
  const [attachments, setAttachments] = useState<Attachment[]>([]);
  const [loadingAttachments, setLoadingAttachments] = useState(true);
  const [uploadingFile, setUploadingFile] = useState(false);
  const [regeneratingSummary, setRegeneratingSummary] = useState(false);
  const [showTimelineEventForm, setShowTimelineEventForm] = useState(false);
  const [entryType, setEntryType] = useState<'comment' | 'event'>('comment');
  const [timelineEventType, setTimelineEventType] = useState<'detected' | 'investigation_started' | 'root_cause_identified' | 'mitigation' | 'timeline_resolved' | 'other'>('other');
  const [timelineEventTime, setTimelineEventTime] = useState('');
  const [timelineEventDescription, setTimelineEventDescription] = useState('');
  const [submittingTimelineEvent, setSubmittingTimelineEvent] = useState(false);
  const [users, setUsers] = useState<User[]>([]);
  const [assigningUser, setAssigningUser] = useState(false);
  const [lightboxImage, setLightboxImage] = useState<string | null>(null);
  const [imageUrls, setImageUrls] = useState<Record<number, string>>({});

  useEffect(() => {
    if (!authLoading && !token) {
      router.push('/login');
    }
  }, [token, authLoading, router]);

  useEffect(() => {
    // Scroll to top when page loads
    window.scrollTo(0, 0);
  }, [id]);

  useEffect(() => {
    if (token && id) {
      fetchIncident();
      fetchActivities();
      fetchAttachments();
    }
  }, [token, id]);

  useEffect(() => {
    // Fetch users list for assignee selection
    if (token && permissions.canEdit) {
      fetchUsers();
    }
  }, [token, permissions.canEdit]);

  const fetchUsers = async () => {
    try {
      const data = await userApi.getAll(token!);
      setUsers(data);
    } catch (err: any) {
      console.error('Failed to fetch users:', err);
    }
  };

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

  const fetchActivities = async () => {
    setLoadingActivities(true);
    try {
      const data = await activityApi.getActivities(token!, parseInt(id));
      setActivities(data);
    } catch (err: any) {
      console.error('Failed to fetch activities:', err);
    } finally {
      setLoadingActivities(false);
    }
  };

  const fetchAttachments = async () => {
    setLoadingAttachments(true);
    try {
      const data = await attachmentApi.getAttachments(token!, parseInt(id));
      setAttachments(data);
    } catch (err: any) {
      console.error('Failed to fetch attachments:', err);
    } finally {
      setLoadingAttachments(false);
    }
  };

  // Fetch image blob URLs for attachments
  useEffect(() => {
    if (token && attachments.length > 0) {
      const fetchImageUrls = async () => {
        const urls: Record<number, string> = {};
        for (const attachment of attachments) {
          if (isImageFile(attachment.file_name)) {
            try {
              const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api';
              const response = await fetch(
                `${apiUrl}/incidents/${id}/attachments/${attachment.id}/download`,
                {
                  headers: {
                    'Authorization': `Bearer ${token}`,
                  },
                }
              );
              if (response.ok) {
                const blob = await response.blob();
                urls[attachment.id] = URL.createObjectURL(blob);
              }
            } catch (err) {
              console.error(`Failed to fetch image ${attachment.id}:`, err);
            }
          }
        }
        setImageUrls(urls);
      };
      fetchImageUrls();

      // Clean up object URLs on unmount
      return () => {
        Object.values(imageUrls).forEach(url => URL.revokeObjectURL(url));
      };
    }
  }, [token, attachments, id]);

  const handleFileUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    setUploadingFile(true);
    try {
      await attachmentApi.uploadAttachment(token!, parseInt(id), file);
      await fetchAttachments(); // Refresh attachments list
      // Reset file input
      e.target.value = '';
    } catch (err: any) {
      alert(err.message || 'Failed to upload file');
    } finally {
      setUploadingFile(false);
    }
  };

  const handleFileDownload = async (attachmentId: number, fileName: string) => {
    try {
      await attachmentApi.downloadAttachment(token!, parseInt(id), attachmentId, fileName);
    } catch (err: any) {
      alert(err.message || 'Failed to download file');
    }
  };

  const handleFileDelete = async (attachmentId: number) => {
    if (!confirm('このファイルを削除しますか？')) {
      return;
    }

    try {
      await attachmentApi.deleteAttachment(token!, parseInt(id), attachmentId);
      await fetchAttachments(); // Refresh attachments list
    } catch (err: any) {
      alert(err.message || 'Failed to delete file');
    }
  };

  const handleAddComment = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newComment.trim()) return;

    setSubmittingComment(true);
    try {
      await activityApi.addComment(token!, parseInt(id), { comment: newComment });
      setNewComment('');
      setShowTimelineEventForm(false); // フォームを閉じる
      await fetchActivities(); // アクティビティを再取得
    } catch (err: any) {
      alert(err.message || 'Failed to add comment');
    } finally {
      setSubmittingComment(false);
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

  const handleRegenerateSummary = async () => {
    if (!confirm('AI要約を再生成しますか？')) {
      return;
    }

    setRegeneratingSummary(true);
    try {
      const result = await incidentApi.regenerateSummary(token!, parseInt(id));
      // Update incident summary
      if (incident) {
        setIncident({ ...incident, summary: result.summary });
      }
    } catch (err: any) {
      alert(err.message || '要約の再生成に失敗しました');
    } finally {
      setRegeneratingSummary(false);
    }
  };

  const handleAddTimelineEvent = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!timelineEventDescription.trim()) return;

    setSubmittingTimelineEvent(true);
    try {
      await activityApi.addTimelineEvent(token!, parseInt(id), {
        event_type: timelineEventType,
        event_time: timelineEventTime ? new Date(timelineEventTime).toISOString() : new Date().toISOString(),
        description: timelineEventDescription,
      });
      setTimelineEventType('other');
      setTimelineEventTime('');
      setTimelineEventDescription('');
      setShowTimelineEventForm(false);
      await fetchActivities(); // Refresh activities
    } catch (err: any) {
      alert(err.message || 'タイムラインイベントの追加に失敗しました');
    } finally {
      setSubmittingTimelineEvent(false);
    }
  };

  const handleAssignIncident = async (assigneeId: number | null) => {
    setAssigningUser(true);
    try {
      const updatedIncident = await incidentApi.assignIncident(token!, parseInt(id), assigneeId);
      // Update incident state with the full response
      setIncident(updatedIncident);
      await fetchActivities(); // Refresh activities to show assignment change
      await fetchIncident(); // Refresh incident to ensure all data is up to date
    } catch (err: any) {
      alert(err.message || '担当者の変更に失敗しました');
    } finally {
      setAssigningUser(false);
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

  const formatFileSize = (bytes: number): string => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + ' ' + sizes[i];
  };

  const isImageFile = (fileName: string): boolean => {
    const imageExtensions = ['.jpg', '.jpeg', '.png', '.gif', '.webp', '.bmp', '.svg'];
    const extension = fileName.toLowerCase().substring(fileName.lastIndexOf('.'));
    return imageExtensions.includes(extension);
  };

  const getImageUrl = (attachmentId: number, fileName: string): string => {
    const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api';
    return `${apiUrl}/incidents/${id}/attachments/${attachmentId}/download`;
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
    <div className="min-h-screen py-8 px-4 sm:px-6 lg:px-8" style={{ background: 'var(--background)' }}>
      <div className="max-w-6xl mx-auto">
        {/* Header */}
        <div className="mb-6 animate-slideDown">
          <button
            onClick={() => router.push('/incidents')}
            className="mb-4 inline-flex items-center font-semibold transition-all"
            style={{ color: 'var(--primary)', fontFamily: 'var(--font-body)' }}
            onMouseEnter={(e) => {
              e.currentTarget.style.color = 'var(--primary-hover)';
              e.currentTarget.style.transform = 'translateX(-4px)';
            }}
            onMouseLeave={(e) => {
              e.currentTarget.style.color = 'var(--primary)';
              e.currentTarget.style.transform = 'translateX(0)';
            }}
          >
            ← インシデント一覧に戻る
          </button>

          <div className="flex flex-col lg:flex-row lg:justify-between lg:items-start gap-4">
            <div className="flex-1">
              <h1
                className="text-3xl md:text-4xl font-bold mb-3"
                style={{
                  color: 'var(--foreground)',
                  fontFamily: 'var(--font-display)',
                }}
              >
                {incident.title}
              </h1>
              <div className="flex flex-wrap gap-2">
                <span
                  className={`px-3 py-1.5 inline-flex text-sm font-bold rounded-full border-2`}
                  style={{
                    background: getSeverityStyle(incident.severity).background,
                    color: getSeverityStyle(incident.severity).color,
                    borderColor: getSeverityStyle(incident.severity).borderColor,
                    fontFamily: 'var(--font-body)'
                  }}
                >
                  {incident.severity.toUpperCase()}
                </span>
                <span
                  className={`px-3 py-1.5 inline-flex text-sm font-bold rounded-full border-2`}
                  style={{
                    background: getStatusStyle(incident.status).background,
                    color: getStatusStyle(incident.status).color,
                    borderColor: getStatusStyle(incident.status).borderColor,
                    fontFamily: 'var(--font-body)'
                  }}
                >
                  {incident.status.charAt(0).toUpperCase() + incident.status.slice(1)}
                </span>
              </div>
            </div>

            <div className="flex flex-wrap gap-2">
              <button
                onClick={() => router.push(`/incidents/${id}/postmortem`)}
                className="px-4 py-2.5 text-white rounded-xl font-semibold transition-all duration-200"
                style={{
                  background: 'linear-gradient(135deg, var(--accent) 0%, var(--accent-hover) 100%)',
                  fontFamily: 'var(--font-body)',
                  boxShadow: '0 4px 12px var(--accent-glow)'
                }}
                onMouseEnter={(e) => {
                  e.currentTarget.style.transform = 'translateY(-2px)';
                  e.currentTarget.style.boxShadow = '0 8px 20px var(--accent-glow)';
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.transform = 'translateY(0)';
                  e.currentTarget.style.boxShadow = '0 4px 12px var(--accent-glow)';
                }}
              >
                Post-Mortem
              </button>
              {canEdit() && (
                <button
                  onClick={() => router.push(`/incidents/${id}/edit`)}
                  className="px-4 py-2.5 text-white rounded-xl font-semibold transition-all duration-200"
                  style={{
                    background: 'linear-gradient(135deg, var(--primary) 0%, var(--primary-dark) 100%)',
                    fontFamily: 'var(--font-body)',
                    boxShadow: '0 4px 12px var(--primary-glow)'
                  }}
                  onMouseEnter={(e) => {
                    e.currentTarget.style.transform = 'translateY(-2px)';
                    e.currentTarget.style.boxShadow = '0 8px 20px var(--primary-glow)';
                  }}
                  onMouseLeave={(e) => {
                    e.currentTarget.style.transform = 'translateY(0)';
                    e.currentTarget.style.boxShadow = '0 4px 12px var(--primary-glow)';
                  }}
                >
                  編集
                </button>
              )}
              {canDelete() && (
                <button
                  onClick={handleDelete}
                  disabled={deleting}
                  className="px-4 py-2.5 text-white rounded-xl font-semibold transition-all duration-200 disabled:opacity-50"
                  style={{
                    background: 'linear-gradient(135deg, var(--error) 0%, var(--error-dark) 100%)',
                    fontFamily: 'var(--font-body)',
                    boxShadow: '0 4px 12px var(--error-glow)'
                  }}
                  onMouseEnter={(e) => {
                    if (!deleting) {
                      e.currentTarget.style.transform = 'translateY(-2px)';
                      e.currentTarget.style.boxShadow = '0 8px 20px var(--error-glow)';
                    }
                  }}
                  onMouseLeave={(e) => {
                    e.currentTarget.style.transform = 'translateY(0)';
                    e.currentTarget.style.boxShadow = '0 4px 12px var(--error-glow)';
                  }}
                >
                  {deleting ? '削除中...' : '削除'}
                </button>
              )}
            </div>
          </div>
        </div>

        {/* Content Sections */}
        <div className="space-y-6">
          {/* Overview Section */}
          <div className="space-y-6 animate-fadeIn">
              {/* Metadata Section */}
              <div
                className="rounded-2xl p-6 border animate-scaleIn"
                style={{
                  background: 'var(--surface)',
                  borderColor: 'var(--border)',
                  boxShadow: 'var(--shadow-lg)'
                }}
              >
                <h2
                  className="text-xl font-bold mb-5"
                  style={{
                    color: 'var(--foreground)',
                    fontFamily: 'var(--font-display)'
                  }}
                >
                  メタデータ
                </h2>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
                  <div>
                    <p className="text-sm font-semibold mb-1.5" style={{ color: 'var(--foreground-secondary)', fontFamily: 'var(--font-body)' }}>検出日時</p>
                    <p className="text-sm font-medium" style={{ color: 'var(--foreground)', fontFamily: 'var(--font-mono)' }}>
                      {new Date(incident.detected_at).toLocaleString('ja-JP')}
                    </p>
                  </div>
                  <div>
                    <p className="text-sm font-semibold mb-1.5" style={{ color: 'var(--foreground-secondary)', fontFamily: 'var(--font-body)' }}>解決日時</p>
                    <p className="text-sm font-medium" style={{ color: 'var(--foreground)', fontFamily: 'var(--font-mono)' }}>
                      {incident.resolved_at
                        ? new Date(incident.resolved_at).toLocaleString('ja-JP')
                        : '未解決'}
                    </p>
                  </div>
                  <div>
                    <p className="text-sm font-semibold mb-1.5" style={{ color: 'var(--foreground-secondary)', fontFamily: 'var(--font-body)' }}>影響範囲</p>
                    <p className="text-sm font-medium" style={{ color: 'var(--foreground)', fontFamily: 'var(--font-body)' }}>{incident.impact_scope || '-'}</p>
                  </div>
                  <div>
                    <p className="text-sm font-semibold mb-1.5" style={{ color: 'var(--foreground-secondary)', fontFamily: 'var(--font-body)' }}>担当者</p>
                    {permissions.canEdit ? (
                      <select
                        value={incident.assignee_id || ''}
                        onChange={(e) => {
                          const value = e.target.value;
                          handleAssignIncident(value === '' ? null : parseInt(value));
                        }}
                        disabled={assigningUser}
                        className="text-sm font-medium border-2 rounded-lg px-3 py-2 focus:outline-none transition-all disabled:opacity-50 disabled:cursor-not-allowed"
                        style={{
                          background: 'var(--surface)',
                          borderColor: 'var(--border)',
                          color: 'var(--foreground)',
                          fontFamily: 'var(--font-body)'
                        }}
                        onFocus={(e) => {
                          e.currentTarget.style.borderColor = 'var(--primary)';
                          e.currentTarget.style.boxShadow = '0 0 0 3px var(--primary-light)';
                        }}
                        onBlur={(e) => {
                          e.currentTarget.style.borderColor = 'var(--border)';
                          e.currentTarget.style.boxShadow = 'none';
                        }}
                      >
                        <option value="">未割り当て</option>
                        {users.map((user) => (
                          <option key={user.id} value={user.id}>
                            {user.name} ({user.email})
                          </option>
                        ))}
                      </select>
                    ) : (
                      <p className="text-sm font-medium" style={{ color: 'var(--foreground)', fontFamily: 'var(--font-body)' }}>
                        {incident.assignee
                          ? `${incident.assignee.name} (${incident.assignee.email})`
                          : '未割り当て'}
                      </p>
                    )}
                  </div>
                  <div>
                    <p className="text-sm font-semibold mb-1.5" style={{ color: 'var(--foreground-secondary)', fontFamily: 'var(--font-body)' }}>作成者</p>
                    <p className="text-sm font-medium" style={{ color: 'var(--foreground)', fontFamily: 'var(--font-body)' }}>
                      {incident.creator.name} ({incident.creator.email})
                    </p>
                  </div>
                  <div>
                    <p className="text-sm font-semibold mb-1.5" style={{ color: 'var(--foreground-secondary)', fontFamily: 'var(--font-body)' }}>作成日時</p>
                    <p className="text-sm font-medium" style={{ color: 'var(--foreground)', fontFamily: 'var(--font-mono)' }}>
                      {new Date(incident.created_at).toLocaleString('ja-JP')}
                    </p>
                  </div>
                </div>
              </div>

              {/* Description Section */}
              <div
                className="rounded-2xl p-6 border animate-scaleIn"
                style={{
                  background: 'var(--surface)',
                  borderColor: 'var(--border)',
                  boxShadow: 'var(--shadow-lg)',
                  animationDelay: '0.1s'
                }}
              >
                <h2
                  className="text-xl font-bold mb-4"
                  style={{
                    color: 'var(--foreground)',
                    fontFamily: 'var(--font-display)'
                  }}
                >
                  説明
                </h2>
                <p
                  className="text-sm leading-relaxed whitespace-pre-wrap"
                  style={{
                    color: 'var(--foreground)',
                    fontFamily: 'var(--font-body)'
                  }}
                >
                  {incident.description}
                </p>
              </div>

              {/* Summary Section */}
              <div
                className="rounded-2xl p-6 border animate-scaleIn"
                style={{
                  background: 'var(--surface)',
                  borderColor: 'var(--border)',
                  boxShadow: 'var(--shadow-lg)',
                  animationDelay: '0.2s'
                }}
              >
                <div className="flex justify-between items-center mb-4">
                  <h2
                    className="text-xl font-bold"
                    style={{
                      color: 'var(--foreground)',
                      fontFamily: 'var(--font-display)'
                    }}
                  >
                    AI要約
                  </h2>
                  {canEdit() && (
                    <button
                      onClick={handleRegenerateSummary}
                      disabled={regeneratingSummary}
                      className="px-4 py-2 text-white rounded-xl text-sm font-semibold transition-all duration-200 disabled:opacity-50"
                      style={{
                        background: 'linear-gradient(135deg, var(--primary) 0%, var(--primary-dark) 100%)',
                        fontFamily: 'var(--font-body)',
                        boxShadow: '0 4px 12px var(--primary-glow)'
                      }}
                      onMouseEnter={(e) => {
                        if (!regeneratingSummary) {
                          e.currentTarget.style.transform = 'translateY(-2px)';
                          e.currentTarget.style.boxShadow = '0 8px 20px var(--primary-glow)';
                        }
                      }}
                      onMouseLeave={(e) => {
                        e.currentTarget.style.transform = 'translateY(0)';
                        e.currentTarget.style.boxShadow = '0 4px 12px var(--primary-glow)';
                      }}
                    >
                      {regeneratingSummary ? '再生成中...' : '要約を再生成'}
                    </button>
                  )}
                </div>
                {incident.summary ? (
                  <p className="text-sm leading-relaxed" style={{ color: 'var(--foreground)', fontFamily: 'var(--font-body)' }}>
                    {incident.summary}
                  </p>
                ) : (
                  <div>
                    <p className="text-sm italic mb-3" style={{ color: 'var(--foreground-secondary)', fontFamily: 'var(--font-body)' }}>
                      AI要約がまだ生成されていません
                    </p>
                    {canEdit() && (
                      <button
                        onClick={handleRegenerateSummary}
                        disabled={regeneratingSummary}
                        className="px-4 py-2 text-white rounded-xl text-sm font-semibold transition-all duration-200 disabled:opacity-50"
                        style={{
                          background: 'linear-gradient(135deg, var(--primary) 0%, var(--primary-dark) 100%)',
                          fontFamily: 'var(--font-body)',
                          boxShadow: '0 4px 12px var(--primary-glow)'
                        }}
                        onMouseEnter={(e) => {
                          if (!regeneratingSummary) {
                            e.currentTarget.style.transform = 'translateY(-2px)';
                            e.currentTarget.style.boxShadow = '0 8px 20px var(--primary-glow)';
                          }
                        }}
                        onMouseLeave={(e) => {
                          e.currentTarget.style.transform = 'translateY(0)';
                          e.currentTarget.style.boxShadow = '0 4px 12px var(--primary-glow)';
                        }}
                      >
                        {regeneratingSummary ? '生成中...' : '要約を生成'}
                      </button>
                    )}
                  </div>
                )}
              </div>

              {/* Tags Section */}
              <div
                className="rounded-2xl p-6 border animate-scaleIn"
                style={{
                  background: 'var(--surface)',
                  borderColor: 'var(--border)',
                  boxShadow: 'var(--shadow-lg)',
                  animationDelay: '0.3s'
                }}
              >
                <h2
                  className="text-xl font-bold mb-4"
                  style={{
                    color: 'var(--foreground)',
                    fontFamily: 'var(--font-display)'
                  }}
                >
                  タグ
                </h2>
                {incident.tags && incident.tags.length > 0 ? (
                  <div className="flex flex-wrap gap-2">
                    {incident.tags.map((tag) => (
                      <span
                        key={tag.id}
                        className="px-4 py-2 rounded-full text-white text-sm font-semibold shadow-sm"
                        style={{ backgroundColor: tag.color, fontFamily: 'var(--font-body)' }}
                      >
                        {tag.name}
                      </span>
                    ))}
                  </div>
                ) : (
                  <p className="text-sm" style={{ color: 'var(--foreground-secondary)', fontFamily: 'var(--font-body)' }}>
                    タグが設定されていません
                  </p>
                )}
              </div>
            </div>

          {/* Attachments Section */}
          <div className="animate-fadeIn">
              <div
                className="rounded-2xl p-6 border"
                style={{
                  background: 'var(--surface)',
                  borderColor: 'var(--border)',
                  boxShadow: 'var(--shadow-lg)'
                }}
              >
                <h2
                  className="text-xl font-bold mb-6"
                  style={{
                    color: 'var(--foreground)',
                    fontFamily: 'var(--font-display)'
                  }}
                >
                  添付ファイル
                </h2>

                {/* Upload Form */}
                <div className="mb-6 p-5 rounded-xl border-2 border-dashed transition-all" style={{ borderColor: 'var(--border)', background: 'var(--gray-50)' }}>
                  <label className="block text-sm font-semibold mb-3" style={{ color: 'var(--foreground)', fontFamily: 'var(--font-body)' }}>
                    ファイルをアップロード
                  </label>
                  <input
                    type="file"
                    onChange={handleFileUpload}
                    disabled={uploadingFile}
                    className="block w-full text-sm border-2 rounded-lg cursor-pointer focus:outline-none transition-all disabled:opacity-50 disabled:cursor-not-allowed file:mr-4 file:py-2.5 file:px-4 file:rounded-lg file:border-0 file:text-sm file:font-semibold file:transition-all"
                    style={{
                      background: 'var(--surface)',
                      borderColor: 'var(--border)',
                      color: 'var(--foreground)'
                    }}
                  />
                  <p className="mt-2 text-xs" style={{ color: 'var(--foreground-secondary)', fontFamily: 'var(--font-body)' }}>
                    対応ファイル: 画像 (jpg, png, gif), PDF, テキスト, アーカイブ (zip, tar, gz) - 最大 50MB
                  </p>
                </div>

                {/* Attachments List */}
                {loadingAttachments ? (
                  <div className="text-center py-8" style={{ color: 'var(--foreground-secondary)', fontFamily: 'var(--font-body)' }}>
                    読み込み中...
                  </div>
                ) : attachments.length === 0 ? (
                  <div className="text-center py-12 rounded-xl" style={{ background: 'var(--gray-50)' }}>
                    <svg className="mx-auto h-12 w-12 mb-3" style={{ color: 'var(--foreground-secondary)' }} fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15.172 7l-6.586 6.586a2 2 0 102.828 2.828l6.414-6.586a4 4 0 00-5.656-5.656l-6.415 6.585a6 6 0 108.486 8.486L20.5 13" />
                    </svg>
                    <p className="text-sm font-medium" style={{ color: 'var(--foreground-secondary)', fontFamily: 'var(--font-body)' }}>
                      添付ファイルはありません
                    </p>
                  </div>
                ) : (
                  <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                    {attachments.map((attachment) => {
                      const isImage = isImageFile(attachment.file_name);
                      const imageUrl = imageUrls[attachment.id];

                      return (
                        <div
                          key={attachment.id}
                          className="border-2 rounded-2xl p-4 transition-all duration-200"
                          style={{
                            background: 'var(--surface)',
                            borderColor: 'var(--border)',
                            boxShadow: 'var(--shadow-sm)'
                          }}
                          onMouseEnter={(e) => {
                            e.currentTarget.style.transform = 'translateY(-4px)';
                            e.currentTarget.style.boxShadow = 'var(--shadow-xl)';
                          }}
                          onMouseLeave={(e) => {
                            e.currentTarget.style.transform = 'translateY(0)';
                            e.currentTarget.style.boxShadow = 'var(--shadow-sm)';
                          }}
                        >
                      {isImage && imageUrl ? (
                        <div
                          className="mb-3 cursor-pointer group relative"
                          onClick={() => setLightboxImage(imageUrl)}
                        >
                          <img
                            src={imageUrl}
                            alt={attachment.file_name}
                            className="w-full h-48 object-cover rounded-md group-hover:opacity-90 transition-opacity"
                            onError={(e) => {
                              console.error('Image load error:', attachment.file_name);
                              e.currentTarget.style.display = 'none';
                            }}
                          />
                          <div className="absolute inset-0 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity bg-black bg-opacity-30 rounded-md">
                            <svg className="w-12 h-12 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0zM10 7v3m0 0v3m0-3h3m-3 0H7" />
                            </svg>
                          </div>
                        </div>
                      ) : isImage ? (
                        <div className="mb-3 flex items-center justify-center h-48 bg-gray-100 rounded-md">
                          <div className="text-center">
                            <svg className="w-12 h-12 text-gray-400 mx-auto mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
                            </svg>
                            <p className="text-xs text-gray-500">読み込み中...</p>
                          </div>
                        </div>
                      ) : (
                        <div className="mb-3 flex items-center justify-center h-48 bg-gray-100 rounded-md">
                          <svg className="w-16 h-16 text-gray-400" fill="currentColor" viewBox="0 0 20 20">
                            <path fillRule="evenodd" d="M8 4a3 3 0 00-3 3v4a5 5 0 0010 0V7a1 1 0 112 0v4a7 7 0 11-14 0V7a5 5 0 0110 0v4a3 3 0 11-6 0V7a1 1 0 012 0v4a1 1 0 102 0V7a3 3 0 00-3-3z" clipRule="evenodd" />
                          </svg>
                        </div>
                      )}
                          <div className="min-w-0">
                            <p className="text-sm font-semibold truncate mb-2" style={{ color: 'var(--foreground)', fontFamily: 'var(--font-body)' }}>
                              {attachment.file_name}
                            </p>
                            <p className="text-xs mb-3" style={{ color: 'var(--foreground-secondary)', fontFamily: 'var(--font-body)' }}>
                              {formatFileSize(attachment.file_size)} • {attachment.user?.name}
                            </p>
                            <div className="flex items-center gap-2">
                              <button
                                onClick={() => handleFileDownload(attachment.id, attachment.file_name)}
                                className="flex-1 px-3 py-2 text-xs font-semibold text-white rounded-lg transition-all duration-200"
                                style={{
                                  background: 'linear-gradient(135deg, var(--primary) 0%, var(--primary-dark) 100%)',
                                  fontFamily: 'var(--font-body)'
                                }}
                                onMouseEnter={(e) => e.currentTarget.style.transform = 'scale(1.05)'}
                                onMouseLeave={(e) => e.currentTarget.style.transform = 'scale(1)'}
                              >
                                ダウンロード
                              </button>
                              {(user?.role === 'admin' || user?.id === attachment.user_id) && (
                                <button
                                  onClick={() => handleFileDelete(attachment.id)}
                                  className="px-3 py-2 text-xs font-semibold text-white rounded-lg transition-all duration-200"
                                  style={{
                                    background: 'linear-gradient(135deg, var(--error) 0%, var(--error-dark) 100%)',
                                    fontFamily: 'var(--font-body)'
                                  }}
                                  onMouseEnter={(e) => e.currentTarget.style.transform = 'scale(1.05)'}
                                  onMouseLeave={(e) => e.currentTarget.style.transform = 'scale(1)'}
                                >
                                  削除
                                </button>
                              )}
                            </div>
                          </div>
                        </div>
                      );
                    })}
                  </div>
                )}
              </div>
            </div>

          {/* Unified Timeline Section */}
          <div className="animate-fadeIn">
              <div
                className="rounded-2xl p-6 border"
                style={{
                  background: 'var(--surface)',
                  borderColor: 'var(--border)',
                  boxShadow: 'var(--shadow-lg)'
                }}
              >
                <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 mb-6">
                  <h2
                    className="text-xl font-bold"
                    style={{
                      color: 'var(--foreground)',
                      fontFamily: 'var(--font-display)'
                    }}
                  >
                    タイムライン & アクティビティ
                  </h2>
                  {canEdit() && (
                    <button
                      onClick={() => {
                        setShowTimelineEventForm(!showTimelineEventForm);
                        if (!showTimelineEventForm) {
                          const now = new Date();
                          const localDateTime = new Date(now.getTime() - now.getTimezoneOffset() * 60000)
                            .toISOString()
                            .slice(0, 16);
                          setTimelineEventTime(localDateTime);
                          setEntryType('comment');
                          setNewComment('');
                          setTimelineEventDescription('');
                        }
                      }}
                      className="px-4 py-2.5 text-white rounded-xl text-sm font-semibold transition-all duration-200"
                      style={{
                        background: showTimelineEventForm
                          ? 'linear-gradient(135deg, var(--secondary) 0%, var(--secondary-dark) 100%)'
                          : 'linear-gradient(135deg, var(--success) 0%, var(--success-dark) 100%)',
                        fontFamily: 'var(--font-body)',
                        boxShadow: '0 4px 12px var(--primary-glow)'
                      }}
                      onMouseEnter={(e) => {
                        e.currentTarget.style.transform = 'translateY(-2px)';
                        e.currentTarget.style.boxShadow = '0 8px 20px var(--primary-glow)';
                      }}
                      onMouseLeave={(e) => {
                        e.currentTarget.style.transform = 'translateY(0)';
                        e.currentTarget.style.boxShadow = '0 4px 12px var(--primary-glow)';
                      }}
                    >
                      {showTimelineEventForm ? 'キャンセル' : '+ 追加'}
                    </button>
                  )}
                </div>

                {/* Unified Entry Form */}
                {showTimelineEventForm && canEdit() && (
                  <form
                    onSubmit={(e) => {
                      e.preventDefault();
                      if (entryType === 'comment') {
                        handleAddComment(e);
                      } else {
                        handleAddTimelineEvent(e);
                      }
                    }}
                    className="mb-6 p-5 rounded-xl border-2 animate-slideDown"
                    style={{
                      background: 'var(--gray-50)',
                      borderColor: 'var(--border)'
                    }}
                  >
                    <div className="mb-4">
                      <label
                        htmlFor="entry_type"
                        className="block text-sm font-semibold mb-2"
                        style={{ color: 'var(--foreground)', fontFamily: 'var(--font-body)' }}
                      >
                        タイプ
                      </label>
                      <select
                        id="entry_type"
                        value={entryType}
                        onChange={(e) => setEntryType(e.target.value as 'comment' | 'event')}
                        className="w-full px-3 py-2 border-2 rounded-lg focus:outline-none transition-all"
                        style={{
                          background: 'var(--surface)',
                          borderColor: 'var(--border)',
                          color: 'var(--foreground)',
                          fontFamily: 'var(--font-body)'
                        }}
                        onFocus={(e) => {
                          e.currentTarget.style.borderColor = 'var(--primary)';
                          e.currentTarget.style.boxShadow = '0 0 0 3px var(--primary-light)';
                        }}
                        onBlur={(e) => {
                          e.currentTarget.style.borderColor = 'var(--border)';
                          e.currentTarget.style.boxShadow = 'none';
                        }}
                      >
                        <option value="comment">💬 コメント</option>
                        <option value="event">⏱️ イベント</option>
                      </select>
                    </div>

                    {entryType === 'event' && (
                      <>
                        <div className="mb-4">
                          <label
                            htmlFor="event_type"
                            className="block text-sm font-semibold mb-2"
                            style={{ color: 'var(--foreground)', fontFamily: 'var(--font-body)' }}
                          >
                            イベントタイプ
                          </label>
                          <select
                            id="event_type"
                            value={timelineEventType}
                            onChange={(e) => setTimelineEventType(e.target.value as any)}
                            className="w-full px-3 py-2 border-2 rounded-lg focus:outline-none transition-all"
                            style={{
                              background: 'var(--surface)',
                              borderColor: 'var(--border)',
                              color: 'var(--foreground)',
                              fontFamily: 'var(--font-body)'
                            }}
                            onFocus={(e) => {
                              e.currentTarget.style.borderColor = 'var(--primary)';
                              e.currentTarget.style.boxShadow = '0 0 0 3px var(--primary-light)';
                            }}
                            onBlur={(e) => {
                              e.currentTarget.style.borderColor = 'var(--border)';
                              e.currentTarget.style.boxShadow = 'none';
                            }}
                          >
                            <option value="detected">検知</option>
                            <option value="investigation_started">調査開始</option>
                            <option value="root_cause_identified">原因特定</option>
                            <option value="mitigation">緩和</option>
                            <option value="timeline_resolved">解決</option>
                            <option value="other">その他</option>
                          </select>
                        </div>
                        <div className="mb-4">
                          <label
                            htmlFor="event_time"
                            className="block text-sm font-semibold mb-2"
                            style={{ color: 'var(--foreground)', fontFamily: 'var(--font-body)' }}
                          >
                            イベント時刻
                          </label>
                          <input
                            type="datetime-local"
                            id="event_time"
                            value={timelineEventTime}
                            onChange={(e) => setTimelineEventTime(e.target.value)}
                            className="w-full px-3 py-2 border-2 rounded-lg focus:outline-none transition-all"
                            style={{
                              background: 'var(--surface)',
                              borderColor: 'var(--border)',
                              color: 'var(--foreground)',
                              fontFamily: 'var(--font-body)'
                            }}
                            onFocus={(e) => {
                              e.currentTarget.style.borderColor = 'var(--primary)';
                              e.currentTarget.style.boxShadow = '0 0 0 3px var(--primary-light)';
                            }}
                            onBlur={(e) => {
                              e.currentTarget.style.borderColor = 'var(--border)';
                              e.currentTarget.style.boxShadow = 'none';
                            }}
                            required
                          />
                        </div>
                      </>
                    )}

                    <div className="mb-4">
                      <label
                        htmlFor={entryType === 'comment' ? 'comment' : 'event_description'}
                        className="block text-sm font-semibold mb-2"
                        style={{ color: 'var(--foreground)', fontFamily: 'var(--font-body)' }}
                      >
                        {entryType === 'comment' ? 'コメント' : '説明'}
                      </label>
                      <textarea
                        id={entryType === 'comment' ? 'comment' : 'event_description'}
                        rows={4}
                        value={entryType === 'comment' ? newComment : timelineEventDescription}
                        onChange={(e) => entryType === 'comment' ? setNewComment(e.target.value) : setTimelineEventDescription(e.target.value)}
                        placeholder={entryType === 'comment' ? 'コメントを入力してください...' : 'イベントの説明を入力してください...'}
                        className="w-full px-3 py-2 border-2 rounded-lg focus:outline-none transition-all"
                        style={{
                          background: 'var(--surface)',
                          borderColor: 'var(--border)',
                          color: 'var(--foreground)',
                          fontFamily: 'var(--font-body)'
                        }}
                        disabled={submittingTimelineEvent || submittingComment}
                        onFocus={(e) => {
                          e.currentTarget.style.borderColor = 'var(--primary)';
                          e.currentTarget.style.boxShadow = '0 0 0 3px var(--primary-light)';
                        }}
                        onBlur={(e) => {
                          e.currentTarget.style.borderColor = 'var(--border)';
                          e.currentTarget.style.boxShadow = 'none';
                        }}
                        required
                      />
                    </div>
                    <div className="flex justify-end">
                      <button
                        type="submit"
                        disabled={submittingTimelineEvent || submittingComment || (entryType === 'comment' ? !newComment.trim() : !timelineEventDescription.trim())}
                        className="px-5 py-2.5 text-white rounded-xl font-semibold transition-all duration-200 disabled:opacity-50"
                        style={{
                          background: entryType === 'comment'
                            ? 'linear-gradient(135deg, var(--primary) 0%, var(--primary-dark) 100%)'
                            : 'linear-gradient(135deg, var(--success) 0%, var(--success-dark) 100%)',
                          fontFamily: 'var(--font-body)',
                          boxShadow: '0 4px 12px var(--primary-glow)'
                        }}
                        onMouseEnter={(e) => {
                          const isValid = entryType === 'comment' ? newComment.trim() : timelineEventDescription.trim();
                          if (!submittingTimelineEvent && !submittingComment && isValid) {
                            e.currentTarget.style.transform = 'translateY(-2px)';
                            e.currentTarget.style.boxShadow = '0 8px 20px var(--primary-glow)';
                          }
                        }}
                        onMouseLeave={(e) => {
                          e.currentTarget.style.transform = 'translateY(0)';
                          e.currentTarget.style.boxShadow = '0 4px 12px var(--primary-glow)';
                        }}
                      >
                        {submittingTimelineEvent || submittingComment ? '送信中...' : (entryType === 'comment' ? 'コメントを投稿' : 'イベントを登録')}
                      </button>
                    </div>
                  </form>
                )}

                {/* Timeline */}
                {loadingActivities ? (
                  <div className="text-center py-8" style={{ color: 'var(--foreground-secondary)', fontFamily: 'var(--font-body)' }}>
                    読み込み中...
                  </div>
                ) : (
                  <Timeline activities={activities} />
                )}
              </div>
            </div>
        </div>
      </div>

      {/* Lightbox Modal */}
      {lightboxImage && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-90"
          onClick={() => setLightboxImage(null)}
        >
          <div className="relative max-w-7xl max-h-screen p-4">
            <button
              onClick={() => setLightboxImage(null)}
              className="absolute top-4 right-4 text-white hover:text-gray-300 transition-colors"
            >
              <svg className="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
            <img
              src={lightboxImage}
              alt="Preview"
              className="max-w-full max-h-screen object-contain"
              onClick={(e) => e.stopPropagation()}
            />
          </div>
        </div>
      )}
    </div>
  );
}
