'use client';

import { useEffect, useState } from 'react';
import { useRouter, useParams } from 'next/navigation';
import { useAuth } from '@/context/AuthContext';
import { incidentApi, activityApi, attachmentApi } from '@/lib/api';
import { Incident, Severity, Status } from '@/types/incident';
import { IncidentActivity } from '@/types/activity';
import { Attachment } from '@/types/attachment';
import Timeline from '@/components/Timeline';

export default function IncidentDetailPage() {
  const { token, user, loading: authLoading } = useAuth();
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

        {/* Metadata Section */}
        <div className="bg-white rounded-lg shadow p-6 mb-6">
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
              <p className="mt-1 text-sm text-gray-900">
                {incident.assignee
                  ? `${incident.assignee.name} (${incident.assignee.email})`
                  : '未割り当て'}
              </p>
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
        <div className="bg-white rounded-lg shadow p-6 mb-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Description</h2>
          <p className="text-sm text-gray-700 whitespace-pre-wrap">{incident.description}</p>
        </div>

        {/* Summary Section */}
        <div className="bg-white rounded-lg shadow p-6 mb-6">
          <div className="flex justify-between items-center mb-4">
            <h2 className="text-lg font-semibold text-gray-900">AI Summary</h2>
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

        {/* Attachments Section */}
        <div className="bg-white rounded-lg shadow p-6 mb-6">
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
            <div className="divide-y divide-gray-200">
              {attachments.map((attachment) => (
                <div key={attachment.id} className="py-3 flex items-center justify-between">
                  <div className="flex items-center space-x-3 flex-1 min-w-0">
                    <svg className="w-8 h-8 text-gray-400 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20">
                      <path fillRule="evenodd" d="M8 4a3 3 0 00-3 3v4a5 5 0 0010 0V7a1 1 0 112 0v4a7 7 0 11-14 0V7a5 5 0 0110 0v4a3 3 0 11-6 0V7a1 1 0 012 0v4a1 1 0 102 0V7a3 3 0 00-3-3z" clipRule="evenodd" />
                    </svg>
                    <div className="min-w-0 flex-1">
                      <p className="text-sm font-medium text-gray-900 truncate">
                        {attachment.file_name}
                      </p>
                      <p className="text-xs text-gray-500">
                        {formatFileSize(attachment.file_size)} • {attachment.user?.name} • {new Date(attachment.created_at).toLocaleString('ja-JP')}
                      </p>
                    </div>
                  </div>
                  <div className="flex items-center space-x-2 ml-4">
                    <button
                      onClick={() => handleFileDownload(attachment.id, attachment.file_name)}
                      className="px-3 py-1 text-sm bg-blue-600 text-white rounded-md hover:bg-blue-700"
                    >
                      ダウンロード
                    </button>
                    {(user?.role === 'admin' || user?.id === attachment.user_id) && (
                      <button
                        onClick={() => handleFileDelete(attachment.id)}
                        className="px-3 py-1 text-sm bg-red-600 text-white rounded-md hover:bg-red-700"
                      >
                        削除
                      </button>
                    )}
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Timeline Section */}
        <div className="bg-white rounded-lg shadow p-6 mb-6">
          <div className="flex justify-between items-center mb-4">
            <h2 className="text-lg font-semibold text-gray-900">タイムライン</h2>
            {canEdit() && (
              <button
                onClick={() => {
                  setShowTimelineEventForm(!showTimelineEventForm);
                  if (!showTimelineEventForm) {
                    // Set default event time to current time
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
      </div>
    </div>
  );
}
