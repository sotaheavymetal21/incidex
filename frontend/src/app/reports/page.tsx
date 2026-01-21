'use client';

import { useState, useEffect, useRef } from 'react';
import { useAuth } from '../../context/AuthContext';
import { reportApi } from '../../lib/api';
import { MonthlyReport } from '../../types/report';

export default function ReportsPage() {
  const { token } = useAuth();
  const [report, setReport] = useState<MonthlyReport | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showCalendar, setShowCalendar] = useState(false);
  const calendarRef = useRef<HTMLDivElement>(null);

  // Date selection state
  const currentDate = new Date();
  const [selectedYear, setSelectedYear] = useState(currentDate.getFullYear());
  const [selectedMonth, setSelectedMonth] = useState(currentDate.getMonth() + 1);
  const [calendarYear, setCalendarYear] = useState(currentDate.getFullYear());

  useEffect(() => {
    if (token) {
      fetchReport();
    }
  }, [token, selectedYear, selectedMonth]);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (calendarRef.current && !calendarRef.current.contains(event.target as Node)) {
        setShowCalendar(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

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

  const handleMonthSelect = (month: number) => {
    setSelectedYear(calendarYear);
    setSelectedMonth(month);
    setShowCalendar(false);
  };

  const handlePreviousMonth = () => {
    if (selectedMonth === 1) {
      setSelectedYear(selectedYear - 1);
      setSelectedMonth(12);
    } else {
      setSelectedMonth(selectedMonth - 1);
    }
  };

  const handleNextMonth = () => {
    // 未来の月には進めない
    const isCurrentMonth = selectedYear === currentDate.getFullYear() && selectedMonth === currentDate.getMonth() + 1;
    if (isCurrentMonth) return;

    if (selectedMonth === 12) {
      setSelectedYear(selectedYear + 1);
      setSelectedMonth(1);
    } else {
      setSelectedMonth(selectedMonth + 1);
    }
  };

  const formatMonth = (year: number, month: number) => {
    return `${year}年${month}月`;
  };

  const months = [
    '1月', '2月', '3月', '4月',
    '5月', '6月', '7月', '8月',
    '9月', '10月', '11月', '12月'
  ];

  const isCurrentMonth = (month: number) => {
    return calendarYear === currentDate.getFullYear() && month === currentDate.getMonth() + 1;
  };

  const isSelectedMonth = (month: number) => {
    return calendarYear === selectedYear && month === selectedMonth;
  };

  const isFutureMonth = (month: number) => {
    return calendarYear > currentDate.getFullYear() ||
      (calendarYear === currentDate.getFullYear() && month > currentDate.getMonth() + 1);
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
    <div className="p-6 max-w-7xl mx-auto bg-gray-50 min-h-screen">
      {/* Header */}
      <div className="mb-6">
        <h1 className="text-3xl font-bold text-gray-900">月次レポート</h1>
        <p className="mt-2 text-base text-gray-700">
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
        <div className="relative" ref={calendarRef}>
          <button
            onClick={() => {
              setCalendarYear(selectedYear);
              setShowCalendar(!showCalendar);
            }}
            className="text-xl font-semibold text-gray-900 hover:text-blue-600 hover:bg-blue-50 px-4 py-2 rounded-md transition-colors flex items-center gap-2"
          >
            {formatMonth(selectedYear, selectedMonth)}
            <svg className={`w-5 h-5 transition-transform ${showCalendar ? 'rotate-180' : ''}`} fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
            </svg>
          </button>

          {/* Calendar Popup */}
          {showCalendar && (
            <div className="absolute top-full left-1/2 transform -translate-x-1/2 mt-2 bg-white border border-gray-200 rounded-lg shadow-lg z-50 p-4 w-72">
              {/* Year Navigation */}
              <div className="flex items-center justify-between mb-4">
                <button
                  onClick={() => setCalendarYear(calendarYear - 1)}
                  className="p-2 hover:bg-gray-100 rounded-md text-gray-600"
                >
                  <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
                  </svg>
                </button>
                <span className="text-lg font-semibold text-gray-900">{calendarYear}年</span>
                <button
                  onClick={() => {
                    if (calendarYear < currentDate.getFullYear()) {
                      setCalendarYear(calendarYear + 1);
                    }
                  }}
                  disabled={calendarYear >= currentDate.getFullYear()}
                  className={`p-2 rounded-md ${
                    calendarYear >= currentDate.getFullYear()
                      ? 'text-gray-300 cursor-not-allowed'
                      : 'hover:bg-gray-100 text-gray-600'
                  }`}
                >
                  <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                  </svg>
                </button>
              </div>

              {/* Month Grid */}
              <div className="grid grid-cols-4 gap-2">
                {months.map((monthName, index) => {
                  const monthNum = index + 1;
                  const isFuture = isFutureMonth(monthNum);
                  const isSelected = isSelectedMonth(monthNum);
                  const isCurrent = isCurrentMonth(monthNum);

                  return (
                    <button
                      key={monthName}
                      onClick={() => !isFuture && handleMonthSelect(monthNum)}
                      disabled={isFuture}
                      className={`py-2 px-1 text-sm rounded-md transition-colors ${
                        isFuture
                          ? 'text-gray-300 cursor-not-allowed'
                          : isSelected
                            ? 'bg-blue-600 text-white font-semibold'
                            : isCurrent
                              ? 'bg-blue-100 text-blue-700 font-medium hover:bg-blue-200'
                              : 'text-gray-700 hover:bg-gray-100'
                      }`}
                    >
                      {monthName}
                    </button>
                  );
                })}
              </div>

              {/* Quick Actions */}
              <div className="mt-4 pt-3 border-t border-gray-200">
                <button
                  onClick={() => {
                    setSelectedYear(currentDate.getFullYear());
                    setSelectedMonth(currentDate.getMonth() + 1);
                    setShowCalendar(false);
                  }}
                  className="w-full py-2 text-sm text-blue-600 hover:bg-blue-50 rounded-md transition-colors"
                >
                  今月に移動
                </button>
              </div>
            </div>
          )}
        </div>
        <button
          onClick={handleNextMonth}
          disabled={selectedYear === currentDate.getFullYear() && selectedMonth === currentDate.getMonth() + 1}
          className={`px-4 py-2 text-sm font-medium rounded-md ${
            selectedYear === currentDate.getFullYear() && selectedMonth === currentDate.getMonth() + 1
              ? 'text-gray-400 bg-gray-100 cursor-not-allowed'
              : 'text-gray-700 bg-gray-100 hover:bg-gray-200'
          }`}
        >
          次月 →
        </button>
      </div>

      {/* Summary Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 mb-6">
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
          <div className="text-sm font-medium text-gray-600">平均解決時間</div>
          <div className="mt-2 text-3xl font-bold text-purple-600">
            {formatHours(report.performance_metrics.average_resolution_time_hours)}
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 mb-6">
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

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
        {/* Severity Breakdown */}
        <div className="bg-white p-6 rounded-lg shadow">
          <h2 className="text-xl font-bold text-gray-900 mb-4">重要度別</h2>
          <div className="space-y-3">
            {Object.entries(report.severity_breakdown).map(([severity, count]) => (
              <div key={severity} className="flex items-center justify-between">
                <div className="flex items-center">
                  <div className={`w-3 h-3 rounded-full mr-3 ${severity === 'critical' ? 'bg-red-500' :
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
                      className={`h-2 rounded-full ${severity === 'critical' ? 'bg-red-500' :
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
        <h2 className="text-xl font-bold text-gray-900 mb-4">タグ</h2>
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
