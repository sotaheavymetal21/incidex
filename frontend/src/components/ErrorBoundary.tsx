'use client';

import React, { Component, ErrorInfo, ReactNode } from 'react';

interface Props {
  children: ReactNode;
  fallback?: ReactNode;
}

interface State {
  hasError: boolean;
  error: Error | null;
}

export class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('ErrorBoundary caught an error:', error, errorInfo);
    // TODO: エラー報告サービスに送信（Sentry, LogRocketなど）
  }

  render() {
    if (this.state.hasError) {
      return this.props.fallback || (
        <div className="min-h-screen flex items-center justify-center p-4" style={{ background: 'var(--background)' }}>
          <div className="max-w-md w-full rounded-xl p-8 border-2" style={{
            background: 'rgba(239, 68, 68, 0.1)',
            borderColor: '#ef4444'
          }}>
            <div className="text-center">
              <div className="text-5xl mb-4">⚠️</div>
              <h2 className="text-2xl font-bold mb-3" style={{ color: '#dc2626' }}>
                エラーが発生しました
              </h2>
              <p className="mb-6" style={{ color: '#991b1b' }}>
                申し訳ございません。予期しないエラーが発生しました。
              </p>

              {this.state.error && process.env.NODE_ENV === 'development' && (
                <div className="mb-6 p-4 rounded-lg text-left overflow-auto" style={{
                  background: '#fee2e2',
                  border: '1px solid #fca5a5',
                  maxHeight: '200px'
                }}>
                  <pre className="text-xs" style={{ color: '#7f1d1d' }}>
                    {this.state.error.toString()}
                  </pre>
                </div>
              )}

              <button
                onClick={() => window.location.reload()}
                className="px-6 py-3 rounded-lg font-medium transition-all"
                style={{
                  background: '#dc2626',
                  color: 'white',
                }}
                onMouseEnter={(e) => {
                  e.currentTarget.style.background = '#b91c1c';
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.background = '#dc2626';
                }}
              >
                ページを再読み込み
              </button>
            </div>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}
