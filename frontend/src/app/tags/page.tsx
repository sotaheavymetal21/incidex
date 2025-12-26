'use client';

import { useState, useEffect } from 'react';
import { useAuth } from '@/context/AuthContext';
import { usePermissions } from '@/hooks/usePermissions';
import { tagApi } from '@/lib/api';
import { Tag } from '@/types/tag';
import { useRouter } from 'next/navigation';

export default function TagsPage() {
  const { token, loading: authLoading } = useAuth();
  const permissions = usePermissions();
  const router = useRouter();
  const [tags, setTags] = useState<Tag[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  // Modal states
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingTag, setEditingTag] = useState<Tag | null>(null);
  const [formData, setFormData] = useState({ name: '', color: '#10b981' });

  useEffect(() => {
    if (authLoading) {
      return; // 認証状態の読み込み中は何もしない
    }

    if (!token) {
      router.push('/login');
      return;
    }

    fetchTags();
  }, [token, authLoading, router]);

  const fetchTags = async () => {
    if (!token) return;
    try {
      setLoading(true);
      const data = await tagApi.getAll(token);
      setTags(data);
    } catch (err: any) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!token) return;

    try {
      if (editingTag) {
        await tagApi.update(token, editingTag.id, formData);
      } else {
        await tagApi.create(token, formData);
      }
      setIsModalOpen(false);
      setEditingTag(null);
      setFormData({ name: '', color: '#10b981' });
      fetchTags();
    } catch (err: any) {
      alert(err.message);
    }
  };

  const handleDelete = async (id: number) => {
    if (!token || !confirm('Are you sure you want to delete this tag?')) return;
    try {
      await tagApi.delete(token, id);
      fetchTags();
    } catch (err: any) {
      alert(err.message);
    }
  };

  const openEditModal = (tag: Tag) => {
    setEditingTag(tag);
    setFormData({ name: tag.name, color: tag.color });
    setIsModalOpen(true);
  };

  const openCreateModal = () => {
    setEditingTag(null);
    setFormData({ name: '', color: '#10b981' });
    setIsModalOpen(true);
  };

  if (authLoading || (loading && !tags.length)) {
    return <div className="p-8 text-center" style={{ color: 'var(--secondary)' }}>Loading...</div>;
  }

  return (
    <div className="min-h-screen py-8 px-4 sm:px-6 lg:px-8" style={{ background: 'var(--background)' }}>
      <div className="container mx-auto max-w-5xl">
        <div className="flex justify-between items-center mb-8 animate-slideDown">
          <div>
            <h1
              className="text-4xl font-bold mb-2"
              style={{
                color: 'var(--foreground)',
                fontFamily: 'var(--font-display)',
                background: 'linear-gradient(135deg, var(--foreground) 0%, var(--primary) 100%)',
                WebkitBackgroundClip: 'text',
                WebkitTextFillColor: 'transparent'
              }}
            >
              タグ管理
            </h1>
            <p className="text-base font-medium" style={{ color: 'var(--foreground-secondary)', fontFamily: 'var(--font-body)' }}>
              インシデントの分類用タグを管理
            </p>
          </div>
          {permissions.canManageTags && (
            <button
              onClick={openCreateModal}
              className="px-5 py-2.5 text-white rounded-xl font-bold transition-all duration-200"
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
              タグを作成
            </button>
          )}
        </div>

        {error && (
          <div className="px-4 py-3 rounded-xl mb-4 border-2" style={{ background: 'var(--error-light)', borderColor: 'var(--error)', color: 'var(--error)' }}>
            {error}
          </div>
        )}

        <div className="rounded-xl shadow-lg overflow-hidden border" style={{ background: 'var(--surface)', borderColor: 'var(--border)' }}>
          <table className="min-w-full">
            <thead style={{ background: 'var(--secondary-light)' }}>
              <tr>
                <th className="px-6 py-3.5 text-left text-xs font-semibold uppercase tracking-wider" style={{ color: 'var(--foreground)' }}>Name</th>
                <th className="px-6 py-3.5 text-left text-xs font-semibold uppercase tracking-wider" style={{ color: 'var(--foreground)' }}>Color</th>
                <th className="px-6 py-3.5 text-right text-xs font-semibold uppercase tracking-wider" style={{ color: 'var(--foreground)' }}>Actions</th>
              </tr>
            </thead>
            <tbody>
              {tags.map((tag, index) => (
                <tr
                  key={tag.id}
                  className="transition-all"
                  style={{
                    borderTop: index > 0 ? `1px solid var(--border)` : 'none'
                  }}
                  onMouseEnter={(e) => e.currentTarget.style.background = 'var(--secondary-light)'}
                  onMouseLeave={(e) => e.currentTarget.style.background = 'transparent'}
                >
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span className="px-3 py-1.5 inline-flex text-sm font-semibold rounded-full text-white shadow-sm" style={{ backgroundColor: tag.color }}>
                      {tag.name}
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="flex items-center">
                      <div className="h-7 w-7 rounded-lg border-2 shadow-sm" style={{ backgroundColor: tag.color, borderColor: 'var(--border)' }}></div>
                      <span className="ml-3 text-sm font-medium" style={{ color: 'var(--secondary)' }}>{tag.color}</span>
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                    {permissions.canManageTags ? (
                      <div className="flex justify-end gap-3">
                        <button
                          onClick={() => openEditModal(tag)}
                          className="transition-colors font-medium"
                          style={{ color: 'var(--primary)' }}
                          onMouseEnter={(e) => e.currentTarget.style.color = 'var(--primary-hover)'}
                          onMouseLeave={(e) => e.currentTarget.style.color = 'var(--primary)'}
                        >
                          Edit
                        </button>
                        <button
                          onClick={() => handleDelete(tag.id)}
                          className="transition-colors font-medium"
                          style={{ color: 'var(--error)' }}
                          onMouseEnter={(e) => e.currentTarget.style.opacity = '0.7'}
                          onMouseLeave={(e) => e.currentTarget.style.opacity = '1'}
                        >
                          Delete
                        </button>
                      </div>
                    ) : (
                      <span style={{ color: 'var(--secondary)' }}>閲覧のみ</span>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        {isModalOpen && (
          <div className="fixed inset-0 bg-black bg-opacity-40 overflow-y-auto h-full w-full flex items-center justify-center z-50 backdrop-blur-sm animate-fadeIn">
            <div
              className="p-8 rounded-2xl w-96 border-2 animate-scaleIn"
              style={{
                background: 'var(--surface)',
                borderColor: 'var(--primary)',
                boxShadow: '0 0 40px var(--primary-glow), 0 20px 40px rgba(0,0,0,0.2)'
              }}
            >
              <h2
                className="text-2xl font-bold mb-6"
                style={{
                  color: 'var(--foreground)',
                  fontFamily: 'var(--font-display)'
                }}
              >
                {editingTag ? 'タグを編集' : 'タグを作成'}
              </h2>
              <form onSubmit={handleSubmit}>
                <div className="mb-5">
                  <label className="block text-sm font-semibold mb-2" style={{ color: 'var(--foreground)' }}>Name</label>
                  <input
                    type="text"
                    value={formData.name}
                    onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                    className="w-full py-2.5 px-3 border-2 rounded-lg focus:outline-none transition-all"
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
                    required
                  />
                </div>
                <div className="mb-6">
                  <label className="block text-sm font-semibold mb-2" style={{ color: 'var(--foreground)' }}>Color</label>
                  <input
                    type="color"
                    value={formData.color}
                    onChange={(e) => setFormData({ ...formData, color: e.target.value })}
                    className="h-12 w-full rounded-lg cursor-pointer border-2"
                    style={{ borderColor: 'var(--border)' }}
                  />
                  <p className="mt-2 text-xs" style={{ color: 'var(--secondary)' }}>選択した色: {formData.color}</p>
                </div>
                <div className="flex gap-3">
                  <button
                    type="button"
                    onClick={() => setIsModalOpen(false)}
                    className="flex-1 px-4 py-2.5 border-2 rounded-xl transition-all duration-200 font-semibold"
                    style={{
                      borderColor: 'var(--border)',
                      color: 'var(--foreground)',
                      background: 'var(--surface)',
                      fontFamily: 'var(--font-body)'
                    }}
                    onMouseEnter={(e) => {
                      e.currentTarget.style.background = 'var(--gray-100)';
                      e.currentTarget.style.borderColor = 'var(--foreground-secondary)';
                      e.currentTarget.style.transform = 'translateY(-2px)';
                    }}
                    onMouseLeave={(e) => {
                      e.currentTarget.style.background = 'var(--surface)';
                      e.currentTarget.style.borderColor = 'var(--border)';
                      e.currentTarget.style.transform = 'translateY(0)';
                    }}
                  >
                    キャンセル
                  </button>
                  <button
                    type="submit"
                    className="flex-1 px-4 py-2.5 text-white rounded-xl transition-all duration-200 font-bold"
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
                    保存
                  </button>
                </div>
              </form>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
