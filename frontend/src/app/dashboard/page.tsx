'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from '@/context/AuthContext';
import { statsApi } from '@/lib/api';
import { DashboardStats, TrendPeriod, TagStats } from '@/types/stats';
import { PieChart, Pie, Cell, LineChart, Line, BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';

const SEVERITY_COLORS = {
  critical: '#dc2626',
  high: '#ea580c',
  medium: '#f59e0b',
  low: '#3b82f6',
};

const STATUS_COLORS = {
  open: '#ef4444',
  investigating: '#f59e0b',
  resolved: '#10b981',
  closed: '#64748b',
};

const SEVERITY_LABELS: Record<string, string> = {
  critical: 'Critical（致命的）',
  high: 'High（高）',
  medium: 'Medium（中）',
  low: 'Low（低）',
};

const STATUS_LABELS: Record<string, string> = {
  open: 'Open（未対応）',
  investigating: 'Investigating（調査中）',
  resolved: 'Resolved（解決済み）',
  closed: 'Closed（完了）',
};

type GraphType = 'pie' | 'timeseries' | 'bar';

export default function DashboardPage() {
  const router = useRouter();
  const { user, token, loading: authLoading } = useAuth();
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [tagStats, setTagStats] = useState<TagStats[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [period, setPeriod] = useState<TrendPeriod>('daily');
  const [graphType, setGraphType] = useState<GraphType>('pie');

  useEffect(() => {
    if (authLoading) {
      return; // 認証状態の読み込み中は何もしない
    }

    if (!token) {
      router.push('/login');
      return;
    }

    const fetchStats = async () => {
      try {
        setLoading(true);
        const [dashboardData, tagStatsData] = await Promise.all([
          statsApi.getDashboardStats(token, period),
          statsApi.getTagStats(token),
        ]);
        setStats(dashboardData);
        setTagStats(tagStatsData.tag_stats);
        setError('');
      } catch (err) {
        setError(err instanceof Error ? err.message : 'データの取得に失敗しました');
      } finally {
        setLoading(false);
      }
    };

    fetchStats();
  }, [token, authLoading, router, period]);

  if (!user) {
    return null;
  }

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center" style={{ background: 'var(--background)' }}>
        <div
          className="text-xl font-semibold"
          style={{ color: 'var(--primary)', fontFamily: 'var(--font-display)' }}
        >
          読み込み中...
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen flex items-center justify-center" style={{ background: 'var(--background)' }}>
        <div
          className="text-xl font-semibold"
          style={{ color: 'var(--error)', fontFamily: 'var(--font-display)' }}
        >
          {error}
        </div>
      </div>
    );
  }

  if (!stats) {
    return null;
  }

  const severityData = Object.entries(stats.by_severity).map(([key, value]) => ({
    name: SEVERITY_LABELS[key] || key,
    value,
    color: SEVERITY_COLORS[key as keyof typeof SEVERITY_COLORS],
  }));

  const statusData = Object.entries(stats.by_status).map(([key, value]) => ({
    name: STATUS_LABELS[key] || key,
    value,
    color: STATUS_COLORS[key as keyof typeof STATUS_COLORS],
    key, // Keep the original key for filtering
  }));

  // Custom tooltip component
  const CustomTooltip = ({ active, payload }: any) => {
    if (active && payload && payload.length) {
      return (
        <div
          className="p-3 rounded-xl"
          style={{
            background: 'var(--surface)',
            border: '2px solid var(--primary)',
            boxShadow: '0 0 20px var(--primary-glow)',
            fontFamily: 'var(--font-body)'
          }}
        >
          <p className="font-bold text-sm" style={{ color: 'var(--foreground)' }}>{payload[0].name}</p>
          <p className="text-sm mt-1" style={{ color: 'var(--foreground-secondary)' }}>
            件数: <span className="font-bold" style={{ color: 'var(--primary)' }}>{payload[0].value}</span>
          </p>
        </div>
      );
    }
    return null;
  };

  // Handle pie chart click - navigate to incidents page with filter
  const handleSeverityClick = (data: any) => {
    const severityKey = Object.keys(SEVERITY_LABELS).find(
      key => SEVERITY_LABELS[key] === data.name
    );
    if (severityKey) {
      router.push(`/incidents?severity=${severityKey}`);
    }
  };

  const handleStatusClick = (data: any) => {
    const statusKey = Object.keys(STATUS_LABELS).find(
      key => STATUS_LABELS[key] === data.name
    );
    if (statusKey) {
      router.push(`/incidents?status=${statusKey}`);
    }
  };

  return (
    <div className="min-h-screen py-8 px-4 sm:px-6 lg:px-8 relative" style={{ background: 'var(--background)' }}>
      <div className="max-w-7xl mx-auto">
        {/* Stats Cards */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
          <div
            onClick={() => router.push('/incidents')}
            className="rounded-2xl p-6 cursor-pointer transition-all duration-300 animate-slideUp stagger-1 card-green-accent"
            style={{
              background: 'var(--surface)',
              border: '1px solid var(--border)'
            }}
            onMouseEnter={(e) => {
              e.currentTarget.style.transform = 'translateY(-4px) scale(1.02)';
              e.currentTarget.style.boxShadow = '0 0 30px var(--primary-glow), var(--shadow-xl)';
            }}
            onMouseLeave={(e) => {
              e.currentTarget.style.transform = 'translateY(0) scale(1)';
              e.currentTarget.style.boxShadow = 'var(--shadow-md)';
            }}
          >
            <div className="flex items-center justify-between">
              <div className="flex-1">
                <p
                  className="text-sm font-semibold mb-2"
                  style={{ color: 'var(--foreground-secondary)', fontFamily: 'var(--font-body)' }}
                >
                  総インシデント数
                </p>
                <p
                  className="text-4xl font-bold"
                  style={{ color: 'var(--foreground)', fontFamily: 'var(--font-display)' }}
                >
                  {stats.total_incidents}
                </p>
                <p className="text-xs mt-2" style={{ color: 'var(--primary)' }}>→ 一覧を表示</p>
              </div>
              <div
                className="w-14 h-14 rounded-xl flex items-center justify-center"
                style={{
                  background: 'linear-gradient(135deg, var(--primary-light) 0%, var(--primary) 100%)',
                  color: 'var(--primary-dark)',
                  boxShadow: '0 4px 12px var(--primary-glow)'
                }}
              >
                <svg className="w-7 h-7" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                </svg>
              </div>
            </div>
          </div>

          <div
            onClick={() => router.push('/incidents?severity=critical')}
            className="rounded-2xl p-6 cursor-pointer transition-all duration-300 animate-slideUp stagger-2"
            style={{
              background: 'var(--surface)',
              border: '1px solid var(--border)',
              borderLeft: '4px solid var(--critical)',
              boxShadow: 'var(--shadow-md)'
            }}
            onMouseEnter={(e) => {
              e.currentTarget.style.transform = 'translateY(-4px) scale(1.02)';
              e.currentTarget.style.boxShadow = '0 0 30px var(--critical-glow), var(--shadow-xl)';
            }}
            onMouseLeave={(e) => {
              e.currentTarget.style.transform = 'translateY(0) scale(1)';
              e.currentTarget.style.boxShadow = 'var(--shadow-md)';
            }}
          >
            <div className="flex items-center justify-between">
              <div className="flex-1">
                <p
                  className="text-sm font-semibold mb-2"
                  style={{ color: 'var(--foreground-secondary)', fontFamily: 'var(--font-body)' }}
                >
                  Critical
                </p>
                <p
                  className="text-4xl font-bold"
                  style={{ color: 'var(--critical)', fontFamily: 'var(--font-display)' }}
                >
                  {stats.by_severity.critical || 0}
                </p>
                <p className="text-xs mt-2" style={{ color: 'var(--critical)' }}>→ 一覧を表示</p>
              </div>
              <div
                className="w-14 h-14 rounded-xl flex items-center justify-center"
                style={{
                  background: 'var(--critical-light)',
                  color: 'var(--critical)',
                  boxShadow: '0 4px 12px var(--critical-glow)'
                }}
              >
                <svg className="w-7 h-7" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                </svg>
              </div>
            </div>
          </div>

          <div
            onClick={() => router.push('/incidents?status=open')}
            className="rounded-2xl p-6 cursor-pointer transition-all duration-300 animate-slideUp stagger-3"
            style={{
              background: 'var(--surface)',
              border: '1px solid var(--border)',
              borderLeft: '4px solid var(--high)',
              boxShadow: 'var(--shadow-md)'
            }}
            onMouseEnter={(e) => {
              e.currentTarget.style.transform = 'translateY(-4px) scale(1.02)';
              e.currentTarget.style.boxShadow = '0 0 30px var(--high-glow), var(--shadow-xl)';
            }}
            onMouseLeave={(e) => {
              e.currentTarget.style.transform = 'translateY(0) scale(1)';
              e.currentTarget.style.boxShadow = 'var(--shadow-md)';
            }}
          >
            <div className="flex items-center justify-between">
              <div className="flex-1">
                <p
                  className="text-sm font-semibold mb-2"
                  style={{ color: 'var(--foreground-secondary)', fontFamily: 'var(--font-body)' }}
                >
                  Open（未対応）
                </p>
                <p
                  className="text-4xl font-bold"
                  style={{ color: 'var(--high)', fontFamily: 'var(--font-display)' }}
                >
                  {stats.by_status.open || 0}
                </p>
                <p className="text-xs mt-2" style={{ color: 'var(--high)' }}>→ 一覧を表示</p>
              </div>
              <div
                className="w-14 h-14 rounded-xl flex items-center justify-center"
                style={{
                  background: 'var(--high-light)',
                  color: 'var(--high)',
                  boxShadow: '0 4px 12px var(--high-glow)'
                }}
              >
                <svg className="w-7 h-7" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
                </svg>
              </div>
            </div>
          </div>

          <div
            onClick={() => router.push('/incidents?status=resolved')}
            className="rounded-2xl p-6 cursor-pointer transition-all duration-300 animate-slideUp stagger-4"
            style={{
              background: 'var(--surface)',
              border: '1px solid var(--border)',
              borderLeft: '4px solid var(--success)',
              boxShadow: 'var(--shadow-md)'
            }}
            onMouseEnter={(e) => {
              e.currentTarget.style.transform = 'translateY(-4px) scale(1.02)';
              e.currentTarget.style.boxShadow = '0 0 30px var(--primary-glow), var(--shadow-xl)';
            }}
            onMouseLeave={(e) => {
              e.currentTarget.style.transform = 'translateY(0) scale(1)';
              e.currentTarget.style.boxShadow = 'var(--shadow-md)';
            }}
          >
            <div className="flex items-center justify-between">
              <div className="flex-1">
                <p
                  className="text-sm font-semibold mb-2"
                  style={{ color: 'var(--foreground-secondary)', fontFamily: 'var(--font-body)' }}
                >
                  Resolved（解決済み）
                </p>
                <p
                  className="text-4xl font-bold"
                  style={{ color: 'var(--success)', fontFamily: 'var(--font-display)' }}
                >
                  {stats.by_status.resolved || 0}
                </p>
                <p className="text-xs mt-2" style={{ color: 'var(--success)' }}>→ 一覧を表示</p>
              </div>
              <div
                className="w-14 h-14 rounded-xl flex items-center justify-center"
                style={{
                  background: 'var(--success-light)',
                  color: 'var(--success)',
                  boxShadow: '0 4px 12px var(--primary-glow)'
                }}
              >
                <svg className="w-7 h-7" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
              </div>
            </div>
          </div>
        </div>

        {/* Graph Type Selector */}
        <div className="mb-6 animate-slideUp">
          <div
            className="inline-flex rounded-2xl p-2 gap-2"
            style={{
              background: 'var(--surface)',
              border: '1px solid var(--border)',
              boxShadow: 'var(--shadow-md)'
            }}
          >
            <button
              onClick={() => setGraphType('pie')}
              className={`px-5 py-2.5 rounded-xl font-semibold transition-all duration-200 ${
                graphType === 'pie' ? 'shadow-lg' : ''
              }`}
              style={{
                background: graphType === 'pie'
                  ? 'linear-gradient(135deg, var(--primary) 0%, var(--primary-dark) 100%)'
                  : 'transparent',
                color: graphType === 'pie' ? 'white' : 'var(--foreground)',
                fontFamily: 'var(--font-body)',
                boxShadow: graphType === 'pie' ? '0 4px 12px var(--primary-glow)' : 'none'
              }}
              onMouseEnter={(e) => {
                if (graphType !== 'pie') {
                  e.currentTarget.style.background = 'var(--primary-light)';
                }
              }}
              onMouseLeave={(e) => {
                if (graphType !== 'pie') {
                  e.currentTarget.style.background = 'transparent';
                }
              }}
            >
              📊 分布グラフ
            </button>
            <button
              onClick={() => setGraphType('timeseries')}
              className={`px-5 py-2.5 rounded-xl font-semibold transition-all duration-200 ${
                graphType === 'timeseries' ? 'shadow-lg' : ''
              }`}
              style={{
                background: graphType === 'timeseries'
                  ? 'linear-gradient(135deg, var(--primary) 0%, var(--primary-dark) 100%)'
                  : 'transparent',
                color: graphType === 'timeseries' ? 'white' : 'var(--foreground)',
                fontFamily: 'var(--font-body)',
                boxShadow: graphType === 'timeseries' ? '0 4px 12px var(--primary-glow)' : 'none'
              }}
              onMouseEnter={(e) => {
                if (graphType !== 'timeseries') {
                  e.currentTarget.style.background = 'var(--primary-light)';
                }
              }}
              onMouseLeave={(e) => {
                if (graphType !== 'timeseries') {
                  e.currentTarget.style.background = 'transparent';
                }
              }}
            >
              📈 時系列グラフ
            </button>
            <button
              onClick={() => setGraphType('bar')}
              className={`px-5 py-2.5 rounded-xl font-semibold transition-all duration-200 ${
                graphType === 'bar' ? 'shadow-lg' : ''
              }`}
              style={{
                background: graphType === 'bar'
                  ? 'linear-gradient(135deg, var(--primary) 0%, var(--primary-dark) 100%)'
                  : 'transparent',
                color: graphType === 'bar' ? 'white' : 'var(--foreground)',
                fontFamily: 'var(--font-body)',
                boxShadow: graphType === 'bar' ? '0 4px 12px var(--primary-glow)' : 'none'
              }}
              onMouseEnter={(e) => {
                if (graphType !== 'bar') {
                  e.currentTarget.style.background = 'var(--primary-light)';
                }
              }}
              onMouseLeave={(e) => {
                if (graphType !== 'bar') {
                  e.currentTarget.style.background = 'transparent';
                }
              }}
            >
              📊 棒グラフ
            </button>
          </div>
        </div>

        {/* Charts Section */}
        {graphType === 'pie' && (
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8 animate-fadeIn">
            {/* Severity Distribution - Pie Chart */}
            <div
              className="rounded-2xl p-6 animate-scaleIn stagger-5"
              style={{
                background: 'var(--surface)',
                border: '1px solid var(--border)',
                boxShadow: 'var(--shadow-lg)'
              }}
            >
              <h2
                className="text-xl font-bold mb-2"
                style={{ color: 'var(--foreground)', fontFamily: 'var(--font-display)' }}
              >
                重要度別分布
              </h2>
              <p className="text-sm mb-6" style={{ color: 'var(--foreground-secondary)', fontFamily: 'var(--font-body)' }}>
                クリックして詳細を表示
              </p>
            <ResponsiveContainer width="100%" height={300}>
              <PieChart>
                <Pie
                  data={severityData}
                  cx="50%"
                  cy="50%"
                  labelLine={false}
                  label={({ name, percent }) => `${name?.split('（')[0] ?? ''}: ${((percent ?? 0) * 100).toFixed(0)}%`}
                  outerRadius={90}
                  fill="#8884d8"
                  dataKey="value"
                  onClick={handleSeverityClick}
                  style={{ cursor: 'pointer', fontFamily: 'var(--font-body)', fontSize: '13px', fontWeight: '600' }}
                  animationBegin={0}
                  animationDuration={800}
                >
                  {severityData.map((entry, index) => (
                    <Cell
                      key={`cell-${index}`}
                      fill={entry.color}
                      className="hover:opacity-80 transition-opacity"
                    />
                  ))}
                </Pie>
                <Tooltip content={<CustomTooltip />} />
              </PieChart>
            </ResponsiveContainer>
          </div>

          {/* Status Distribution - Pie Chart */}
          <div
            className="rounded-2xl p-6 animate-scaleIn stagger-6"
            style={{
              background: 'var(--surface)',
              border: '1px solid var(--border)',
              boxShadow: 'var(--shadow-lg)'
            }}
          >
            <h2
              className="text-xl font-bold mb-2"
              style={{ color: 'var(--foreground)', fontFamily: 'var(--font-display)' }}
            >
              ステータス別分布
            </h2>
            <p className="text-sm mb-6" style={{ color: 'var(--foreground-secondary)', fontFamily: 'var(--font-body)' }}>
              クリックして詳細を表示
            </p>
            <ResponsiveContainer width="100%" height={300}>
              <PieChart>
                <Pie
                  data={statusData}
                  cx="50%"
                  cy="50%"
                  labelLine={false}
                  label={({ name, percent }) => `${name?.split('（')[0] ?? ''}: ${((percent ?? 0) * 100).toFixed(0)}%`}
                  outerRadius={90}
                  fill="#8884d8"
                  dataKey="value"
                  onClick={handleStatusClick}
                  style={{ cursor: 'pointer', fontFamily: 'var(--font-body)', fontSize: '13px', fontWeight: '600' }}
                  animationBegin={0}
                  animationDuration={800}
                >
                  {statusData.map((entry, index) => (
                    <Cell
                      key={`cell-${index}`}
                      fill={entry.color}
                      className="hover:opacity-80 transition-opacity"
                    />
                  ))}
                </Pie>
                <Tooltip content={<CustomTooltip />} />
              </PieChart>
            </ResponsiveContainer>
          </div>
        </div>
        )}

        {/* Timeseries Graph */}
        {graphType === 'timeseries' && (
          <div className="mb-8 animate-fadeIn">
            {/* Period Selector */}
            <div className="mb-6">
              <div
                className="inline-flex rounded-xl p-1.5 gap-1.5"
                style={{
                  background: 'var(--surface)',
                  border: '1px solid var(--border)',
                  boxShadow: 'var(--shadow-sm)'
                }}
              >
                {(['daily', 'weekly', 'monthly'] as TrendPeriod[]).map((p) => (
                  <button
                    key={p}
                    onClick={() => setPeriod(p)}
                    className={`px-4 py-2 rounded-lg font-semibold text-sm transition-all duration-200`}
                    style={{
                      background: period === p
                        ? 'linear-gradient(135deg, var(--primary) 0%, var(--primary-dark) 100%)'
                        : 'transparent',
                      color: period === p ? 'white' : 'var(--foreground)',
                      fontFamily: 'var(--font-body)',
                      boxShadow: period === p ? '0 2px 8px var(--primary-glow)' : 'none'
                    }}
                  >
                    {p === 'daily' ? '日別' : p === 'weekly' ? '週別' : '月別'}
                  </button>
                ))}
              </div>
            </div>

            <div
              className="rounded-2xl p-6"
              style={{
                background: 'var(--surface)',
                border: '1px solid var(--border)',
                boxShadow: 'var(--shadow-lg)'
              }}
            >
              <h2
                className="text-xl font-bold mb-2"
                style={{ color: 'var(--foreground)', fontFamily: 'var(--font-display)' }}
              >
                インシデント発生推移
              </h2>
              <p className="text-sm mb-6" style={{ color: 'var(--foreground-secondary)', fontFamily: 'var(--font-body)' }}>
                {period === 'daily' && '過去30日間の'}
                {period === 'weekly' && '過去12週間の'}
                {period === 'monthly' && '過去12ヶ月の'}
                インシデント発生件数
              </p>
              <ResponsiveContainer width="100%" height={400}>
                <LineChart data={stats.trend_data}>
                  <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" />
                  <XAxis
                    dataKey="date"
                    stroke="var(--foreground)"
                    style={{ fontSize: '12px', fontFamily: 'var(--font-body)' }}
                  />
                  <YAxis
                    stroke="var(--foreground)"
                    style={{ fontSize: '12px', fontFamily: 'var(--font-body)' }}
                  />
                  <Tooltip content={<CustomTooltip />} />
                  <Legend
                    wrapperStyle={{ fontFamily: 'var(--font-body)', fontSize: '13px' }}
                  />
                  <Line
                    type="monotone"
                    dataKey="count"
                    name="件数"
                    stroke="var(--primary)"
                    strokeWidth={3}
                    dot={{ fill: 'var(--primary)', r: 4 }}
                    activeDot={{ r: 6 }}
                    animationDuration={800}
                  />
                </LineChart>
              </ResponsiveContainer>
            </div>
          </div>
        )}

        {/* Bar Graph */}
        {graphType === 'bar' && (
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8 animate-fadeIn">
            {/* Severity Bar Chart */}
            <div
              className="rounded-2xl p-6"
              style={{
                background: 'var(--surface)',
                border: '1px solid var(--border)',
                boxShadow: 'var(--shadow-lg)'
              }}
            >
              <h2
                className="text-xl font-bold mb-2"
                style={{ color: 'var(--foreground)', fontFamily: 'var(--font-display)' }}
              >
                重要度別件数
              </h2>
              <p className="text-sm mb-6" style={{ color: 'var(--foreground-secondary)', fontFamily: 'var(--font-body)' }}>
                深刻度ごとのインシデント数
              </p>
              <ResponsiveContainer width="100%" height={300}>
                <BarChart data={severityData}>
                  <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" />
                  <XAxis
                    dataKey="name"
                    stroke="var(--foreground)"
                    style={{ fontSize: '11px', fontFamily: 'var(--font-body)' }}
                    tickFormatter={(value) => value.split('（')[0]}
                  />
                  <YAxis
                    stroke="var(--foreground)"
                    style={{ fontSize: '12px', fontFamily: 'var(--font-body)' }}
                  />
                  <Tooltip content={<CustomTooltip />} />
                  <Bar
                    dataKey="value"
                    name="件数"
                    radius={[8, 8, 0, 0]}
                    animationDuration={800}
                  >
                    {severityData.map((entry, index) => (
                      <Cell key={`cell-${index}`} fill={entry.color} />
                    ))}
                  </Bar>
                </BarChart>
              </ResponsiveContainer>
            </div>

            {/* Status Bar Chart */}
            <div
              className="rounded-2xl p-6"
              style={{
                background: 'var(--surface)',
                border: '1px solid var(--border)',
                boxShadow: 'var(--shadow-lg)'
              }}
            >
              <h2
                className="text-xl font-bold mb-2"
                style={{ color: 'var(--foreground)', fontFamily: 'var(--font-display)' }}
              >
                ステータス別件数
              </h2>
              <p className="text-sm mb-6" style={{ color: 'var(--foreground-secondary)', fontFamily: 'var(--font-body)' }}>
                ステータスごとのインシデント数
              </p>
              <ResponsiveContainer width="100%" height={300}>
                <BarChart data={statusData}>
                  <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" />
                  <XAxis
                    dataKey="name"
                    stroke="var(--foreground)"
                    style={{ fontSize: '11px', fontFamily: 'var(--font-body)' }}
                    tickFormatter={(value) => value.split('（')[0]}
                  />
                  <YAxis
                    stroke="var(--foreground)"
                    style={{ fontSize: '12px', fontFamily: 'var(--font-body)' }}
                  />
                  <Tooltip content={<CustomTooltip />} />
                  <Bar
                    dataKey="value"
                    name="件数"
                    radius={[8, 8, 0, 0]}
                    animationDuration={800}
                  >
                    {statusData.map((entry, index) => (
                      <Cell key={`cell-${index}`} fill={entry.color} />
                    ))}
                  </Bar>
                </BarChart>
              </ResponsiveContainer>
            </div>
          </div>
        )}

        {/* Trend Chart (kept for compatibility) */}
        <div
          className="rounded-2xl p-6 mb-8 animate-fadeIn"
          style={{
            background: 'var(--surface)',
            border: '1px solid var(--border)',
            boxShadow: 'var(--shadow-lg)'
          }}
        >
          <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between mb-6">
            <div>
              <h2
                className="text-xl font-bold mb-1"
                style={{ color: 'var(--foreground)', fontFamily: 'var(--font-display)' }}
              >
                インシデント発生トレンド
              </h2>
              <p className="text-sm" style={{ color: 'var(--foreground-secondary)', fontFamily: 'var(--font-body)' }}>
                期間ごとの発生推移
              </p>
            </div>
            <div className="flex space-x-2 mt-4 sm:mt-0">
              {(['daily', 'weekly', 'monthly'] as TrendPeriod[]).map((p) => (
                <button
                  key={p}
                  onClick={() => setPeriod(p)}
                  className="px-4 py-2 rounded-xl text-sm font-semibold transition-all"
                  style={{
                    background: period === p ? 'linear-gradient(135deg, var(--primary) 0%, var(--primary-dark) 100%)' : 'var(--secondary-light)',
                    color: period === p ? 'white' : 'var(--foreground)',
                    fontFamily: 'var(--font-body)',
                    boxShadow: period === p ? '0 4px 12px var(--primary-glow)' : 'none'
                  }}
                  onMouseEnter={(e) => {
                    if (period !== p) {
                      e.currentTarget.style.background = 'var(--primary-light)';
                    }
                  }}
                  onMouseLeave={(e) => {
                    if (period !== p) {
                      e.currentTarget.style.background = 'var(--secondary-light)';
                    }
                  }}
                >
                  {p === 'daily' ? '日次' : p === 'weekly' ? '週次' : '月次'}
                </button>
              ))}
            </div>
          </div>
          <ResponsiveContainer width="100%" height={400}>
            <LineChart data={stats.trend_data}>
              <defs>
                <linearGradient id="colorCount" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%" stopColor="var(--primary)" stopOpacity={0.3}/>
                  <stop offset="95%" stopColor="var(--primary)" stopOpacity={0}/>
                </linearGradient>
              </defs>
              <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" />
              <XAxis
                dataKey="date"
                tick={{ fontSize: 12, fontFamily: 'var(--font-body)', fill: 'var(--foreground-secondary)' }}
              />
              <YAxis
                tick={{ fontSize: 12, fontFamily: 'var(--font-body)', fill: 'var(--foreground-secondary)' }}
              />
              <Tooltip content={<CustomTooltip />} />
              <Line
                type="monotone"
                dataKey="count"
                stroke="var(--primary)"
                strokeWidth={3}
                name="インシデント数"
                dot={{ fill: 'var(--primary)', strokeWidth: 2, r: 5 }}
                activeDot={{ r: 7, fill: 'var(--primary-dark)' }}
                animationBegin={0}
                animationDuration={1000}
                fill="url(#colorCount)"
              />
            </LineChart>
          </ResponsiveContainer>
        </div>

        {/* Tag Statistics */}
        {tagStats.length > 0 && (
          <div
            className="rounded-2xl p-6 mb-8 animate-fadeIn"
            style={{
              background: 'var(--surface)',
              border: '1px solid var(--border)',
              boxShadow: 'var(--shadow-lg)'
            }}
          >
            <h2
              className="text-xl font-bold mb-6"
              style={{ color: 'var(--foreground)', fontFamily: 'var(--font-display)' }}
            >
              タグ別インシデント統計
            </h2>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              {tagStats.map((tag, index) => (
                <div
                  key={tag.tag_id}
                  className="rounded-xl p-4 transition-all duration-300"
                  style={{
                    background: 'var(--gray-50)',
                    border: '1px solid var(--border)',
                    animationDelay: `${index * 0.05}s`
                  }}
                  onMouseEnter={(e) => {
                    e.currentTarget.style.transform = 'translateY(-2px)';
                    e.currentTarget.style.boxShadow = 'var(--shadow-md)';
                  }}
                  onMouseLeave={(e) => {
                    e.currentTarget.style.transform = 'translateY(0)';
                    e.currentTarget.style.boxShadow = 'none';
                  }}
                >
                  <div className="flex items-center justify-between mb-3">
                    <div className="flex items-center">
                      <span
                        className="w-4 h-4 rounded-full mr-2"
                        style={{ backgroundColor: tag.tag_color, boxShadow: `0 0 8px ${tag.tag_color}` }}
                      />
                      <span
                        className="font-semibold text-sm"
                        style={{ color: 'var(--foreground)', fontFamily: 'var(--font-body)' }}
                      >
                        {tag.tag_name}
                      </span>
                    </div>
                    <span
                      className="text-2xl font-bold"
                      style={{ color: 'var(--foreground)', fontFamily: 'var(--font-display)' }}
                    >
                      {tag.count}
                    </span>
                  </div>
                  <div className="w-full rounded-full h-2" style={{ background: 'var(--gray-200)' }}>
                    <div
                      className="h-2 rounded-full transition-all duration-500"
                      style={{
                        backgroundColor: tag.tag_color,
                        width: `${tag.percentage}%`,
                      }}
                    />
                  </div>
                  <div className="mt-2 text-xs text-right" style={{ color: 'var(--foreground-secondary)', fontFamily: 'var(--font-mono)' }}>
                    {tag.percentage.toFixed(1)}%
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Recent Incidents */}
        <div
          className="rounded-2xl overflow-hidden animate-fadeIn"
          style={{
            background: 'var(--surface)',
            border: '1px solid var(--border)',
            boxShadow: 'var(--shadow-lg)'
          }}
        >
          <div className="px-6 py-5" style={{ borderBottom: '1px solid var(--border)' }}>
            <h2
              className="text-xl font-bold"
              style={{ color: 'var(--foreground)', fontFamily: 'var(--font-display)' }}
            >
              最近のインシデント
            </h2>
          </div>
          <div className="overflow-x-auto">
            <table className="min-w-full">
              <thead style={{ background: 'var(--gray-50)' }}>
                <tr>
                  <th
                    className="px-6 py-4 text-left text-xs font-bold uppercase tracking-wider"
                    style={{ color: 'var(--foreground-secondary)', fontFamily: 'var(--font-body)' }}
                  >
                    タイトル
                  </th>
                  <th
                    className="px-6 py-4 text-left text-xs font-bold uppercase tracking-wider"
                    style={{ color: 'var(--foreground-secondary)', fontFamily: 'var(--font-body)' }}
                  >
                    重要度
                  </th>
                  <th
                    className="px-6 py-4 text-left text-xs font-bold uppercase tracking-wider"
                    style={{ color: 'var(--foreground-secondary)', fontFamily: 'var(--font-body)' }}
                  >
                    ステータス
                  </th>
                  <th
                    className="px-6 py-4 text-left text-xs font-bold uppercase tracking-wider"
                    style={{ color: 'var(--foreground-secondary)', fontFamily: 'var(--font-body)' }}
                  >
                    検出日時
                  </th>
                </tr>
              </thead>
              <tbody>
                {stats.recent_incidents.length === 0 ? (
                  <tr>
                    <td
                      colSpan={4}
                      className="px-6 py-8 text-center"
                      style={{ color: 'var(--foreground-secondary)', fontFamily: 'var(--font-body)' }}
                    >
                      インシデントがありません
                    </td>
                  </tr>
                ) : (
                  stats.recent_incidents.map((incident, index) => (
                    <tr
                      key={incident.id}
                      onClick={() => router.push(`/incidents/${incident.id}`)}
                      className="transition-all duration-200 cursor-pointer"
                      style={{ borderTop: '1px solid var(--border)' }}
                      onMouseEnter={(e) => {
                        e.currentTarget.style.background = 'var(--primary-light)';
                      }}
                      onMouseLeave={(e) => {
                        e.currentTarget.style.background = 'transparent';
                      }}
                    >
                      <td className="px-6 py-4">
                        <div
                          className="text-sm font-semibold"
                          style={{ color: 'var(--foreground)', fontFamily: 'var(--font-body)' }}
                        >
                          {incident.title}
                        </div>
                      </td>
                      <td className="px-6 py-4">
                        <span
                          className="px-3 py-1 inline-flex text-xs font-bold rounded-full"
                          style={{
                            background: SEVERITY_COLORS[incident.severity as keyof typeof SEVERITY_COLORS] + '20',
                            color: SEVERITY_COLORS[incident.severity as keyof typeof SEVERITY_COLORS],
                            border: `2px solid ${SEVERITY_COLORS[incident.severity as keyof typeof SEVERITY_COLORS]}`,
                            fontFamily: 'var(--font-body)'
                          }}
                        >
                          {SEVERITY_LABELS[incident.severity]}
                        </span>
                      </td>
                      <td className="px-6 py-4">
                        <span
                          className="px-3 py-1 inline-flex text-xs font-bold rounded-full"
                          style={{
                            background: STATUS_COLORS[incident.status as keyof typeof STATUS_COLORS] + '20',
                            color: STATUS_COLORS[incident.status as keyof typeof STATUS_COLORS],
                            border: `2px solid ${STATUS_COLORS[incident.status as keyof typeof STATUS_COLORS]}`,
                            fontFamily: 'var(--font-body)'
                          }}
                        >
                          {STATUS_LABELS[incident.status]}
                        </span>
                      </td>
                      <td className="px-6 py-4">
                        <span
                          className="text-sm"
                          style={{ color: 'var(--foreground-secondary)', fontFamily: 'var(--font-mono)' }}
                        >
                          {new Date(incident.detected_at).toLocaleString('ja-JP')}
                        </span>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  );
}
