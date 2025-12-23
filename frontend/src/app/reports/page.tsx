'use client';

import { useState, useEffect } from 'react';
import { useAuth } from '../../context/AuthContext';
import { reportApi } from '../../lib/api';
import { MonthlyReport } from '../../types/report';

export default function ReportsPage() {
  const { token } = useAuth();
  const [report, setReport] = useState<MonthlyReport | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Date selection state
  const currentDate = new Date();
  const [selectedYear, setSelectedYear] = useState(currentDate.getFullYear());
  const [selectedMonth, setSelectedMonth] = useState(currentDate.getMonth() + 1);

  useEffect(() => {
    if (token) {
      fetchReport();
    }
  }, [token, selectedYear, selectedMonth]);

  const fetchReport = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await reportApi.getMonthlyReport(token!, selectedYear, selectedMonth);
      setReport(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'レポートの取得に失敗しました');
    } finally {
      setLoading(false);
    }
  };

  const handlePreviousMonth = () => {
    if (selectedMonth === 1) {
      setSelectedMonth(12);
      setSelectedYear(selectedYear - 1);
    } else {
      setSelectedMonth(selectedMonth - 1);
    }
  };

  const handleNextMonth = () => {
    if (selectedMonth === 12) {
      setSelectedMonth(1);
      setSelectedYear(selectedYear + 1);
    } else {
      setSelectedMonth(selectedMonth + 1);
    }
  };

  const formatMonth = (year: number, month: number) => {
    return `${year}年${month}月`;
  };

  const getSeverityLabel = (severity: string) => {
    const labels: Record<string, string> = {
      critical: 'クリティカル',
      high: '高',
      medium: '中',
      low: '低',
    };
    return labels[severity] || severity;
  };

  const getStatusLabel = (status: string) => {
    const labels: Record<string, string> = {
      open: '未対応',
      investigating: '調査中',
      resolved: '解決済み',
      closed: 'クローズ',
    };
    return labels[status] || status;
  };

  const formatHours = (hours: number) => {
    if (hours < 24) {
      return `${hours.toFixed(1)}時間`;
    }
    const days = Math.floor(hours / 24);
    const remainingHours = hours % 24;
    return `${days}日${remainingHours.toFixed(1)}時間`;
  };

  const getChangeIcon = (value: number) => {
    if (value > 0) return '↑';
    if (value < 0) return '↓';
    return '→';
  };

  const getChangeColor = (value: number) => {
    if (value > 0) return 'text-red-600';
    if (value < 0) return 'text-green-600';
    return 'text-gray-600';
  };

  if (loading) {
    return (
      <div className="p-6">
        <div className="flex justify-center items-center h-64">
          <div className="text-gray-500">読み込み中...</div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-6">
        <div className="bg-red-50 border border-red-200 text-red-700 p-4 rounded-lg">
          {error}
        </div>
      </div>
    );
  }

  if (!report) {
    return (
      <div className="p-6">
        <div className="text-gray-500">レポートがありません</div>
      </div>
    );
  }

  return (
    <div className="p-6 max-w-7xl mx-auto">
      {/* Header */}
      <div className="mb-6">
        <h1 className="text-3xl font-bold text-gray-900">月次レポート</h1>
        <p className="mt-2 text-sm text-gray-600">
          インシデント管理の月次統計とパフォーマンスメトリクス
        </p>
      </div>

      {/* Month Selector */}
      <div className="mb-6 flex items-center justify-between bg-white p-4 rounded-lg shadow">
        <button
          onClick={handlePreviousMonth}
          className="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 hover:bg-gray-200 rounded-md"
        >
          ← 前月
        </button>
        <div className="text-xl font-semibold text-gray-900">
          {formatMonth(selectedYear, selectedMonth)}
        </div>
        <button
          onClick={handleNextMonth}
          className="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 hover:bg-gray-200 rounded-md"
        >
          次月 →
        </button>
      </div>

      {/* Summary Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-4 mb-6">
        <div className="bg-white p-6 rounded-lg shadow">
          <div className="text-sm font-medium text-gray-600">総インシデント</div>
          <div className="mt-2 text-3xl font-bold text-gray-900">
            {report.summary.total_incidents}
          </div>
          {report.comparison && (
            <div className={`mt-2 text-sm ${getChangeColor(report.comparison.total_incidents_change)}`}>
              {getChangeIcon(report.comparison.total_incidents_change)}{' '}
              {Math.abs(report.comparison.total_incidents_change)} ({report.comparison.total_incidents_change_percent.toFixed(1)}%)
            </div>
          )}
        </div>

        <div className="bg-white p-6 rounded-lg shadow">
          <div className="text-sm font-medium text-gray-600">解決済み</div>
          <div className="mt-2 text-3xl font-bold text-green-600">
            {report.summary.resolved_incidents}
          </div>
          {report.comparison && (
            <div className={`mt-2 text-sm ${getChangeColor(report.comparison.resolved_incidents_change)}`}>
              {getChangeIcon(report.comparison.resolved_incidents_change)}{' '}
              {Math.abs(report.comparison.resolved_incidents_change)} ({report.comparison.resolved_incidents_change_percent.toFixed(1)}%)
            </div>
          )}
        </div>

        <div className="bg-white p-6 rounded-lg shadow">
          <div className="text-sm font-medium text-gray-600">未解決</div>
          <div className="mt-2 text-3xl font-bold text-orange-600">
            {report.summary.open_incidents}
          </div>
        </div>

        <div className="bg-white p-6 rounded-lg shadow">
          <div className="text-sm font-medium text-gray-600">クリティカル</div>
          <div className="mt-2 text-3xl font-bold text-red-600">
            {report.summary.critical_incidents}
          </div>
        </div>

        <div className="bg-white p-6 rounded-lg shadow">
          <div className="text-sm font-medium text-gray-600">解決率</div>
          <div className="mt-2 text-3xl font-bold text-blue-600">
            {report.summary.total_incidents > 0
              ? ((report.summary.resolved_incidents / report.summary.total_incidents) * 100).toFixed(1)
              : 0}%
          </div>
        </div>
      </div>

      {/* Performance Metrics */}
      <div className="bg-white p-6 rounded-lg shadow mb-6">
        <h2 className="text-xl font-bold text-gray-900 mb-4">パフォーマンスメトリクス</h2>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
          <div>
            <div className="text-sm font-medium text-gray-600">平均解決時間</div>
            <div className="mt-2 text-2xl font-bold text-gray-900">
              {formatHours(report.performance_metrics.average_resolution_time_hours)}
            </div>
          </div>
          <div>
            <div className="text-sm font-medium text-gray-600">中央値解決時間</div>
            <div className="mt-2 text-2xl font-bold text-gray-900">
              {formatHours(report.performance_metrics.median_resolution_time_hours)}
            </div>
          </div>
          <div>
            <div className="text-sm font-medium text-gray-600">SLA準拠率</div>
            <div className="mt-2 text-2xl font-bold text-gray-900">
              {report.performance_metrics.sla_compliance_rate.toFixed(1)}%
            </div>
          </div>
          <div>
            <div className="text-sm font-medium text-gray-600">平均応答時間</div>
            <div className="mt-2 text-2xl font-bold text-gray-900">
              {formatHours(report.performance_metrics.mean_time_to_acknowledge_hours)}
            </div>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
        {/* Severity Breakdown */}
        <div className="bg-white p-6 rounded-lg shadow">
          <h2 className="text-xl font-bold text-gray-900 mb-4">重要度別</h2>
          <div className="space-y-3">
            {Object.entries(report.severity_breakdown).map(([severity, count]) => (
              <div key={severity} className="flex items-center justify-between">
                <div className="flex items-center">
                  <div className={`w-3 h-3 rounded-full mr-3 ${
                    severity === 'critical' ? 'bg-red-500' :
                    severity === 'high' ? 'bg-orange-500' :
                    severity === 'medium' ? 'bg-yellow-500' :
                    'bg-green-500'
                  }`}></div>
                  <span className="text-gray-700">{getSeverityLabel(severity)}</span>
                </div>
                <div className="flex items-center">
                  <span className="text-gray-900 font-semibold mr-3">{count}</span>
                  <div className="w-32 bg-gray-200 rounded-full h-2">
                    <div
                      className={`h-2 rounded-full ${
                        severity === 'critical' ? 'bg-red-500' :
                        severity === 'high' ? 'bg-orange-500' :
                        severity === 'medium' ? 'bg-yellow-500' :
                        'bg-green-500'
                      }`}
                      style={{
                        width: `${(count / report.summary.total_incidents) * 100}%`,
                      }}
                    ></div>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Status Breakdown */}
        <div className="bg-white p-6 rounded-lg shadow">
          <h2 className="text-xl font-bold text-gray-900 mb-4">ステータス別</h2>
          <div className="space-y-3">
            {Object.entries(report.status_breakdown).map(([status, count]) => (
              <div key={status} className="flex items-center justify-between">
                <span className="text-gray-700">{getStatusLabel(status)}</span>
                <div className="flex items-center">
                  <span className="text-gray-900 font-semibold mr-3">{count}</span>
                  <div className="w-32 bg-gray-200 rounded-full h-2">
                    <div
                      className="bg-blue-500 h-2 rounded-full"
                      style={{
                        width: `${(count / report.summary.total_incidents) * 100}%`,
                      }}
                    ></div>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Top Tags */}
      <div className="bg-white p-6 rounded-lg shadow mb-6">
        <h2 className="text-xl font-bold text-gray-900 mb-4">よく使われるタグ</h2>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-4">
          {report.top_tags.map((tag) => (
            <div key={tag.tag_id} className="bg-gray-50 p-4 rounded-lg">
              <div className="text-sm font-medium text-gray-600">{tag.tag_name}</div>
              <div className="mt-1 text-2xl font-bold text-gray-900">{tag.count}</div>
            </div>
          ))}
        </div>
      </div>

      {/* Daily Trend */}
      <div className="bg-white p-6 rounded-lg shadow">
        <h2 className="text-xl font-bold text-gray-900 mb-4">日別トレンド</h2>
        <div className="overflow-x-auto">
          <div className="flex items-end space-x-1 h-48">
            {report.daily_trend.map((day) => {
              const maxCount = Math.max(...report.daily_trend.map(d => d.count));
              const height = maxCount > 0 ? (day.count / maxCount) * 100 : 0;
              return (
                <div key={day.date} className="flex-1 flex flex-col items-center">
                  <div
                    className="w-full bg-blue-500 hover:bg-blue-600 rounded-t transition-colors cursor-pointer"
                    style={{ height: `${height}%` }}
                    title={`${new Date(day.date).toLocaleDateString('ja-JP')}: ${day.count}件`}
                  ></div>
                  <div className="text-xs text-gray-500 mt-2 transform -rotate-45 origin-left">
                    {new Date(day.date).getDate()}
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      </div>
    </div>
  );
}
