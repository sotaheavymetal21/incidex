'use client';

import { useState } from 'react';

export default function SeverityGuide() {
  const [isExpanded, setIsExpanded] = useState(false);

  const severityLevels = [
    {
      level: 'critical',
      label: 'Critical（致命的）',
      color: { background: '#fee2e2', color: '#dc2626', borderColor: '#dc2626' },
      bgColor: { background: '#fef2f2' },
      criteria: [
        'サービス全体が停止している、または停止する可能性が高い',
        'データ損失や重大なセキュリティ侵害が発生している',
        '全ユーザーまたは大多数のユーザーに影響がある',
        '決済や課金など、ビジネスクリティカルな機能が完全に停止',
      ],
      examples: [
        'データベース接続の完全な喪失',
        '決済システムの障害',
        'セキュリティ侵害によるデータ漏洩',
        'サーバーのメモリリークによる全サービス停止',
      ],
      responseTime: '即時対応（15分以内）',
    },
    {
      level: 'high',
      label: 'High（高）',
      color: { background: '#ffedd5', color: '#ea580c', borderColor: '#ea580c' },
      bgColor: { background: '#fff7ed' },
      criteria: [
        '主要機能が正常に動作していない',
        '多くのユーザーに影響がある（30%以上）',
        'パフォーマンスが著しく低下している',
        'セキュリティリスクが存在する',
      ],
      examples: [
        'API応答時間の大幅な遅延',
        'モバイルアプリの頻繁なクラッシュ',
        'バックアップジョブの失敗',
        '不正アクセスの試行検知',
      ],
      responseTime: '1時間以内',
    },
    {
      level: 'medium',
      label: 'Medium（中）',
      color: { background: '#fef3c7', color: '#f59e0b', borderColor: '#f59e0b' },
      bgColor: { background: '#fffbeb' },
      criteria: [
        '一部の機能に問題がある',
        '限定的なユーザーに影響がある（5-30%程度）',
        '代替手段が存在する',
        'パフォーマンスに軽微な影響がある',
      ],
      examples: [
        '画像アップロード機能の間欠的なエラー',
        'メール通知の遅延',
        'APIレート制限の誤動作',
        'CSVエクスポートの文字化け',
      ],
      responseTime: '4時間以内',
    },
    {
      level: 'low',
      label: 'Low（低）',
      color: { background: '#dbeafe', color: '#3b82f6', borderColor: '#3b82f6' },
      bgColor: { background: '#eff6ff' },
      criteria: [
        'マイナーな問題や改善要望',
        '影響範囲が非常に限定的',
        'ユーザー体験に軽微な影響',
        '通常業務には支障がない',
      ],
      examples: [
        'UI表示の軽微な崩れ',
        '開発環境のSSL証明書更新',
        'タイムゾーン表示のズレ',
        '監視アラートの誤検知',
      ],
      responseTime: '1営業日以内',
    },
  ];

  return (
    <div className="rounded-xl shadow-lg mb-6 border" style={{ background: 'var(--surface)', borderColor: 'var(--border)' }}>
      <button
        onClick={() => setIsExpanded(!isExpanded)}
        className="w-full px-6 py-4 flex justify-between items-center transition-all rounded-t-xl"
        style={{ color: 'var(--foreground)' }}
        onMouseEnter={(e) => e.currentTarget.style.background = 'var(--secondary-light)'}
        onMouseLeave={(e) => e.currentTarget.style.background = 'transparent'}
      >
        <div className="flex items-center gap-2">
          <svg
            className="w-5 h-5"
            style={{ color: 'var(--primary)' }}
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
            />
          </svg>
          <h2 className="text-lg font-semibold">
            Severity（深刻度）の設定基準
          </h2>
        </div>
        <svg
          className={`w-5 h-5 transition-transform ${
            isExpanded ? 'rotate-180' : ''
          }`}
          style={{ color: 'var(--secondary)' }}
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M19 9l-7 7-7-7"
          />
        </svg>
      </button>

      {isExpanded && (
        <div className="px-6 pb-6 border-t" style={{ borderColor: 'var(--border)' }}>
          <p className="text-sm mb-6 mt-4" style={{ color: 'var(--secondary)' }}>
            インシデントの深刻度は、以下の基準に基づいて設定してください。適切な深刻度を設定することで、優先順位付けと迅速な対応が可能になります。
          </p>

          <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
            {severityLevels.map((severity) => (
              <div
                key={severity.level}
                className="border-2 rounded-xl p-4"
                style={{
                  ...severity.bgColor,
                  borderColor: severity.color.borderColor,
                }}
              >
                <div className="flex items-center justify-between mb-3">
                  <span
                    className="px-3 py-1 inline-flex text-sm font-semibold rounded-full border-2"
                    style={{
                      ...severity.color,
                      fontWeight: '600',
                    }}
                  >
                    {severity.label}
                  </span>
                  <span className="text-xs font-medium" style={{ color: 'var(--secondary)' }}>
                    対応目安: {severity.responseTime}
                  </span>
                </div>

                <div className="space-y-3">
                  <div>
                    <h4 className="text-sm font-semibold mb-2" style={{ color: 'var(--foreground)' }}>
                      設定基準:
                    </h4>
                    <ul className="space-y-1">
                      {severity.criteria.map((criterion, index) => (
                        <li
                          key={index}
                          className="text-xs flex items-start"
                          style={{ color: 'var(--foreground)' }}
                        >
                          <span className="mr-2 mt-0.5">•</span>
                          <span>{criterion}</span>
                        </li>
                      ))}
                    </ul>
                  </div>

                  <div>
                    <h4 className="text-sm font-semibold mb-2" style={{ color: 'var(--foreground)' }}>
                      具体例:
                    </h4>
                    <ul className="space-y-1">
                      {severity.examples.map((example, index) => (
                        <li
                          key={index}
                          className="text-xs flex items-start"
                          style={{ color: 'var(--secondary)' }}
                        >
                          <span className="mr-2 mt-0.5">-</span>
                          <span>{example}</span>
                        </li>
                      ))}
                    </ul>
                  </div>
                </div>
              </div>
            ))}
          </div>

          <div className="mt-6 p-4 rounded-xl border-2" style={{
            background: 'var(--info-light)',
            borderColor: 'var(--info)',
          }}>
            <h4 className="text-sm font-semibold mb-2" style={{ color: 'var(--info)' }}>
              💡 Tips
            </h4>
            <ul className="space-y-1 text-xs" style={{ color: 'var(--info)' }}>
              <li>
                • 深刻度は状況に応じて変更できます。調査が進むにつれて適切なレベルに調整してください。
              </li>
              <li>
                • 迷った場合は、影響範囲（ユーザー数）とビジネスへの影響度を最優先で考慮してください。
              </li>
              <li>
                • セキュリティに関連するインシデントは、1段階上の深刻度を設定することを推奨します。
              </li>
            </ul>
          </div>
        </div>
      )}
    </div>
  );
}
