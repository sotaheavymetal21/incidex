'use client';

import { useState, useEffect } from 'react';
import { useAuth } from '@/context/AuthContext';
import { useRouter, useParams } from 'next/navigation';
import { postMortemApi, actionItemApi, userApi, incidentApi } from '@/lib/api';
import { PostMortem, FiveWhysAnalysis } from '@/types/postmortem';
import { ActionItem, Priority, ActionStatus } from '@/types/actionitem';
import { User, Incident } from '@/types/incident';

export default function PostMortemPage() {
  const { token, user, loading: authLoading } = useAuth();
  const router = useRouter();
  const params = useParams();
  const incidentId = parseInt(params.id as string);

  // State management
  const [postMortem, setPostMortem] = useState<PostMortem | null>(null);
  const [incident, setIncident] = useState<Incident | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [actionItems, setActionItems] = useState<ActionItem[]>([]);
  const [users, setUsers] = useState<User[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [aiLoading, setAiLoading] = useState(false);

  // Form state
  const [rootCause, setRootCause] = useState('');
  const [impactAnalysis, setImpactAnalysis] = useState('');
  const [whatWentWell, setWhatWentWell] = useState('');
  const [whatWentWrong, setWhatWentWrong] = useState('');
  const [lessonsLearned, setLessonsLearned] = useState('');
  const [fiveWhys, setFiveWhys] = useState<FiveWhysAnalysis>({
    why1: '',
    why2: '',
    why3: '',
    why4: '',
    why5: '',
  });

  // Action Item form state
  const [showActionItemForm, setShowActionItemForm] = useState(false);
  const [editingActionItem, setEditingActionItem] = useState<ActionItem | null>(null);
  const [actionItemForm, setActionItemForm] = useState({
    title: '',
    description: '',
    assignee_id: 0,
    priority: 'medium' as Priority,
    status: 'pending' as ActionStatus,
    due_date: '',
    related_links: '',
  });

  useEffect(() => {
    if (authLoading) return;
    if (!token) {
      router.push('/login');
      return;
    }
    fetchData();
  }, [token, authLoading, incidentId]);

  const fetchData = async () => {
    try {
      setLoading(true);
      setError(null);

      // Fetch incident
      const incidentData = await incidentApi.getById(token!, incidentId);
      setIncident(incidentData);

      // Fetch users for assignee selection
      const usersData = await userApi.getAll(token!);
      setUsers(usersData);

      // Try to fetch existing post-mortem
      try {
        const pmData = await postMortemApi.getByIncidentId(token!, incidentId);
        setPostMortem(pmData);

        // Populate form with existing data
        setRootCause(pmData.root_cause || '');
        setImpactAnalysis(pmData.impact_analysis || '');
        setWhatWentWell(pmData.what_went_well || '');
        setWhatWentWrong(pmData.what_went_wrong || '');
        setLessonsLearned(pmData.lessons_learned || '');

        // Parse five whys
        if (pmData.five_whys_analysis) {
          try {
            const parsedFiveWhys = JSON.parse(pmData.five_whys_analysis);
            setFiveWhys(parsedFiveWhys);
          } catch (e) {
            console.error('Failed to parse five whys:', e);
          }
        }

        // Fetch action items
        const actionItemsData = await actionItemApi.getByPostMortemId(token!, pmData.id);
        setActionItems(actionItemsData);
      } catch (err: any) {
        // Post-mortem doesn't exist yet, that's okay
        if (!err.message.includes('404') && !err.message.includes('not found')) {
          console.error('Error fetching post-mortem:', err);
        }
      }
    } catch (err: any) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleAISuggestion = async () => {
    try {
      setAiLoading(true);
      const response = await postMortemApi.generateAISuggestion(token!, incidentId);
      setRootCause(response.suggestion);
    } catch (err: any) {
      alert('AI提案の生成に失敗しました: ' + err.message);
    } finally {
      setAiLoading(false);
    }
  };

  const handleSave = async () => {
    try {
      setSaving(true);
      setError(null);

      const data = {
        root_cause: rootCause,
        impact_analysis: impactAnalysis,
        what_went_well: whatWentWell,
        what_went_wrong: whatWentWrong,
        lessons_learned: lessonsLearned,
        five_whys_analysis: fiveWhys,
      };

      if (postMortem) {
        // Update existing
        const updated = await postMortemApi.update(token!, postMortem.id, data);
        setPostMortem(updated);
        alert('Post-Mortemを保存しました');
      } else {
        // Create new
        const created = await postMortemApi.create(token!, {
          incident_id: incidentId,
          ...data,
        });
        setPostMortem(created);
        alert('Post-Mortemを作成しました');
      }
    } catch (err: any) {
      setError(err.message);
      alert('保存に失敗しました: ' + err.message);
    } finally {
      setSaving(false);
    }
  };

  const handlePublish = async () => {
    if (!postMortem) {
      alert('まず Post-Mortem を保存してください');
      return;
    }

    if (!confirm('Post-Mortemを公開しますか？公開後は編集できません。')) {
      return;
    }

    try {
      setSaving(true);
      const published = await postMortemApi.publish(token!, postMortem.id);
      setPostMortem(published);
      alert('Post-Mortemを公開しました');
    } catch (err: any) {
      alert('公開に失敗しました: ' + err.message);
    } finally {
      setSaving(false);
    }
  };

  const handleCreateActionItem = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!postMortem) {
      alert('まず Post-Mortem を保存してください');
      return;
    }

    try {
      const data = {
        post_mortem_id: postMortem.id,
        title: actionItemForm.title,
        description: actionItemForm.description,
        assignee_id: actionItemForm.assignee_id || undefined,
        priority: actionItemForm.priority,
        due_date: actionItemForm.due_date ? new Date(actionItemForm.due_date).toISOString() : undefined,
        related_links: actionItemForm.related_links,
      };

      if (editingActionItem) {
        // Update existing
        const updated = await actionItemApi.update(token!, editingActionItem.id, {
          ...data,
          status: actionItemForm.status,
        });
        setActionItems(actionItems.map(item => item.id === updated.id ? updated : item));
        alert('アクションアイテムを更新しました');
      } else {
        // Create new
        const created = await actionItemApi.create(token!, data);
        setActionItems([...actionItems, created]);
        alert('アクションアイテムを作成しました');
      }

      // Reset form
      setActionItemForm({
        title: '',
        description: '',
        assignee_id: 0,
        priority: 'medium',
        status: 'pending',
        due_date: '',
        related_links: '',
      });
      setShowActionItemForm(false);
      setEditingActionItem(null);
    } catch (err: any) {
      alert('操作に失敗しました: ' + err.message);
    }
  };

  const handleEditActionItem = (item: ActionItem) => {
    setEditingActionItem(item);
    setActionItemForm({
      title: item.title,
      description: item.description,
      assignee_id: item.assignee_id || 0,
      priority: item.priority,
      status: item.status,
      due_date: item.due_date ? item.due_date.substring(0, 16) : '',
      related_links: item.related_links,
    });
    setShowActionItemForm(true);
  };

  const handleDeleteActionItem = async (id: number) => {
    if (!confirm('このアクションアイテムを削除しますか？')) {
      return;
    }

    try {
      await actionItemApi.delete(token!, id);
      setActionItems(actionItems.filter(item => item.id !== id));
      alert('アクションアイテムを削除しました');
    } catch (err: any) {
      alert('削除に失敗しました: ' + err.message);
    }
  };

  const getPriorityColor = (priority: Priority) => {
    switch (priority) {
      case 'high':
        return 'bg-red-100 text-red-800';
      case 'medium':
        return 'bg-yellow-100 text-yellow-800';
      case 'low':
        return 'bg-green-100 text-green-800';
    }
  };

  const getStatusColor = (status: ActionStatus) => {
    switch (status) {
      case 'pending':
        return 'bg-gray-100 text-gray-800';
      case 'in_progress':
        return 'bg-blue-100 text-blue-800';
      case 'completed':
        return 'bg-green-100 text-green-800';
    }
  };

  const isEditable = !postMortem || postMortem.status === 'draft';

  if (authLoading || loading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="text-lg text-gray-600">読み込み中...</div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50 py-8 px-4">
      <div className="max-w-5xl mx-auto">
        {/* Header */}
        <div className="mb-6">
          <button
            onClick={() => router.push(`/incidents/${incidentId}`)}
            className="text-blue-600 hover:text-blue-800 mb-4 flex items-center"
          >
            ← インシデント詳細に戻る
          </button>
          <div className="flex justify-between items-start">
            <div>
              <h1 className="text-3xl font-bold text-gray-900">Post-Mortem</h1>
              {incident && (
                <p className="text-gray-900 mt-2">
                  インシデント: {incident.title}
                </p>
              )}
              {postMortem && (
                <div className="mt-2 flex items-center gap-2">
                  <span className={`px-3 py-1 rounded-full text-sm font-medium ${
                    postMortem.status === 'published'
                      ? 'bg-green-100 text-green-800'
                      : 'bg-gray-100 text-gray-800'
                  }`}>
                    {postMortem.status === 'published' ? '公開済み' : 'ドラフト'}
                  </span>
                </div>
              )}
            </div>
          </div>
        </div>

        {error && (
          <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded mb-6">
            {error}
          </div>
        )}

        {/* Root Cause Section */}
        <section className="bg-white rounded-lg shadow p-6 mb-6">
          <div className="flex justify-between items-center mb-4">
            <h2 className="text-xl font-semibold text-gray-900">Root Cause (根本原因)</h2>
            {isEditable && (
              <button
                onClick={handleAISuggestion}
                disabled={aiLoading}
                className="px-4 py-2 bg-purple-600 text-white rounded-md hover:bg-purple-700 disabled:bg-gray-400"
              >
                {aiLoading ? 'AI提案生成中...' : 'AI提案を取得'}
              </button>
            )}
          </div>
          {postMortem?.ai_root_cause_suggestion && (
            <div className="bg-purple-50 border border-purple-200 rounded-md p-4 mb-4">
              <p className="text-sm font-medium text-purple-800">AI提案:</p>
              <p className="text-sm text-purple-700 mt-1">{postMortem.ai_root_cause_suggestion}</p>
            </div>
          )}
          <textarea
            value={rootCause}
            onChange={(e) => setRootCause(e.target.value)}
            disabled={!isEditable}
            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-100 disabled:text-gray-700"
            rows={4}
            placeholder="インシデントの根本原因を記述してください..."
          />
        </section>

        {/* Impact Analysis */}
        <section className="bg-white rounded-lg shadow p-6 mb-6">
          <h2 className="text-xl font-semibold text-gray-900 mb-4">Impact Analysis (影響分析)</h2>
          <textarea
            value={impactAnalysis}
            onChange={(e) => setImpactAnalysis(e.target.value)}
            disabled={!isEditable}
            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-100 disabled:text-gray-700"
            rows={4}
            placeholder="インシデントの影響範囲と程度を分析してください..."
          />
        </section>

        {/* What Went Well/Wrong */}
        <section className="bg-white rounded-lg shadow p-6 mb-6">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div>
              <h2 className="text-xl font-semibold text-gray-900 mb-4">
                What Went Well (うまくいったこと)
              </h2>
              <textarea
                value={whatWentWell}
                onChange={(e) => setWhatWentWell(e.target.value)}
                disabled={!isEditable}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-100 disabled:text-gray-700"
                rows={6}
                placeholder="対応でうまくいったことを記述してください..."
              />
            </div>
            <div>
              <h2 className="text-xl font-semibold text-gray-900 mb-4">
                What Went Wrong (問題点)
              </h2>
              <textarea
                value={whatWentWrong}
                onChange={(e) => setWhatWentWrong(e.target.value)}
                disabled={!isEditable}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-100 disabled:text-gray-700"
                rows={6}
                placeholder="対応で問題だった点を記述してください..."
              />
            </div>
          </div>
        </section>

        {/* Lessons Learned */}
        <section className="bg-white rounded-lg shadow p-6 mb-6">
          <h2 className="text-xl font-semibold text-gray-900 mb-4">Lessons Learned (教訓)</h2>
          <textarea
            value={lessonsLearned}
            onChange={(e) => setLessonsLearned(e.target.value)}
            disabled={!isEditable}
            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-100 disabled:text-gray-700"
            rows={4}
            placeholder="今回のインシデントから得られた教訓を記述してください..."
          />
        </section>

        {/* 5 Whys Analysis */}
        <section className="bg-white rounded-lg shadow p-6 mb-6">
          <h2 className="text-xl font-semibold text-gray-900 mb-4">5 Whys Analysis (5回のなぜ)</h2>
          <div className="space-y-4">
            {[1, 2, 3, 4, 5].map((index) => (
              <div key={index}>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Why {index}
                </label>
                <input
                  type="text"
                  value={fiveWhys[`why${index}` as keyof FiveWhysAnalysis]}
                  onChange={(e) =>
                    setFiveWhys({ ...fiveWhys, [`why${index}`]: e.target.value })
                  }
                  disabled={!isEditable}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-100 disabled:text-gray-700"
                  placeholder={`なぜ ${index === 1 ? 'この問題が発生したのか' : '前の答えなのか'}？`}
                />
              </div>
            ))}
          </div>
        </section>

        {/* Action Items */}
        <section className="bg-white rounded-lg shadow p-6 mb-6">
          <div className="flex justify-between items-center mb-4">
            <h2 className="text-xl font-semibold text-gray-900">
              Action Items (アクションアイテム)
            </h2>
            {postMortem && isEditable && (
              <button
                onClick={() => {
                  setShowActionItemForm(true);
                  setEditingActionItem(null);
                  setActionItemForm({
                    title: '',
                    description: '',
                    assignee_id: 0,
                    priority: 'medium',
                    status: 'pending',
                    due_date: '',
                    related_links: '',
                  });
                }}
                className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
              >
                + 追加
              </button>
            )}
          </div>

          {/* Action Items List */}
          <div className="space-y-4 mb-4">
            {actionItems.map((item) => (
              <div key={item.id} className="border rounded-lg p-4 hover:bg-gray-50">
                <div className="flex justify-between items-start">
                  <div className="flex-1">
                    <h3 className="font-semibold text-gray-900">{item.title}</h3>
                    {item.description && (
                      <p className="text-sm text-gray-900 mt-1">{item.description}</p>
                    )}
                    <div className="flex gap-2 mt-2 flex-wrap">
                      <span className={`px-2 py-1 text-xs rounded font-medium ${getPriorityColor(item.priority)}`}>
                        {item.priority}
                      </span>
                      <span className={`px-2 py-1 text-xs rounded font-medium ${getStatusColor(item.status)}`}>
                        {item.status}
                      </span>
                      {item.assignee && (
                        <span className="px-2 py-1 text-xs bg-gray-100 text-gray-800 rounded">
                          担当: {item.assignee.name}
                        </span>
                      )}
                      {item.due_date && (
                        <span className="px-2 py-1 text-xs bg-gray-100 text-gray-800 rounded">
                          期限: {new Date(item.due_date).toLocaleDateString('ja-JP')}
                        </span>
                      )}
                    </div>
                  </div>
                  {isEditable && (
                    <div className="flex gap-2 ml-4">
                      <button
                        onClick={() => handleEditActionItem(item)}
                        className="text-blue-600 hover:text-blue-800 text-sm"
                      >
                        編集
                      </button>
                      <button
                        onClick={() => handleDeleteActionItem(item.id)}
                        className="text-red-600 hover:text-red-800 text-sm"
                      >
                        削除
                      </button>
                    </div>
                  )}
                </div>
              </div>
            ))}

            {actionItems.length === 0 && (
              <p className="text-gray-500 text-center py-8">
                アクションアイテムはまだありません
              </p>
            )}
          </div>

          {/* Action Item Form */}
          {showActionItemForm && (
            <form onSubmit={handleCreateActionItem} className="border-t pt-4">
              <h3 className="font-semibold text-gray-900 mb-4">
                {editingActionItem ? 'アクションアイテムを編集' : '新しいアクションアイテム'}
              </h3>
              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    タイトル *
                  </label>
                  <input
                    type="text"
                    value={actionItemForm.title}
                    onChange={(e) =>
                      setActionItemForm({ ...actionItemForm, title: e.target.value })
                    }
                    required
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                    placeholder="アクションアイテムのタイトル"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    説明
                  </label>
                  <textarea
                    value={actionItemForm.description}
                    onChange={(e) =>
                      setActionItemForm({ ...actionItemForm, description: e.target.value })
                    }
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                    rows={3}
                    placeholder="詳細な説明"
                  />
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      優先度
                    </label>
                    <select
                      value={actionItemForm.priority}
                      onChange={(e) =>
                        setActionItemForm({
                          ...actionItemForm,
                          priority: e.target.value as Priority,
                        })
                      }
                      className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                    >
                      <option value="high">High</option>
                      <option value="medium">Medium</option>
                      <option value="low">Low</option>
                    </select>
                  </div>

                  {editingActionItem && (
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        ステータス
                      </label>
                      <select
                        value={actionItemForm.status}
                        onChange={(e) =>
                          setActionItemForm({
                            ...actionItemForm,
                            status: e.target.value as ActionStatus,
                          })
                        }
                        className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                      >
                        <option value="pending">Pending</option>
                        <option value="in_progress">In Progress</option>
                        <option value="completed">Completed</option>
                      </select>
                    </div>
                  )}

                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      担当者
                    </label>
                    <select
                      value={actionItemForm.assignee_id}
                      onChange={(e) =>
                        setActionItemForm({
                          ...actionItemForm,
                          assignee_id: parseInt(e.target.value),
                        })
                      }
                      className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                    >
                      <option value={0}>未割り当て</option>
                      {users.map((u) => (
                        <option key={u.id} value={u.id}>
                          {u.name}
                        </option>
                      ))}
                    </select>
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      期限
                    </label>
                    <input
                      type="datetime-local"
                      value={actionItemForm.due_date}
                      onChange={(e) =>
                        setActionItemForm({ ...actionItemForm, due_date: e.target.value })
                      }
                      className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                    />
                  </div>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    関連リンク
                  </label>
                  <input
                    type="text"
                    value={actionItemForm.related_links}
                    onChange={(e) =>
                      setActionItemForm({ ...actionItemForm, related_links: e.target.value })
                    }
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                    placeholder="関連するリンク（カンマ区切り）"
                  />
                </div>

                <div className="flex gap-2">
                  <button
                    type="submit"
                    className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
                  >
                    {editingActionItem ? '更新' : '作成'}
                  </button>
                  <button
                    type="button"
                    onClick={() => {
                      setShowActionItemForm(false);
                      setEditingActionItem(null);
                    }}
                    className="px-4 py-2 bg-gray-300 text-gray-700 rounded-md hover:bg-gray-400"
                  >
                    キャンセル
                  </button>
                </div>
              </div>
            </form>
          )}
        </section>

        {/* Save/Publish Buttons */}
        <div className="flex justify-end gap-4">
          {isEditable && (
            <>
              <button
                onClick={handleSave}
                disabled={saving}
                className="px-6 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:bg-gray-400"
              >
                {saving ? '保存中...' : '保存'}
              </button>
              {postMortem && (
                <button
                  onClick={handlePublish}
                  disabled={saving}
                  className="px-6 py-2 bg-green-600 text-white rounded-md hover:bg-green-700 disabled:bg-gray-400"
                >
                  公開
                </button>
              )}
            </>
          )}
        </div>
      </div>
    </div>
  );
}
