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
  const [formData, setFormData] = useState({ name: '', color: '#808080' });

  useEffect(() => {
    if (!authLoading && !token) {
      router.push('/login');
    } else if (token) {
      fetchTags();
    }
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
      setFormData({ name: '', color: '#808080' });
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
    setFormData({ name: '', color: '#808080' });
    setIsModalOpen(true);
  };

  if (authLoading || (loading && !tags.length)) {
    return <div className="p-8 text-center">Loading...</div>;
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold">Tag Management</h1>
        {permissions.canManageTags && (
          <button
            onClick={openCreateModal}
            className="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700"
          >
            Create Tag
          </button>
        )}
      </div>

      {error && <div className="text-red-500 mb-4">{error}</div>}

      <div className="bg-white shadow-md rounded-lg overflow-hidden">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Name</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Color</th>
              <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
            </tr>
          </thead>
          <tbody className="bg-white divide-y divide-gray-200">
            {tags.map((tag) => (
              <tr key={tag.id}>
                <td className="px-6 py-4 whitespace-nowrap">
                  <span className="px-2 inline-flex text-xs leading-5 font-semibold rounded-full bg-gray-100" style={{ color: tag.color }}>
                    {tag.name}
                  </span>
                </td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <div className="flex items-center">
                    <div className="h-6 w-6 rounded border border-gray-300" style={{ backgroundColor: tag.color }}></div>
                    <span className="ml-2 text-sm text-gray-500">{tag.color}</span>
                  </div>
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                  {permissions.canManageTags ? (
                    <>
                      <button onClick={() => openEditModal(tag)} className="text-indigo-600 hover:text-indigo-900 mr-4">Edit</button>
                      <button onClick={() => handleDelete(tag.id)} className="text-red-600 hover:text-red-900">Delete</button>
                    </>
                  ) : (
                    <span className="text-gray-400">閲覧のみ</span>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {isModalOpen && (
        <div className="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full flex items-center justify-center">
          <div className="bg-white p-8 rounded-lg shadow-xl w-96">
            <h2 className="text-xl font-bold mb-4">{editingTag ? 'Edit Tag' : 'Create Tag'}</h2>
            <form onSubmit={handleSubmit}>
              <div className="mb-4">
                <label className="block text-gray-700 text-sm font-bold mb-2">Name</label>
                <input
                  type="text"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  className="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline"
                  required
                />
              </div>
              <div className="mb-6">
                <label className="block text-gray-700 text-sm font-bold mb-2">Color</label>
                <input
                  type="color"
                  value={formData.color}
                  onChange={(e) => setFormData({ ...formData, color: e.target.value })}
                  className="h-10 w-full rounded cursor-pointer"
                />
              </div>
              <div className="flex justify-end">
                <button
                  type="button"
                  onClick={() => setIsModalOpen(false)}
                  className="bg-gray-500 text-white px-4 py-2 rounded mr-2 hover:bg-gray-600"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700"
                >
                  Save
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
