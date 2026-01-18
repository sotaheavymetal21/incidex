'use client';

import { useAuth } from '@/context/AuthContext';
import { authApi } from '@/lib/api';
import { useRouter } from 'next/navigation';
import { useState } from 'react';

export default function SignupPage() {
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [employeeNumber, setEmployeeNumber] = useState('');
  const [department, setDepartment] = useState('');
  const [error, setError] = useState('');
  const router = useRouter();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    // Validate password strength
    if (password.length < 8) {
      setError('パスワードは8文字以上である必要があります');
      return;
    }
    if (!/[A-Z]/.test(password)) {
      setError('パスワードには大文字を含める必要があります');
      return;
    }
    if (!/[a-z]/.test(password)) {
      setError('パスワードには小文字を含める必要があります');
      return;
    }
    if (!/[0-9]/.test(password)) {
      setError('パスワードには数字を含める必要があります');
      return;
    }

    try {
      await authApi.register(name, email, password, employeeNumber, department);
      // Automatically redirect to login or login directly
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
      {/* Animated gradient orbs */}
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
        {/* Logo Section */}
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

        {/* Signup Card */}
        <div
          className="rounded-2xl p-8 relative overflow-hidden animate-scaleIn stagger-1"
          style={{
            background: 'var(--surface)',
            boxShadow: '0 0 30px var(--primary-glow), 0 20px 40px rgba(0,0,0,0.08)',
            border: '1px solid var(--border)'
          }}
        >
          {/* Decorative corner accents */}
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
                    氏名
                  </label>
                  <input
                    id="name"
                    type="text"
                    required
                    className="block w-full rounded-xl px-4 py-3.5 focus:outline-none transition-all"
                    style={{
                      background: 'var(--gray-50)',
                      border: '2px solid var(--border)',
                      color: 'var(--foreground)',
                      fontFamily: 'var(--font-body)'
                    }}
                    onFocus={(e) => {
                      e.currentTarget.style.borderColor = 'var(--primary)';
                      e.currentTarget.style.boxShadow = '0 0 0 3px var(--primary-light)';
                    }}
                    onBlur={(e) => {
                      e.currentTarget.style.borderColor = 'var(--border)';
                      e.currentTarget.style.boxShadow = 'none';
                    }}
                    placeholder="山田 太郎"
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                  />
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
                    メールアドレス
                  </label>
                  <input
                    id="email"
                    type="email"
                    required
                    className="block w-full rounded-xl px-4 py-3.5 focus:outline-none transition-all"
                    style={{
                      background: 'var(--gray-50)',
                      border: '2px solid var(--border)',
                      color: 'var(--foreground)',
                      fontFamily: 'var(--font-body)'
                    }}
                    onFocus={(e) => {
                      e.currentTarget.style.borderColor = 'var(--primary)';
                      e.currentTarget.style.boxShadow = '0 0 0 3px var(--primary-light)';
                    }}
                    onBlur={(e) => {
                      e.currentTarget.style.borderColor = 'var(--border)';
                      e.currentTarget.style.boxShadow = 'none';
                    }}
                    placeholder="user@example.com"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                  />
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
                    社員番号
                  </label>
                  <input
                    id="employeeNumber"
                    type="text"
                    required
                    className="block w-full rounded-xl px-4 py-3.5 focus:outline-none transition-all"
                    style={{
                      background: 'var(--gray-50)',
                      border: '2px solid var(--border)',
                      color: 'var(--foreground)',
                      fontFamily: 'var(--font-body)'
                    }}
                    onFocus={(e) => {
                      e.currentTarget.style.borderColor = 'var(--primary)';
                      e.currentTarget.style.boxShadow = '0 0 0 3px var(--primary-light)';
                    }}
                    onBlur={(e) => {
                      e.currentTarget.style.borderColor = 'var(--border)';
                      e.currentTarget.style.boxShadow = 'none';
                    }}
                    placeholder="EMP-001"
                    value={employeeNumber}
                    onChange={(e) => setEmployeeNumber(e.target.value)}
                  />
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
                    所属部署
                  </label>
                  <input
                    id="department"
                    type="text"
                    required
                    className="block w-full rounded-xl px-4 py-3.5 focus:outline-none transition-all"
                    style={{
                      background: 'var(--gray-50)',
                      border: '2px solid var(--border)',
                      color: 'var(--foreground)',
                      fontFamily: 'var(--font-body)'
                    }}
                    onFocus={(e) => {
                      e.currentTarget.style.borderColor = 'var(--primary)';
                      e.currentTarget.style.boxShadow = '0 0 0 3px var(--primary-light)';
                    }}
                    onBlur={(e) => {
                      e.currentTarget.style.borderColor = 'var(--border)';
                      e.currentTarget.style.boxShadow = 'none';
                    }}
                    placeholder="開発部"
                    value={department}
                    onChange={(e) => setDepartment(e.target.value)}
                  />
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
                    パスワード
                  </label>
                  <input
                    id="password"
                    type="password"
                    required
                    minLength={8}
                    className="block w-full rounded-xl px-4 py-3.5 focus:outline-none transition-all"
                    style={{
                      background: 'var(--gray-50)',
                      border: '2px solid var(--border)',
                      color: 'var(--foreground)',
                      fontFamily: 'var(--font-body)'
                    }}
                    onFocus={(e) => {
                      e.currentTarget.style.borderColor = 'var(--primary)';
                      e.currentTarget.style.boxShadow = '0 0 0 3px var(--primary-light)';
                    }}
                    onBlur={(e) => {
                      e.currentTarget.style.borderColor = 'var(--border)';
                      e.currentTarget.style.boxShadow = 'none';
                    }}
                    placeholder="••••••••"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                  />
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

        {/* Footer */}
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
