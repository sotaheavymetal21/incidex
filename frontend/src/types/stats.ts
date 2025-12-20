import { Incident } from './incident';

export interface DashboardStats {
  total_incidents: number;
  by_severity: Record<string, number>;
  by_status: Record<string, number>;
  recent_incidents: Incident[];
  trend_data: TrendDataPoint[];
}

export interface TrendDataPoint {
  date: string;
  count: number;
}

export type TrendPeriod = 'daily' | 'weekly' | 'monthly';
