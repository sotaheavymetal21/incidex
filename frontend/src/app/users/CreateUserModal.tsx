'use client';

import { useState } from 'react';
import { useAuth } from '@/context/AuthContext';
import { userApi } from '@/lib/api';
import { Role, CreateUserRequest } from '@/types/user';
import { generateSecurePassword, copyToClipboard } from '@/utils/password';
import {
  validateName,
  validateEmail,
  validatePassword,
  validateEmployeeNumber,
  validateDepartment,
  ValidationLimits,
} from '@/utils/validation';

interface CreateUserModalProps {
  onClose: () => void;
  onSuccess: () => void;
}

interface FieldErrors {
  name?: string;
  email?: string;
  password?: string;
  employee_number?: string;
  department?: string;
}

export default function CreateUserModal({ onClose, onSuccess }: CreateUserModalProps) {
  const { token } = useAuth();
  const [formData, setFormData] = useState<CreateUserRequest>({
    email: '',
    password: '',
    name: '',
    role: 'viewer',
  });
  const [showPassword, setShowPassword] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState('');
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({});

  const handleGeneratePassword = () => {
    const newPassword = generateSecurePassword(16);
    setFormData({ ...formData, password: newPassword });
    setShowPassword(true);
  };

  const handleCopyPassword = async () => {
    try {
      await copyToClipboard(formData.password);
      alert('パスワードをクリップボードにコピーしました');
    } catch (err) {
      alert('パスワードのコピーに失敗しました');
    }
  };

  const validateField = (field: keyof FieldErrors, value: string): string | undefined => {
    switch (field) {
      case 'name': {
        const result = validateName(value);
        return result.isValid ? undefined : result.error;
      }
      case 'email': {
        const result = validateEmail(value);
        return result.isValid ? undefined : result.error;
      }
      case 'password': {
        // 管理者作成なので、緩いパスワード要件
        const result = validatePassword(value, false);
        return result.isValid ? undefined : result.error;
      }
      case 'employee_number': {
        const result = validateEmployeeNumber(value);
        return result.isValid ? undefined : result.error;
      }
      case 'department': {
        const result = validateDepartment(value);
        return result.isValid ? undefined : result.error;
      }
    }
    return undefined;
  };

  const handleFieldChange = (field: keyof CreateUserRequest, value: string) => {
    setFormData({ ...formData, [field]: value });
    if (field in fieldErrors) {
      const error = validateField(field as keyof FieldErrors, value);
      setFieldErrors((prev) => ({ ...prev, [field]: error }));
    }
  };

  const handleFieldBlur = (field: keyof FieldErrors, value: string) => {
    const error = validateField(field, value);
    setFieldErrors((prev) => ({ ...prev, [field]: error }));
  };

  const validateForm = (): boolean => {
    const errors: FieldErrors = {
      name: validateField('name', formData.name),
      email: validateField('email', formData.email),
      password: validateField('password', formData.password),
      employee_number: validateField('employee_number', formData.employee_number || ''),
      department: validateField('department', formData.department || ''),
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

    setSubmitting(true);
    try {
      await userApi.create(token!, {
        ...formData,
        name: formData.name.trim(),
        email: formData.email.trim(),
        employee_number: formData.employee_number?.trim(),
        department: formData.department?.trim(),
      });
      onSuccess();
    } catch (err: any) {
      setError(err.message || 'ユーザーの作成に失敗しました');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full flex items-center justify-center z-50">
      <div className="bg-white p-8 rounded-lg shadow-xl w-96">
        <h2 className="text-2xl font-bold mb-4 text-gray-900">ユーザーを作成</h2>

        {error && (
          <div className="mb-4 p-3 bg-red-100 border border-red-400 text-red-700 rounded">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit}>
          <div className="mb-4">
            <label htmlFor="name" className="block text-sm font-medium text-gray-700 mb-1">
              名前 <span className="text-red-500">*</span>
            </label>
            <input
              type="text"
              id="name"
              value={formData.name}
              onChange={(e) => handleFieldChange('name', e.target.value)}
              onBlur={(e) => handleFieldBlur('name', e.target.value)}
              maxLength={ValidationLimits.NAME_MAX_LENGTH}
              className={`w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 text-gray-900 ${
                fieldErrors.name ? 'border-red-500' : 'border-gray-300'
              }`}
              required
              disabled={submitting}
            />
            {fieldErrors.name && (
              <p className="mt-1 text-xs text-red-500">{fieldErrors.name}</p>
            )}
            <p className="mt-1 text-xs text-gray-500">{ValidationLimits.NAME_MAX_LENGTH}文字以内</p>
          </div>

          <div className="mb-4">
            <label htmlFor="email" className="block text-sm font-medium text-gray-700 mb-1">
              メールアドレス <span className="text-red-500">*</span>
            </label>
            <input
              type="email"
              id="email"
              value={formData.email}
              onChange={(e) => handleFieldChange('email', e.target.value)}
              onBlur={(e) => handleFieldBlur('email', e.target.value)}
              maxLength={ValidationLimits.EMAIL_MAX_LENGTH}
              className={`w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 text-gray-900 ${
                fieldErrors.email ? 'border-red-500' : 'border-gray-300'
              }`}
              required
              disabled={submitting}
            />
            {fieldErrors.email && (
              <p className="mt-1 text-xs text-red-500">{fieldErrors.email}</p>
            )}
          </div>

          <div className="mb-4">
            <label htmlFor="employee_number" className="block text-sm font-medium text-gray-700 mb-1">
              社員番号
            </label>
            <input
              type="text"
              id="employee_number"
              value={formData.employee_number || ''}
              onChange={(e) => handleFieldChange('employee_number', e.target.value)}
              onBlur={(e) => handleFieldBlur('employee_number', e.target.value)}
              maxLength={ValidationLimits.EMPLOYEE_NUMBER_MAX_LENGTH}
              className={`w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 text-gray-900 ${
                fieldErrors.employee_number ? 'border-red-500' : 'border-gray-300'
              }`}
              placeholder="例: EMP-001"
              disabled={submitting}
            />
            {fieldErrors.employee_number && (
              <p className="mt-1 text-xs text-red-500">{fieldErrors.employee_number}</p>
            )}
            <p className="mt-1 text-xs text-gray-500">英数字とハイフンのみ、{ValidationLimits.EMPLOYEE_NUMBER_MAX_LENGTH}文字以内</p>
          </div>

          <div className="mb-4">
            <label htmlFor="department" className="block text-sm font-medium text-gray-700 mb-1">
              所属部署
            </label>
            <input
              type="text"
              id="department"
              value={formData.department || ''}
              onChange={(e) => handleFieldChange('department', e.target.value)}
              onBlur={(e) => handleFieldBlur('department', e.target.value)}
              maxLength={ValidationLimits.DEPARTMENT_MAX_LENGTH}
              className={`w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 text-gray-900 ${
                fieldErrors.department ? 'border-red-500' : 'border-gray-300'
              }`}
              placeholder="例: 開発部"
              disabled={submitting}
            />
            {fieldErrors.department && (
              <p className="mt-1 text-xs text-red-500">{fieldErrors.department}</p>
            )}
            <p className="mt-1 text-xs text-gray-500">{ValidationLimits.DEPARTMENT_MAX_LENGTH}文字以内</p>
          </div>

          <div className="mb-4">
            <label htmlFor="password" className="block text-sm font-medium text-gray-700 mb-1">
              パスワード <span className="text-red-500">*</span>
            </label>
            <div className="relative">
              <input
                type={showPassword ? 'text' : 'password'}
                id="password"
                value={formData.password}
                onChange={(e) => handleFieldChange('password', e.target.value)}
                onBlur={(e) => handleFieldBlur('password', e.target.value)}
                className={`w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 text-gray-900 ${
                  fieldErrors.password ? 'border-red-500' : 'border-gray-300'
                }`}
                required
                minLength={ValidationLimits.PASSWORD_MIN_LENGTH_ADMIN}
                disabled={submitting}
              />
              <button
                type="button"
                onClick={() => setShowPassword(!showPassword)}
                className="absolute right-3 top-2 text-sm text-gray-600 hover:text-gray-800"
              >
                {showPassword ? '隠す' : '表示'}
              </button>
            </div>
            {fieldErrors.password && (
              <p className="mt-1 text-xs text-red-500">{fieldErrors.password}</p>
            )}
            <div className="mt-2 flex gap-2">
              <button
                type="button"
                onClick={handleGeneratePassword}
                className="text-sm px-3 py-1 bg-gray-600 text-white rounded hover:bg-gray-700 disabled:opacity-50"
                disabled={submitting}
              >
                ランダム生成
              </button>
              {formData.password && (
                <button
                  type="button"
                  onClick={handleCopyPassword}
                  className="text-sm px-3 py-1 bg-gray-600 text-white rounded hover:bg-gray-700 disabled:opacity-50"
                  disabled={submitting}
                >
                  コピー
                </button>
              )}
            </div>
            <p className="mt-1 text-xs text-gray-500">最低{ValidationLimits.PASSWORD_MIN_LENGTH_ADMIN}文字必要です</p>
          </div>

          <div className="mb-6">
            <label htmlFor="role" className="block text-sm font-medium text-gray-700 mb-1">
              権限
            </label>
            <select
              id="role"
              value={formData.role}
              onChange={(e) => setFormData({ ...formData, role: e.target.value as Role })}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 text-gray-900"
              required
              disabled={submitting}
            >
              <option value="admin">管理者</option>
              <option value="editor">編集者</option>
              <option value="viewer">閲覧者</option>
            </select>
          </div>

          <div className="flex justify-end gap-2">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 text-gray-700 bg-gray-200 rounded-md hover:bg-gray-300 disabled:opacity-50"
              disabled={submitting}
            >
              キャンセル
            </button>
            <button
              type="submit"
              className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
              disabled={submitting}
            >
              {submitting ? '作成中...' : '作成'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
