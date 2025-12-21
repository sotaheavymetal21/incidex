'use client';

import { useState, useEffect } from 'react';
import { useAuth } from '@/context/AuthContext';
import { templateApi } from '@/lib/api';
import { IncidentTemplate } from '@/types/template';
import { useRouter } from 'next/navigation';

export default function TemplatesPage() {
  const { token } = useAuth();
  const router = useRouter();
  const [templates, setTemplates] = useState<IncidentTemplate[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!token) {
      router.push('/login');
      return;
    }

    fetchTemplates();
  }, [token]);

  const fetchTemplates = async () => {
    if (!token) return;

    try {
      setLoading(true);
      const data = await templateApi.getAll(token);
      setTemplates(data);
    } catch (err: any) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (id: number) => {
    if (!token) return;
    if (!confirm('このテンプレートを削除してもよろしいですか？')) return;

    try {
      await templateApi.delete(token, id);
      fetchTemplates();
    } catch (err: any) {
      alert(`削除に失敗しました: ${err.message}`);
    }
  };

  const getSeverityBadge = (severity: string) => {
    const colors: Record<string, string> = {
      critical: 'bg-red-100 text-red-800',
      high: 'bg-orange-100 text-orange-800',
      medium: 'bg-yellow-100 text-yellow-800',
      low: 'bg-green-100 text-green-800',
    };

    const labels: Record<string, string> = {
      critical: '致命的',
      high: '高',
      medium: '中',
      low: '低',
    };

    return (
      <span className={`px-2 py-1 rounded text-xs font-semibold ${colors[severity]}`}>
        {labels[severity]}
      </span>
    );
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-gray-600">読み込み中...</div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50 py-8">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        {/* ヘッダー */}
        <div className="mb-8 flex justify-between items-center">
          <div>
            <h1 className="text-3xl font-bold text-gray-900">インシデントテンプレート</h1>
            <p className="mt-2 text-gray-600">
              よくあるインシデントのテンプレートを管理します
            </p>
          </div>
          <button
            onClick={() => router.push('/templates/create')}
            className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition"
          >
            新規テンプレート作成
          </button>
        </div>

        {/* エラーメッセージ */}
        {error && (
          <div className="mb-6 bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded">
            {error}
          </div>
        )}

        {/* テンプレート一覧 */}
        {templates.length === 0 ? (
          <div className="bg-white shadow rounded-lg p-8 text-center">
            <p className="text-gray-500 mb-4">テンプレートがまだありません</p>
            <button
              onClick={() => router.push('/templates/create')}
              className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition"
            >
              最初のテンプレートを作成
            </button>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {templates.map((template) => (
              <div
                key={template.id}
                className="bg-white shadow rounded-lg p-6 hover:shadow-lg transition"
              >
                <div className="flex justify-between items-start mb-4">
                  <h3 className="text-lg font-semibold text-gray-900 flex-1">
                    {template.name}
                  </h3>
                  {template.is_public && (
                    <span className="px-2 py-1 bg-blue-100 text-blue-800 text-xs rounded font-semibold ml-2">
                      公開
                    </span>
                  )}
                </div>

                <p className="text-sm text-gray-600 mb-4 line-clamp-2">
                  {template.description || 'テンプレートの説明はありません'}
                </p>

                <div className="mb-4">
                  <div className="text-xs text-gray-500 mb-1">重要度</div>
                  {getSeverityBadge(template.severity)}
                </div>

                {template.tags.length > 0 && (
                  <div className="mb-4 flex flex-wrap gap-1">
                    {template.tags.map((tag) => (
                      <span
                        key={tag.id}
                        className="px-2 py-1 rounded text-xs"
                        style={{ backgroundColor: tag.color + '20', color: tag.color }}
                      >
                        {tag.name}
                      </span>
                    ))}
                  </div>
                )}

                <div className="flex items-center justify-between text-xs text-gray-500 mb-4">
                  <span>使用回数: {template.usage_count}</span>
                  {template.creator && <span>作成者: {template.creator.name}</span>}
                </div>

                <div className="flex gap-2">
                  <button
                    onClick={() => router.push(`/incidents/create?template=${template.id}`)}
                    className="flex-1 px-3 py-2 bg-green-600 text-white text-sm rounded hover:bg-green-700 transition"
                  >
                    このテンプレートで作成
                  </button>
                  <button
                    onClick={() => router.push(`/templates/${template.id}/edit`)}
                    className="px-3 py-2 bg-gray-200 text-gray-700 text-sm rounded hover:bg-gray-300 transition"
                  >
                    編集
                  </button>
                  <button
                    onClick={() => handleDelete(template.id)}
                    className="px-3 py-2 bg-red-100 text-red-700 text-sm rounded hover:bg-red-200 transition"
                  >
                    削除
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
