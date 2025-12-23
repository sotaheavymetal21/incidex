'use client';

import { useState } from 'react';
import { useAuth } from '@/context/AuthContext';
import { userApi } from '@/lib/api';
import { User, UpdatePasswordRequest } from '@/types/user';

interface ChangePasswordModalProps {
  user: User;
  onClose: () => void;
  onSuccess: () => void;
}

export default function ChangePasswordModal({ user, onClose, onSuccess }: ChangePasswordModalProps) {
  const { token } = useAuth();
  const [formData, setFormData] = useState<UpdatePasswordRequest>({
    old_password: '',
    new_password: '',
  });
  const [confirmPassword, setConfirmPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!token) return;

    // Validate passwords match
    if (formData.new_password !== confirmPassword) {
      setError('新しいパスワードが一致しません');
      return;
    }

    // Validate password length
    if (formData.new_password.length < 6) {
      setError('パスワードは6文字以上である必要があります');
      return;
    }

    try {
      setLoading(true);
      setError('');
      await userApi.updatePassword(token, user.id, formData);
      alert('パスワードが正常に変更されました');
      onSuccess();
    } catch (err: any) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full flex items-center justify-center z-50">
      <div className="bg-white p-8 rounded-lg shadow-xl w-96">
        <h2 className="text-xl font-bold mb-4">パスワード変更</h2>
        <p className="text-sm text-gray-600 mb-4">
          ユーザー: <span className="font-semibold">{user.name}</span>
        </p>

        {error && (
          <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-4">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit}>
          <div className="mb-4">
            <label className="block text-gray-700 text-sm font-bold mb-2">
              現在のパスワード
            </label>
            <input
              type="password"
              value={formData.old_password}
              onChange={(e) => setFormData({ ...formData, old_password: e.target.value })}
              className="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline"
              required
              autoComplete="current-password"
            />
          </div>

          <div className="mb-4">
            <label className="block text-gray-700 text-sm font-bold mb-2">
              新しいパスワード
            </label>
            <input
              type="password"
              value={formData.new_password}
              onChange={(e) => setFormData({ ...formData, new_password: e.target.value })}
              className="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline"
              required
              minLength={6}
              autoComplete="new-password"
            />
            <p className="text-xs text-gray-500 mt-1">最低6文字</p>
          </div>

          <div className="mb-6">
            <label className="block text-gray-700 text-sm font-bold mb-2">
              新しいパスワード（確認）
            </label>
            <input
              type="password"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              className="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline"
              required
              minLength={6}
              autoComplete="new-password"
            />
          </div>

          <div className="flex justify-end">
            <button
              type="button"
              onClick={onClose}
              className="bg-gray-500 text-white px-4 py-2 rounded mr-2 hover:bg-gray-600"
              disabled={loading}
            >
              キャンセル
            </button>
            <button
              type="submit"
              className="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700 disabled:bg-blue-300"
              disabled={loading}
            >
              {loading ? '変更中...' : 'パスワード変更'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
