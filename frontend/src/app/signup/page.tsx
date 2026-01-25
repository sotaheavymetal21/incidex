'use client';

import { useAuth } from '@/context/AuthContext';
import { authApi } from '@/lib/api';
import { useRouter } from 'next/navigation';
import { useState } from 'react';
import {
  validateName,
  validateEmail,
  validatePassword,
  validateEmployeeNumber,
  validateDepartment,
  ValidationLimits,
} from '@/utils/validation';

interface FieldErrors {
  name?: string;
  email?: string;
  password?: string;
  employeeNumber?: string;
  department?: string;
}

export default function SignupPage() {
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [employeeNumber, setEmployeeNumber] = useState('');
  const [department, setDepartment] = useState('');
  const [error, setError] = useState('');
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({});
  const router = useRouter();

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
        const result = validatePassword(value, true);
        return result.isValid ? undefined : result.error;
      }
      case 'employeeNumber': {
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

  const handleFieldChange = (field: keyof FieldErrors, value: string, setter: (v: string) => void) => {
    setter(value);
    if (fieldErrors[field]) {
      const error = validateField(field, value);
      setFieldErrors((prev) => ({ ...prev, [field]: error }));
    }
  };

  const handleFieldBlur = (field: keyof FieldErrors, value: string) => {
    const error = validateField(field, value);
    setFieldErrors((prev) => ({ ...prev, [field]: error }));
  };

  const validateForm = (): boolean => {
    const errors: FieldErrors = {
      name: validateField('name', name),
      email: validateField('email', email),
      password: validateField('password', password),
      employeeNumber: validateField('employeeNumber', employeeNumber),
      department: validateField('department', department),
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

    try {
      await authApi.register(name.trim(), email.trim(), password, employeeNumber.trim(), department.trim());
      router.push('/login');
    } catch (err: any) {
      setError(err.message);
    }
  };

  return (
    <div
      className="flex min-h-screen items-center justify-center px-4 relative overflow-hidden"
      style={{
        background: 'linear-gradient(135deg, #f0fdf4 0%, #d1fae5 50%, #a7f3d0 100%)'
      }}
    >
      {/* アニメーショングラデーションの装飾 */}
      <div
        className="absolute top-20 left-20 w-96 h-96 rounded-full opacity-20 blur-3xl"
        style={{
          background: 'radial-gradient(circle, var(--primary) 0%, transparent 70%)',
          animation: 'glow 4s ease-in-out infinite'
        }}
      />
      <div
        className="absolute bottom-20 right-20 w-96 h-96 rounded-full opacity-20 blur-3xl"
        style={{
          background: 'radial-gradient(circle, var(--accent) 0%, transparent 70%)',
          animation: 'glow 4s ease-in-out infinite 2s'
        }}
      />

      <div className="w-full max-w-md relative z-10 animate-slideUp">
        {/* ロゴセクション */}
        <div className="flex flex-col items-center mb-12 animate-fadeIn">
          <div
            className="mb-6 p-6 rounded-3xl shadow-2xl relative overflow-hidden"
            style={{
              background: 'linear-gradient(135deg, var(--surface) 0%, var(--primary-light) 100%)',
              boxShadow: '0 0 40px var(--primary-glow), 0 20px 40px rgba(0,0,0,0.1)'
            }}
          >
            <img
              src="/incidex_full_logo.jpg"
              alt="Incidex"
              className="h-20 w-auto relative z-10"
            />
            <div
              className="absolute inset-0 opacity-30"
              style={{
                background: 'linear-gradient(45deg, transparent 30%, var(--primary-light) 50%, transparent 70%)',
                backgroundSize: '200% 200%',
                animation: 'shimmer 3s ease-in-out infinite'
              }}
            />
          </div>
          <h1
            className="text-3xl font-bold mb-2"
            style={{
              color: 'var(--foreground)',
              fontFamily: 'var(--font-display)',
              textShadow: '0 2px 10px rgba(16, 185, 129, 0.1)'
            }}
          >
            アカウント作成
          </h1>
          <p
            className="text-base font-medium tracking-wide"
            style={{
              color: 'var(--primary-dark)',
              fontFamily: 'var(--font-display)'
            }}
          >
            新規ユーザー登録
          </p>
        </div>

        {/* サインアップカード */}
        <div
          className="rounded-2xl p-8 relative overflow-hidden animate-scaleIn stagger-1"
          style={{
            background: 'var(--surface)',
            boxShadow: '0 0 30px var(--primary-glow), 0 20px 40px rgba(0,0,0,0.08)',
            border: '1px solid var(--border)'
          }}
        >
          {/* 装飾用のコーナーアクセント */}
          <div
            className="absolute top-0 left-0 w-20 h-20 opacity-30"
            style={{
              background: 'linear-gradient(135deg, var(--primary-light) 0%, transparent 100%)',
              borderTopLeftRadius: '1rem'
            }}
          />
          <div
            className="absolute bottom-0 right-0 w-20 h-20 opacity-30"
            style={{
              background: 'linear-gradient(315deg, var(--primary-light) 0%, transparent 100%)',
              borderBottomRightRadius: '1rem'
            }}
          />

          <div className="relative z-10">
            <form className="space-y-6" onSubmit={handleSubmit}>
              <div className="space-y-5">
                <div>
                  <label
                    htmlFor="name"
                    className="block text-sm font-semibold mb-2"
                    style={{
                      color: 'var(--foreground)',
                      fontFamily: 'var(--font-body)'
                    }}
                  >
                    氏名 <span style={{ color: 'var(--error)' }}>*</span>
                  </label>
                  <input
                    id="name"
                    type="text"
                    required
                    maxLength={ValidationLimits.NAME_MAX_LENGTH}
                    className="block w-full rounded-xl px-4 py-3.5 focus:outline-none transition-all"
                    style={{
                      background: 'var(--gray-50)',
                      border: `2px solid ${fieldErrors.name ? 'var(--error)' : 'var(--border)'}`,
                      color: 'var(--foreground)',
                      fontFamily: 'var(--font-body)'
                    }}
                    onFocus={(e) => {
                      e.currentTarget.style.borderColor = 'var(--primary)';
                      e.currentTarget.style.boxShadow = '0 0 0 3px var(--primary-light)';
                    }}
                    onBlur={(e) => {
                      handleFieldBlur('name', e.target.value);
                      e.currentTarget.style.borderColor = fieldErrors.name ? 'var(--error)' : 'var(--border)';
                      e.currentTarget.style.boxShadow = 'none';
                    }}
                    placeholder="山田 太郎"
                    value={name}
                    onChange={(e) => handleFieldChange('name', e.target.value, setName)}
                  />
                  {fieldErrors.name && (
                    <p className="mt-1 text-xs" style={{ color: 'var(--error)' }}>{fieldErrors.name}</p>
                  )}
                  <p className="mt-1 text-xs" style={{ color: 'var(--muted)' }}>{ValidationLimits.NAME_MAX_LENGTH}文字以内</p>
                </div>
                <div>
                  <label
                    htmlFor="email"
                    className="block text-sm font-semibold mb-2"
                    style={{
                      color: 'var(--foreground)',
                      fontFamily: 'var(--font-body)'
                    }}
                  >
                    メールアドレス <span style={{ color: 'var(--error)' }}>*</span>
                  </label>
                  <input
                    id="email"
                    type="email"
                    required
                    maxLength={ValidationLimits.EMAIL_MAX_LENGTH}
                    className="block w-full rounded-xl px-4 py-3.5 focus:outline-none transition-all"
                    style={{
                      background: 'var(--gray-50)',
                      border: `2px solid ${fieldErrors.email ? 'var(--error)' : 'var(--border)'}`,
                      color: 'var(--foreground)',
                      fontFamily: 'var(--font-body)'
                    }}
                    onFocus={(e) => {
                      e.currentTarget.style.borderColor = 'var(--primary)';
                      e.currentTarget.style.boxShadow = '0 0 0 3px var(--primary-light)';
                    }}
                    onBlur={(e) => {
                      handleFieldBlur('email', e.target.value);
                      e.currentTarget.style.borderColor = fieldErrors.email ? 'var(--error)' : 'var(--border)';
                      e.currentTarget.style.boxShadow = 'none';
                    }}
                    placeholder="user@example.com"
                    value={email}
                    onChange={(e) => handleFieldChange('email', e.target.value, setEmail)}
                  />
                  {fieldErrors.email && (
                    <p className="mt-1 text-xs" style={{ color: 'var(--error)' }}>{fieldErrors.email}</p>
                  )}
                </div>
                <div>
                  <label
                    htmlFor="employeeNumber"
                    className="block text-sm font-semibold mb-2"
                    style={{
                      color: 'var(--foreground)',
                      fontFamily: 'var(--font-body)'
                    }}
                  >
                    社員番号 <span style={{ color: 'var(--error)' }}>*</span>
                  </label>
                  <input
                    id="employeeNumber"
                    type="text"
                    required
                    maxLength={ValidationLimits.EMPLOYEE_NUMBER_MAX_LENGTH}
                    className="block w-full rounded-xl px-4 py-3.5 focus:outline-none transition-all"
                    style={{
                      background: 'var(--gray-50)',
                      border: `2px solid ${fieldErrors.employeeNumber ? 'var(--error)' : 'var(--border)'}`,
                      color: 'var(--foreground)',
                      fontFamily: 'var(--font-body)'
                    }}
                    onFocus={(e) => {
                      e.currentTarget.style.borderColor = 'var(--primary)';
                      e.currentTarget.style.boxShadow = '0 0 0 3px var(--primary-light)';
                    }}
                    onBlur={(e) => {
                      handleFieldBlur('employeeNumber', e.target.value);
                      e.currentTarget.style.borderColor = fieldErrors.employeeNumber ? 'var(--error)' : 'var(--border)';
                      e.currentTarget.style.boxShadow = 'none';
                    }}
                    placeholder="EMP-001"
                    value={employeeNumber}
                    onChange={(e) => handleFieldChange('employeeNumber', e.target.value, setEmployeeNumber)}
                  />
                  {fieldErrors.employeeNumber && (
                    <p className="mt-1 text-xs" style={{ color: 'var(--error)' }}>{fieldErrors.employeeNumber}</p>
                  )}
                  <p className="mt-1 text-xs" style={{ color: 'var(--muted)' }}>英数字とハイフンのみ、{ValidationLimits.EMPLOYEE_NUMBER_MAX_LENGTH}文字以内</p>
                </div>
                <div>
                  <label
                    htmlFor="department"
                    className="block text-sm font-semibold mb-2"
                    style={{
                      color: 'var(--foreground)',
                      fontFamily: 'var(--font-body)'
                    }}
                  >
                    所属部署 <span style={{ color: 'var(--error)' }}>*</span>
                  </label>
                  <input
                    id="department"
                    type="text"
                    required
                    maxLength={ValidationLimits.DEPARTMENT_MAX_LENGTH}
                    className="block w-full rounded-xl px-4 py-3.5 focus:outline-none transition-all"
                    style={{
                      background: 'var(--gray-50)',
                      border: `2px solid ${fieldErrors.department ? 'var(--error)' : 'var(--border)'}`,
                      color: 'var(--foreground)',
                      fontFamily: 'var(--font-body)'
                    }}
                    onFocus={(e) => {
                      e.currentTarget.style.borderColor = 'var(--primary)';
                      e.currentTarget.style.boxShadow = '0 0 0 3px var(--primary-light)';
                    }}
                    onBlur={(e) => {
                      handleFieldBlur('department', e.target.value);
                      e.currentTarget.style.borderColor = fieldErrors.department ? 'var(--error)' : 'var(--border)';
                      e.currentTarget.style.boxShadow = 'none';
                    }}
                    placeholder="開発部"
                    value={department}
                    onChange={(e) => handleFieldChange('department', e.target.value, setDepartment)}
                  />
                  {fieldErrors.department && (
                    <p className="mt-1 text-xs" style={{ color: 'var(--error)' }}>{fieldErrors.department}</p>
                  )}
                  <p className="mt-1 text-xs" style={{ color: 'var(--muted)' }}>{ValidationLimits.DEPARTMENT_MAX_LENGTH}文字以内</p>
                </div>
                <div>
                  <label
                    htmlFor="password"
                    className="block text-sm font-semibold mb-2"
                    style={{
                      color: 'var(--foreground)',
                      fontFamily: 'var(--font-body)'
                    }}
                  >
                    パスワード <span style={{ color: 'var(--error)' }}>*</span>
                  </label>
                  <input
                    id="password"
                    type="password"
                    required
                    minLength={ValidationLimits.PASSWORD_MIN_LENGTH}
                    className="block w-full rounded-xl px-4 py-3.5 focus:outline-none transition-all"
                    style={{
                      background: 'var(--gray-50)',
                      border: `2px solid ${fieldErrors.password ? 'var(--error)' : 'var(--border)'}`,
                      color: 'var(--foreground)',
                      fontFamily: 'var(--font-body)'
                    }}
                    onFocus={(e) => {
                      e.currentTarget.style.borderColor = 'var(--primary)';
                      e.currentTarget.style.boxShadow = '0 0 0 3px var(--primary-light)';
                    }}
                    onBlur={(e) => {
                      handleFieldBlur('password', e.target.value);
                      e.currentTarget.style.borderColor = fieldErrors.password ? 'var(--error)' : 'var(--border)';
                      e.currentTarget.style.boxShadow = 'none';
                    }}
                    placeholder="••••••••"
                    value={password}
                    onChange={(e) => handleFieldChange('password', e.target.value, setPassword)}
                  />
                  {fieldErrors.password && (
                    <p className="mt-1 text-xs" style={{ color: 'var(--error)' }}>{fieldErrors.password}</p>
                  )}
                  <div
                    className="mt-2 text-xs space-y-1"
                    style={{
                      color: 'var(--muted)',
                      fontFamily: 'var(--font-body)'
                    }}
                  >
                    <p className="font-medium" style={{ color: 'var(--foreground)' }}>パスワード要件:</p>
                    <ul className="list-disc list-inside ml-1 space-y-0.5">
                      <li>8文字以上</li>
                      <li>大文字（A-Z）を含む</li>
                      <li>小文字（a-z）を含む</li>
                      <li>数字（0-9）を含む</li>
                    </ul>
                  </div>
                </div>
              </div>

              {error && (
                <div
                  className="rounded-xl p-4 animate-slideDown"
                  style={{
                    background: 'var(--error-light)',
                    border: '2px solid var(--error)',
                    borderLeft: '4px solid var(--error)'
                  }}
                >
                  <p
                    className="text-sm font-semibold"
                    style={{
                      color: 'var(--error)',
                      fontFamily: 'var(--font-body)'
                    }}
                  >
                    {error}
                  </p>
                </div>
              )}

              <button
                type="submit"
                className="w-full flex justify-center rounded-xl px-4 py-4 text-base font-bold text-white focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 transition-all"
                style={{
                  background: 'linear-gradient(135deg, var(--primary) 0%, var(--primary-dark) 100%)',
                  outlineColor: 'var(--primary)',
                  fontFamily: 'var(--font-display)',
                  boxShadow: '0 4px 12px var(--primary-glow)'
                }}
                onMouseEnter={(e) => {
                  e.currentTarget.style.transform = 'translateY(-2px)';
                  e.currentTarget.style.boxShadow = '0 8px 24px var(--primary-glow)';
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.transform = 'translateY(0)';
                  e.currentTarget.style.boxShadow = '0 4px 12px var(--primary-glow)';
                }}
              >
                アカウント作成
              </button>
            </form>

            <div className="mt-8 text-center">
              <a
                href="/login"
                className="text-sm font-semibold hover:underline transition-all inline-block"
                style={{
                  color: 'var(--primary)',
                  fontFamily: 'var(--font-body)'
                }}
                onMouseEnter={(e) => {
                  e.currentTarget.style.color = 'var(--primary-dark)';
                  e.currentTarget.style.transform = 'scale(1.05)';
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.color = 'var(--primary)';
                  e.currentTarget.style.transform = 'scale(1)';
                }}
              >
                すでにアカウントをお持ちの方はこちら →
              </a>
            </div>
          </div>
        </div>

        {/* フッター */}
        <p
          className="mt-8 text-center text-sm font-medium animate-fadeIn stagger-2"
          style={{
            color: 'var(--primary-dark)',
            fontFamily: 'var(--font-body)'
          }}
        >
          © 2026 Incidex. All rights reserved.
        </p>
      </div>

      <style jsx>{`
        @keyframes shimmer {
          0% { background-position: 0% 50%; }
          50% { background-position: 100% 50%; }
          100% { background-position: 0% 50%; }
        }
      `}</style>
    </div>
  );
}
