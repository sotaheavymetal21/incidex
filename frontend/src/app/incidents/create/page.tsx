'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from '@/context/AuthContext';
import { incidentApi, tagApi, userApi } from '@/lib/api';
import { Severity, Status, User } from '@/types/incident';
import { Tag } from '@/types/tag';
import {
  validateIncidentTitle,
  validateIncidentDescription,
  validateImpactScope,
  validateDatetime,
  ValidationLimits,
} from '@/utils/validation';

interface FieldErrors {
  title?: string;
  description?: string;
  impactScope?: string;
  detectedAt?: string;
}

function CreateIncidentForm() {
  const { token, loading: authLoading } = useAuth();
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const [tags, setTags] = useState<Tag[]>([]);
  const [users, setUsers] = useState<User[]>([]);
  const [error, setError] = useState('');
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({});

  // Form state
  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [severity, setSeverity] = useState<Severity>('medium');
  const [impactScope, setImpactScope] = useState('');
  const [detectedAt, setDetectedAt] = useState('');
  const [assigneeId, setAssigneeId] = useState<number | ''>('');
  const [selectedTagIds, setSelectedTagIds] = useState<number[]>([]);

  useEffect(() => {
    if (!authLoading && !token) {
      router.push('/login');
    }
  }, [token, authLoading, router]);

  useEffect(() => {
    if (token) {
      fetchTags();
      fetchUsers();
      // Set default detected_at to current time
      const now = new Date();
      const localDateTime = new Date(now.getTime() - now.getTimezoneOffset() * 60000)
        .toISOString()
        .slice(0, 16);
      setDetectedAt(localDateTime);
    }
  }, [token]);

  const fetchTags = async () => {
    try {
      const fetchedTags = await tagApi.getAll(token!);
      setTags(fetchedTags);
    } catch (err) {
      console.error('Failed to fetch tags:', err);
    }
  };

  const fetchUsers = async () => {
    try {
      const fetchedUsers = await userApi.getAll(token!);
      setUsers(fetchedUsers);
    } catch (err) {
      console.error('Failed to fetch users:', err);
    }
  };

  const handleTagToggle = (tagId: number) => {
    setSelectedTagIds((prev) =>
      prev.includes(tagId) ? prev.filter((id) => id !== tagId) : [...prev, tagId]
    );
  };

  const validateField = (field: keyof FieldErrors, value: string): string | undefined => {
    switch (field) {
      case 'title': {
        const result = validateIncidentTitle(value);
        return result.isValid ? undefined : result.error;
      }
      case 'description': {
        const result = validateIncidentDescription(value);
        return result.isValid ? undefined : result.error;
      }
      case 'impactScope': {
        const result = validateImpactScope(value);
        return result.isValid ? undefined : result.error;
      }
      case 'detectedAt': {
        const result = validateDatetime(value, true, '検出日時');
        return result.isValid ? undefined : result.error;
      }
    }
    return undefined;
  };

  const handleFieldBlur = (field: keyof FieldErrors, value: string) => {
    const error = validateField(field, value);
    setFieldErrors((prev) => ({ ...prev, [field]: error }));
  };

  const validateForm = (): boolean => {
    const errors: FieldErrors = {
      title: validateField('title', title),
      description: validateField('description', description),
      impactScope: validateField('impactScope', impactScope),
      detectedAt: validateField('detectedAt', detectedAt),
    };
    setFieldErrors(errors);
    return !Object.values(errors).some((e) => e !== undefined);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    if (!validateForm()) {
      return;
    }

    setLoading(true);
    try {
      const data = {
        title: title.trim(),
        description: description.trim(),
        severity,
        status: 'open' as Status,
        impact_scope: impactScope.trim(),
        detected_at: new Date(detectedAt).toISOString(),
        assignee_id: assigneeId || undefined,
        tag_ids: selectedTagIds,
      };

      const incident = await incidentApi.create(token!, data);
      router.push(`/incidents/${incident.id}`);
    } catch (err: any) {
      setError(err.message || 'インシデントの作成に失敗しました');
    } finally {
      setLoading(false);
    }
  };

  if (authLoading || !token) {
    return (
      <div className="min-h-screen flex items-center justify-center" style={{ background: 'var(--background)' }}>
        <div style={{ color: 'var(--secondary)' }}>Loading...</div>
      </div>
    );
  }

  return (
    <div className="min-h-screen py-8 px-4 sm:px-6 lg:px-8" style={{ background: 'var(--background)' }}>
      <div className="max-w-3xl mx-auto">
        <div className="mb-6">
          <button
            onClick={() => router.push('/incidents')}
            className="mb-4 inline-flex items-center transition-colors"
            style={{ color: 'var(--primary)' }}
            onMouseEnter={(e) => e.currentTarget.style.color = 'var(--primary-hover)'}
            onMouseLeave={(e) => e.currentTarget.style.color = 'var(--primary)'}
          >
            ← 一覧に戻る
          </button>
          <h1 className="text-3xl font-bold" style={{ color: 'var(--foreground)' }}>新規インシデント作成</h1>
        </div>

        {error && (
          <div className="px-4 py-3 rounded-xl mb-4 border-2" style={{ background: 'var(--error-light)', borderColor: 'var(--error)', color: 'var(--error)' }}>
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="rounded-xl shadow-lg p-6 border" style={{ background: 'var(--surface)', borderColor: 'var(--border)' }}>
          {/* Title */}
          <div className="mb-5">
            <label htmlFor="title" className="block text-sm font-semibold mb-2" style={{ color: 'var(--foreground)' }}>
              タイトル <span style={{ color: 'var(--error)' }}>*</span>
            </label>
            <input
              type="text"
              id="title"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              maxLength={ValidationLimits.TITLE_MAX_LENGTH}
              required
              className="w-full px-3 py-2.5 border-2 rounded-lg focus:outline-none transition-all"
              style={{
                background: 'var(--surface)',
                borderColor: fieldErrors.title ? 'var(--error)' : 'var(--border)',
                color: 'var(--foreground)'
              }}
              onFocus={(e) => {
                e.currentTarget.style.borderColor = 'var(--primary)';
                e.currentTarget.style.boxShadow = '0 0 0 3px var(--primary-light)';
              }}
              onBlur={(e) => {
                handleFieldBlur('title', e.target.value);
                e.currentTarget.style.borderColor = fieldErrors.title ? 'var(--error)' : 'var(--border)';
                e.currentTarget.style.boxShadow = 'none';
              }}
            />
            {fieldErrors.title && (
              <p className="mt-1 text-xs" style={{ color: 'var(--error)' }}>{fieldErrors.title}</p>
            )}
            <p className="mt-1 text-xs" style={{ color: 'var(--secondary)' }}>{ValidationLimits.TITLE_MAX_LENGTH}文字以内</p>
          </div>

          {/* Description */}
          <div className="mb-5">
            <label htmlFor="description" className="block text-sm font-semibold mb-2" style={{ color: 'var(--foreground)' }}>
              説明 <span style={{ color: 'var(--error)' }}>*</span>
            </label>
            <textarea
              id="description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              required
              rows={6}
              className="w-full px-3 py-2.5 border-2 rounded-lg focus:outline-none transition-all"
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

          {/* Severity */}
          <div className="mb-5">
            <label htmlFor="severity" className="block text-sm font-semibold mb-2" style={{ color: 'var(--foreground)' }}>
              重要度 <span style={{ color: 'var(--error)' }}>*</span>
            </label>
            <select
              id="severity"
              value={severity}
              onChange={(e) => setSeverity(e.target.value as Severity)}
              required
              className="w-full px-3 py-2.5 border-2 rounded-lg focus:outline-none transition-all"
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
            >
              <option value="low">🟢 Low - 軽微な問題（通常業務時間内で対応）</option>
              <option value="medium">🟡 Medium - 機能劣化あり（4時間以内に対応開始）</option>
              <option value="high">🟠 High - 主要機能に重大な影響（1時間以内に対応開始）</option>
              <option value="critical">🔴 Critical - サービス停止・全体障害（即座に対応）</option>
            </select>
            <p className="mt-1.5 text-xs" style={{ color: 'var(--secondary)' }}>
              詳細な基準は <a href="/docs/severity-guidelines.md" target="_blank" className="transition-colors" style={{ color: 'var(--primary)' }} onMouseEnter={(e) => e.currentTarget.style.textDecoration = 'underline'} onMouseLeave={(e) => e.currentTarget.style.textDecoration = 'none'}>Severityガイドライン</a> を参照してください
            </p>
          </div>

          {/* Impact Scope */}
          <div className="mb-5">
            <label htmlFor="impactScope" className="block text-sm font-semibold mb-2" style={{ color: 'var(--foreground)' }}>
              影響範囲
            </label>
            <input
              type="text"
              id="impactScope"
              value={impactScope}
              onChange={(e) => setImpactScope(e.target.value)}
              maxLength={500}
              className="w-full px-3 py-2.5 border-2 rounded-lg focus:outline-none transition-all"
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

          {/* Detected At */}
          <div className="mb-5">
            <label htmlFor="detectedAt" className="block text-sm font-semibold mb-2" style={{ color: 'var(--foreground)' }}>
              検出日時 <span style={{ color: 'var(--error)' }}>*</span>
            </label>
            <input
              type="datetime-local"
              id="detectedAt"
              value={detectedAt}
              onChange={(e) => setDetectedAt(e.target.value)}
              required
              className="w-full px-3 py-2.5 border-2 rounded-lg focus:outline-none transition-all"
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

          {/* Assignee */}
          <div className="mb-5">
            <label htmlFor="assignee" className="block text-sm font-semibold mb-2" style={{ color: 'var(--foreground)' }}>
              担当者
            </label>
            <select
              id="assignee"
              value={assigneeId}
              onChange={(e) => setAssigneeId(e.target.value ? parseInt(e.target.value) : '')}
              className="w-full px-3 py-2.5 border-2 rounded-lg focus:outline-none transition-all"
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
            >
              <option value="">未割り当て</option>
              {users.map((user) => (
                <option key={user.id} value={user.id}>
                  {user.name} ({user.email})
                </option>
              ))}
            </select>
          </div>

          {/* Tags */}
          <div className="mb-6">
            <label className="block text-sm font-semibold mb-3" style={{ color: 'var(--foreground)' }}>タグ</label>
            <div className="flex flex-wrap gap-2">
              {tags.map((tag) => (
                <button
                  key={tag.id}
                  type="button"
                  onClick={() => handleTagToggle(tag.id)}
                  className="px-3 py-1.5 rounded-full text-sm border-2 transition-all shadow-sm"
                  style={
                    selectedTagIds.includes(tag.id)
                      ? { backgroundColor: tag.color, borderColor: tag.color, color: 'white' }
                      : { background: 'var(--surface)', borderColor: 'var(--border)', color: 'var(--foreground)' }
                  }
                  onMouseEnter={(e) => {
                    if (!selectedTagIds.includes(tag.id)) {
                      e.currentTarget.style.borderColor = tag.color;
                      e.currentTarget.style.color = tag.color;
                    }
                  }}
                  onMouseLeave={(e) => {
                    if (!selectedTagIds.includes(tag.id)) {
                      e.currentTarget.style.borderColor = 'var(--border)';
                      e.currentTarget.style.color = 'var(--foreground)';
                    }
                  }}
                >
                  {tag.name}
                </button>
              ))}
            </div>
          </div>

          {/* Submit Buttons */}
          <div className="flex gap-4">
            <button
              type="submit"
              disabled={loading}
              className="flex-1 px-4 py-2.5 text-white rounded-lg shadow-lg transition-all disabled:opacity-50 font-medium"
              style={{ background: 'var(--primary)' }}
              onMouseEnter={(e) => !loading && (e.currentTarget.style.background = 'var(--primary-hover)')}
              onMouseLeave={(e) => e.currentTarget.style.background = 'var(--primary)'}
            >
              {loading ? '作成中...' : 'インシデントを作成'}
            </button>
            <button
              type="button"
              onClick={() => router.push('/incidents')}
              className="px-6 py-2.5 border-2 rounded-lg transition-all font-medium"
              style={{ borderColor: 'var(--border)', color: 'var(--foreground)', background: 'var(--surface)' }}
              onMouseEnter={(e) => {
                e.currentTarget.style.background = 'var(--secondary-light)';
                e.currentTarget.style.borderColor = 'var(--secondary)';
              }}
              onMouseLeave={(e) => {
                e.currentTarget.style.background = 'var(--surface)';
                e.currentTarget.style.borderColor = 'var(--border)';
              }}
            >
              キャンセル
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

export default function CreateIncidentPage() {
  return <CreateIncidentForm />;
}
