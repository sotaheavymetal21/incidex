'use client';

import { useState, useEffect } from 'react';
import { useAuth } from '@/context/AuthContext';
import { notificationApi } from '@/lib/api';
import { NotificationSetting } from '@/types/notification';
import { useRouter } from 'next/navigation';

export default function NotificationSettingsPage() {
  const { token } = useAuth();
  const router = useRouter();
  const [settings, setSettings] = useState<NotificationSetting | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  useEffect(() => {
    if (!token) {
      router.push('/login');
      return;
    }

    fetchSettings();
  }, [token]);

  const fetchSettings = async () => {
    if (!token) return;

    try {
      setLoading(true);
      const data = await notificationApi.getMySettings(token);
      setSettings(data);
    } catch (err: any) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleSave = async () => {
    if (!token || !settings) return;

    try {
      setSaving(true);
      setError(null);
      setSuccess(null);

      await notificationApi.updateMySettings(token, settings);
      setSuccess('通知設定を保存しました');

      // 3秒後に成功メッセージをクリア
      setTimeout(() => setSuccess(null), 3000);
    } catch (err: any) {
      setError(err.message);
    } finally {
      setSaving(false);
    }
  };

  const handleToggle = (field: keyof NotificationSetting) => {
    if (!settings) return;

    setSettings({
      ...settings,
      [field]: !settings[field],
    });
  };

  const handleSlackWebhookChange = (value: string) => {
    if (!settings) return;

    setSettings({
      ...settings,
      slack_webhook: value,
    });
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-gray-600">読み込み中...</div>
      </div>
    );
  }

  if (!settings) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-red-600">通知設定の読み込みに失敗しました</div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50 py-8">
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
        {/* ヘッダー */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900">通知設定</h1>
          <p className="mt-2 text-gray-600">
            インシデントに関する通知の受け取り方法を設定します
          </p>
        </div>

        {/* エラーメッセージ */}
        {error && (
          <div className="mb-6 bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded">
            {error}
          </div>
        )}

        {/* 成功メッセージ */}
        {success && (
          <div className="mb-6 bg-green-50 border border-green-200 text-green-700 px-4 py-3 rounded">
            {success}
          </div>
        )}

        <div className="bg-white shadow rounded-lg">
          {/* 通知チャネル */}
          <div className="px-6 py-5 border-b border-gray-200">
            <h2 className="text-lg font-semibold text-gray-900 mb-4">通知チャネル</h2>

            <div className="space-y-4">
              {/* Email通知 */}
              <div className="flex items-center justify-between">
                <div>
                  <h3 className="text-sm font-medium text-gray-900">Email通知</h3>
                  <p className="text-sm text-gray-500">登録したメールアドレスに通知を送信</p>
                </div>
                <button
                  type="button"
                  onClick={() => handleToggle('email_enabled')}
                  className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${
                    settings.email_enabled ? 'bg-blue-600' : 'bg-gray-200'
                  }`}
                >
                  <span
                    className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${
                      settings.email_enabled ? 'translate-x-6' : 'translate-x-1'
                    }`}
                  />
                </button>
              </div>

              {/* Slack通知 */}
              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <div>
                    <h3 className="text-sm font-medium text-gray-900">Slack通知</h3>
                    <p className="text-sm text-gray-500">Slackチャンネルに通知を送信</p>
                  </div>
                  <button
                    type="button"
                    onClick={() => handleToggle('slack_enabled')}
                    className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${
                      settings.slack_enabled ? 'bg-blue-600' : 'bg-gray-200'
                    }`}
                  >
                    <span
                      className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${
                        settings.slack_enabled ? 'translate-x-6' : 'translate-x-1'
                      }`}
                    />
                  </button>
                </div>

                {settings.slack_enabled && (
                  <div className="ml-0 mt-2">
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      Slack Webhook URL
                    </label>
                    <input
                      type="url"
                      value={settings.slack_webhook || ''}
                      onChange={(e) => handleSlackWebhookChange(e.target.value)}
                      placeholder="https://hooks.slack.com/services/..."
                      className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                    />
                    <p className="mt-1 text-xs text-gray-500">
                      Slackの Incoming Webhook URL を入力してください
                    </p>
                  </div>
                )}
              </div>
            </div>
          </div>

          {/* 通知イベント */}
          <div className="px-6 py-5">
            <h2 className="text-lg font-semibold text-gray-900 mb-4">通知イベント</h2>

            <div className="space-y-3">
              <NotificationToggle
                label="インシデント作成"
                description="新しいインシデントが作成された時"
                enabled={settings.notify_on_incident_created}
                onChange={() => handleToggle('notify_on_incident_created')}
              />

              <NotificationToggle
                label="担当者割り当て"
                description="自分がインシデントの担当者に割り当てられた時"
                enabled={settings.notify_on_assigned}
                onChange={() => handleToggle('notify_on_assigned')}
              />

              <NotificationToggle
                label="コメント追加"
                description="関係するインシデントに新しいコメントが追加された時"
                enabled={settings.notify_on_comment}
                onChange={() => handleToggle('notify_on_comment')}
              />

              <NotificationToggle
                label="ステータス変更"
                description="関係するインシデントのステータスが変更された時"
                enabled={settings.notify_on_status_change}
                onChange={() => handleToggle('notify_on_status_change')}
              />

              <NotificationToggle
                label="重要度変更"
                description="関係するインシデントの重要度が変更された時"
                enabled={settings.notify_on_severity_change}
                onChange={() => handleToggle('notify_on_severity_change')}
              />

              <NotificationToggle
                label="インシデント解決"
                description="関係するインシデントが解決された時"
                enabled={settings.notify_on_resolved}
                onChange={() => handleToggle('notify_on_resolved')}
              />

              <NotificationToggle
                label="エスカレーション"
                description="インシデントがエスカレートされた時"
                enabled={settings.notify_on_escalation}
                onChange={() => handleToggle('notify_on_escalation')}
              />
            </div>
          </div>

          {/* 保存ボタン */}
          <div className="px-6 py-4 bg-gray-50 border-t border-gray-200 flex justify-end space-x-3">
            <button
              type="button"
              onClick={() => router.push('/incidents')}
              className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50"
            >
              キャンセル
            </button>
            <button
              type="button"
              onClick={handleSave}
              disabled={saving}
              className="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-md hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {saving ? '保存中...' : '設定を保存'}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}

function NotificationToggle({
  label,
  description,
  enabled,
  onChange,
}: {
  label: string;
  description: string;
  enabled: boolean;
  onChange: () => void;
}) {
  return (
    <div className="flex items-center justify-between py-2">
      <div className="flex-1">
        <h3 className="text-sm font-medium text-gray-900">{label}</h3>
        <p className="text-sm text-gray-500">{description}</p>
      </div>
      <button
        type="button"
        onClick={onChange}
        className={`ml-4 relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${
          enabled ? 'bg-blue-600' : 'bg-gray-200'
        }`}
      >
        <span
          className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${
            enabled ? 'translate-x-6' : 'translate-x-1'
          }`}
        />
      </button>
    </div>
  );
}
