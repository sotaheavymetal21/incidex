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
import Tabs, { Tab } from '@/components/Tabs';
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
              <button
                onClick={() => router.push(`/incidents/${id}/postmortem`)}
                className="px-4 py-2 bg-purple-600 text-white rounded-md hover:bg-purple-700"
              >
                Post-Mortem
              </button>
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

        {/* Tabs Section */}
        <Tabs
          tabs={[
            {
              id: 'overview',
              label: '概要',
              icon: (
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                </svg>
              ),
              content: (
                <div className="space-y-6">
                  {/* Metadata Section */}
                  <div className="bg-white rounded-lg shadow p-6">
                    <h2 className="text-lg font-semibold text-gray-900 mb-4">メタデータ</h2>
                    <div className="grid grid-cols-2 gap-4">
                      <div>
                        <p className="text-sm font-medium text-gray-500">検出日時</p>
                        <p className="mt-1 text-sm text-gray-900">
                          {new Date(incident.detected_at).toLocaleString('ja-JP')}
                        </p>
                      </div>
                      <div>
                        <p className="text-sm font-medium text-gray-500">解決日時</p>
                        <p className="mt-1 text-sm text-gray-900">
                          {incident.resolved_at
                            ? new Date(incident.resolved_at).toLocaleString('ja-JP')
                            : '未解決'}
                        </p>
                      </div>
                      <div>
                        <p className="text-sm font-medium text-gray-500">影響範囲</p>
                        <p className="mt-1 text-sm text-gray-900">{incident.impact_scope || '-'}</p>
                      </div>
                      <div>
                        <p className="text-sm font-medium text-gray-500">担当者</p>
                        {permissions.canEdit ? (
                          <select
                            value={incident.assignee_id || ''}
                            onChange={(e) => {
                              const value = e.target.value;
                              handleAssignIncident(value === '' ? null : parseInt(value));
                            }}
                            disabled={assigningUser}
                            className="mt-1 text-sm text-gray-900 border border-gray-300 rounded-md px-2 py-1 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
                          >
                            <option value="">未割り当て</option>
                            {users.map((user) => (
                              <option key={user.id} value={user.id}>
                                {user.name} ({user.email})
                              </option>
                            ))}
                          </select>
                        ) : (
                          <p className="mt-1 text-sm text-gray-900">
                            {incident.assignee
                              ? `${incident.assignee.name} (${incident.assignee.email})`
                              : '未割り当て'}
                          </p>
                        )}
                      </div>
                      <div>
                        <p className="text-sm font-medium text-gray-500">作成者</p>
                        <p className="mt-1 text-sm text-gray-900">
                          {incident.creator.name} ({incident.creator.email})
                        </p>
                      </div>
                      <div>
                        <p className="text-sm font-medium text-gray-500">作成日時</p>
                        <p className="mt-1 text-sm text-gray-900">
                          {new Date(incident.created_at).toLocaleString('ja-JP')}
                        </p>
                      </div>
                    </div>
                  </div>

                  {/* Description Section */}
                  <div className="bg-white rounded-lg shadow p-6">
                    <h2 className="text-lg font-semibold text-gray-900 mb-4">説明</h2>
                    <p className="text-sm text-gray-700 whitespace-pre-wrap">{incident.description}</p>
                  </div>

                  {/* Summary Section */}
                  <div className="bg-white rounded-lg shadow p-6">
                    <div className="flex justify-between items-center mb-4">
                      <h2 className="text-lg font-semibold text-gray-900">AI要約</h2>
                      {canEdit() && (
                        <button
                          onClick={handleRegenerateSummary}
                          disabled={regeneratingSummary}
                          className="px-3 py-1 text-sm bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
                        >
                          {regeneratingSummary ? '再生成中...' : '要約を再生成'}
                        </button>
                      )}
                    </div>
                    {incident.summary ? (
                      <p className="text-sm text-gray-700">{incident.summary}</p>
                    ) : (
                      <div>
                        <p className="text-sm text-gray-500 italic mb-2">
                          AI要約がまだ生成されていません
                        </p>
                        {canEdit() && (
                          <button
                            onClick={handleRegenerateSummary}
                            disabled={regeneratingSummary}
                            className="px-3 py-1 text-sm bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
                          >
                            {regeneratingSummary ? '生成中...' : '要約を生成'}
                          </button>
                        )}
                      </div>
                    )}
                  </div>

                  {/* Tags Section */}
                  <div className="bg-white rounded-lg shadow p-6">
                    <h2 className="text-lg font-semibold text-gray-900 mb-4">タグ</h2>
                    {incident.tags && incident.tags.length > 0 ? (
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
                      <p className="text-sm text-gray-500">タグが設定されていません</p>
                    )}
                  </div>
                </div>
              ),
            },
            {
              id: 'attachments',
              label: '添付ファイル',
              icon: (
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15.172 7l-6.586 6.586a2 2 0 102.828 2.828l6.414-6.586a4 4 0 00-5.656-5.656l-6.415 6.585a6 6 0 108.486 8.486L20.5 13" />
                </svg>
              ),
              content: (
                <div className="bg-white rounded-lg shadow p-6">
                  <h2 className="text-lg font-semibold text-gray-900 mb-4">添付ファイル</h2>

                  {/* Upload Form */}
                  <div className="mb-6">
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      ファイルをアップロード
                    </label>
                    <input
                      type="file"
                      onChange={handleFileUpload}
                      disabled={uploadingFile}
                      className="block w-full text-sm text-gray-900 border border-gray-300 rounded-md cursor-pointer focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed file:mr-4 file:py-2 file:px-4 file:rounded-md file:border-0 file:text-sm file:font-semibold file:bg-blue-50 file:text-blue-700 hover:file:bg-blue-100"
                    />
                    <p className="mt-1 text-xs text-gray-500">
                      対応ファイル: 画像 (jpg, png, gif), PDF, テキスト, アーカイブ (zip, tar, gz) - 最大 50MB
                    </p>
                  </div>

                  {/* Attachments List */}
                  {loadingAttachments ? (
                    <div className="text-center py-4 text-gray-500">読み込み中...</div>
                  ) : attachments.length === 0 ? (
                    <p className="text-sm text-gray-500">添付ファイルはありません</p>
                  ) : (
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                      {attachments.map((attachment) => {
                        const isImage = isImageFile(attachment.file_name);
                        const imageUrl = imageUrls[attachment.id];

                        return (
                          <div key={attachment.id} className="border border-gray-200 rounded-lg p-3 hover:shadow-md transition-shadow">
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
                              <p className="text-sm font-medium text-gray-900 truncate mb-1">
                                {attachment.file_name}
                              </p>
                              <p className="text-xs text-gray-500 mb-3">
                                {formatFileSize(attachment.file_size)} • {attachment.user?.name}
                              </p>
                              <div className="flex items-center space-x-2">
                                <button
                                  onClick={() => handleFileDownload(attachment.id, attachment.file_name)}
                                  className="flex-1 px-3 py-1 text-xs bg-blue-600 text-white rounded-md hover:bg-blue-700"
                                >
                                  ダウンロード
                                </button>
                                {(user?.role === 'admin' || user?.id === attachment.user_id) && (
                                  <button
                                    onClick={() => handleFileDelete(attachment.id)}
                                    className="px-3 py-1 text-xs bg-red-600 text-white rounded-md hover:bg-red-700"
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
              ),
            },
            {
              id: 'timeline',
              label: 'タイムライン & コメント',
              icon: (
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
              ),
              content: (
                <div className="bg-white rounded-lg shadow p-6">
                  <div className="flex justify-between items-center mb-4">
                    <h2 className="text-lg font-semibold text-gray-900">タイムライン</h2>
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
                          }
                        }}
                        className="px-3 py-1 text-sm bg-green-600 text-white rounded-md hover:bg-green-700"
                      >
                        {showTimelineEventForm ? 'キャンセル' : 'イベントを追加'}
                      </button>
                    )}
                  </div>

                  {/* Timeline Event Form */}
                  {showTimelineEventForm && canEdit() && (
                    <form onSubmit={handleAddTimelineEvent} className="mb-6 p-4 bg-gray-50 rounded-lg border border-gray-200">
                      <div className="mb-4">
                        <label htmlFor="event_type" className="block text-sm font-medium text-gray-700 mb-1">
                          イベントタイプ
                        </label>
                        <select
                          id="event_type"
                          value={timelineEventType}
                          onChange={(e) => setTimelineEventType(e.target.value as any)}
                          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 text-gray-900"
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
                        <label htmlFor="event_time" className="block text-sm font-medium text-gray-700 mb-1">
                          イベント時刻
                        </label>
                        <input
                          type="datetime-local"
                          id="event_time"
                          value={timelineEventTime}
                          onChange={(e) => setTimelineEventTime(e.target.value)}
                          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 text-gray-900"
                          required
                        />
                      </div>
                      <div className="mb-4">
                        <label htmlFor="event_description" className="block text-sm font-medium text-gray-700 mb-1">
                          説明
                        </label>
                        <textarea
                          id="event_description"
                          rows={3}
                          value={timelineEventDescription}
                          onChange={(e) => setTimelineEventDescription(e.target.value)}
                          placeholder="イベントの説明を入力してください..."
                          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 text-gray-900 placeholder-gray-500"
                          disabled={submittingTimelineEvent}
                          required
                        />
                      </div>
                      <div className="flex justify-end">
                        <button
                          type="submit"
                          disabled={submittingTimelineEvent || !timelineEventDescription.trim()}
                          className="px-4 py-2 bg-green-600 text-white rounded-md hover:bg-green-700 disabled:opacity-50 disabled:cursor-not-allowed"
                        >
                          {submittingTimelineEvent ? '追加中...' : 'イベントを追加'}
                        </button>
                      </div>
                    </form>
                  )}

                  {/* Comment Form */}
                  <form onSubmit={handleAddComment} className="mb-6">
                    <div className="mb-2">
                      <label htmlFor="comment" className="block text-sm font-medium text-gray-700 mb-1">
                        コメントを追加
                      </label>
                      <textarea
                        id="comment"
                        rows={3}
                        value={newComment}
                        onChange={(e) => setNewComment(e.target.value)}
                        placeholder="コメントを入力してください..."
                        className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 text-gray-900 placeholder-gray-500"
                        disabled={submittingComment}
                      />
                    </div>
                    <div className="flex justify-end">
                      <button
                        type="submit"
                        disabled={submittingComment || !newComment.trim()}
                        className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
                      >
                        {submittingComment ? '送信中...' : 'コメントを投稿'}
                      </button>
                    </div>
                  </form>

                  {/* Timeline */}
                  {loadingActivities ? (
                    <div className="text-center py-4 text-gray-500">読み込み中...</div>
                  ) : (
                    <Timeline activities={activities} />
                  )}
                </div>
              ),
            },
          ]}
          defaultTab="overview"
        />
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
